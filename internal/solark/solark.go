package solark

import (
	"encoding/binary"
	"fmt"

	"wombatt/internal/modbus"
)

// RealtimeData holds the values from the "Real-time Running Data" table.
type RealtimeData struct {
	DayActivePowerWh       int16  `modbus:"60" name:"Day Active Power" unit:"kWh" multiplier:"0.1"`
	TotalActivePowerWhLow  uint16 `modbus:"63" name:"Total Active Power Low" unit:"kWh" multiplier:"0.1"`
	TotalActivePowerWhHigh uint16 `modbus:"64" name:"Total Active Power High" unit:"kWh" multiplier:"0.1"`
	GridFrequency          uint16 `modbus:"79" name:"Grid Frequency" unit:"Hz" multiplier:"0.01"`
	DCDCTemp               int16  `modbus:"90" name:"DC/DC Transformer Temperature" unit:"°C" multiplier:"0.1"`
	IGBTHSCTemp            int16  `modbus:"91" name:"IGBT Heat Sink Temperature" unit:"°C" multiplier:"0.1"`
	FaultInfoWord1         uint16 `modbus:"103" name:"Fault Information Word 1" flags:"GFDI_Relay_Failure,Grid_Mode_changed,DC_OverCurr_Fault,SW_AC_OverCurr_Fault,GFCI_Failure,HW_Ac_OverCurr_Fault,Tz_Dc_OverCurr_Fault,Tz_EmergStop_Fault,Tz_GFCI_OC_Fault,DC_Insulation_ISO_Fault,BusUnbalance_Fault,Parallel_Fault,AC_Overload_Fault,AC_WU_OverVolt_Fault,AC_VW_OverVolt_Fault,AC_UV_OverVolt_Fault,Parallel_Aux_Fault,AC_OverFreq_Fault,AC_UnderFreq_Fault,DC_VoltHigh_Fault,DC_VoltLow_Fault,AC_U_GridCurr_High_Fault,Button_Manual_OFF,AC_B_InductCurr_High_Fault,Arc_Fault,Heatsink_HighTemp_Fault"`
	FaultInfoWord2         uint16 `modbus:"104" name:"Fault Information Word 2"`
	FaultInfoWord3         uint16 `modbus:"105" name:"Fault Information Word 3"`
	FaultInfoWord4         uint16 `modbus:"106" name:"Fault Information Word 4"`
	CorrectedBattCapacity  uint16 `modbus:"107" name:"Corrected Battery Capacity" unit:"AH"`
	DailyPVPow             uint16 `modbus:"108" name:"Daily PV Power" unit:"kWh" multiplier:"0.1"`
	DCVoltage1             uint16 `modbus:"109" name:"DC Voltage 1" unit:"V" multiplier:"0.1"`
	DCCurrent1             uint16 `modbus:"110" name:"DC Current 1" unit:"A" multiplier:"0.1"`
	DCVoltage2             uint16 `modbus:"111" name:"DC Voltage 2" unit:"V" multiplier:"0.1"`
	DCCurrent2             uint16 `modbus:"112" name:"DC Current 2" unit:"A" multiplier:"0.1"`

	GridSideVoltageL1N           uint16 `modbus:"150" name:"Grid Side Voltage L1-N" unit:"V" multiplier:"0.1"`
	GridSideVoltageL2N           uint16 `modbus:"151" name:"Grid Side Voltage L2-N" unit:"V" multiplier:"0.1"`
	GridSideVoltageL1L2          uint16 `modbus:"152" name:"Grid Side Voltage L1-L2" unit:"V" multiplier:"0.1"`
	VoltageMiddleRelayL1L2       uint16 `modbus:"153" name:"Voltage at Middle Side of Relay L1-L2" unit:"V" multiplier:"0.1"`
	InverterOutputVoltageL1N     uint16 `modbus:"154" name:"Inverter Output Voltage L1-N" unit:"V" multiplier:"0.1"`
	InverterOutputVoltageL2N     uint16 `modbus:"155" name:"Inverter Output Voltage L2-N" unit:"V" multiplier:"0.1"`
	InverterOutputVoltageL1L2    uint16 `modbus:"156" name:"Inverter Output Voltage L1-L2" unit:"V" multiplier:"0.1"`
	LoadVoltageL1                uint16 `modbus:"157" name:"Load Voltage L1" unit:"V" multiplier:"0.1"`
	LoadVoltageL2                uint16 `modbus:"158" name:"Load Voltage L2" unit:"V" multiplier:"0.1"`
	GridSideCurrentL1            int16  `modbus:"160" name:"Grid Side Current L1" unit:"A" multiplier:"0.01"`
	GridSideCurrentL2            int16  `modbus:"161" name:"Grid Side Current L2" unit:"A" multiplier:"0.01"`
	GridExternalLimiterCurrentL1 int16  `modbus:"162" name:"Grid External Limiter Current L1" unit:"A" multiplier:"0.01"`
	GridExternalLimiterCurrentL2 int16  `modbus:"163" name:"Grid External Limiter Current L2" unit:"A" multiplier:"0.01"`
	InverterOutputCurrentL1      int16  `modbus:"164" name:"Inverter Output Current L1" unit:"A" multiplier:"0.01"`
	InverterOutputCurrentL2      int16  `modbus:"165" name:"Inverter Output Current L2" unit:"A" multiplier:"0.01"`
	GenACOutputPowerInput        int16  `modbus:"166" name:"Gen or AC Coupled Power Input" unit:"W"`
	GridSideL1Power              int16  `modbus:"167" name:"Grid Side L1 Power" unit:"W"`
	GridSideL2Power              int16  `modbus:"168" name:"Grid Side L2 Power" unit:"W"`
	TotalPowerGridSideL1L2       int16  `modbus:"169" name:"Total Power of Grid Side L1-L2" unit:"W"`
	GridExternalLimiter1Power    int16  `modbus:"170" name:"Grid External Limiter 1 Power (CT1)" unit:"W"`
	GridExternalLimiter2Power    int16  `modbus:"171" name:"Grid External Limiter 2 Power (CT2)" unit:"W"`
	GridExternalTotalPower       int16  `modbus:"172" name:"Grid External Total Power" unit:"W"`
	InverterOutputsL1Power       int16  `modbus:"173" name:"Inverter Outputs L1 Power" unit:"W"`
	InverterOutputsL2Power       int16  `modbus:"174" name:"Inverter Outputs L2 Power" unit:"W"`
	InverterOutputTotalPower     int16  `modbus:"175" name:"Inverter Output Total Power" unit:"W"`
	LoadSideL1Power              int16  `modbus:"176" name:"Load Side L1 Power" unit:"W"`
	LoadSideL2Power              int16  `modbus:"177" name:"Load Side L2 Power" unit:"W"`
	LoadSideTotalPower           int16  `modbus:"178" name:"Load Side Total Power" unit:"W"`
	LoadCurrentL1                uint16 `modbus:"179" name:"Load Current L1" unit:"A" multiplier:"0.01"`
	LoadCurrentL2                uint16 `modbus:"180" name:"Load Current L2" unit:"A" multiplier:"0.01"`
	GenPortVoltageL1L2           uint16 `modbus:"181" name:"Gen Port Voltage L1-L2" unit:"V"`
	BatteryTemperature           int16  `modbus:"182" name:"Battery Temperature" unit:"°C" multiplier:"0.1"`
	BatteryVoltage               uint16 `modbus:"183" name:"Battery Voltage" unit:"V" multiplier:"0.01"`
	BatteryCapacitySOC           uint16 `modbus:"184" name:"Battery Capacity SOC" unit:"%"`
	PV1InputPower                uint16 `modbus:"186" name:"PV1 Input Power" unit:"W"`
	PV2InputPower                uint16 `modbus:"187" name:"PV2 Input Power" unit:"W"`
	BatteryOutputPower           int16  `modbus:"190" name:"Battery Output Power" unit:"W"`
	BatteryOutputCurrent         int16  `modbus:"191" name:"Battery Output Current" unit:"A" multiplier:"0.01"`
	LoadFrequency                uint16 `modbus:"192" name:"Load Frequency" unit:"Hz" multiplier:"0.01"`
	InverterOutputFrequency      uint16 `modbus:"193" name:"Inverter Output Frequency" unit:"Hz" multiplier:"0.01"`
	GridSideRelayStatus          uint16 `modbus:"194" name:"Grid Side Relay Status" values:"1:Open (Disconnect),2:Closed"`
	GeneratorSideRelayStatus     uint16 `modbus:"195" name:"Generator Side Relay Status" values:"0:Open,1:Closed,2:No Connection,3:Closed when Generator is on"`
	GeneratorRelayFrequency      uint16 `modbus:"196" name:"Generator Relay Frequency" unit:"Hz" multiplier:"0.01"`
}

// ReadRealtimeData reads the real-time running data from the Solark inverter.
func ReadRealtimeData(reader modbus.RegisterReader, id uint8) (*RealtimeData, error) {
	// The Solark protocol document indicates that registers are read using function code 0x03 (Read Multiple Holding Registers).
	// The registers are not contiguous, so we'll need to make multiple calls.
	// We'll read in blocks to minimize Modbus requests.

	// Block 1: Addr 60-64 (Day Active Power Wh, Total Active Power Wh)
	data1, err := reader.ReadHoldingRegisters(id, 60, 5) // 60, 61, 62, 63, 64
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 60-64: %w", err)
	}

	// Block 2: Addr 79-91 (Grid Frequency, DC/DC Temp, IGBT Temp)
	data2, err := reader.ReadHoldingRegisters(id, 79, 13) // 79-91
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 79-91: %w", err)
	}

	// Block 3: Addr 103-112 (Fault Info, Batt Capacity, Daily PV, DC Volt/Current)
	data3, err := reader.ReadHoldingRegisters(id, 103, 10) // 103-112
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 103-112: %w", err)
	}

	// Block 4: Addr 150-184 (Various Voltages, Currents, Powers, Battery Info)
	data4, err := reader.ReadHoldingRegisters(id, 150, 35) // 150-184
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 150-184: %w", err)
	}

	// Block 5: Addr 186-196 (PV Power, Battery Output, Frequencies, Relay Status)
	data5, err := reader.ReadHoldingRegisters(id, 186, 11) // 186-196
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 186-196: %w", err)
	}

	rtd := &RealtimeData{}

	// Populate rtd from data1
	rtd.DayActivePowerWh = int16(binary.BigEndian.Uint16(data1[0:2]))
	rtd.TotalActivePowerWhLow = binary.BigEndian.Uint16(data1[6:8])
	rtd.TotalActivePowerWhHigh = binary.BigEndian.Uint16(data1[8:10])

	// Populate rtd from data2
	rtd.GridFrequency = binary.BigEndian.Uint16(data2[0:2])
	rtd.DCDCTemp = int16(binary.BigEndian.Uint16(data2[22:24]))
	rtd.IGBTHSCTemp = int16(binary.BigEndian.Uint16(data2[24:26]))

	// Populate rtd from data3
	rtd.FaultInfoWord1 = binary.BigEndian.Uint16(data3[0:2])
	rtd.FaultInfoWord2 = binary.BigEndian.Uint16(data3[2:4])
	rtd.FaultInfoWord3 = binary.BigEndian.Uint16(data3[4:6])
	rtd.FaultInfoWord4 = binary.BigEndian.Uint16(data3[6:8])
	rtd.CorrectedBattCapacity = binary.BigEndian.Uint16(data3[8:10])
	rtd.DailyPVPow = binary.BigEndian.Uint16(data3[10:12])
	rtd.DCVoltage1 = binary.BigEndian.Uint16(data3[12:14])
	rtd.DCCurrent1 = binary.BigEndian.Uint16(data3[14:16])
	rtd.DCVoltage2 = binary.BigEndian.Uint16(data3[16:18])
	rtd.DCCurrent2 = binary.BigEndian.Uint16(data3[18:20])

	// Populate rtd from data4
	rtd.GridSideVoltageL1N = binary.BigEndian.Uint16(data4[0:2])
	rtd.GridSideVoltageL2N = binary.BigEndian.Uint16(data4[2:4])
	rtd.GridSideVoltageL1L2 = binary.BigEndian.Uint16(data4[4:6])
	rtd.VoltageMiddleRelayL1L2 = binary.BigEndian.Uint16(data4[6:8])
	rtd.InverterOutputVoltageL1N = binary.BigEndian.Uint16(data4[8:10])
	rtd.InverterOutputVoltageL2N = binary.BigEndian.Uint16(data4[10:12])
	rtd.InverterOutputVoltageL1L2 = binary.BigEndian.Uint16(data4[12:14])
	rtd.LoadVoltageL1 = binary.BigEndian.Uint16(data4[14:16])
	rtd.LoadVoltageL2 = binary.BigEndian.Uint16(data4[16:18])
	rtd.GridSideCurrentL1 = int16(binary.BigEndian.Uint16(data4[20:22]))
	rtd.GridSideCurrentL2 = int16(binary.BigEndian.Uint16(data4[22:24]))
	rtd.GridExternalLimiterCurrentL1 = int16(binary.BigEndian.Uint16(data4[24:26]))
	rtd.GridExternalLimiterCurrentL2 = int16(binary.BigEndian.Uint16(data4[26:28]))
	rtd.InverterOutputCurrentL1 = int16(binary.BigEndian.Uint16(data4[28:30]))
	rtd.InverterOutputCurrentL2 = int16(binary.BigEndian.Uint16(data4[30:32]))
	rtd.GenACOutputPowerInput = int16(binary.BigEndian.Uint16(data4[32:34]))
	rtd.GridSideL1Power = int16(binary.BigEndian.Uint16(data4[34:36]))
	rtd.GridSideL2Power = int16(binary.BigEndian.Uint16(data4[36:38]))
	rtd.TotalPowerGridSideL1L2 = int16(binary.BigEndian.Uint16(data4[38:40]))
	rtd.GridExternalLimiter1Power = int16(binary.BigEndian.Uint16(data4[40:42]))
	rtd.GridExternalLimiter2Power = int16(binary.BigEndian.Uint16(data4[42:44]))
	rtd.GridExternalTotalPower = int16(binary.BigEndian.Uint16(data4[44:46]))
	rtd.InverterOutputsL1Power = int16(binary.BigEndian.Uint16(data4[46:48]))
	rtd.InverterOutputsL2Power = int16(binary.BigEndian.Uint16(data4[48:50]))
	rtd.InverterOutputTotalPower = int16(binary.BigEndian.Uint16(data4[50:52]))
	rtd.LoadSideL1Power = int16(binary.BigEndian.Uint16(data4[52:54]))
	rtd.LoadSideL2Power = int16(binary.BigEndian.Uint16(data4[54:56]))
	rtd.LoadSideTotalPower = int16(binary.BigEndian.Uint16(data4[56:58]))
	rtd.LoadCurrentL1 = binary.BigEndian.Uint16(data4[58:60])
	rtd.LoadCurrentL2 = binary.BigEndian.Uint16(data4[60:62])
	rtd.GenPortVoltageL1L2 = binary.BigEndian.Uint16(data4[62:64])
	rtd.BatteryTemperature = int16(binary.BigEndian.Uint16(data4[64:66]))
	rtd.BatteryVoltage = binary.BigEndian.Uint16(data4[66:68])
	rtd.BatteryCapacitySOC = binary.BigEndian.Uint16(data4[68:70])

	// Populate rtd from data5
	rtd.PV1InputPower = binary.BigEndian.Uint16(data5[0:2])
	rtd.PV2InputPower = binary.BigEndian.Uint16(data5[2:4])
	rtd.BatteryOutputPower = int16(binary.BigEndian.Uint16(data5[8:10]))
	rtd.BatteryOutputCurrent = int16(binary.BigEndian.Uint16(data5[10:12]))
	rtd.LoadFrequency = binary.BigEndian.Uint16(data5[12:14])
	rtd.InverterOutputFrequency = binary.BigEndian.Uint16(data5[14:16])
	rtd.GridSideRelayStatus = binary.BigEndian.Uint16(data5[16:18])
	rtd.GeneratorSideRelayStatus = binary.BigEndian.Uint16(data5[18:20])
	rtd.GeneratorRelayFrequency = binary.BigEndian.Uint16(data5[20:22])

	return rtd, nil
}

// TotalActivePowerWh combines the low and high words into a single uint32.
func (r *RealtimeData) TotalActivePowerWh() uint32 {
	return uint32(r.TotalActivePowerWhHigh)<<16 | uint32(r.TotalActivePowerWhLow)
}

// IntrinsicAttributes holds the values from the "Intrinsic Attribute Table".
type IntrinsicAttributes struct {
	SNByte01 uint16 `modbus:"3"`
	SNByte02 uint16 `modbus:"4"`
	SNByte03 uint16 `modbus:"5"`
	SNByte04 uint16 `modbus:"6"`
	SNByte05 uint16 `modbus:"7"`
}

// ReadIntrinsicAttributes reads the intrinsic attributes from the Solark inverter.
func ReadIntrinsicAttributes(reader modbus.RegisterReader, id uint8) (*IntrinsicAttributes, error) {
	// The serial number is spread across registers 3-7.
	data, err := reader.ReadHoldingRegisters(id, 3, 5) // Registers 3, 4, 5, 6, 7
	if err != nil {
		return nil, fmt.Errorf("failed to read registers 3-7: %w", err)
	}

	ia := &IntrinsicAttributes{}
	ia.SNByte01 = binary.BigEndian.Uint16(data[0:2])
	ia.SNByte02 = binary.BigEndian.Uint16(data[2:4])
	ia.SNByte03 = binary.BigEndian.Uint16(data[4:6])
	ia.SNByte04 = binary.BigEndian.Uint16(data[6:8])
	ia.SNByte05 = binary.BigEndian.Uint16(data[8:10])

	return ia, nil
}

// SerialNumber combines the SN bytes into a single string.
func (ia *IntrinsicAttributes) SerialNumber() string {
	// The serial number is ten ASCII characters, two per register.
	// Each register holds two ASCII characters, so we need to convert uint16 to two bytes.
	// The PDF says "SN byte 01", "SN byte 02" etc. which implies individual bytes.
	// However, Modbus registers are 16-bit. Assuming each register holds two ASCII characters.
	// The example "AH12345678" implies 10 characters.
	// Register 3: SN byte 01, SN byte 02
	// Register 4: SN byte 03, SN byte 04
	// Register 5: SN byte 05, SN byte 06
	// Register 6: SN byte 07, SN byte 08
	// Register 7: SN byte 09, SN byte 10

	snBytes := make([]byte, 0, 10)
	snBytes = append(snBytes, byte(ia.SNByte01>>8), byte(ia.SNByte01&0xFF))
	snBytes = append(snBytes, byte(ia.SNByte02>>8), byte(ia.SNByte02&0xFF))
	snBytes = append(snBytes, byte(ia.SNByte03>>8), byte(ia.SNByte03&0xFF))
	snBytes = append(snBytes, byte(ia.SNByte04>>8), byte(ia.SNByte04&0xFF))
	snBytes = append(snBytes, byte(ia.SNByte05>>8), byte(ia.SNByte05&0xFF))

	return string(snBytes)
}
