package eg4_18kpv

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

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
				err = fmt.Errorf("unknown eg4_18kpv command: %s", cmd)
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

// RealtimeData holds the values from the "Register Mapping Table" (Input Registers) from EG4-18KPV-12LV-Modbus-Protocol-input-registers.csv.
type RealtimeData struct {
	State                 uint16 `modbus:"0" name:"State" values:"0x00:Standby;0x01:Fault;0x02:Programming;0x04:PV on-grid mode;0x08:PV Charge mode;0x0C:PV Charge+on-grid mode;0x10:Battery on-grid mode;0x14:PV+Battery on-grid mode;0x20:AC Charge mode;0x28:PV+AC charge mode;0x40:Battery off-grid mode;0x80:PV off-grid mode;0xC0:PV+battery off-grid mode;0x88:PV charge +off-grid mode"`
	Vpv1                  uint16 `modbus:"1" name:"PV1 Voltage" unit:"V" multiplier:"0.1"`
	Vpv2                  uint16 `modbus:"2" name:"PV2 Voltage" unit:"V" multiplier:"0.1"`
	Vpv3                  uint16 `modbus:"3" name:"PV3 Voltage" unit:"V" multiplier:"0.1"`
	Vbat                  uint16 `modbus:"4" name:"Battery Voltage" unit:"V" multiplier:"0.1"`
	SOC                   uint16 `modbus:"5" name:"Battery Capacity" unit:"%"`
	SOH                   uint16 `modbus:"6" name:"State of Health" unit:"%"`
	InternalFault         uint16 `modbus:"7" name:"Internal Fault"`
	Ppv1                  uint16 `modbus:"8" name:"PV1 Power" unit:"W"`
	Ppv2                  uint16 `modbus:"9" name:"PV2 Power" unit:"W"`
	Pcharge               uint16 `modbus:"10" name:"Charging Power" unit:"W"`
	Pdischarge            uint16 `modbus:"11" name:"Discharge Power" unit:"W"`
	VacR                  uint16 `modbus:"12" name:"R-phase Mains Voltage" unit:"V" multiplier:"0.1"`
	VacS                  uint16 `modbus:"13" name:"S-phase Mains Voltage" unit:"V" multiplier:"0.1"`
	VacT                  uint16 `modbus:"14" name:"T-phase Mains Voltage" unit:"V" multiplier:"0.1"`
	Fac                   uint16 `modbus:"15" name:"Mains Frequency" unit:"Hz" multiplier:"0.01"`
	Pinv                  uint16 `modbus:"16" name:"Inverter Output Power (Grid Port)" unit:"W"`
	Prec                  uint16 `modbus:"17" name:"AC Charging Rectified Power" unit:"W"`
	LinvRMS               uint16 `modbus:"18" name:"Inverter Current RMS" unit:"A" multiplier:"0.01"`
	PF                    uint16 `modbus:"19" name:"Power Factor" multiplier:"0.001"` // Special calculation needed for display
	VepsR                 uint16 `modbus:"20" name:"R-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	VepsS                 uint16 `modbus:"21" name:"S-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	VepsT                 uint16 `modbus:"22" name:"T-phase Off-grid Output Voltage" unit:"V" multiplier:"0.1"`
	Feps                  uint16 `modbus:"23" name:"Off-grid Output Frequency" unit:"Hz" multiplier:"0.01"`
	Peps                  uint16 `modbus:"24" name:"Off-grid Inverter Power" unit:"W"`
	Seps                  uint16 `modbus:"25" name:"Off-grid Apparent Power" unit:"VA"`
	Ptogrid               uint16 `modbus:"26" name:"Export Power to Grid" unit:"W"`
	Ptouser               uint16 `modbus:"27" name:"Import Power from Grid" unit:"W"`
	Epv1Day               uint16 `modbus:"28" name:"PV1 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	Epv2Day               uint16 `modbus:"29" name:"PV2 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	Epv3Day               uint16 `modbus:"30" name:"PV3 Power Generation Today" unit:"kWh" multiplier:"0.1"`
	EinvDay               uint16 `modbus:"31" name:"Today's Grid-connected Inverter Output Energy" unit:"kWh" multiplier:"0.1"`
	ErecDay               uint16 `modbus:"32" name:"Today's AC Charging Rectified Energy" unit:"kWh" multiplier:"0.1"`
	EchgDay               uint16 `modbus:"33" name:"Charged Energy Today" unit:"kWh" multiplier:"0.1"`
	EdischgDay            uint16 `modbus:"34" name:"Discharged Energy Today" unit:"kWh" multiplier:"0.1"`
	EepsDay               uint16 `modbus:"35" name:"Off-grid Output Energy Today" unit:"kWh" multiplier:"0.1"`
	EtogridDay            uint16 `modbus:"36" name:"Today's Export Energy to Grid" unit:"kWh" multiplier:"0.1"`
	EtouserDay            uint16 `modbus:"37" name:"Today's Import Energy from Grid" unit:"kWh" multiplier:"0.1"`
	Vbus1                 uint16 `modbus:"38" name:"Bus 1 Voltage" unit:"V" multiplier:"0.1"`
	Vbus2                 uint16 `modbus:"39" name:"Bus 2 Voltage" unit:"V" multiplier:"0.1"`
	Epv1All               uint32 `modbus:"40" name:"PV1 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`           // L and H words
	Epv2All               uint32 `modbus:"42" name:"PV2 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`           // L and H words
	Epv3All               uint32 `modbus:"44" name:"PV3 Cumulative Power Generation" unit:"kWh" multiplier:"0.1"`           // L and H words
	EinvAll               uint32 `modbus:"46" name:"Inverter Accumulative Output Energy" unit:"kWh" multiplier:"0.1"`       // L and H words
	ErecAll               uint32 `modbus:"48" name:"AC Charging Accumulative Rectified Energy" unit:"kWh" multiplier:"0.1"` // L and H words
	EchgAll               uint32 `modbus:"50" name:"Cumulative Charge Energy Level" unit:"kWh" multiplier:"0.1"`            // L and H words
	EdischgAll            uint32 `modbus:"52" name:"Cumulative Discharge Energy" unit:"kWh" multiplier:"0.1"`               // L and H words
	EepsAll               uint32 `modbus:"54" name:"Cumulative Off-grid Inverter Power" unit:"kWh" multiplier:"0.1"`        // L and H words
	EtogridAll            uint32 `modbus:"56" name:"Cumulative Export Energy to Grid" unit:"kWh" multiplier:"0.1"`          // L and H words
	EtouserAll            uint32 `modbus:"58" name:"Cumulative Import Energy from Grid" unit:"kWh" multiplier:"0.1"`        // L and H words
	FaultCode             uint32 `modbus:"60" name:"Fault Code"`                                                            // L and H words
	WarningCode           uint32 `modbus:"62" name:"Warning Code"`                                                          // L and H words
	Tinner                uint16 `modbus:"64" name:"Internal Ring Temperature" unit:"°C"`
	Tradiator1            uint16 `modbus:"65" name:"Radiator Temperature 1" unit:"°C"`
	Tradiator2            uint16 `modbus:"66" name:"Radiator Temperature 2" unit:"°C"`
	Tbat                  uint16 `modbus:"67" name:"Battery Temperature" unit:"°C"`
	RunningTime           uint32 `modbus:"69" name:"Running Time" unit:"Second"` // L and H words
	AutoTest              uint16 `modbus:"71" name:"Auto Test"`                  // Bitfield: AutoTestStart (Bit0-3), ubAutoTestStatus (Bit4-7), ubAutoTestStep (Bit8-11)
	WAutoTestLimit        uint16 `modbus:"72" name:"Auto Test Limit" unit:"V/Hz" multiplier:"0.1"`
	UwAutoTestDefaultTime uint16 `modbus:"73" name:"Auto Test Default Time" unit:"ms"`
	UwAutoTestTripValue   uint16 `modbus:"74" name:"Auto Test Trip Value" unit:"V/Hz" multiplier:"0.1"`
	UwAutoTestTripTime    uint16 `modbus:"75" name:"Auto Test Trip Time" unit:"ms"`
	ACInputType           uint16 `modbus:"77" name:"AC Input Type"`
	MaxChgCurr            uint16 `modbus:"81" name:"BMS Limited Maximum Charging Current" unit:"A" multiplier:"0.01"`
	MaxDischgCurr         uint16 `modbus:"82" name:"BMS Limited Maximum Discharge Current" unit:"A" multiplier:"0.01"`
	ChargeVoltRef         uint16 `modbus:"83" name:"BMS Recommended Charging Voltage" unit:"V" multiplier:"0.1"`
	DischgCutVolt         uint16 `modbus:"84" name:"BMS Recommends Discharge Cut-off Voltage" unit:"V" multiplier:"0.1"`
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
	BatStatusINV          uint16 `modbus:"95" name:"Inverter Summarizes Lithium Battery Status Information"`
	BatParallelNum        uint16 `modbus:"96" name:"Number of Batteries in Parallel"`
	BatCapacity           uint16 `modbus:"97" name:"Battery Capacity" unit:"Ah"`
	BatCurrentBMS         int16  `modbus:"98" name:"Battery Current (BMS)" unit:"A" multiplier:"0.01"`
	FaultCodeBMS          uint16 `modbus:"99" name:"Fault Code (BMS)"`
	WarningCodeBMS        uint16 `modbus:"100" name:"Warning Code (BMS)"`
	MaxCellVoltBMS        uint16 `modbus:"101" name:"Maximum Cell Voltage (BMS)" unit:"V" multiplier:"0.001"`
	MinCellVoltBMS        uint16 `modbus:"102" name:"Minimum Cell Voltage (BMS)" unit:"V" multiplier:"0.001"`
	MaxCellTempBMS        int16  `modbus:"103" name:"Maximum Monomer Temperature (BMS)" unit:"°C" multiplier:"0.1"`
	MinCellTempBMS        int16  `modbus:"104" name:"Minimum Monomer Temperature (BMS)" unit:"°C" multiplier:"0.1"`
	BMSFWUpdateState      uint16 `modbus:"105" name:"BMS Firmware Update State"`
	CycleCntBMS           uint16 `modbus:"106" name:"Number of Charge and Discharge Cycles (BMS)"`
	BatVoltSampleINV      uint16 `modbus:"107" name:"Inverter Battery Voltage Sampling" unit:"V" multiplier:"0.1"`
	T1                    uint16 `modbus:"108" name:"12K BT Temperature" unit:"°C" multiplier:"0.1"`
	T2                    uint16 `modbus:"109" name:"Reserved Temperature 2" unit:"°C" multiplier:"0.1"`
	T3                    uint16 `modbus:"110" name:"Reserved Temperature 3" unit:"°C" multiplier:"0.1"`
	T4                    uint16 `modbus:"111" name:"Reserved Temperature 4" unit:"°C" multiplier:"0.1"`
	T5                    uint16 `modbus:"112" name:"Reserved Temperature 5" unit:"°C" multiplier:"0.1"`
	ParallelInfo          uint16 `modbus:"113" name:"Parallel Information"` // Bitfield: MasterOrSlave (Bit0-1), SingleOrThree Phase (Bit2-3), Resvd (Bit4-7), Parallel Num (Bit8-16)
	VBusP                 uint16 `modbus:"120" name:"Half Bus Voltage" unit:"V" multiplier:"0.1"`
	GenVolt               uint16 `modbus:"121" name:"Generator Voltage" unit:"V" multiplier:"0.1"`
	GenFreq               uint16 `modbus:"122" name:"Generator Frequency" unit:"Hz" multiplier:"0.01"`
	GenPower              uint16 `modbus:"123" name:"Generator Power" unit:"W"`
	EgenDay               uint16 `modbus:"124" name:"Daily Energy of Generator" unit:"kWh" multiplier:"0.1"`
	EgenAll               uint32 `modbus:"125" name:"Total Generator Energy" unit:"kWh" multiplier:"0.1"` // L and H words
	EPSVoltL1N            uint16 `modbus:"127" name:"Voltage of EPS L1N" unit:"V" multiplier:"0.1"`
	EPSVoltL2N            uint16 `modbus:"128" name:"Voltage of EPS L2N" unit:"V" multiplier:"0.1"`
	PepsL1N               uint16 `modbus:"129" name:"Active Power of EPS L1N" unit:"W"`
	PepsL2N               uint16 `modbus:"130" name:"Active Power of EPS L2N" unit:"W"`
	SepsL1N               uint16 `modbus:"131" name:"Apparent Power of EPS L1N" unit:"VA"`
	SepsL2N               uint16 `modbus:"132" name:"Apparent Power of EPS L2N" unit:"VA"`
	EepsL1NDay            uint16 `modbus:"133" name:"Daily Energy of EPSL1N" unit:"kWh" multiplier:"0.1"`
	EepsL2NDay            uint16 `modbus:"134" name:"Daily Energy of EPSL2N" unit:"kWh" multiplier:"0.1"`
	EepsL1NAll            uint32 `modbus:"135" name:"Total EPSL1N Energy" unit:"kWh" multiplier:"0.1"` // L and H words
	EepsL2NAll            uint32 `modbus:"137" name:"Total EPSL2N Energy" unit:"kWh" multiplier:"0.1"` // L and H words
	AFCICurrCH1           uint16 `modbus:"140" name:"AFCI Current CH1" unit:"mA"`
	AFCICurrCH2           uint16 `modbus:"141" name:"AFCI Current CH2" unit:"mA"`
	AFCICurrCH3           uint16 `modbus:"142" name:"AFCI Current CH3" unit:"mA"`
	AFCICurrCH4           uint16 `modbus:"143" name:"AFCI Current CH4" unit:"mA"`
	AFCIFlag              uint16 `modbus:"144" name:"AFCI Flag"` // Bitfield: ArcAlarmCH1 (Bit0), ArcAlarmCH2 (Bit1), ArcAlarmCH3 (Bit2), ArcAlarmCH4 (Bit3), SelfTestResultCH1 (Bit4), SelfTestResultCH2 (Bit5), SelfTestResultCH3 (Bit6), SelfTestResultCH4 (Bit7), rsvd (Bit8-15)
	AFCIArcCH1            uint16 `modbus:"145" name:"AFCI Real Time Arc CH1"`
	AFCIArcCH2            uint16 `modbus:"146" name:"AFCI Real Time Arc CH2"`
	AFCIArcCH3            uint16 `modbus:"147" name:"AFCI Real Time Arc CH3"`
	AFCIArcCH4            uint16 `modbus:"148" name:"AFCI Real Time Arc CH4"`
	AFCIMaxArcCH1         uint16 `modbus:"149" name:"AFCI Max Arc CH1"`
	AFCIMaxArcCH2         uint16 `modbus:"150" name:"AFCI Max Arc CH2"`
	AFCIMaxArcCH3         uint16 `modbus:"151" name:"AFCI Max Arc CH3"`
	AFCIMaxArcCH4         uint16 `modbus:"152" name:"AFCI Max Arc CH4"`
}

// ReadRealtimeData reads the real-time running data from the EG4 18kPV inverter.
func ReadRealtimeData(reader modbus.RegisterReader, id uint8) (*RealtimeData, error) {
	// The EG4 18kPV protocol document indicates that registers are read using function code 0x04 (Read Input Registers).
	// We'll read in multiple blocks to avoid reading too many registers in a single call.

	// Block 1: Registers 0x0000 to 0x0044 (0 to 68 decimal) - 69 registers
	data1, err := reader.ReadInputRegisters(id, 0, 69) // 69 registers * 2 bytes/register = 138 bytes
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 0-68: %w", err)
	}

	// Block 2: Registers 0x0045 to 0x0078 (69 to 120 decimal) - 52 registers
	data2, err := reader.ReadInputRegisters(id, 69, 52) // 52 registers * 2 bytes/register = 104 bytes
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 69-120: %w", err)
	}

	// Block 3: Registers 0x0079 to 0x0098 (121 to 152 decimal) - 32 registers
	data3, err := reader.ReadInputRegisters(id, 121, 32) // 32 registers * 2 bytes/register = 64 bytes
	if err != nil {
		return nil, fmt.Errorf("failed to read input registers 121-152: %w", err)
	}

	rtd := &RealtimeData{}

	// Populate rtd from data1
	rtd.State = binary.BigEndian.Uint16(data1[0:2])
	rtd.Vpv1 = binary.BigEndian.Uint16(data1[2:4])
	rtd.Vpv2 = binary.BigEndian.Uint16(data1[4:6])
	rtd.Vpv3 = binary.BigEndian.Uint16(data1[6:8])
	rtd.Vbat = binary.BigEndian.Uint16(data1[8:10])
	rtd.SOC = binary.BigEndian.Uint16(data1[10:12])
	rtd.SOH = binary.BigEndian.Uint16(data1[12:14])
	rtd.InternalFault = binary.BigEndian.Uint16(data1[14:16])
	rtd.Ppv1 = binary.BigEndian.Uint16(data1[16:18])
	rtd.Ppv2 = binary.BigEndian.Uint16(data1[18:20])
	rtd.Pcharge = binary.BigEndian.Uint16(data1[20:22])
	rtd.Pdischarge = binary.BigEndian.Uint16(data1[22:24])
	rtd.VacR = binary.BigEndian.Uint16(data1[24:26])
	rtd.VacS = binary.BigEndian.Uint16(data1[26:28])
	rtd.VacT = binary.BigEndian.Uint16(data1[28:30])
	rtd.Fac = binary.BigEndian.Uint16(data1[30:32])
	rtd.Pinv = binary.BigEndian.Uint16(data1[32:34])
	rtd.Prec = binary.BigEndian.Uint16(data1[34:36])
	rtd.LinvRMS = binary.BigEndian.Uint16(data1[36:38])
	rtd.PF = binary.BigEndian.Uint16(data1[38:40])
	rtd.VepsR = binary.BigEndian.Uint16(data1[40:42])
	rtd.VepsS = binary.BigEndian.Uint16(data1[42:44])
	rtd.VepsT = binary.BigEndian.Uint16(data1[44:46])
	rtd.Feps = binary.BigEndian.Uint16(data1[46:48])
	rtd.Peps = binary.BigEndian.Uint16(data1[48:50])
	rtd.Seps = binary.BigEndian.Uint16(data1[50:52])
	rtd.Ptogrid = binary.BigEndian.Uint16(data1[52:54])
	rtd.Ptouser = binary.BigEndian.Uint16(data1[54:56])
	rtd.Epv1Day = binary.BigEndian.Uint16(data1[56:58])
	rtd.Epv2Day = binary.BigEndian.Uint16(data1[58:60])
	rtd.Epv3Day = binary.BigEndian.Uint16(data1[60:62])
	rtd.EinvDay = binary.BigEndian.Uint16(data1[62:64])
	rtd.ErecDay = binary.BigEndian.Uint16(data1[64:66])
	rtd.EchgDay = binary.BigEndian.Uint16(data1[66:68])
	rtd.EdischgDay = binary.BigEndian.Uint16(data1[68:70])
	rtd.EepsDay = binary.BigEndian.Uint16(data1[70:72])
	rtd.EtogridDay = binary.BigEndian.Uint16(data1[72:74])
	rtd.EtouserDay = binary.BigEndian.Uint16(data1[74:76])
	rtd.Vbus1 = binary.BigEndian.Uint16(data1[76:78])
	rtd.Vbus2 = binary.BigEndian.Uint16(data1[78:80])
	rtd.Epv1All = binary.BigEndian.Uint32(data1[80:84])
	rtd.Epv2All = binary.BigEndian.Uint32(data1[84:88])
	rtd.Epv3All = binary.BigEndian.Uint32(data1[88:92])
	rtd.EinvAll = binary.BigEndian.Uint32(data1[92:96])
	rtd.ErecAll = binary.BigEndian.Uint32(data1[96:100])
	rtd.EchgAll = binary.BigEndian.Uint32(data1[100:104])
	rtd.EdischgAll = binary.BigEndian.Uint32(data1[104:108])
	rtd.EepsAll = binary.BigEndian.Uint32(data1[108:112])
	rtd.EtogridAll = binary.BigEndian.Uint32(data1[112:116])
	rtd.EtouserAll = binary.BigEndian.Uint32(data1[116:120])
	rtd.FaultCode = binary.BigEndian.Uint32(data1[120:124])
	rtd.WarningCode = binary.BigEndian.Uint32(data1[124:128])
	rtd.Tinner = binary.BigEndian.Uint16(data1[128:130])
	rtd.Tradiator1 = binary.BigEndian.Uint16(data1[130:132])
	rtd.Tradiator2 = binary.BigEndian.Uint16(data1[132:134])
	rtd.Tbat = binary.BigEndian.Uint16(data1[134:136])

	// Populate rtd from data2
	rtd.RunningTime = binary.BigEndian.Uint32(data2[0:4])
	rtd.AutoTest = binary.BigEndian.Uint16(data2[4:6])
	rtd.WAutoTestLimit = binary.BigEndian.Uint16(data2[6:8])
	rtd.UwAutoTestDefaultTime = binary.BigEndian.Uint16(data2[8:10])
	rtd.UwAutoTestTripValue = binary.BigEndian.Uint16(data2[10:12])
	rtd.UwAutoTestTripTime = binary.BigEndian.Uint16(data2[12:14])
	rtd.ACInputType = binary.BigEndian.Uint16(data2[16:18])
	rtd.MaxChgCurr = binary.BigEndian.Uint16(data2[24:26])
	rtd.MaxDischgCurr = binary.BigEndian.Uint16(data2[26:28])
	rtd.ChargeVoltRef = binary.BigEndian.Uint16(data2[28:30])
	rtd.DischgCutVolt = binary.BigEndian.Uint16(data2[30:32])
	rtd.BatStatus0BMS = binary.BigEndian.Uint16(data2[32:34])
	rtd.BatStatus1BMS = binary.BigEndian.Uint16(data2[34:36])
	rtd.BatStatus2BMS = binary.BigEndian.Uint16(data2[36:38])
	rtd.BatStatus3BMS = binary.BigEndian.Uint16(data2[38:40])
	rtd.BatStatus4BMS = binary.BigEndian.Uint16(data2[40:42])
	rtd.BatStatus5BMS = binary.BigEndian.Uint16(data2[42:44])
	rtd.BatStatus6BMS = binary.BigEndian.Uint16(data2[44:46])
	rtd.BatStatus7BMS = binary.BigEndian.Uint16(data2[46:48])
	rtd.BatStatus8BMS = binary.BigEndian.Uint16(data2[48:50])
	rtd.BatStatus9BMS = binary.BigEndian.Uint16(data2[50:52])
	rtd.BatStatusINV = binary.BigEndian.Uint16(data2[52:54])
	rtd.BatParallelNum = binary.BigEndian.Uint16(data2[54:56])
	rtd.BatCapacity = binary.BigEndian.Uint16(data2[56:58])
	rtd.BatCurrentBMS = int16(binary.BigEndian.Uint16(data2[58:60]))
	rtd.FaultCodeBMS = binary.BigEndian.Uint16(data2[60:62])
	rtd.WarningCodeBMS = binary.BigEndian.Uint16(data2[62:64])
	rtd.MaxCellVoltBMS = binary.BigEndian.Uint16(data2[64:66])
	rtd.MinCellVoltBMS = binary.BigEndian.Uint16(data2[66:68])
	rtd.MaxCellTempBMS = int16(binary.BigEndian.Uint16(data2[68:70]))
	rtd.MinCellTempBMS = int16(binary.BigEndian.Uint16(data2[70:72]))
	rtd.BMSFWUpdateState = binary.BigEndian.Uint16(data2[72:74])
	rtd.CycleCntBMS = binary.BigEndian.Uint16(data2[74:76])
	rtd.BatVoltSampleINV = binary.BigEndian.Uint16(data2[76:78])
	rtd.T1 = binary.BigEndian.Uint16(data2[78:80])
	rtd.T2 = binary.BigEndian.Uint16(data2[80:82])
	rtd.T3 = binary.BigEndian.Uint16(data2[82:84])
	rtd.T4 = binary.BigEndian.Uint16(data2[84:86])
	rtd.T5 = binary.BigEndian.Uint16(data2[86:88])
	rtd.ParallelInfo = binary.BigEndian.Uint16(data2[88:90])
	rtd.VBusP = binary.BigEndian.Uint16(data2[102:104])

	// Populate rtd from data3
	rtd.GenVolt = binary.BigEndian.Uint16(data3[0:2])
	rtd.GenFreq = binary.BigEndian.Uint16(data3[2:4])
	rtd.GenPower = binary.BigEndian.Uint16(data3[4:6])
	rtd.EgenDay = binary.BigEndian.Uint16(data3[6:8])
	rtd.EgenAll = binary.BigEndian.Uint32(data3[8:12])
	rtd.EPSVoltL1N = binary.BigEndian.Uint16(data3[12:14])
	rtd.EPSVoltL2N = binary.BigEndian.Uint16(data3[14:16])
	rtd.PepsL1N = binary.BigEndian.Uint16(data3[16:18])
	rtd.PepsL2N = binary.BigEndian.Uint16(data3[18:20])
	rtd.SepsL1N = binary.BigEndian.Uint16(data3[20:22])
	rtd.SepsL2N = binary.BigEndian.Uint16(data3[22:24])
	rtd.EepsL1NDay = binary.BigEndian.Uint16(data3[24:26])
	rtd.EepsL2NDay = binary.BigEndian.Uint16(data3[26:28])
	rtd.EepsL1NAll = binary.BigEndian.Uint32(data3[28:32])
	rtd.EepsL2NAll = binary.BigEndian.Uint32(data3[32:36])
	rtd.AFCICurrCH1 = binary.BigEndian.Uint16(data3[38:40])
	rtd.AFCICurrCH2 = binary.BigEndian.Uint16(data3[40:42])
	rtd.AFCICurrCH3 = binary.BigEndian.Uint16(data3[42:44])
	rtd.AFCICurrCH4 = binary.BigEndian.Uint16(data3[44:46])
	rtd.AFCIFlag = binary.BigEndian.Uint16(data3[46:48])
	rtd.AFCIArcCH1 = binary.BigEndian.Uint16(data3[48:50])
	rtd.AFCIArcCH2 = binary.BigEndian.Uint16(data3[50:52])
	rtd.AFCIArcCH3 = binary.BigEndian.Uint16(data3[52:54])
	rtd.AFCIArcCH4 = binary.BigEndian.Uint16(data3[54:56])
	rtd.AFCIMaxArcCH1 = binary.BigEndian.Uint16(data3[56:58])
	rtd.AFCIMaxArcCH2 = binary.BigEndian.Uint16(data3[58:60])
	rtd.AFCIMaxArcCH3 = binary.BigEndian.Uint16(data3[60:62])
	rtd.AFCIMaxArcCH4 = binary.BigEndian.Uint16(data3[62:64])

	return rtd, nil
}

// WriteTo writes the data to the given writer.
func WriteTo(writer io.Writer, data any) {
	cb := func(info map[string]string, val any) {
		fmt.Fprintf(writer, "%s: %v%s\n", info["desc"], val, info["unit"])
	}
	common.TraverseStruct(data, cb)
}
