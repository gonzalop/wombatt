package bms

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"wombatt/internal/modbus"
)

const (
	modbusBasicInfoAddress       uint16 = 0
	modbusBasicInfoRegisterCount uint8  = 39

	modbusExtraInfoAddress       uint16 = 105
	modbusExtraInfoRegisterCount uint8  = 23
)

type EG4LLv2 struct {
}

func NewEG4LLv2() BMS {
	return &EG4LLv2{}
}

func (*EG4LLv2) InfoInstance() any {
	return &EG4BatteryInfo{}
}

func (*EG4LLv2) DefaultProtocol(deviceType string) string {
	switch deviceType {
	case "tcp":
		return modbus.TCPProtocol
	default:
		return modbus.RTUProtocol
	}
}

func (e *EG4LLv2) ReadInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var info EG4ModbusBatteryInfo
	if data, err := readIntoStruct(&info, reader, timeout, id, modbusBasicInfoAddress, modbusBasicInfoRegisterCount); err != nil {
		return nil, err
	} else if len(data) != (int(modbusBasicInfoRegisterCount) * 2) {
		slog.Debug("data received", "data", hex.EncodeToString(data))
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(data), int(modbusBasicInfoRegisterCount)*2)
	}
	result := EG4BatteryInfo{EG4ModbusBatteryInfo: info}
	result.FullCapacity /= 3600 // FullCapacity is in mAs -> 3600000 == 100Ah
	updateVoltageStats(result.CellVoltages, &result.VoltageStats)
	return &result, nil
}

func (e *EG4LLv2) ReadExtraInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var extra EG4ModbusExtraBatteryInfo
	if data, err := readIntoStruct(&extra, reader, timeout, id, modbusExtraInfoAddress, modbusExtraInfoRegisterCount); err != nil {
		return nil, err
	} else if len(data) != (int(modbusExtraInfoRegisterCount) * 2) {
		slog.Debug("extra data received", "data", hex.EncodeToString(data))
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(data), int(modbusExtraInfoRegisterCount)*2)
	}
	return &extra, nil
}

type EG4ModbusBatteryInfo struct {
	// The following fields must be in the same order as the Modbus registers available
	// starting at address 0 and reading 39 registers.
	// Reference at https://eg4electronics.com/backend/wp-content/uploads/2023/06/EG4-LL-MODBUS-Communication-Protocol.pdf
	Voltage            uint16     `name:"battery_voltage" dclass:"voltage" unit:"V" multiplier:"0.01" precision:"2"`
	Current            int16      `name:"current" dclass:"current" unit:"A" multiplier:"0.01" precision:"2"`
	CellVoltages       [16]uint16 `name:"cell_%d_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
	PCBTemp            int16      `name:"pcb_temp" dclass:"temperature" unit:"°C"`
	MaxTemp            int16      `name:"max_temp" dclass:"temperature" unit:"°C"` // MaxTemp and AvgTemp seem to be swapped in the PDF doc.
	AvgTemp            int16      `name:"avg_temp" dclass:"temperature" unit:"°C" precision:"1"`
	CapRemaining       uint16     `name:"cap_remaining" unit:"%"`
	MaxChargingCurrent uint16     `name:"max_charging_current" dclass:"current" unit:"A" icon:"mdi:current-dc"`
	SOH                uint16     `name:"soh" unit:"%"`
	SOC                uint16     `name:"soc" dclass:"battery" unit:"%"`
	Status             uint16     `name:"status" values:"0:inactive/stand by,1:inactive/charging,2:inactive/discharging,4:inactive/protect,8:inactive/charging limit,32768:active/stand by,32769:active/charging,32770:active/discharging,32772:active/protect,32776:active/charging limit"`
	Warning            uint16     `name:"warning" flags:"0x8000,0x4000,float stopped,low capacity,discharge under temp,charge under temp,discharge over temp,charge over temp,MOS overheating,abnormal ambient temp,discharge overcurrent,charge overcurrent,cell undervoltage,pack undervoltage,cell overvoltage,pack overvoltage"`
	Protection         uint16     `name:"protection" flags:"0x8000,0x4000,discharge short circuit,low capacity,discharge under temp,charge under temp,discharge over temp,charge over temp,MOS overheating,abnormal ambient temp,discharge overcurrent, charge overcurrent,cell undervoltage, pack undervoltage,cell overvoltage,pack overvoltage"`
	ErrorCode          uint16     `name:"error_code" flags:"0x8000,0x4000,0x2000,0x1000,0x0800,0x0400,0x0200,0x0100,0x0080,0x0040,0x0020,cell unbalance,0x0008,current flow error,temperature error,voltage error"`
	CycleCounts        uint32     `name:"cycle_counts" icon:"mdi:battery-sync"`
	FullCapacity       uint32     `name:"full_capacity" unit:"mAh"`
	Temp1              int8       `name:"temp1" dclass:"temperature" unit:"°C" precision:"1"`
	Temp2              int8       `name:"temp2" dclass:"temperature" unit:"°C" precision:"1"`
	Temp3              int8       `name:"temp3" dclass:"temperature" unit:"°C" precision:"1"`
	Temp4              int8       `name:"temp4" dclass:"temperature" unit:"°C" precision:"1"`
	Temp5              int8       `name:"temp5"` // Always 0
	Temp6              int8       `name:"temp6"` // Always 0
	CellNum            uint16     `name:"cell_num"`
	DesignedCapacity   uint16     `name:"designed_capacity" unit:"Ah" multiplier:"0.1" precision:"1"`
	CellBalanceStatus  uint16     `name:"cell_balance_status" flags:"cell 16 unbalanced,cell 15 unbalanced,cell 14 unbalanced,cell 13 unbalanced,cell 12 unbalanced,cell 11 unbalanced,cell 10 unbalanced,cell 9 unbalanced,cell 8 unbalanced,cell 7 unbalanced,cell 6 unbalanced,cell 5 unbalanced,cell 4 unbalanced,cell 3 unbalanced,cell 2 unbalanced,cell 1 unbalanced"`
	// end of Modbus fields.
}

type EG4BatteryInfo struct {
	EG4ModbusBatteryInfo
	VoltageStats
}

type EG4ModbusExtraBatteryInfo struct {
	Model           [24]byte `name:"model" type:"string"`
	FirmwareVersion [6]byte  `name:"firmware_version" type:"string"`
	Serial          [16]byte `name:"serial" type:"string"`
}
