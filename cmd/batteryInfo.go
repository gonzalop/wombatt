package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"wombatt/internal/batteries"
	"wombatt/internal/common"
	"wombatt/internal/modbus"

	"go.bug.st/serial"
)

type BatteryInfoCmd struct {
	Address     string        `required:"" short:"p" help:"Serial port or address used for communication"`
	IDs         []uint        `short:"i" name:"battery-ids" help:"IDs of the batteries to get info from. Default: 1 thru 64"`
	ReadTimeout time.Duration `short:"t" default:"500ms" help:"Timeout when reading from serial ports"`
	BaudRate    uint          `short:"B" default:"9600" help:"Baud rate"`
	BatteryType BatteryType   `default:"EG4LLv2" help:"Battery type" enum:"${battery_types}"`
	Protocol    string        `default:"auto" enum:"auto,RTU,TCP,"`
	DeviceType  string        `short:"T" default:"serial" enum:"${device_types}" help:"Device type"`
}

func (cmd *BatteryInfoCmd) Run(globals *Globals) error {
	if len(cmd.IDs) == 0 {
		cmd.IDs = make([]uint, 64)
		for i := range cmd.IDs {
			cmd.IDs[i] = uint(i) + 1
		}
	}
	portOptions := &common.PortOptions{
		Address: cmd.Address,
		Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:    common.DeviceTypeFromString[cmd.DeviceType],
	}
	battery := batteries.Instance(string(cmd.BatteryType))
	port := common.OpenPortOrFatal(portOptions)
	reader, err := modbus.ReaderFromProtocol(port, cmd.Protocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer reader.Close()
	for _, id := range cmd.IDs {
		binfo, err := battery.ReadInfo(reader, uint8(id), cmd.ReadTimeout)
		if err != nil {
			if err := port.ReopenWithBackoff(); err != nil {
				log.Fatalf("error reopening port: %v", err)
				return err
			}
			continue
		}
		extra, err := battery.ReadExtraInfo(reader, uint8(id), cmd.ReadTimeout)
		if err != nil {
			if err := port.ReopenWithBackoff(); err != nil {
				log.Fatalf("error reopening port: %v", err)
				return err
			}
			continue
		}
		fmt.Printf("Battery #%d\n===========\n", id)
		writeBatteryInfo(binfo)
		if extra != nil {
			writeBatteryInfo(extra)
		}
		fmt.Println()
	}
	return nil
}

func writeBatteryInfo(bi any) {
	f := func(info map[string]string, value interface{}) {
		name := info["name"]
		unit := info["unit"]
		name = strings.ReplaceAll(name, "_", " ")
		fmt.Printf("%s: %v%s\n", name, value, unit)
	}
	common.TraverseStruct(bi, f)
}
