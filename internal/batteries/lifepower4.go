package batteries

import (
	"time"

	"wombatt/internal/modbus"
)

type LFP4 struct {
}

func NewLFP4() Battery {
	return &LFP4{}
}

func (*LFP4) InfoInstance() any {
	return &LFP4AnalogValueBatteryInfo{}
}

func (*LFP4) DefaultProtocol() string {
	return "lifepower4"
}

func (*LFP4) ReadInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	var result LFP4AnalogValueBatteryInfo
	if err := readIntoStruct2(&result, reader, timeout, id, 0, 39); err != nil {
		return nil, err
	}
	return &result, nil
}

func (*LFP4) ReadExtraInfo(reader modbus.RegisterReader, id uint8, timeout time.Duration) (any, error) {
	return nil, nil
}

type LFP4AnalogValueBatteryInfo struct {
	// https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
	DataFlag          uint8      `name:"alarm_flag" flags:"0x80,0x40,0x20,0x10,0x08,0x04,no unread alarms,unread alarms"`
	_                 uint8      `name:"pack_group"`
	NumberOfCells     uint8      `skip:"1"` // 16 or 8, if it's 8, loading data will not work!
	CellVoltages      [16]uint16 `name:"cell_%d_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	_                 uint8      `skip:"1"` // always 4
	CellTemps         [4]uint16  `name:"cell_temp_%d" dclass:"temperature" unit:"°K" multiplier:"0.1"`
	EnvTemp           uint16     `name:"environment_temp" dclass:"temperature" unit:"°K" multiplier:"0.1"`
	MOSFETTemp        uint16     `name:"mosfet_temp" dclass:"temperature" unit:"°K" multiplier:"0.1"`
	PackCurrent       int16      `name:"pack_current" dclass:"current" unit:"A" multiplier:"0.01"`
	PackVoltage       int16      `name:"pack_voltage" dclass:"voltage" unit:"V" multiplier:"0.01"`
	CapRemaining      uint16     `name:"remaining_capacity:" unit:"Ah" multiplier:"0.01"`
	FullCapacity      uint16     `name:"full_capacity:" unit:"Ah" multiplier:"0.01"`
	CycleCounts       uint16     `name:"cycle_counts" icon:"mdi:battery-sync"`
	UserDefined       uint8      `name:"user_defined"`
	SOH               uint16     `name:"soh" unit:"%"`
	SOC               uint16     `name:"soc" dclass:"battery" unit:"%"`
	MaxCellVoltage    uint16     `name:"max_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MinCellVoltage    uint16     `name:"min_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	CellVoltageDiff   uint16     `name:"diff_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MaxCellTemp       uint16     `name:"max_cell_temp" dclass:"temperature" unit:"°K" multiplier:"0.1"`
	MinCellTemp       uint16     `name:"min_cell_temp" dclass:"temperature" unit:"°K" multiplier:"0.1"`
	CumChargingCap    uint32     `name:"cumulative_charging_capacity" dclass:"current" unit:"A" multiplier:"0.01"`
	CumDischargeCap   uint32     `name:"cumulative_discharge_capacity" dclass:"current" unit:"A" multiplier:"0.01"`
	CumChargingPower  uint32     `name:"cumulative_charging_power" dclass:"power" unit:"kWh" multiplier:"0.001"`
	CumDischargePower uint32     `name:"cumulative_discharge_power" dclass:"power" unit:"kWh" multiplier:"0.001"`
	CumChargingTime   uint32     `name:"cumulative_charging_time" unit:"h"`
	CumDischargeTime  uint32     `name:"cumulative_discharge_time" unit:"h"`
	CumChargingTimes  uint16     `name:"cumulative_charging_times" unit:"h"`
	CumDischargeTimes uint16     `name:"cumulative_discharge_times" unit:"h"`
}
