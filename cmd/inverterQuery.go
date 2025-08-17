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
	"wombatt/internal/eg4_18kpv"
	"wombatt/internal/eg4_6000xp"
	"wombatt/internal/pi30"
	"wombatt/internal/solark"
)

// InverterType represents the type of inverter protocol to use.
type InverterType string

const (
	InverterTypePI30      InverterType = "pi30"
	InverterTypeSolark    InverterType = "solark"
	InverterTypeEG418KPV  InverterType = "eg4_18kpv"
	InverterTypeEG46000XP InverterType = "eg4_6000xp"
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
	InverterType InverterType  `short:"I" default:"pi30" enum:"pi30,solark,eg4_18kpv,eg4_6000xp" help:"Type of inverter protocol (pi30, solark, eg4_18kpv, eg4_6000xp)"`
	Protocol     string        `short:"R" default:"auto" enum:"ModbusRTU,ModbusTCP,auto" help:"Modbus protocol (auto, ModbusRTU, ModbusTCP)"`
	ModbusID     int           `short:"i" default:"1" help:"Modbus slave ID"`
}

func (cmd *InverterQueryCmd) Run(globals *Globals) error {
	ctx := context.Background()
	var failed error
	for _, dev := range cmd.Address {
		port, err := common.NewPort(dev, int(cmd.BaudRate), cmd.DataBits, cmd.StopBits, cmd.Parity, cmd.DeviceType)
		if err != nil {
			failed = errors.Join(failed, err)
			continue
		}
		defer port.Close()

		tctx, cancel := context.WithTimeout(ctx, cmd.ReadTimeout)

		var results []any
		var errs []error

		switch cmd.InverterType {
		case InverterTypePI30:
			results, errs = pi30.RunCommands(tctx, port, cmd.Command)
		case InverterTypeSolark:
			results, errs = solark.RunCommands(tctx, port, cmd.Protocol, uint8(cmd.ModbusID), cmd.Command)
		case InverterTypeEG418KPV:
			results, errs = eg4_18kpv.RunCommands(tctx, port, cmd.Protocol, uint8(cmd.ModbusID), cmd.Command)
		case InverterTypeEG46000XP:
			results, errs = eg4_6000xp.RunCommands(tctx, port, cmd.Protocol, uint8(cmd.ModbusID), cmd.Command)
		default:
			cancel()
			port.Close()
			failed = errors.Join(failed, fmt.Errorf("unsupported inverter type: %s", cmd.InverterType))
			continue
		}

		cancel()
		if results == nil && len(errs) == 1 {
			port.Close()
			failed = errors.Join(failed, fmt.Errorf("error running commands on port %s: %w", dev, errs[0]))
			continue
		}
		for i, res := range results {
			command := cmd.Command[i]
			if errs[i] != nil {
				port.Close()
				failed = errors.Join(failed, fmt.Errorf("error running %s on port %s: %w", command, dev, errs[i]))
				continue
			}
			fmt.Printf("Device: %s, Command: %s\n%s\n", dev, command, strings.Repeat("=", 40))
			common.WriteTo(os.Stdout, res)
		}
		port.Close()
	}
	if failed != nil {
		log.Fatal(failed)
	}
	return nil
}
