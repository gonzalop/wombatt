package batteries

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/modbus"

	"golang.org/x/exp/slices"
)

type Battery interface {
	ReadInfo(common.Port, uint8, time.Duration) (any, error)
	ReadExtraInfo(common.Port, uint8, time.Duration) (any, error)
	InfoInstance() any
}

func Instance(batteryType string) Battery {
	switch batteryType {
	case "EG4LLv2":
		return NewEG4LLv2()
	default:
		log.Fatalf("Unsupported battery type: %v", batteryType)
	}
	return nil
}

type EG4LLv2 struct {
}

func NewEG4LLv2() Battery {
	return &EG4LLv2{}
}

func (*EG4LLv2) InfoInstance() any {
	return &EG4ModbusBatteryInfo{}
}

func (*EG4LLv2) ReadInfo(port common.Port, id uint8, timeout time.Duration) (any, error) {
	var info EG4ModbusBatteryInfo
	if err := readIntoStruct(&info, port, timeout, id, 0, 39); err != nil {
		return nil, err
	}
	result := EG4BatteryInfo{EG4ModbusBatteryInfo: info}
	result.FullCapacity /= 3600 // FullCapacity is in mAs -> 3600000 == 100Ah
	updateDerivedFields(&result)
	return &result, nil
}

func (*EG4LLv2) ReadExtraInfo(port common.Port, id uint8, timeout time.Duration) (any, error) {
	var extra EG4ModbusExtraBatteryInfo
	err := readIntoStruct(&extra, port, timeout, id, 105, 23)
	if err != nil {
		return nil, err
	}
	return &extra, nil
}

func readIntoStruct(result any, port common.Port, timeout time.Duration, id uint8, address uint16, count uint8) error {
	frame, err := readWithTimeout(port, timeout, id, address, count)
	if err != nil {
		return err
	}

	data := frame.Data()
	if len(data) != (int(count) * 2) {
		log.Printf("%s\n", hex.EncodeToString(data))
		return fmt.Errorf("unexpected data length: got %d, want 78", len(data))
	}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		log.Printf("error binary read: %v\n", err)
		return err
	}
	return nil
}

func readWithTimeout(port common.Port, timeout time.Duration, id uint8, start uint16, count uint8) (*modbus.RTUFrame, error) {
	var frame *modbus.RTUFrame
	var err error
	result := make(chan struct{}, 1)
	go func() {
		frame, err = modbus.ReadRegisters(port, id, start, count)
		result <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	case <-result:
		return frame, err
	}
}

func updateDerivedFields(bi *EG4BatteryInfo) {
	var sum uint
	voltages := make([]uint16, 16)
	for i := 0; i < 16; i++ {
		mv := bi.CellVoltages[i]
		voltages[i] = mv
		if i == 0 {
			bi.MinVoltage = mv
			bi.MaxVoltage = mv
		}
		sum += uint(mv)
		if bi.MinVoltage > mv {
			bi.MinVoltage = mv
		}
		if bi.MaxVoltage < mv {
			bi.MaxVoltage = mv
		}
	}
	bi.MeanVoltage = uint16(sum / 16)
	slices.Sort(voltages)
	bi.MedianVoltage = (voltages[7] + voltages[8]) / 2
}

type EG4ModbusBatteryInfo struct {
	// The following fields must be in the same order as the Modbus registers available
	// starting at address 0 and reading 39 registers.
	// Reference at https://eg4electronics.com/wp-content/uploads/2023/06/EG4-LL-MODBUS-Communication-Protocol.pdf
	Voltage            uint16     `name:"battery_voltage" dclass:"voltage" unit:"V" multiplier:"0.01"`
	Current            int16      `name:"current" dclass:"current" unit:"A" multiplier:"0.01"`
	CellVoltages       [16]uint16 `name:"cell_%d_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	PCBTemp            int16      `name:"pcb_temp" dclass:"temperature" unit:"°C"`
	MaxTemp            int16      `name:"max_temp" dclass:"temperature" unit:"°C"` // MaxTemp and AvgTemp seem to be swapped in the PDF doc.
	AvgTemp            int16      `name:"avg_temp" dclass:"temperature" unit:"°C"`
	CapRemaining       uint16     `name:"cap_remaining" unit:"%"`
	MaxChargingCurrent uint16     `name:"max_charging_current" dclass:"current" unit:"A"`
	SOH                uint16     `name:"soh" unit:"%"`
	SOC                uint16     `name:"soc" dclass:"battery" unit:"%"`
	Status             uint16     `name:"status" values:"0:inactive/stand by,1:inactive/charging,2:inactive/discharging,4:inactive/protect,8:inactive/charging limit,32768:active/stand by,32769:active/charging,32770:active/discharging,32772:active/protect,32776:active/charging limit"`
	Warning            uint16     `name:"warning" flags:"0x8000,0x4000,float stopped,low capacity,discharge under temp,charge under temp,discharge over temp,charge over temp,MOS overheating,abnormal ambient temp,discharge overcurrent,charge overcurrent,cell undervoltage,pack undervoltage,cell overvoltage,pack overvoltage"`
	Protection         uint16     `name:"protection" flags:"0x8000,0x4000,discharge short circuit,low capacity,discharge under temp,charge under temp,discharge over temp,charge over temp,MOS overheating,abnormal ambient temp,discharge overcurrent, charge overcurrent,cell undervoltage, pack undervoltage,cell overvoltage,pack overvoltage"`
	ErrorCode          uint16     `name:"error_code" flags:"0x8000,0x4000,0x2000,0x1000,0x0800,0x0400,0x0200,0x0100,0x0080,0x0040,0x0020,cell unbalance,0x0008,current flow error,temperature error,voltage error"`
	CycleCounts        uint32     `name:"cycle_counts" icon:"mdi:battery-sync"`
	FullCapacity       uint32     `name:"full_capacity" unit:"mAh"`
	Temp1              int8       `name:"temp1" dclass:"temperature" unit:"°C"`
	Temp2              int8       `name:"temp2" dclass:"temperature" unit:"°C"`
	Temp3              int8       `name:"temp3" dclass:"temperature" unit:"°C"`
	Temp4              int8       `name:"temp4" dclass:"temperature" unit:"°C"`
	Temp5              int8       `name:"temp5"` // Always 0
	Temp6              int8       `name:"temp6"` // Always 0
	CellNum            uint16     `name:"cell_num"`
	DesignedCapacity   uint16     `name:"designed_capacity" unit:"Ah" multiplier:"0.1"`
	CellBalanceStatus  uint16     `name:"cell_balance_status" flags:"cell 16 unbalanced,cell 15 unbalanced,cell 14 unbalanced,cell 13 unbalanced,cell 12 unbalanced,cell 11 unbalanced,cell 10 unbalanced,cell 9 unbalanced,cell 8 unbalanced,cell 7 unbalanced,cell 6 unbalanced,cell 5 unbalanced,cell 4 unbalanced,cell 3 unbalanced,cell 2 unbalanced,cell 1 unbalanced"`
	// end of Modbus fields.
}

type EG4BatteryInfo struct {
	EG4ModbusBatteryInfo

	// Derived data
	MaxVoltage    uint16 `name:"max_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MinVoltage    uint16 `name:"min_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MeanVoltage   uint16 `name:"mean_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MedianVoltage uint16 `name:"median_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
}

type EG4ModbusExtraBatteryInfo struct {
	Model           [24]byte `name:"model" type:"string"`
	FirmwareVersion [6]byte  `name:"firmware_version" type:"string"`
	Serial          [16]byte `name:"serial" type:"string"`
}
