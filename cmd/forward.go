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
	f := NewForward(cmd)
	if err := f.Init(); err != nil {
		log.Fatalf("error initializing: %v\n", err)
	}
	f.RunForever()
	return nil
}

type Forward struct {
	ForwardCmd

	controller  common.Port
	subordinate common.Port
}

func NewForward(cmd *ForwardCmd) *Forward {
	return &Forward{ForwardCmd: *cmd}
}

func (f *Forward) Init() error {
	opts := &common.PortOptions{
		Address: f.Controller,
		Mode:    &serial.Mode{BaudRate: int(f.BaudRate)},
		Type:    common.DeviceTypeFromString[f.DeviceType],
	}
	opts.Address = f.Controller
	port, err := common.OpenPort(opts)
	if err != nil {
		return err
	}
	f.controller = port

	opts.Address = f.Subordinate
	port, err = common.OpenPort(opts)
	if err != nil {
		f.controller.Close()
		return err
	}
	f.subordinate = port
	return nil
}

func (f *Forward) RunForever() {
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
			readWrite(f.controller, f.subordinate, filepath.Base(f.Controller), f.Subordinate)
		}
	}()

	for {
		readWrite(f.subordinate, f.controller, filepath.Base(f.Subordinate), f.Controller)
	}
}
