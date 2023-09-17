package cmd

import (
	"context"
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

func (iq *InverterQueryCmd) Run(globals *Globals) error {
	ctx := context.Background()
	for _, dev := range iq.SerialPorts {
		portOptions := &common.PortOptions{
			Name: dev,
			Mode: &serial.Mode{BaudRate: int(iq.BaudRate)},
			Type: common.DeviceTypeFromString[iq.DeviceType],
		}
		port, err := common.OpenPort(portOptions)
		if err != nil {
			log.Printf("error opening %s: %v\n", dev, err)
			continue
		}
		tctx, cancel := context.WithTimeout(ctx, iq.ReadTimeout)
		results, errors := pi30.RunCommands(tctx, port, iq.Commands)
		cancel()
		if results == nil && len(errors) == 1 {
			fmt.Printf("error running commands on port %s: %v\n", dev, errors[0])
			port.Close()
			continue
		}
		for i, res := range results {
			cmd := iq.Commands[i]
			if errors[i] != nil {
				fmt.Printf("error running %s on port %s: %v\n", cmd, dev, errors[i])
				port.Close()
				continue
			}
			fmt.Printf("Device: %s, Command: %s\n%s\n", dev, cmd, strings.Repeat("=", 40))
			pi30.WriteTo(os.Stdout, res)
		}
		port.Close()
	}
	return nil
}
