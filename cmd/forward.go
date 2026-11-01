package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"wombatt/internal/common"

	"go.bug.st/serial"
)

// reopenPause is how long forward waits after reopening a port that failed.
var reopenPause = 500 * time.Millisecond

type ForwardCmd struct {
	Controller  string `name:"controller-port" required:"" help:"Serial port or address of the controller"`
	Subordinate string `name:"subordinate-port" required:"" help:"Serial port or address of the subordinate device"`
	BaudRate    uint   `short:"B" default:"9600" help:"Baud rate"`
	DeviceType  string `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
}

func (cmd *ForwardCmd) Run(globals *Globals, ctx context.Context) error {
	controller, subordinate, err := cmd.openPorts()
	if err != nil {
		log.Fatalf("error initializing: %v\n", err)
	}
	// Closing the ports unblocks any pending read once we are asked to stop.
	context.AfterFunc(ctx, func() {
		controller.Close()
		subordinate.Close()
	})
	cmd.run(ctx, controller, subordinate)
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

func (cmd *ForwardCmd) run(ctx context.Context, controller, subordinate common.Port) {
	read := func(p common.Port) ([]byte, error) {
		b := make([]byte, 128)
		n, err := p.Read(b)
		return b[0:n], err
	}

	reopenOnError := func(err error, p common.Port, d, op string) {
		if ctx.Err() != nil {
			return
		}
		slog.Error(fmt.Sprintf("error %s", op), "error", err, "file", d)
		if err := p.ReopenWithBackoff(ctx); err != nil {
			slog.Error("error reopening", "error", err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(reopenPause):
		}
	}

	readWrite := func(from, to common.Port, fname, tname string) {
		data, err := read(from)
		if common.IsTimeout(err) && len(data) == 0 {
			// An idle line is normal: keep waiting instead of reopening the port.
			return
		}
		if len(data) > 0 {
			// Forward whatever arrived even if the read also returned an error.
			slog.Info("writing data", "file", fname, "data", hex.EncodeToString(data))
			if _, werr := to.Write(data); werr != nil {
				reopenOnError(werr, to, tname, "writing")
				return
			}
		}
		if err != nil {
			reopenOnError(err, from, fname, "reading")
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for ctx.Err() == nil {
			readWrite(controller, subordinate, filepath.Base(cmd.Controller), cmd.Subordinate)
		}
	}()

	go func() {
		defer wg.Done()
		for ctx.Err() == nil {
			readWrite(subordinate, controller, filepath.Base(cmd.Subordinate), cmd.Controller)
		}
	}()

	wg.Wait()
}
