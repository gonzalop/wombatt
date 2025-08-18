package bms

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"wombatt/internal/modbus"
)

const (
	paceBasicInfoAddress       uint16 = 0
	paceBasicInfoRegisterCount uint8  = 37

	paceExtraInfoAddress       uint16 = 150
	paceExtraInfoRegisterCount uint8  = 30
)

type Pace struct {
}

func NewPace() BMS {
	return &Pace{}
}

func (*Pace) InfoInstance() any {
	return &PaceBatteryInfo{}
}

func (*Pace) DefaultProtocol(deviceType string) string {
	switch deviceType {
	case "tcp":
		return modbus.TCPProtocol
	default:
		return modbus.RTUProtocol
	}
}

func (e *Pace) ReadInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var info PaceModbusBatteryInfo
	if data, err := readIntoStruct(&info, reader, timeout, id, paceBasicInfoAddress, paceBasicInfoRegisterCount); err != nil {
		return nil, err
	} else if len(data) != (int(paceBasicInfoRegisterCount) * 2) {
		slog.Debug("data received", "data", hex.EncodeToString(data))
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(data), int(paceBasicInfoRegisterCount)*2)
	}
	result := PaceBatteryInfo{PaceModbusBatteryInfo: info}
	updateVoltageStats(result.CellVoltages, &result.VoltageStats)
	return &result, nil
}

func (e *Pace) ReadExtraInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var extra PaceModbusExtraBatteryInfo
	if data, err := readIntoStruct(&extra, reader, timeout, id, paceExtraInfoAddress, paceExtraInfoRegisterCount); err != nil {
		return nil, err
	} else if len(data) != (int(paceExtraInfoRegisterCount) * 2) {
		slog.Debug("extra data received", "data", hex.EncodeToString(data))
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(data), int(paceExtraInfoRegisterCount)*2)
	}
	return &extra, nil
}

type PaceModbusBatteryInfo struct {
	// The following fields must be in the same order as the Modbus registers available
	// starting at address 0 and reading 39 registers.
	// Reference at https://github.com/gonzalop/wombatt/blob/main/docs/ref/PACE-BMS-Modbus-Protocol-for-RS485-V1.3-20170627.pdf
	Current           int16      `name:"current" dclass:"current" unit:"A" multiplier:"0.01" precision:"2" icon:"mdi:current-dc"`
	Voltage           uint16     `name:"battery_voltage" dclass:"voltage" unit:"V" multiplier:"0.01" precision:"2"`
	SOC               uint16     `name:"soc" dclass:"battery" unit:"%"`
	SOH               uint16     `name:"soh" unit:"%"`
	RemainingCapacity uint16     `name:"remaining_capacity" unit:"Ah" multiplier:"0.01" precision:"2"`
	FullCapacity      uint16     `name:"full_capacity" unit:"Ah" multiplier:"0.01" precision:"2"`
	DesignCapacity    uint16     `name:"design_capacity" unit:"Ah" multiplier:"0.01" precision:"2"`
	CycleCounts       uint16     `name:"cycle_counts" icon:"mdi:battery-sync"`
	_                 int16      // Reserved
	WarningFlag       uint16     `name:"warning_flag" flags:"SOC low,MOSFET high temp,environment low temp,discharging low temp,charging low temp,discharging high temp,discharging low temp,0x80,0x40,discharging overcurrent,charging overcurrent,pack low voltage,pack overvoltage,cell low voltage,cell overvoltage"`
	ProtectionFlag    uint16     `name:"protection_flag" flags:"0x8000,environment low temp,environment high temp,MOSFET high temp,discharging low temp,charging low temp,discharging high temp,charging high temp,charger overvoltage,short circuit,discharging over current,charging over current,pack low voltage,pack overvoltage,cell low voltage,cell overvoltage"`
	StatusFlag        uint16     `name:"status_flag" flags:"heater,charger inversed,0x2000,charging limiter,discharging MOSFET,charging MOSFET,discharge,charge,0x0080,0x0040,front end sampling comms fault,battery cell fault,0x0008,temp sensor fault,discharging MOSFET fault,charging MOSTFET fault"`
	BalanceStatus     uint16     `name:"balance_status"`
	_                 int16      // Reserved
	_                 int16      // Reserved
	CellVoltages      [16]uint16 `name:"cell_%d_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
	CellTemps         [4]int16   `name:"cell_temp_%d" dclass:"temperature" unit:"°C" multiplier:"0.1" precision:"1"`
	MOSFETTemp        int16      `name:"mosfet_temp" dclass:"temperature" unit:"°C" multiplier:"0.1" precision:"1"`      // Might be 0°C always
	EnvTemp           int16      `name:"environment_temp" dclass:"temperature" unit:"°C" multiplier:"0.1" precision:"1"` // Might be 0°C always
	// Note: there are more documented R/W registers starting at address 60. See the doc.
}

type PaceBatteryInfo struct {
	PaceModbusBatteryInfo
	VoltageStats
}

type PaceModbusExtraBatteryInfo struct {
	Version [20]byte `name:"firmware_version" type:"string"`
	ModelSN [20]byte `name:"model_sn" type:"string"`
	PackSN  [20]byte `name:"pack_sn" type:"string"`
}
