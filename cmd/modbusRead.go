package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/modbus"

	"go.bug.st/serial"
)

type ModbusReadCmd struct {
	Address     string        `short:"p" required:"" help:"Port or TCP address used for communication"`
	ID          uint8         `required:"" help:"Device ID"`
	Start       uint16        `required:"" help:"Start address of the first register to read"`
	Count       uint8         `required:"" help:"Number of registers to read"`
	BaudRate    uint          `short:"B" default:"9600" help:"Baud rate"`
	ReadTimeout time.Duration `short:"t" default:"500ms" help:"Timeout when reading from serial ports"`
	Protocol    string        `default:"auto" enum:"${protocols}" help:"One of ${protocols}"`
	DeviceType  string        `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
}

func (cmd *ModbusReadCmd) Run(globals *Globals) error {
	if cmd.ID == 0 || cmd.ID > 247 {
		log.Fatal("id must be between 1 and 247")
	}
	if cmd.Count > 125 {
		log.Fatal("count must be between <= 125")
	}
	portOptions := &common.PortOptions{
		Address: cmd.Address,
		Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:    common.DeviceTypeFromString[cmd.DeviceType],
	}
	port := common.OpenPortOrFatal(portOptions)
	reader, err := modbus.ReaderFromProtocol(port, cmd.Protocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	data, err := reader.ReadRegisters(cmd.ID, cmd.Start, cmd.Count)
	if err != nil {
		log.Printf("Error reading registers %v: %v\n", cmd.Address, err)
		log.Fatal(err.Error())
	}
	fmt.Printf("%v ID#%d:\n%s\n", cmd.Address, cmd.ID, hex.Dump(data))
	return nil
}
