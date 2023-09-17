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
	SerialPort  string        `short:"p" required:"" help:"Serial port used for communication"`
	ID          uint8         `required:"" help:"RTU device ID"`
	Start       uint16        `required:"" help:"Start address of the first register to read"`
	Count       uint8         `required:"" help:"Number of registers to read"`
	BaudRate    uint          `short:"B" default:"9600" help:"Timeout when reading from serial ports"`
	ReadTimeout time.Duration `short:"t" default:"500ms" help:"Baud rate"`
	Protocol    string        `default:"ModbusRTU" enum:"ModbusRTU"`
}

func (mbr *ModbusReadCmd) Run(globals *Globals) error {
	if mbr.ID == 0 || mbr.ID > 247 {
		log.Fatal("id must be between 1 and 247")
	}
	if mbr.Count > 125 {
		log.Fatal("count must be between <= 125")
	}
	portOptions := &common.PortOptions{
		Name: mbr.SerialPort,
		Mode: &serial.Mode{BaudRate: int(mbr.BaudRate)},
	}
	port := common.OpenPortOrFatal(portOptions)
	frame, err := modbus.ReadRegisters(port, mbr.ID, mbr.Start, mbr.Count)
	if err != nil {
		log.Printf("Error reading registers %v: %v\n", mbr.SerialPort, err)
		return nil
	}
	fmt.Printf("%v:\n%s\n", mbr.SerialPort, hex.Dump(frame.RawData()))
	return nil
}
