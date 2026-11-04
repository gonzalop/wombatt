package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"wombatt/internal/common"
)

type InverterQueryCmd struct {
	Address      []string      `short:"p" required:"" help:"Ports or addresses used for communication with the inverters"`
	Command      []string      `short:"c" required:"" help:"Commands to send to the inverters"`
	BaudRate     uint          `short:"B" default:"2400" help:"Baud rate"`
	DataBits     int           `help:"Number of data bits for serial port" default:"8"`
	StopBits     int           `help:"Number of stop bits for serial port" default:"1"`
	Parity       string        `help:"Parity for serial port (N, E, O)" default:"N"`
	ReadTimeout  time.Duration `short:"t" default:"5s" help:"Per inverter timeout for processing all the commands being sent"`
	DeviceType   string        `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
	InverterType string        `short:"I" default:"pi30" enum:"${inverter_types}" help:"Type of inverter protocol. One of ${inverter_types}"`
	Protocol     string        `short:"R" default:"auto" enum:"ModbusRTU,ModbusTCP,auto" help:"Modbus protocol (auto, ModbusRTU, ModbusTCP)"`
	ModbusID     int           `short:"i" default:"1" help:"Modbus slave ID"`
}

func (cmd *InverterQueryCmd) Run(globals *Globals, ctx context.Context) error {
	driver := inverterDrivers[cmd.InverterType] // kong's enum only allows known types.
	var failed error
	for _, dev := range cmd.Address {
		if ctx.Err() != nil {
			return nil
		}
		err := func(dev string) error {
			port, err := common.NewPort(dev, cmd.DeviceType, int(cmd.BaudRate), cmd.DataBits, cmd.StopBits, cmd.Parity)
			if err != nil {
				return err
			}
			defer port.Close()

			tctx, cancel := context.WithTimeout(ctx, cmd.ReadTimeout)
			defer cancel()

			results, errs := driver.run(tctx, port, cmd.Protocol, uint8(cmd.ModbusID), cmd.Command)
			return processQueryResults(dev, cmd.Command, results, errs, os.Stdout)
		}(dev)
		if err != nil {
			failed = errors.Join(failed, err)
		}
	}
	if failed != nil {
		log.Fatal(failed)
	}
	return nil
}

func processQueryResults(dev string, commands []string, results []any, errs []error, out io.Writer) error {
	if results == nil && len(errs) > 0 {
		return fmt.Errorf("error running commands on port %s: %w", dev, errors.Join(errs...))
	}
	for i, res := range results {
		command := commands[i]
		if errs[i] != nil {
			return fmt.Errorf("error running %s on port %s: %w", command, dev, errs[i])
		}
		fmt.Fprintf(out, "Device: %s, Command: %s\n%s\n", dev, command, strings.Repeat("=", 40))
		common.WriteTo(out, res)
	}
	return nil
}
