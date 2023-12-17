package bms

import (
	"time"

	"wombatt/internal/modbus"
)

const (
	cmdGetAnalogValue uint8 = 0x42
	cmdGetAlarmInfo   uint8 = 0x44
)

type LFP4 struct {
}

func NewLFP4() BMS {
	return &LFP4{}
}

func (*LFP4) InfoInstance() any {
	return &LFP4AnalogValueBatteryInfo{}
}

func (*LFP4) DefaultProtocol(_ string) string {
	return "lifepower4"
}

func (*LFP4) ReadInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var result LFP4AnalogValueBatteryInfo
	if err := readIntoStruct(&result, reader, timeout, id, 0 /*ignored*/, cmdGetAnalogValue); err != nil {
		return nil, err
	}
	return &result, nil
}

func (*LFP4) ReadExtraInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var extra LFP4AlarmInfo
	err := readIntoStruct(&extra, reader, timeout, id, 0 /*ignored*/, cmdGetAlarmInfo)
	if err != nil {
		return nil, err
	}
	return &extra, nil
}

type LFP4AnalogValueBatteryInfo struct {
	// https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
	DataFlag          uint8      `name:"alarm_flag" flags:"0x80,0x40,0x20,0x10,0x08,0x04,no unread alarms,unread alarms"`
	_                 uint8      `name:"pack_group"`
	NumberOfCells     uint8      `skip:"1"` // 16 or 8, if it's 8, loading data will not work!
	CellVoltages      [16]uint16 `name:"cell_%d_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	_                 uint8      `skip:"1"` // always 4
	CellTemps         [4]uint16  `name:"cell_temp_%d" dclass:"temperature" unit:"K" multiplier:"0.1"`
	EnvTemp           uint16     `name:"environment_temp" dclass:"temperature" unit:"K" multiplier:"0.1"`
	MOSFETTemp        uint16     `name:"mosfet_temp" dclass:"temperature" unit:"K" multiplier:"0.1"`
	PackCurrent       int16      `name:"pack_current" dclass:"current" unit:"A" multiplier:"0.01"`
	PackVoltage       int16      `name:"pack_voltage" dclass:"voltage" unit:"V" multiplier:"0.01"`
	CapRemaining      uint16     `name:"remaining_capacity" unit:"Ah" multiplier:"0.01"`
	FullCapacity      uint16     `name:"full_capacity" unit:"Ah" multiplier:"0.01"`
	CycleCounts       uint16     `name:"cycle_counts" icon:"mdi:battery-sync"`
	UserDefined       uint8      `name:"user_defined"`
	SOC               uint16     `name:"soc" dclass:"battery" unit:"%"`
	SOH               uint16     `name:"soh" unit:"%"`
	MaxCellVoltage    uint16     `name:"max_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MinCellVoltage    uint16     `name:"min_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	CellVoltageDiff   uint16     `name:"diff_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MaxCellTemp       uint16     `name:"max_cell_temp" dclass:"temperature" unit:"K" multiplier:"0.1"`
	MinCellTemp       uint16     `name:"min_cell_temp" dclass:"temperature" unit:"K" multiplier:"0.1"`
	CumChargingCap    uint32     `name:"cumulative_charging_capacity" dclass:"current" unit:"A" multiplier:"0.01"`
	CumDischargeCap   uint32     `name:"cumulative_discharge_capacity" dclass:"current" unit:"A" multiplier:"0.01"`
	CumChargingPower  uint32     `name:"cumulative_charging_power" dclass:"power" unit:"kW" multiplier:"0.001"`
	CumDischargePower uint32     `name:"cumulative_discharge_power" dclass:"power" unit:"kW" multiplier:"0.001"`
	CumChargingTime   uint32     `name:"cumulative_charging_time" unit:"h"`
	CumDischargeTime  uint32     `name:"cumulative_discharge_time" unit:"h"`
	CumChargingTimes  uint16     `name:"cumulative_charging_times" unit:"h"`
	CumDischargeTimes uint16     `name:"cumulative_discharge_times" unit:"h"`
}

type LFP4AlarmInfo struct {
	DataFlag               uint8     `name:"alarm_flag" flags:"0x80,0x40,0x20,0x10,0x08,0x04,no unread alarms,unread alarms"`
	_                      uint8     `name:"pack_group"`
	NumberOfCells          uint8     `skip:"1"` // 16 or 8, if it's 8, loading data will not work!
	CellVoltageAlarmStatus [16]uint8 `name:"cell_%d_alarm_status" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	_                      uint8     `skip:"1"` // always 4
	CellTemps              [4]uint8  `name:"cell_temp_%d" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	EnvTempAlarmStatus     uint8     `name:"environment_temp_alarm_status" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	MOSFETTempAlarmStatus  uint8     `name:"mosfet_temp_alarm_status" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	PackCurrentAlarmStatus uint8     `name:"pack_current_alarm_status" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	PackVoltageAlarmStatus uint8     `name:"pack_voltage_alarm_status" flags:"0x80,0x40,0x20,0x10,0x08,0x04,upper limit alarm,lower limit alarm"`
	UserDefined            uint8     `name:"user_defined"`
	BalanceEventCode       uint8     `name:"balance_event_code" flags:"0x80,discharge MOS fault alarm,charge MOS fault alarm,cell voltage difference alarm,0x08,0x04,0x02,balance module"`
	VoltageEventCode       uint8     `name:"voltage_event_code" flags:"pack UV protection,pack UV alarm,pack OV protection,pack OV alarm,cell UV protection,cell UV alarm,cell OV protection,cell OV alarm"`
	TempEventCode          uint16    `name:"temperature_event_code" flags:"0x8000,0x4000,fire alarm event,MOSFET high temperature protection,environment low temparature protection,environment low temparature alarm,environment high temperature protection,environment high temperature alarm,discharge low temperature protection,discharge low temperature alarm,discharge high temperature protection,discharge high temperature alarm,charge low temperature protection,charge low temperature alarm,charge high temperature protection,charge high temperature alarm"`
	CurrentEventCode       uint8     `name:"current_event_code" flags:"output SC lockout,discharge level 2 OC lockout,output SC protection,discharge level 2 OC protection,discharge OC protection,discharge OC alarm,charge OC protection,charge OC alarm"`
	RemainingCapacityAlarm uint8     `name:"current_event_code" flags:"0x80,0x40,0x20,0x10,0x08,0x04,0x02,SOC low alarm"`
	FETStatusCode          uint8     `name:"fet_status_code" flags:"0x80,0x40,0x20,0x10,heater,charge current limiter,charge MOS,discharge MOS"`
	SystemStatusCode       uint8     `name:"system_status_code" flags:"0x80,0x40,0x20,0x10,standby,0x04,charging,discharging"`
	BalanceStatusCode      uint32    `name:"balance_status_code" flags:"0x80000000,0x40000000,0x20000000,0x10000000,0x08000000,0x04000000,0x02000000,0x01000000,0x00800000,0x00400000,0x00200000,0x00100000,0x00080000,0x00040000,0x00020000,0x00010000,cell 16 equalization on,cell 15 equalization on,cell 14 equalization on,cell 13 equalization on,cell 12 equalization on,cell 11 equalization on,cell 10 equalization on,cell 9 equalization on,cell 8 equalization on,cell 7 equalization on,cell 6 equalization on,cell 5 equalization on,cell 4 equalization on,cell 3 equalization on,cell 2 equalization on,cell 1 equalization on"`
	_                      uint8     `skip:"1"`
}
