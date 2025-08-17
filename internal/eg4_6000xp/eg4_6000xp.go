package eg4_6000xp

import (
	"context"
	"encoding/binary"
	"fmt"

	"wombatt/internal/common"
	"wombatt/internal/modbus"
)

func RunCommands(ctx context.Context, port common.Port, protocol string, id uint8, commands []string) ([]any, []error) {
	reader, err := modbus.Reader(port, protocol, "")
	if err != nil {
		var errors []error
		for range commands {
			errors = append(errors, err)
		}
		return nil, errors
	}
	var results []any
	var errors []error

	for _, cmd := range commands {
		type data struct {
			res any
			err error
		}
		ch := make(chan *data, 1)

		go func(cmd string) {
			var res any
			var err error
			switch cmd {
			case "RealtimeData":
				res, err = ReadRealtimeData(reader, id)
			default:
				err = fmt.Errorf("unknown eg4_6000xp command: %s", cmd)
			}
			ch <- &data{res, err}
		}(cmd)

		select {
		case <-ctx.Done():
			results = append(results, nil)
			errors = append(errors, ctx.Err())
		case d := <-ch:
			results = append(results, d.res)
			errors = append(errors, d.err)
		}
	}
	return results, errors
}

// RealtimeData holds the values from the "Register Mapping Table" (Input Registers) from 6kXP-Modbus-20231028.pdf.
type RealtimeData struct {
	State                 uint16 `modbus:"0" name:"State" values:"0x00:Standby;0x01:Fault;0x02:Programming;0x04:PV connected to grid;0x08:PV charging;0x0C:PV charging connected to the grid;0x10:The battery connect to the grid;0x14:(PV+ battery) connected to the grid;0x20:AC charging;0x28:(PV+AC) charging;0x40:The battery is off-grid;0x60:Off-grid + battery charging;0x80:PV off-grid;0xC0:(PV+ battery) off-grid;0x88:PV charging + off-grid"`
	Vpv1                  uint16 `modbus:"1" name:"PV1 Voltage" unit:"V" multiplier:"0.1"`
	Vpv2                  uint16 `modbus:"2" name:"PV2 Voltage" unit:"V" multiplier:"0.1"`
	Vpv3                  uint16 `modbus:"3" name:"PV3 Voltage" unit:"V" multiplier:"0.1"`
	Vbat                  uint16 `modbus:"4" name:"Battery Voltage" unit:"V" multiplier:"0.1"`
	SOC                   uint16 `modbus:"5" name:"Battery Capacity" unit:"%"`
	SOH                   uint16 `modbus:"6" name:"State of Health" unit:"%"`
	Ppv1                  uint16 `modbus:"7" name:"PV1 Power" unit:"W"`
	Ppv2                  uint16 `modbus:"8" name:"PV2 Power" unit:"W"`
	Ppv3                  uint16 `modbus:"9" name:"PV3 Power" unit:"W"`
	Pcharge               uint16 `modbus:"10" name:"Charging Power" unit:"W"`
	Pdischarge            uint16 `modbus:"11" name:"Discharge Power" unit:"W"`
	VacR                  uint16 `modbus:"12" name:"R-phase Utility Grid Voltage" unit:"V" multiplier:"0.1"`
	VacS                  uint16 `modbus:"13" name:"S-phase Utility Grid Voltage" unit:"V" multiplier:"0.1"`
	VacT                  uint16 `modbus:"14" name:"T-phase Utility Grid Voltage" unit:"V" multiplier:"0.1"`
	Fac                   uint16 `modbus:"15" name:"Utility Grid Frequency" unit:"Hz" multiplier:"0.01"`
	Pinv                  uint16 `modbus:"16" name:"On-grid Inverter Power" unit:"W"`
	Prec                  uint16 `modbus:"17" name:"AC Charging Rectification Power" unit:"W"`
	LinvRMS               uint16 `modbus:"18" name:"Inverter RMS Current Output" unit:"A" multiplier:"0.01"`
	PF                    uint16 `modbus:"19" name:"Power Factor" multiplier:"0.001"`
	VepsR                 uint16 `modbus:"20" name:"R-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	VepsS                 uint16 `modbus:"21" name:"S-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	VepsT                 uint16 `modbus:"22" name:"T-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	Feps                  uint16 `modbus:"23" name:"Off-grid Output Frequency" unit:"Hz" multiplier:"0.01"`
	Peps                  uint16 `modbus:"24" name:"Off-grid Inverter Power" unit:"W"`
	Seps                  uint16 `modbus:"25" name:"Off-grid Apparent Power" unit:"VA"`
	Ptogrid               uint16 `modbus:"26" name:"User On-grid Power" unit:"W"`
	Ptouser               uint16 `modbus:"27" name:"Grid Power Capacity" unit:"W"`
	Epv1Day               uint16 `modbus:"28" name:"PV1 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	Epv2Day               uint16 `modbus:"29" name:"PV2 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	Epv3Day               uint16 `modbus:"30" name:"PV3 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	EinvDay               uint16 `modbus:"31" name:"Today's On-grid Inverter Output Energy" unit:"kWh" multiplier:"0.1"`
	ErecDay               uint16 `modbus:"32" name:"Today's AC Charging Rectifier Energy" unit:"kWh" multiplier:"0.1"`
	EchgDay               uint16 `modbus:"33" name:"Energy Charge Today" unit:"kWh" multiplier:"0.1"`
	EdischgDay            uint16 `modbus:"34" name:"Energy Discharge Today" unit:"kWh" multiplier:"0.1"`
	EepsDay               uint16 `modbus:"35" name:"Today's Off-grid Output Energy" unit:"kWh" multiplier:"0.1"`
	EtogridDay            uint16 `modbus:"36" name:"Today's Export to Grid Energy" unit:"kWh" multiplier:"0.1"`
	EtouserDay            uint16 `modbus:"37" name:"Electricity Supplied to User from the Grid Today" unit:"kWh" multiplier:"0.1"`
	Vbus1                 uint16 `modbus:"38" name:"Voltage of Bus 1" unit:"V" multiplier:"0.1"`
	Vbus2                 uint16 `modbus:"39" name:"Voltage of Bus 2" unit:"V" multiplier:"0.1"`
	Epv1All               uint32 `modbus:"40" name:"PV1 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`
	Epv2All               uint32 `modbus:"42" name:"PV2 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`
	Epv3All               uint32 `modbus:"44" name:"PV3 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`
	EinvAll               uint32 `modbus:"46" name:"Inverter Output Accumulated Power" unit:"kWh" multiplier:"0.1"`
	ErecAll               uint32 `modbus:"48" name:"AC Charging Accumulates Rectified Power" unit:"kWh" multiplier:"0.1"`
	EchgAll               uint32 `modbus:"50" name:"Cumulative Charge Energy" unit:"kWh" multiplier:"0.1"`
	EdischgAll            uint32 `modbus:"52" name:"Cumulative Discharge Charge Energy" unit:"kWh" multiplier:"0.1"`
	EepsAll               uint32 `modbus:"54" name:"Cumulative Inverter Off-grid Output Energy" unit:"kWh" multiplier:"0.1"`
	EtogridAll            uint32 `modbus:"56" name:"Accumulate Export Energy" unit:"kWh" multiplier:"0.1"`
	EtouserAll            uint32 `modbus:"58" name:"Cumulative Import Energy" unit:"kWh" multiplier:"0.1"`
	FaultCode             uint32 `modbus:"60" name:"Fault Code"`
	WarningCode           uint32 `modbus:"62" name:"Warning Code"`
	Tinner                uint16 `modbus:"64" name:"Internal Temperature" unit:"°C"`
	Tradiator1            uint16 `modbus:"65" name:"Radiator Temperature 1" unit:"°C"`
	Tradiator2            uint16 `modbus:"66" name:"Radiator Temperature 2" unit:"°C"`
	Tbat                  uint16 `modbus:"67" name:"Battery Temperature" unit:"°C"`
	_                     uint16 `modbus:"68"` // Placeholder for unused register 68
	RunningTime           uint32 `modbus:"69" name:"Runtime Duration" unit:"s"`
	AutoTestInfo          uint16 `modbus:"71" name:"Auto Test Info"`
	WAutoTestLimit        uint16 `modbus:"72" name:"Auto Test Limit"`
	UwAutoTestDefaultTime uint16 `modbus:"73" name:"Auto Test Default Time"`
	UwAutoTestTripValue   uint16 `modbus:"74" name:"Auto Test Trip Value"`
	UwAutoTestTripTime    uint16 `modbus:"75" name:"Auto Test Trip Time"`
	_                     uint16 `modbus:"76"` // Placeholder for unused register 76
	ACInputType           uint16 `modbus:"77" name:"AC Input Type"`
	_                     uint16 `modbus:"78"` // Placeholder for unused register 78
	_                     uint16 `modbus:"79"` // Placeholder for unused register 79
	BatTypeAndBrand       uint16 `modbus:"80" name:"Battery Type and Brand"`
	MaxChgCurr            uint16 `modbus:"81" name:"Max Charging Current (BMS)" unit:"A" multiplier:"0.01"`
	MaxDischgCurr         uint16 `modbus:"82" name:"Max Discharging Current (BMS)" unit:"A" multiplier:"0.01"`
	ChargeVoltRef         uint16 `modbus:"83" name:"Recommended Charging Voltage (BMS)" unit:"V" multiplier:"0.1"`
	DischgCutVolt         uint16 `modbus:"84" name:"Recommended Discharging Cut-off Voltage (BMS)" unit:"V" multiplier:"0.1"`
	BatStatus0BMS         uint16 `modbus:"85" name:"BMS Status Information 0"`
	BatStatus1BMS         uint16 `modbus:"86" name:"BMS Status Information 1"`
	BatStatus2BMS         uint16 `modbus:"87" name:"BMS Status Information 2"`
	BatStatus3BMS         uint16 `modbus:"88" name:"BMS Status Information 3"`
	BatStatus4BMS         uint16 `modbus:"89" name:"BMS Status Information 4"`
	BatStatus5BMS         uint16 `modbus:"90" name:"BMS Status Information 5"`
	BatStatus6BMS         uint16 `modbus:"91" name:"BMS Status Information 6"`
	BatStatus7BMS         uint16 `modbus:"92" name:"BMS Status Information 7"`
	BatStatus8BMS         uint16 `modbus:"93" name:"BMS Status Information 8"`
	BatStatus9BMS         uint16 `modbus:"94" name:"BMS Status Information 9"`
	BatStatusINV          uint16 `modbus:"95" name:"Inverter Aggregates Lithium Battery Status"`
	BatParallelNum        uint16 `modbus:"96" name:"Number of Batteries in Parallel"`
	BatCapacity           uint16 `modbus:"97" name:"Battery Capacity" unit:"Ah"`
	BatCurrentBMS         int16  `modbus:"98" name:"Battery Current (BMS)" unit:"A" multiplier:"0.01"`
	FaultCodeBMS          uint16 `modbus:"99" name:"BMS Fault Code"`
	WarningCodeBMS        uint16 `modbus:"100" name:"BMS Warning Code"`
	MaxCellVoltBMS        uint16 `modbus:"101" name:"Max Cell Voltage (BMS)" unit:"V" multiplier:"0.001"`
	MinCellVoltBMS        uint16 `modbus:"102" name:"Min Cell Voltage (BMS)" unit:"V" multiplier:"0.001"`
	MaxCellTempBMS        int16  `modbus:"103" name:"Max Cell Temperature (BMS)" unit:"°C" multiplier:"0.1"`
	MinCellTempBMS        int16  `modbus:"104" name:"Min Cell Temperature (BMS)" unit:"°C" multiplier:"0.1"`
	BMSFWUpdateState      uint16 `modbus:"105" name:"BMS Firmware Update State"`
	CycleCntBMS           uint16 `modbus:"106" name:"BMS Cycle Count"`
	BatVoltSampleINV      uint16 `modbus:"107" name:"Inverter Samples the Battery Voltage" unit:"V" multiplier:"0.1"`
	_                     uint16 `modbus:"108"` // Placeholder for unused register 108
	_                     uint16 `modbus:"109"` // Placeholder for unused register 109
	_                     uint16 `modbus:"110"` // Placeholder for unused register 110
	_                     uint16 `modbus:"111"` // Placeholder for unused register 111
	_                     uint16 `modbus:"112"` // Placeholder for unused register 112
	ParallelInfo          uint16 `modbus:"113" name:"Parallel Info"`
	OnGridLoadPower       uint16 `modbus:"114" name:"Load Power of the 12k Inverter" unit:"W"`
	_                     uint16 `modbus:"115"` // Placeholder for unused register 115
	_                     uint16 `modbus:"116"` // Placeholder for unused register 116
	_                     uint16 `modbus:"117"` // Placeholder for unused register 117
	_                     uint16 `modbus:"118"` // Placeholder for unused register 118
	_                     uint16 `modbus:"119"` // Placeholder for unused register 119
	VBusP                 uint16 `modbus:"120" name:"Half BUS Voltage" unit:"V" multiplier:"0.1"`
	GenVolt               uint16 `modbus:"121" name:"Generator Voltage" unit:"V" multiplier:"0.1"`
	GenFreq               uint16 `modbus:"122" name:"Generator Frequency" unit:"Hz" multiplier:"0.01"`
	GenPower              uint16 `modbus:"123" name:"Generator Power" unit:"W"`
	EgenDay               uint16 `modbus:"124" name:"Energy of Generator Today" unit:"kWh" multiplier:"0.1"`
	EgenAll               uint32 `modbus:"125" name:"Total Generator Energy" unit:"kWh" multiplier:"0.1"`
	EPSVoltL1N            uint16 `modbus:"127" name:"Voltage of EPS L1N" unit:"V" multiplier:"0.1"`
	EPSVoltL2N            uint16 `modbus:"128" name:"Voltage of EPS L2N" unit:"V" multiplier:"0.1"`
	PepsL1N               uint16 `modbus:"129" name:"Active Power of EPS L1N" unit:"W"`
	PepsL2N               uint16 `modbus:"130" name:"Active Power of EPS L2N" unit:"W"`
	SepsL1N               uint16 `modbus:"131" name:"Apparent Power of EPS L1N" unit:"VA"`
	SepsL2N               uint16 `modbus:"132" name:"Apparent Power of EPS L2N" unit:"VA"`
	EepsL1NDay            uint16 `modbus:"133" name:"Daily Energy of EPSL1N" unit:"kWh" multiplier:"0.1"`
	EepsL2NDay            uint16 `modbus:"134" name:"Daily Energy of EPSL2N" unit:"kWh" multiplier:"0.1"`
	EepsL1NAll            uint32 `modbus:"135" name:"Total EPSL1N Energy" unit:"kWh" multiplier:"0.1"`
	EepsL2NAll            uint32 `modbus:"137" name:"Total EPSL2N Energy" unit:"kWh" multiplier:"0.1"`
	_                     uint16 `modbus:"139"` // Placeholder for unused register 139
	AFCICurrCH1           uint16 `modbus:"140" name:"AFCI Current CH1" unit:"mA"`
	AFCICurrCH2           uint16 `modbus:"141" name:"AFCI Current CH2" unit:"mA"`
	AFCICurrCH3           uint16 `modbus:"142" name:"AFCI Current CH3" unit:"mA"`
	AFCICurrCH4           uint16 `modbus:"143" name:"AFCI Current CH4" unit:"mA"`
	AFCIFlag              uint16 `modbus:"144" name:"AFCI Flag"`
	AFCIArcCH1            uint16 `modbus:"145" name:"AFCI Arc CH1"`
	AFCIArcCH2            uint16 `modbus:"146" name:"AFCI Arc CH2"`
	AFCIArcCH3            uint16 `modbus:"147" name:"AFCI Arc CH3"`
	AFCIArcCH4            uint16 `modbus:"148" name:"AFCI Arc CH4"`
	AFCIMaxArcCH1         uint16 `modbus:"149" name:"AFCI Max Arc CH1"`
	AFCIMaxArcCH2         uint16 `modbus:"150" name:"AFCI Max Arc CH2"`
	AFCIMaxArcCH3         uint16 `modbus:"151" name:"AFCI Max Arc CH3"`
	AFCIMaxArcCH4         uint16 `modbus:"152" name:"AFCI Max Arc CH4"`
	ACCouplePower         uint16 `modbus:"153" name:"AC Coupled Inverter Power" unit:"W"`
	_                     uint16 `modbus:"154"` // Placeholder for unused register 154
	_                     uint16 `modbus:"155"` // Placeholder for unused register 155
	_                     uint16 `modbus:"156"` // Placeholder for unused register 156
	_                     uint16 `modbus:"157"` // Placeholder for unused register 157
	_                     uint16 `modbus:"158"` // Placeholder for unused register 158
	_                     uint16 `modbus:"159"` // Placeholder for unused register 159
	_                     uint16 `modbus:"160"` // Placeholder for unused register 160
	_                     uint16 `modbus:"161"` // Placeholder for unused register 161
	_                     uint16 `modbus:"162"` // Placeholder for unused register 162
	_                     uint16 `modbus:"163"` // Placeholder for unused register 163
	_                     uint16 `modbus:"164"` // Placeholder for unused register 164
	_                     uint16 `modbus:"165"` // Placeholder for unused register 165
	_                     uint16 `modbus:"166"` // Placeholder for unused register 166
	_                     uint16 `modbus:"167"` // Placeholder for unused register 167
	_                     uint16 `modbus:"168"` // Placeholder for unused register 168
	_                     uint16 `modbus:"169"` // Placeholder for unused register 169
	Pload                 uint16 `modbus:"170" name:"Pload" unit:"W"`
	EloadDay              uint16 `modbus:"171" name:"Eload Day" unit:"kWh" multiplier:"0.1"`
	EloadAll              uint32 `modbus:"172" name:"Eload All" unit:"kWh" multiplier:"0.1"`
	SwitchState           uint16 `modbus:"174" name:"Switch State"`
	_                     uint16 `modbus:"175"` // Placeholder for unused register 175
	_                     uint16 `modbus:"176"` // Placeholder for unused register 176
	_                     uint16 `modbus:"177"` // Placeholder for unused register 177
	_                     uint16 `modbus:"178"` // Placeholder for unused register 178
	_                     uint16 `modbus:"179"` // Placeholder for unused register 179
	PinvS                 uint16 `modbus:"180" name:"On-grid Inverter Power (S-phase)" unit:"W"`
	PinvT                 uint16 `modbus:"181" name:"On-grid Inverter Power (T-phase)" unit:"W"`
	PrecS                 uint16 `modbus:"182" name:"Charging Rectification Power (S-phase)" unit:"W"`
	PrecT                 uint16 `modbus:"183" name:"Charging Rectification Power (T-phase)" unit:"W"`
	PtogridS              uint16 `modbus:"184" name:"User On-grid Power (S-phase)" unit:"W"`
	PtogridT              uint16 `modbus:"185" name:"User On-grid Power (T-phase)" unit:"W"`
	PtouserS              uint16 `modbus:"186" name:"Grid Supply Power (S-phase)" unit:"W"`
	PtouserT              uint16 `modbus:"187" name:"Grid Supply Power (T-phase)" unit:"W"`
	GenPowerS             uint16 `modbus:"188" name:"Generator Power (S-phase)" unit:"W"`
	GenPowerT             uint16 `modbus:"189" name:"Generator Power (T-phase)" unit:"W"`
	LinvRMSS              uint16 `modbus:"190" name:"Inverter RMS Current (S-phase)" unit:"A" multiplier:"0.01"`
	LinvRMST              uint16 `modbus:"191" name:"Inverter RMS Current (T-phase)" unit:"A" multiplier:"0.01"`
	PFS                   uint16 `modbus:"192" name:"Power Factor (S-phase)" multiplier:"0.001"`
	PFT                   uint16 `modbus:"193" name:"Power Factor (T-phase)" multiplier:"0.001"`
	_                     uint16 `modbus:"194"` // Placeholder for unused register 194
	_                     uint16 `modbus:"195"` // Placeholder for unused register 195
	_                     uint16 `modbus:"196"` // Placeholder for unused register 196
	_                     uint16 `modbus:"197"` // Placeholder for unused register 197
	_                     uint16 `modbus:"198"` // Placeholder for unused register 198
	_                     uint16 `modbus:"199"` // Placeholder for unused register 199
}

// ReadRealtimeData reads the real-time running data from the EG4 6000XP inverter.
func ReadRealtimeData(reader modbus.RegisterReader, id uint8) (*RealtimeData, error) {
	// The EG4 6000XP protocol document indicates that registers are read using function code 0x04 (Read Input Registers).
	// We'll read in multiple blocks to avoid reading too many registers in a single call.

	// Block 1: Registers 0-39
	data1, err := reader.ReadInputRegisters(id, 0, 40)
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 0-39: %w", err)
	}

	// Block 2: Registers 40-79
	data2, err := reader.ReadInputRegisters(id, 40, 40)
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 40-79: %w", err)
	}

	// Block 3: Registers 80-119
	data3, err := reader.ReadInputRegisters(id, 80, 40)
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 80-119: %w", err)
	}

	// Block 4: Registers 120-159
	data4, err := reader.ReadInputRegisters(id, 120, 40)
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 120-159: %w", err)
	}

	// Block 5: Registers 160-199
	data5, err := reader.ReadInputRegisters(id, 160, 40)
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 160-199: %w", err)
	}

	rtd := &RealtimeData{}

	// Populate rtd from data1
	rtd.State = binary.LittleEndian.Uint16(data1[0:2])
	rtd.Vpv1 = binary.LittleEndian.Uint16(data1[2:4])
	rtd.Vpv2 = binary.LittleEndian.Uint16(data1[4:6])
	rtd.Vpv3 = binary.LittleEndian.Uint16(data1[6:8])
	rtd.Vbat = binary.LittleEndian.Uint16(data1[8:10])
	rtd.SOC = binary.LittleEndian.Uint16(data1[10:12])
	rtd.SOH = binary.LittleEndian.Uint16(data1[12:14])
	rtd.Ppv1 = binary.LittleEndian.Uint16(data1[14:16])
	rtd.Ppv2 = binary.LittleEndian.Uint16(data1[16:18])
	rtd.Ppv3 = binary.LittleEndian.Uint16(data1[18:20])
	rtd.Pcharge = binary.LittleEndian.Uint16(data1[20:22])
	rtd.Pdischarge = binary.LittleEndian.Uint16(data1[22:24])
	rtd.VacR = binary.LittleEndian.Uint16(data1[24:26])
	rtd.VacS = binary.LittleEndian.Uint16(data1[26:28])
	rtd.VacT = binary.LittleEndian.Uint16(data1[28:30])
	rtd.Fac = binary.LittleEndian.Uint16(data1[30:32])
	rtd.Pinv = binary.LittleEndian.Uint16(data1[32:34])
	rtd.Prec = binary.LittleEndian.Uint16(data1[34:36])
	rtd.LinvRMS = binary.LittleEndian.Uint16(data1[36:38])
	rtd.PF = binary.LittleEndian.Uint16(data1[38:40])
	rtd.VepsR = binary.LittleEndian.Uint16(data1[40:42])
	rtd.VepsS = binary.LittleEndian.Uint16(data1[42:44])
	rtd.VepsT = binary.LittleEndian.Uint16(data1[44:46])
	rtd.Feps = binary.LittleEndian.Uint16(data1[46:48])
	rtd.Peps = binary.LittleEndian.Uint16(data1[48:50])
	rtd.Seps = binary.LittleEndian.Uint16(data1[50:52])
	rtd.Ptogrid = binary.LittleEndian.Uint16(data1[52:54])
	rtd.Ptouser = binary.LittleEndian.Uint16(data1[54:56])
	rtd.Epv1Day = binary.LittleEndian.Uint16(data1[56:58])
	rtd.Epv2Day = binary.LittleEndian.Uint16(data1[58:60])
	rtd.Epv3Day = binary.LittleEndian.Uint16(data1[60:62])
	rtd.EinvDay = binary.LittleEndian.Uint16(data1[62:64])
	rtd.ErecDay = binary.LittleEndian.Uint16(data1[64:66])
	rtd.EchgDay = binary.LittleEndian.Uint16(data1[66:68])
	rtd.EdischgDay = binary.LittleEndian.Uint16(data1[68:70])
	rtd.EepsDay = binary.LittleEndian.Uint16(data1[70:72])
	rtd.EtogridDay = binary.LittleEndian.Uint16(data1[72:74])
	rtd.EtouserDay = binary.LittleEndian.Uint16(data1[74:76])
	rtd.Vbus1 = binary.LittleEndian.Uint16(data1[76:78])
	rtd.Vbus2 = binary.LittleEndian.Uint16(data1[78:80])

	// Populate rtd from data2
	rtd.Epv1All = binary.LittleEndian.Uint32(data2[0:4])
	rtd.Epv2All = binary.LittleEndian.Uint32(data2[4:8])
	rtd.Epv3All = binary.LittleEndian.Uint32(data2[8:12])
	rtd.EinvAll = binary.LittleEndian.Uint32(data2[12:16])
	rtd.ErecAll = binary.LittleEndian.Uint32(data2[16:20])
	rtd.EchgAll = binary.LittleEndian.Uint32(data2[20:24])
	rtd.EdischgAll = binary.LittleEndian.Uint32(data2[24:28])
	rtd.EepsAll = binary.LittleEndian.Uint32(data2[28:32])
	rtd.EtogridAll = binary.LittleEndian.Uint32(data2[32:36])
	rtd.EtouserAll = binary.LittleEndian.Uint32(data2[36:40])
	rtd.FaultCode = binary.LittleEndian.Uint32(data2[40:44])
	rtd.WarningCode = binary.LittleEndian.Uint32(data2[44:48])
	rtd.Tinner = binary.LittleEndian.Uint16(data2[48:50])
	rtd.Tradiator1 = binary.LittleEndian.Uint16(data2[50:52])
	rtd.Tradiator2 = binary.LittleEndian.Uint16(data2[52:54])
	rtd.Tbat = binary.LittleEndian.Uint16(data2[54:56])
	rtd.RunningTime = binary.LittleEndian.Uint32(data2[58:62])
	rtd.AutoTestInfo = binary.LittleEndian.Uint16(data2[62:64])
	rtd.WAutoTestLimit = binary.LittleEndian.Uint16(data2[64:66])
	rtd.UwAutoTestDefaultTime = binary.LittleEndian.Uint16(data2[66:68])
	rtd.UwAutoTestTripValue = binary.LittleEndian.Uint16(data2[68:70])
	rtd.UwAutoTestTripTime = binary.LittleEndian.Uint16(data2[70:72])
	rtd.ACInputType = binary.LittleEndian.Uint16(data2[74:76])

	// Populate rtd from data3
	rtd.BatTypeAndBrand = binary.LittleEndian.Uint16(data3[0:2])
	rtd.MaxChgCurr = binary.LittleEndian.Uint16(data3[2:4])
	rtd.MaxDischgCurr = binary.LittleEndian.Uint16(data3[4:6])
	rtd.ChargeVoltRef = binary.LittleEndian.Uint16(data3[6:8])
	rtd.DischgCutVolt = binary.LittleEndian.Uint16(data3[8:10])
	rtd.BatStatus0BMS = binary.LittleEndian.Uint16(data3[10:12])
	rtd.BatStatus1BMS = binary.LittleEndian.Uint16(data3[12:14])
	rtd.BatStatus2BMS = binary.LittleEndian.Uint16(data3[14:16])
	rtd.BatStatus3BMS = binary.LittleEndian.Uint16(data3[16:18])
	rtd.BatStatus4BMS = binary.LittleEndian.Uint16(data3[18:20])
	rtd.BatStatus5BMS = binary.LittleEndian.Uint16(data3[20:22])
	rtd.BatStatus6BMS = binary.LittleEndian.Uint16(data3[22:24])
	rtd.BatStatus7BMS = binary.LittleEndian.Uint16(data3[24:26])
	rtd.BatStatus8BMS = binary.LittleEndian.Uint16(data3[26:28])
	rtd.BatStatus9BMS = binary.LittleEndian.Uint16(data3[28:30])
	rtd.BatStatusINV = binary.LittleEndian.Uint16(data3[30:32])
	rtd.BatParallelNum = binary.LittleEndian.Uint16(data3[32:34])
	rtd.BatCapacity = binary.LittleEndian.Uint16(data3[34:36])
	rtd.BatCurrentBMS = int16(binary.LittleEndian.Uint16(data3[36:38]))
	rtd.FaultCodeBMS = binary.LittleEndian.Uint16(data3[38:40])
	rtd.WarningCodeBMS = binary.LittleEndian.Uint16(data3[40:42])
	rtd.MaxCellVoltBMS = binary.LittleEndian.Uint16(data3[42:44])
	rtd.MinCellVoltBMS = binary.LittleEndian.Uint16(data3[44:46])
	rtd.MaxCellTempBMS = int16(binary.LittleEndian.Uint16(data3[46:48]))
	rtd.MinCellTempBMS = int16(binary.LittleEndian.Uint16(data3[48:50]))
	rtd.BMSFWUpdateState = binary.LittleEndian.Uint16(data3[50:52])
	rtd.CycleCntBMS = binary.LittleEndian.Uint16(data3[52:54])
	rtd.BatVoltSampleINV = binary.LittleEndian.Uint16(data3[54:56])
	rtd.ParallelInfo = binary.LittleEndian.Uint16(data3[66:68])
	rtd.OnGridLoadPower = binary.LittleEndian.Uint16(data3[68:70])

	// Populate rtd from data4
	rtd.VBusP = binary.LittleEndian.Uint16(data4[0:2])
	rtd.GenVolt = binary.LittleEndian.Uint16(data4[2:4])
	rtd.GenFreq = binary.LittleEndian.Uint16(data4[4:6])
	rtd.GenPower = binary.LittleEndian.Uint16(data4[6:8])
	rtd.EgenDay = binary.LittleEndian.Uint16(data4[8:10])
	rtd.EgenAll = binary.LittleEndian.Uint32(data4[10:14])
	rtd.EPSVoltL1N = binary.LittleEndian.Uint16(data4[14:16])
	rtd.EPSVoltL2N = binary.LittleEndian.Uint16(data4[16:18])
	rtd.PepsL1N = binary.LittleEndian.Uint16(data4[18:20])
	rtd.PepsL2N = binary.LittleEndian.Uint16(data4[20:22])
	rtd.SepsL1N = binary.LittleEndian.Uint16(data4[22:24])
	rtd.SepsL2N = binary.LittleEndian.Uint16(data4[24:26])
	rtd.EepsL1NDay = binary.LittleEndian.Uint16(data4[26:28])
	rtd.EepsL2NDay = binary.LittleEndian.Uint16(data4[28:30])
	rtd.EepsL1NAll = binary.LittleEndian.Uint32(data4[30:34])
	rtd.EepsL2NAll = binary.LittleEndian.Uint32(data4[34:38])
	rtd.AFCICurrCH1 = binary.LittleEndian.Uint16(data4[40:42])
	rtd.AFCICurrCH2 = binary.LittleEndian.Uint16(data4[42:44])
	rtd.AFCICurrCH3 = binary.LittleEndian.Uint16(data4[44:46])
	rtd.AFCICurrCH4 = binary.LittleEndian.Uint16(data4[46:48])
	rtd.AFCIFlag = binary.LittleEndian.Uint16(data4[48:50])
	rtd.AFCIArcCH1 = binary.LittleEndian.Uint16(data4[50:52])
	rtd.AFCIArcCH2 = binary.LittleEndian.Uint16(data4[52:54])
	rtd.AFCIArcCH3 = binary.LittleEndian.Uint16(data4[54:56])
	rtd.AFCIArcCH4 = binary.LittleEndian.Uint16(data4[56:58])
	rtd.AFCIMaxArcCH1 = binary.LittleEndian.Uint16(data4[58:60])
	rtd.AFCIMaxArcCH2 = binary.LittleEndian.Uint16(data4[60:62])
	rtd.AFCIMaxArcCH3 = binary.LittleEndian.Uint16(data4[62:64])
	rtd.AFCIMaxArcCH4 = binary.LittleEndian.Uint16(data4[64:66])
	rtd.ACCouplePower = binary.LittleEndian.Uint16(data4[66:68])

	// Populate rtd from data5
	rtd.Pload = binary.LittleEndian.Uint16(data5[20:22])
	rtd.EloadDay = binary.LittleEndian.Uint16(data5[22:24])
	rtd.EloadAll = binary.LittleEndian.Uint32(data5[24:28])
	rtd.SwitchState = binary.LittleEndian.Uint16(data5[28:30])
	rtd.PinvS = binary.LittleEndian.Uint16(data5[40:42])
	rtd.PinvT = binary.LittleEndian.Uint16(data5[42:44])
	rtd.PrecS = binary.LittleEndian.Uint16(data5[44:46])
	rtd.PrecT = binary.LittleEndian.Uint16(data5[46:48])
	rtd.PtogridS = binary.LittleEndian.Uint16(data5[48:50])
	rtd.PtogridT = binary.LittleEndian.Uint16(data5[50:52])
	rtd.PtouserS = binary.LittleEndian.Uint16(data5[52:54])
	rtd.PtouserT = binary.LittleEndian.Uint16(data5[54:56])
	rtd.GenPowerS = binary.LittleEndian.Uint16(data5[56:58])
	rtd.GenPowerT = binary.LittleEndian.Uint16(data5[58:60])
	rtd.LinvRMSS = binary.LittleEndian.Uint16(data5[60:62])
	rtd.LinvRMST = binary.LittleEndian.Uint16(data5[62:64])
	rtd.PFS = binary.LittleEndian.Uint16(data5[64:66])
	rtd.PFT = binary.LittleEndian.Uint16(data5[66:68])

	return rtd, nil
}
