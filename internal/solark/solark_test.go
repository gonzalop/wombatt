package solark

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

	// Mock data for RealtimeData
	// Block 1: Addr 60-64 (Day Active Power Wh, Total Active Power Wh)
	// DayActivePowerWh: 100 (0x0064)
	// TotalActivePowerWhLow: 1000 (0x03E8)
	// TotalActivePowerWhHigh: 0 (0x0000)
	data1 := make([]byte, 10)
	binary.BigEndian.PutUint16(data1[0:2], 100)  // DayActivePowerWh
	binary.BigEndian.PutUint16(data1[6:8], 1000) // TotalActivePowerWhLow
	binary.BigEndian.PutUint16(data1[8:10], 0)   // TotalActivePowerWhHigh

	// Block 2: Addr 79-91 (Grid Frequency, DC/DC Temp, IGBT Temp)
	// GridFrequency: 5000 (50.00 Hz) (0x1388)
	// DCDCTemp: 250 (25.0 C) (0x00FA)
	// IGBTHSCTemp: 300 (30.0 C) (0x012C)
	data2 := make([]byte, 26)
	binary.BigEndian.PutUint16(data2[0:2], 5000)  // GridFrequency
	binary.BigEndian.PutUint16(data2[22:24], 250) // DCDCTemp
	binary.BigEndian.PutUint16(data2[24:26], 300) // IGBTHSCTemp

	// Block 3: Addr 103-112 (Fault Info, Batt Capacity, Daily PV, DC Volt/Current)
	// FaultInfoWord1: 1 (0x0001) - GFDI_Relay_Failure
	// CorrectedBattCapacity: 100 (0x0064)
	// DailyPVPow: 500 (50.0 kWh) (0x01F4)
	// DCVoltage1: 5000 (500.0 V) (0x1388)
	// DCCurrent1: 1000 (100.0 A) (0x03E8)
	// DCVoltage2: 4000 (400.0 V) (0x0FA0)
	// DCCurrent2: 500 (50.0 A) (0x01F4)
	data3 := make([]byte, 20)
	binary.BigEndian.PutUint16(data3[0:2], 1)      // FaultInfoWord1
	binary.BigEndian.PutUint16(data3[8:10], 100)   // CorrectedBattCapacity
	binary.BigEndian.PutUint16(data3[10:12], 500)  // DailyPVPow
	binary.BigEndian.PutUint16(data3[12:14], 5000) // DCVoltage1
	binary.BigEndian.PutUint16(data3[14:16], 1000) // DCCurrent1
	binary.BigEndian.PutUint16(data3[16:18], 4000) // DCVoltage2
	binary.BigEndian.PutUint16(data3[18:20], 500)  // DCCurrent2

	// Block 4: Addr 150-184 (Various Voltages, Currents, Powers, Battery Info)
	// GridSideVoltageL1N: 2400 (240.0 V)
	// BatteryTemperature: 250 (25.0 C)
	// BatteryVoltage: 5000 (50.00 V)
	// BatteryCapacitySOC: 90 (90%)
	data4 := make([]byte, 70)
	binary.BigEndian.PutUint16(data4[0:2], 2400)   // GridSideVoltageL1N
	binary.BigEndian.PutUint16(data4[64:66], 250)  // BatteryTemperature
	binary.BigEndian.PutUint16(data4[66:68], 5000) // BatteryVoltage
	binary.BigEndian.PutUint16(data4[68:70], 90)   // BatteryCapacitySOC

	// Block 5: Addr 186-196 (PV Power, Battery Output, Frequencies, Relay Status)
	// PV1InputPower: 1000 (1000 W)
	// PV2InputPower: 500 (500 W)
	// BatteryOutputPower: -200 (0xFF38) (-200 W)
	// BatteryOutputCurrent: -50 (0xFFCE) (-5.0 A)
	// LoadFrequency: 6000 (60.00 Hz)
	// InverterOutputFrequency: 6000 (60.00 Hz)
	// GridSideRelayStatus: 2 (Closed)
	// GeneratorSideRelayStatus: 3 (Closed when Generator is on)
	// GeneratorRelayFrequency: 6000 (60.00 Hz)
	data5 := make([]byte, 22)
	binary.BigEndian.PutUint16(data5[0:2], 1000)     // PV1InputPower
	binary.BigEndian.PutUint16(data5[2:4], 500)      // PV2InputPower
	binary.BigEndian.PutUint16(data5[8:10], 0xFF38)  // BatteryOutputPower (signed -200 as uint16)
	binary.BigEndian.PutUint16(data5[10:12], 0xFFCE) // BatteryOutputCurrent (signed -50 as uint16)
	binary.BigEndian.PutUint16(data5[12:14], 6000)   // LoadFrequency
	binary.BigEndian.PutUint16(data5[14:16], 6000)   // InverterOutputFrequency
	binary.BigEndian.PutUint16(data5[16:18], 2)      // GridSideRelayStatus
	binary.BigEndian.PutUint16(data5[18:20], 3)      // GeneratorSideRelayStatus
	binary.BigEndian.PutUint16(data5[20:22], 6000)   // GeneratorRelayFrequency

	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(60), uint8(5)).Return(data1, nil).Once()
	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(79), uint8(13)).Return(data2, nil).Once()
	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(103), uint8(10)).Return(data3, nil).Once()
	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(150), uint8(35)).Return(data4, nil).Once()
	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(186), uint8(11)).Return(data5, nil).Once()

	rtd, err := ReadRealtimeData(mockReader, 1)
	assert.NoError(t, err)
	assert.NotNil(t, rtd)

	assert.Equal(t, int16(100), rtd.DayActivePowerWh)
	assert.Equal(t, uint32(1000), rtd.TotalActivePowerWh())
	assert.Equal(t, uint16(5000), rtd.GridFrequency)
	assert.Equal(t, int16(250), rtd.DCDCTemp)
	assert.Equal(t, int16(300), rtd.IGBTHSCTemp)
	assert.Equal(t, uint16(1), rtd.FaultInfoWord1)
	assert.Equal(t, uint16(100), rtd.CorrectedBattCapacity)
	assert.Equal(t, uint16(500), rtd.DailyPVPow)
	assert.Equal(t, uint16(5000), rtd.DCVoltage1)
	assert.Equal(t, uint16(1000), rtd.DCCurrent1)
	assert.Equal(t, uint16(4000), rtd.DCVoltage2)
	assert.Equal(t, uint16(500), rtd.DCCurrent2)
	assert.Equal(t, uint16(2400), rtd.GridSideVoltageL1N)
	assert.Equal(t, int16(250), rtd.BatteryTemperature)
	assert.Equal(t, uint16(5000), rtd.BatteryVoltage)
	assert.Equal(t, uint16(90), rtd.BatteryCapacitySOC)
	assert.Equal(t, uint16(1000), rtd.PV1InputPower)
	assert.Equal(t, uint16(500), rtd.PV2InputPower)
	assert.Equal(t, int16(-200), rtd.BatteryOutputPower)
	assert.Equal(t, int16(-50), rtd.BatteryOutputCurrent)
	assert.Equal(t, uint16(6000), rtd.LoadFrequency)
	assert.Equal(t, uint16(6000), rtd.InverterOutputFrequency)
	assert.Equal(t, uint16(2), rtd.GridSideRelayStatus)
	assert.Equal(t, uint16(3), rtd.GeneratorSideRelayStatus)
	assert.Equal(t, uint16(6000), rtd.GeneratorRelayFrequency)

	mockReader.AssertExpectations(t)
}

func TestReadIntrinsicAttributes(t *testing.T) {
	mockReader := new(MockRegisterReader)

	// Mock data for IntrinsicAttributes (Serial Number: "AH12345678")
	data := make([]byte, 10)
	binary.BigEndian.PutUint16(data[0:2], 0x4148)  // 'A', 'H'
	binary.BigEndian.PutUint16(data[2:4], 0x3132)  // '1', '2'
	binary.BigEndian.PutUint16(data[4:6], 0x3334)  // '3', '4'
	binary.BigEndian.PutUint16(data[6:8], 0x3536)  // '5', '6'
	binary.BigEndian.PutUint16(data[8:10], 0x3738) // '7', '8'

	mockReader.On("ReadHoldingRegisters", uint8(1), uint16(3), uint8(5)).Return(data, nil).Once()

	ia, err := ReadIntrinsicAttributes(mockReader, 1)
	assert.NoError(t, err)
	assert.NotNil(t, ia)

	assert.Equal(t, "AH12345678", ia.SerialNumber())

	mockReader.AssertExpectations(t)
}
