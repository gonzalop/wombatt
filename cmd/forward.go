package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"time"

	"wombatt/internal/common"

	"go.bug.st/serial"
)

type ForwardCmd struct {
	Controller  string `name:"controller-port" required:"" help:"Serial port or address of the controller"`
	Subordinate string `name:"subordinate-port" required:"" help:"Serial port or address of the subordinate device"`
	BaudRate    uint   `short:"B" default:"9600" help:"Baud rate"`
	DeviceType  string `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
}

func (cmd *ForwardCmd) Run(globals *Globals) error {
	controller, subordinate, err := cmd.openPorts()
	if err != nil {
		log.Fatalf("error initializing: %v\n", err)
	}
	cmd.runForever(controller, subordinate)
	return nil
}

func (cmd *ForwardCmd) openPorts() (common.Port, common.Port, error) {
	opts := &common.PortOptions{
		Address: cmd.Controller,
		Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:    common.DeviceTypeFromString[cmd.DeviceType],
	}
	controller, err := common.OpenPort(opts)
	if err != nil {
		return nil, nil, err
	}

	opts.Address = cmd.Subordinate
	subordinate, err := common.OpenPort(opts)
	if err != nil {
		controller.Close()
		return nil, nil, err
	}
	return controller, subordinate, nil
}

func (cmd *ForwardCmd) runForever(controller, subordinate common.Port) {
	read := func(p common.Port) ([]byte, error) {
		b := make([]byte, 128)
		n, err := p.Read(b)
		return b[0:n], err
	}

	reopenOnError := func(err error, p common.Port, d, op string) {
		slog.Error(fmt.Sprintf("error %s", op), "error", err, "file", d)
		if err := p.ReopenWithBackoff(); err != nil {
			slog.Error("error reopening", "error", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	readWrite := func(from, to common.Port, fname, tname string) {
		data, err := read(from)
		if err != nil {
			reopenOnError(err, from, fname, "reading")
			return
		}
		slog.Info("writing data", "file", fname, "data", hex.EncodeToString(data))
		_, err = to.Write(data)
		if err != nil {
			reopenOnError(err, to, tname, "writing")
			return
		}
	}

	go func() {
		for {
			readWrite(controller, subordinate, filepath.Base(cmd.Controller), cmd.Subordinate)
		}
	}()

	for {
		readWrite(subordinate, controller, filepath.Base(cmd.Subordinate), cmd.Controller)
	}
}
