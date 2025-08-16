package eg4_18kpv

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRegisterReader is a mock implementation of the modbus.RegisterReader interface.
type MockRegisterReader struct {
	mock.Mock
}

func (m *MockRegisterReader) ReadHoldingRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	args := m.Called(id, start, count)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRegisterReader) ReadInputRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	args := m.Called(id, start, count)
	return args.Get(0).([]byte), args.Error(1)
}

func TestReadRealtimeData(t *testing.T) {
	mockReader := new(MockRegisterReader)

	// Mock data for RealtimeData (Input Registers 0x0000-0x0098, total 153 registers)
	// Split into 3 blocks as per ReadRealtimeData function

	// Block 1: Registers 0x0000 to 0x0044 (0 to 68 decimal) - 69 registers
	data1 := make([]byte, 69*2) // 138 bytes
	// Populate data1
	binary.BigEndian.PutUint16(data1[0*2:2], 1)        // State: 1
	binary.BigEndian.PutUint16(data1[1*2:4], 3000)     // Vpv1: 300.0V
	binary.BigEndian.PutUint16(data1[2*2:6], 2800)     // Vpv2: 280.0V
	binary.BigEndian.PutUint16(data1[3*2:8], 2500)     // Vpv3: 250.0V
	binary.BigEndian.PutUint16(data1[4*2:10], 480)     // Vbat: 48.0V
	binary.BigEndian.PutUint16(data1[5*2:12], 95)      // SOC: 95%
	binary.BigEndian.PutUint16(data1[6*2:14], 100)     // SOH: 100%
	binary.BigEndian.PutUint16(data1[7*2:16], 0)       // InternalFault: 0
	binary.BigEndian.PutUint16(data1[8*2:18], 1500)    // Ppv1: 1500W
	binary.BigEndian.PutUint16(data1[9*2:20], 1120)    // Ppv2: 1120W
	binary.BigEndian.PutUint16(data1[10*2:22], 500)    // Pcharge: 500W
	binary.BigEndian.PutUint16(data1[11*2:24], 200)    // Pdischarge: 200W
	binary.BigEndian.PutUint16(data1[12*2:26], 2300)   // VacR: 230.0V
	binary.BigEndian.PutUint16(data1[13*2:28], 2310)   // VacS: 231.0V
	binary.BigEndian.PutUint16(data1[14*2:30], 2320)   // VacT: 232.0V
	binary.BigEndian.PutUint16(data1[15*2:32], 5000)   // Fac: 50.00Hz
	binary.BigEndian.PutUint16(data1[16*2:34], 3000)   // Pinv: 3000W
	binary.BigEndian.PutUint16(data1[17*2:36], 1000)   // Prec: 1000W
	binary.BigEndian.PutUint16(data1[18*2:38], 1000)   // LinvRMS: 10.00A
	binary.BigEndian.PutUint16(data1[19*2:40], 950)    // PF: 950
	binary.BigEndian.PutUint16(data1[20*2:42], 2200)   // VepsR: 220.0V
	binary.BigEndian.PutUint16(data1[21*2:44], 2210)   // VepsS: 221.0V
	binary.BigEndian.PutUint16(data1[22*2:46], 2220)   // VepsT: 222.0V
	binary.BigEndian.PutUint16(data1[23*2:48], 6000)   // Feps: 60.00Hz
	binary.BigEndian.PutUint16(data1[24*2:50], 2500)   // Peps: 2500W
	binary.BigEndian.PutUint16(data1[25*2:52], 2800)   // Seps: 2800VA
	binary.BigEndian.PutUint16(data1[26*2:54], 1500)   // Ptogrid: 1500W
	binary.BigEndian.PutUint16(data1[27*2:56], 1000)   // Ptouser: 1000W
	binary.BigEndian.PutUint16(data1[28*2:58], 105)    // Epv1Day: 10.5kWh
	binary.BigEndian.PutUint16(data1[29*2:60], 82)     // Epv2Day: 8.2kWh
	binary.BigEndian.PutUint16(data1[30*2:62], 51)     // Epv3Day: 5.1kWh
	binary.BigEndian.PutUint16(data1[31*2:64], 203)    // EinvDay: 20.3kWh
	binary.BigEndian.PutUint16(data1[32*2:66], 78)     // ErecDay: 7.8kWh
	binary.BigEndian.PutUint16(data1[33*2:68], 121)    // EchgDay: 12.1kWh
	binary.BigEndian.PutUint16(data1[34*2:70], 95)     // EdischgDay: 9.5kWh
	binary.BigEndian.PutUint16(data1[35*2:72], 156)    // EepsDay: 15.6kWh
	binary.BigEndian.PutUint16(data1[36*2:74], 67)     // EtogridDay: 6.7kWh
	binary.BigEndian.PutUint16(data1[37*2:76], 43)     // EtouserDay: 4.3kWh
	binary.BigEndian.PutUint16(data1[38*2:78], 4000)   // Vbus1: 400.0V
	binary.BigEndian.PutUint16(data1[39*2:80], 3900)   // Vbus2: 390.0V
	binary.BigEndian.PutUint32(data1[40*2:84], 10005)  // Epv1All: 1000.5kWh
	binary.BigEndian.PutUint32(data1[42*2:88], 8002)   // Epv2All: 800.2kWh
	binary.BigEndian.PutUint32(data1[44*2:92], 5001)   // Epv3All: 500.1kWh
	binary.BigEndian.PutUint32(data1[46*2:96], 20003)  // EinvAll: 2000.3kWh
	binary.BigEndian.PutUint32(data1[48*2:100], 7008)  // ErecAll: 700.8kWh
	binary.BigEndian.PutUint32(data1[50*2:104], 12001) // EchgAll: 1200.1kWh
	binary.BigEndian.PutUint32(data1[52*2:108], 9005)  // EdischgAll: 900.5kWh
	binary.BigEndian.PutUint32(data1[54*2:112], 15006) // EepsAll: 1500.6kWh
	binary.BigEndian.PutUint32(data1[56*2:116], 6007)  // EtogridAll: 600.7kWh
	binary.BigEndian.PutUint32(data1[58*2:120], 4003)  // EtouserAll: 400.3kWh
	binary.BigEndian.PutUint32(data1[60*2:124], 1)     // FaultCode: 1
	binary.BigEndian.PutUint32(data1[62*2:128], 2)     // WarningCode: 2
	binary.BigEndian.PutUint16(data1[64*2:130], 50)    // Tinner: 50C
	binary.BigEndian.PutUint16(data1[65*2:132], 45)    // Tradiator1: 45C
	binary.BigEndian.PutUint16(data1[66*2:134], 40)    // Tradiator2: 40C
	binary.BigEndian.PutUint16(data1[67*2:136], 25)    // Tbat: 25C

	// Block 2: Registers 0x0045 to 0x0078 (69 to 120 decimal) - 52 registers
	data2 := make([]byte, 52*2) // 104 bytes
	// Populate data2
	binary.BigEndian.PutUint32(data2[0*2:4], 3600)    // RunningTime: 3600 seconds
	binary.BigEndian.PutUint16(data2[2*2:6], 0)       // AutoTest: 0
	binary.BigEndian.PutUint16(data2[3*2:8], 100)     // WAutoTestLimit: 10.0V
	binary.BigEndian.PutUint16(data2[4*2:10], 1000)   // UwAutoTestDefaultTime: 1000ms
	binary.BigEndian.PutUint16(data2[5*2:12], 50)     // UwAutoTestTripValue: 5.0V
	binary.BigEndian.PutUint16(data2[6*2:14], 500)    // UwAutoTestTripTime: 500ms
	binary.BigEndian.PutUint16(data2[8*2:18], 0)      // ACInputType: 0
	binary.BigEndian.PutUint16(data2[12*2:26], 5000)  // MaxChgCurr: 50.00A
	binary.BigEndian.PutUint16(data2[13*2:28], 6000)  // MaxDischgCurr: 60.00A
	binary.BigEndian.PutUint16(data2[14*2:30], 500)   // ChargeVoltRef: 50.0V
	binary.BigEndian.PutUint16(data2[15*2:32], 450)   // DischgCutVolt: 45.0V
	binary.BigEndian.PutUint16(data2[16*2:34], 0)     // BatStatus0BMS: 0
	binary.BigEndian.PutUint16(data2[17*2:36], 0)     // BatStatus1BMS: 0
	binary.BigEndian.PutUint16(data2[18*2:38], 0)     // BatStatus2BMS: 0
	binary.BigEndian.PutUint16(data2[19*2:40], 0)     // BatStatus3BMS: 0
	binary.BigEndian.PutUint16(data2[20*2:42], 0)     // BatStatus4BMS: 0
	binary.BigEndian.PutUint16(data2[21*2:44], 0)     // BatStatus5BMS: 0
	binary.BigEndian.PutUint16(data2[22*2:46], 0)     // BatStatus6BMS: 0
	binary.BigEndian.PutUint16(data2[23*2:48], 0)     // BatStatus7BMS: 0
	binary.BigEndian.PutUint16(data2[24*2:50], 0)     // BatStatus8BMS: 0
	binary.BigEndian.PutUint16(data2[25*2:52], 0)     // BatStatus9BMS: 0
	binary.BigEndian.PutUint16(data2[26*2:54], 0)     // BatStatusINV: 0
	binary.BigEndian.PutUint16(data2[27*2:56], 1)     // BatParallelNum: 1
	binary.BigEndian.PutUint16(data2[28*2:58], 100)   // BatCapacity: 100Ah
	binary.BigEndian.PutUint16(data2[29*2:60], 1000)  // BatCurrentBMS: 10.00A (raw value)
	binary.BigEndian.PutUint16(data2[30*2:62], 0)     // FaultCodeBMS: 0
	binary.BigEndian.PutUint16(data2[31*2:64], 0)     // WarningCodeBMS: 0
	binary.BigEndian.PutUint16(data2[32*2:66], 3600)  // MaxCellVoltBMS: 3.600V
	binary.BigEndian.PutUint16(data2[33*2:68], 3200)  // MinCellVoltBMS: 3.200V
	binary.BigEndian.PutUint16(data2[34*2:70], 300)   // MaxCellTempBMS: 30.0C (raw value)
	binary.BigEndian.PutUint16(data2[35*2:72], 200)   // MinCellTempBMS: 20.0C (raw value)
	binary.BigEndian.PutUint16(data2[36*2:74], 2)     // BMSFWUpdateState: 2
	binary.BigEndian.PutUint16(data2[37*2:76], 500)   // CycleCntBMS: 500
	binary.BigEndian.PutUint16(data2[38*2:78], 490)   // BatVoltSampleINV: 49.0V
	binary.BigEndian.PutUint16(data2[39*2:80], 280)   // T1: 28.0C
	binary.BigEndian.PutUint16(data2[40*2:82], 270)   // T2: 27.0C
	binary.BigEndian.PutUint16(data2[41*2:84], 260)   // T3: 26.0C
	binary.BigEndian.PutUint16(data2[42*2:86], 250)   // T4: 25.0C
	binary.BigEndian.PutUint16(data2[43*2:88], 240)   // T5: 24.0C
	binary.BigEndian.PutUint16(data2[44*2:90], 1)     // ParallelInfo: 1
	binary.BigEndian.PutUint16(data2[51*2:104], 2000) // VBusP: 200.0V

	// Block 3: Registers 0x0079 to 0x0098 (121 to 152 decimal) - 32 registers
	data3 := make([]byte, 32*2) // 64 bytes
	// Populate data3
	binary.BigEndian.PutUint16(data3[0*2:2], 2400)   // GenVolt: 240.0V
	binary.BigEndian.PutUint16(data3[1*2:4], 6000)   // GenFreq: 60.00Hz
	binary.BigEndian.PutUint16(data3[2*2:6], 5000)   // GenPower: 5000W
	binary.BigEndian.PutUint16(data3[3*2:8], 100)    // EgenDay: 10.0kWh
	binary.BigEndian.PutUint32(data3[4*2:12], 10000) // EgenAll: 1000.0kWh
	binary.BigEndian.PutUint16(data3[6*2:14], 2200)  // EPSVoltL1N: 220.0V
	binary.BigEndian.PutUint16(data3[7*2:16], 2210)  // EPSVoltL2N: 221.0V
	binary.BigEndian.PutUint16(data3[8*2:18], 1000)  // PepsL1N: 1000W
	binary.BigEndian.PutUint16(data3[9*2:20], 800)   // PepsL2N: 800W
	binary.BigEndian.PutUint16(data3[10*2:22], 1200) // SepsL1N: 1200VA
	binary.BigEndian.PutUint16(data3[11*2:24], 1000) // SepsL2N: 1000VA
	binary.BigEndian.PutUint16(data3[12*2:26], 50)   // EepsL1NDay: 5.0kWh
	binary.BigEndian.PutUint16(data3[13*2:28], 40)   // EepsL2NDay: 4.0kWh
	binary.BigEndian.PutUint32(data3[14*2:32], 5000) // EepsL1NAll: 500.0kWh
	binary.BigEndian.PutUint32(data3[16*2:36], 4000) // EepsL2NAll: 400.0kWh
	binary.BigEndian.PutUint16(data3[19*2:40], 100)  // AFCICurrCH1: 100mA
	binary.BigEndian.PutUint16(data3[20*2:42], 80)   // AFCICurrCH2: 80mA
	binary.BigEndian.PutUint16(data3[21*2:44], 60)   // AFCICurrCH3: 60mA
	binary.BigEndian.PutUint16(data3[22*2:46], 40)   // AFCICurrCH4: 40mA
	binary.BigEndian.PutUint16(data3[23*2:48], 0)    // AFCIFlag: 0
	binary.BigEndian.PutUint16(data3[24*2:50], 10)   // AFCIArcCH1: 10
	binary.BigEndian.PutUint16(data3[25*2:52], 8)    // AFCIArcCH2: 8
	binary.BigEndian.PutUint16(data3[26*2:54], 6)    // AFCIArcCH3: 6
	binary.BigEndian.PutUint16(data3[27*2:56], 4)    // AFCIArcCH4: 4
	binary.BigEndian.PutUint16(data3[28*2:58], 20)   // AFCIMaxArcCH1: 20
	binary.BigEndian.PutUint16(data3[29*2:60], 15)   // AFCIMaxArcCH2: 15
	binary.BigEndian.PutUint16(data3[30*2:62], 10)   // AFCIMaxArcCH3: 10
	binary.BigEndian.PutUint16(data3[31*2:64], 5)    // AFCIMaxArcCH4: 5

	mockReader.On("ReadInputRegisters", uint8(0), uint16(0), uint8(69)).Return(data1, nil).Once()
	mockReader.On("ReadInputRegisters", uint8(0), uint16(69), uint8(52)).Return(data2, nil).Once()
	mockReader.On("ReadInputRegisters", uint8(0), uint16(121), uint8(32)).Return(data3, nil).Once()

	rtd, err := ReadRealtimeData(mockReader, 0)
	assert.NoError(t, err)
	assert.NotNil(t, rtd)

	// Assertions for data1 fields
	assert.Equal(t, uint16(1), rtd.State)
	assert.Equal(t, uint16(3000), rtd.Vpv1)
	assert.Equal(t, uint16(2800), rtd.Vpv2)
	assert.Equal(t, uint16(2500), rtd.Vpv3)
	assert.Equal(t, uint16(480), rtd.Vbat)
	assert.Equal(t, uint16(95), rtd.SOC)
	assert.Equal(t, uint16(100), rtd.SOH)
	assert.Equal(t, uint16(0), rtd.InternalFault)
	assert.Equal(t, uint16(1500), rtd.Ppv1)
	assert.Equal(t, uint16(1120), rtd.Ppv2)
	assert.Equal(t, uint16(500), rtd.Pcharge)
	assert.Equal(t, uint16(200), rtd.Pdischarge)
	assert.Equal(t, uint16(2300), rtd.VacR)
	assert.Equal(t, uint16(2310), rtd.VacS)
	assert.Equal(t, uint16(2320), rtd.VacT)
	assert.Equal(t, uint16(5000), rtd.Fac)
	assert.Equal(t, uint16(3000), rtd.Pinv)
	assert.Equal(t, uint16(1000), rtd.Prec)
	assert.Equal(t, uint16(1000), rtd.LinvRMS)
	assert.Equal(t, uint16(950), rtd.PF)
	assert.Equal(t, uint16(2200), rtd.VepsR)
	assert.Equal(t, uint16(2210), rtd.VepsS)
	assert.Equal(t, uint16(2220), rtd.VepsT)
	assert.Equal(t, uint16(6000), rtd.Feps)
	assert.Equal(t, uint16(2500), rtd.Peps)
	assert.Equal(t, uint16(2800), rtd.Seps)
	assert.Equal(t, uint16(1500), rtd.Ptogrid)
	assert.Equal(t, uint16(1000), rtd.Ptouser)
	assert.Equal(t, uint16(105), rtd.Epv1Day)
	assert.Equal(t, uint16(82), rtd.Epv2Day)
	assert.Equal(t, uint16(51), rtd.Epv3Day)
	assert.Equal(t, uint16(203), rtd.EinvDay)
	assert.Equal(t, uint16(78), rtd.ErecDay)
	assert.Equal(t, uint16(121), rtd.EchgDay)
	assert.Equal(t, uint16(95), rtd.EdischgDay)
	assert.Equal(t, uint16(156), rtd.EepsDay)
	assert.Equal(t, uint16(67), rtd.EtogridDay)
	assert.Equal(t, uint16(43), rtd.EtouserDay)
	assert.Equal(t, uint16(4000), rtd.Vbus1)
	assert.Equal(t, uint16(3900), rtd.Vbus2)
	assert.Equal(t, uint32(10005), rtd.Epv1All)
	assert.Equal(t, uint32(8002), rtd.Epv2All)
	assert.Equal(t, uint32(5001), rtd.Epv3All)
	assert.Equal(t, uint32(20003), rtd.EinvAll)
	assert.Equal(t, uint32(7008), rtd.ErecAll)
	assert.Equal(t, uint32(12001), rtd.EchgAll)
	assert.Equal(t, uint32(9005), rtd.EdischgAll)
	assert.Equal(t, uint32(15006), rtd.EepsAll)
	assert.Equal(t, uint32(6007), rtd.EtogridAll)
	assert.Equal(t, uint32(4003), rtd.EtouserAll)
	assert.Equal(t, uint32(1), rtd.FaultCode)
	assert.Equal(t, uint32(2), rtd.WarningCode)
	assert.Equal(t, uint16(50), rtd.Tinner)
	assert.Equal(t, uint16(45), rtd.Tradiator1)
	assert.Equal(t, uint16(40), rtd.Tradiator2)
	assert.Equal(t, uint16(25), rtd.Tbat)

	// Assertions for data2 fields
	assert.Equal(t, uint32(3600), rtd.RunningTime)
	assert.Equal(t, uint16(0), rtd.AutoTest)
	assert.Equal(t, uint16(100), rtd.WAutoTestLimit)
	assert.Equal(t, uint16(1000), rtd.UwAutoTestDefaultTime)
	assert.Equal(t, uint16(50), rtd.UwAutoTestTripValue)
	assert.Equal(t, uint16(500), rtd.UwAutoTestTripTime)
	assert.Equal(t, uint16(0), rtd.ACInputType)
	assert.Equal(t, uint16(5000), rtd.MaxChgCurr)
	assert.Equal(t, uint16(6000), rtd.MaxDischgCurr)
	assert.Equal(t, uint16(500), rtd.ChargeVoltRef)
	assert.Equal(t, uint16(450), rtd.DischgCutVolt)
	assert.Equal(t, uint16(0), rtd.BatStatus0BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus1BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus2BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus3BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus4BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus5BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus6BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus7BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus8BMS)
	assert.Equal(t, uint16(0), rtd.BatStatus9BMS)
	assert.Equal(t, uint16(0), rtd.BatStatusINV)
	assert.Equal(t, uint16(1), rtd.BatParallelNum)
	assert.Equal(t, uint16(100), rtd.BatCapacity)
	assert.Equal(t, int16(1000), rtd.BatCurrentBMS)
	assert.Equal(t, uint16(0), rtd.FaultCodeBMS)
	assert.Equal(t, uint16(0), rtd.WarningCodeBMS)
	assert.Equal(t, uint16(3600), rtd.MaxCellVoltBMS)
	assert.Equal(t, uint16(3200), rtd.MinCellVoltBMS)
	assert.Equal(t, int16(300), rtd.MaxCellTempBMS)
	assert.Equal(t, int16(200), rtd.MinCellTempBMS)
	assert.Equal(t, uint16(2), rtd.BMSFWUpdateState)
	assert.Equal(t, uint16(500), rtd.CycleCntBMS)
	assert.Equal(t, uint16(490), rtd.BatVoltSampleINV)
	assert.Equal(t, uint16(280), rtd.T1)
	assert.Equal(t, uint16(270), rtd.T2)
	assert.Equal(t, uint16(260), rtd.T3)
	assert.Equal(t, uint16(250), rtd.T4)
	assert.Equal(t, uint16(240), rtd.T5)
	assert.Equal(t, uint16(1), rtd.ParallelInfo)
	assert.Equal(t, uint16(2000), rtd.VBusP)

	// Assertions for data3 fields
	assert.Equal(t, uint16(2400), rtd.GenVolt)
	assert.Equal(t, uint16(6000), rtd.GenFreq)
	assert.Equal(t, uint16(5000), rtd.GenPower)
	assert.Equal(t, uint16(100), rtd.EgenDay)
	assert.Equal(t, uint32(10000), rtd.EgenAll)
	assert.Equal(t, uint16(2200), rtd.EPSVoltL1N)
	assert.Equal(t, uint16(2210), rtd.EPSVoltL2N)
	assert.Equal(t, uint16(1000), rtd.PepsL1N)
	assert.Equal(t, uint16(800), rtd.PepsL2N)
	assert.Equal(t, uint16(1200), rtd.SepsL1N)
	assert.Equal(t, uint16(1000), rtd.SepsL2N)
	assert.Equal(t, uint16(50), rtd.EepsL1NDay)
	assert.Equal(t, uint16(40), rtd.EepsL2NDay)
	assert.Equal(t, uint32(5000), rtd.EepsL1NAll)
	assert.Equal(t, uint32(4000), rtd.EepsL2NAll)
	assert.Equal(t, uint16(100), rtd.AFCICurrCH1)
	assert.Equal(t, uint16(80), rtd.AFCICurrCH2)
	assert.Equal(t, uint16(60), rtd.AFCICurrCH3)
	assert.Equal(t, uint16(40), rtd.AFCICurrCH4)
	assert.Equal(t, uint16(0), rtd.AFCIFlag)
	assert.Equal(t, uint16(10), rtd.AFCIArcCH1)
	assert.Equal(t, uint16(8), rtd.AFCIArcCH2)
	assert.Equal(t, uint16(6), rtd.AFCIArcCH3)
	assert.Equal(t, uint16(4), rtd.AFCIArcCH4)
	assert.Equal(t, uint16(20), rtd.AFCIMaxArcCH1)
	assert.Equal(t, uint16(15), rtd.AFCIMaxArcCH2)
	assert.Equal(t, uint16(10), rtd.AFCIMaxArcCH3)
	assert.Equal(t, uint16(5), rtd.AFCIMaxArcCH4)

	mockReader.AssertExpectations(t)
}
