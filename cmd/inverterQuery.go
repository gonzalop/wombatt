package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/pi30"

	"go.bug.st/serial"
)

type InverterQueryCmd struct {
	SerialPorts []string      `short:"p" required:"" help:"Serial ports used for communication with the inverters"`
	Commands    []string      `short:"c" required:"" help:"Commands to send to the inverters"`
	BaudRate    uint          `short:"B" default:"2400" help:"Baud rate"`
	ReadTimeout time.Duration `short:"t" default:"5s" help:"Per inverter timeout for processing all the commands being sent"`
	DeviceType  string        `short:"T" default:"serial" enum:"${device_types}" help:"Device type"`
}

func (cmd *InverterQueryCmd) Run(globals *Globals) error {
	ctx := context.Background()
	var failed error
	for _, dev := range cmd.SerialPorts {
		portOptions := &common.PortOptions{
			Address: dev,
			Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
			Type:    common.DeviceTypeFromString[cmd.DeviceType],
		}
		port, err := common.OpenPort(portOptions)
		if err != nil {
			failed = errors.Join(failed, err)
			continue
		}
		tctx, cancel := context.WithTimeout(ctx, cmd.ReadTimeout)
		results, errs := pi30.RunCommands(tctx, port, cmd.Commands)
		cancel()
		if results == nil && len(errs) == 1 {
			port.Close()
			failed = errors.Join(failed, fmt.Errorf("error running commands on port %s: %w", dev, errs[0]))
			continue
		}
		for i, res := range results {
			cmd := cmd.Commands[i]
			if errs[i] != nil {
				port.Close()
				failed = errors.Join(failed, fmt.Errorf("error running %s on port %s: %w", cmd, dev, errs[i]))
				continue
			}
			fmt.Printf("Device: %s, Command: %s\n%s\n", dev, cmd, strings.Repeat("=", 40))
			pi30.WriteTo(os.Stdout, res)
		}
		port.Close()
	}
	if failed != nil {
		log.Fatal(failed)
	}
	return nil
}
