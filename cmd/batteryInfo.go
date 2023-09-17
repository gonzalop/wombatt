package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"wombatt/internal/batteries"
	"wombatt/internal/common"

	"go.bug.st/serial"
)

type BatteryInfoCmd struct {
	SerialPort  string        `required:"" short:"p" help:"Serial port used for communication"`
	IDs         []uint        `short:"i" name:"battery-ids" help:"IDs of the batteries to get info from. Default: 1 thru 64"`
	ReadTimeout time.Duration `short:"t" default:"500ms" help:"Timeout when reading from serial ports"`
	BaudRate    uint          `short:"B" default:"9600" help:"Baud rate"`
	BatteryType BatteryType   `default:"EG4LLv2" help:"Battery type" enum:"${battery_types}"`
}

func (bi *BatteryInfoCmd) Run(globals *Globals) error {
	if len(bi.IDs) == 0 {
		bi.IDs = make([]uint, 64)
		for i := range bi.IDs {
			bi.IDs[i] = uint(i) + 1
		}
	}
	portOptions := &common.PortOptions{
		Name: bi.SerialPort,
		Mode: &serial.Mode{BaudRate: int(bi.BaudRate)},
	}
	battery := batteries.Instance(string(bi.BatteryType))
	port := common.OpenPortOrFatal(portOptions)
	for _, id := range bi.IDs {
		binfo, err := battery.ReadInfo(port, uint8(id), bi.ReadTimeout)
		if err != nil {
			log.Printf("error getting info of ID#%d: %v\n", id, err)
			if err := port.ReopenWithBackoff(); err != nil {
				log.Fatalf("error reopening port: %v", err)
				return err
			}
			continue
		}
		extra, err := battery.ReadExtraInfo(port, uint8(id), bi.ReadTimeout)
		if err != nil {
			log.Printf("error getting extra info of ID#%d: %v\n", id, err)
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
