package bms_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bms_pkg "wombatt/internal/bms"
	"wombatt/internal/modbus"
)

type mockModbusClient struct {
	mock.Mock
	modbus.RegisterReader // This line ensures the import is used
}

func (m *mockModbusClient) ReadRegisters(address, quantity uint16) ([]byte, error) {
	args := m.Called(address, quantity)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockModbusClient) ReadHoldingRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	args := m.Called(id, start, count)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockModbusClient) ReadInputRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	args := m.Called(id, start, count)
	return args.Get(0).([]byte), args.Error(1)
}

func TestPaceBMS_ReadInfo(t *testing.T) {
	client := new(mockModbusClient)
	bms := bms_pkg.NewPace()

	expectedInfo := bms_pkg.PaceBatteryInfo{
		PaceModbusBatteryInfo: bms_pkg.PaceModbusBatteryInfo{
			Current:           1000,
			Voltage:           5120,
			SOC:               80,
			SOH:               95,
			RemainingCapacity: 5000,
			FullCapacity:      6000,
			DesignCapacity:    6000,
			CycleCounts:       100,
			WarningFlag:       0x01,
			ProtectionFlag:    0x02,
			StatusFlag:        0x04,
			BalanceStatus:     0x08,
			CellVoltages:      [16]uint16{3200, 3201, 3202, 3203, 3204, 3205, 3206, 3207, 3208, 3209, 3210, 3211, 3212, 3213, 3214, 3215},
			CellTemps:         [4]int16{250, 260, 270, 280},
			MOSFETTemp:        300,
			EnvTemp:           200,
		},
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, &expectedInfo.PaceModbusBatteryInfo)
	assert.NoError(t, err)
	response := buf.Bytes()

	client.On("ReadHoldingRegisters", uint8(0x01), uint16(0x0000), uint8(len(response)/2)).Return(response, nil)

	info, err := bms.ReadInfo(t.Context(), client, 0x01, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, info)

	pInfo, ok := info.(*bms_pkg.PaceBatteryInfo)
	assert.True(t, ok)
	assert.NotNil(t, pInfo)

	// Assertions for the fields
	assert.Equal(t, int16(1000), pInfo.Current)
	assert.Equal(t, uint16(5120), pInfo.Voltage)
	assert.Equal(t, uint16(80), pInfo.SOC)
	assert.Equal(t, uint16(95), pInfo.SOH)
	assert.Equal(t, uint16(5000), pInfo.RemainingCapacity)
	assert.Equal(t, uint16(6000), pInfo.FullCapacity)
	assert.Equal(t, uint16(6000), pInfo.DesignCapacity)
	assert.Equal(t, uint16(100), pInfo.CycleCounts)
	assert.Equal(t, uint16(0x01), pInfo.WarningFlag)
	assert.Equal(t, uint16(0x02), pInfo.ProtectionFlag)
	assert.Equal(t, uint16(0x04), pInfo.StatusFlag)
	assert.Equal(t, uint16(0x08), pInfo.BalanceStatus)

	for i := range 16 {
		assert.Equal(t, uint16(3200+i), pInfo.CellVoltages[i])
	}

	assert.Equal(t, int16(250), pInfo.CellTemps[0])
	assert.Equal(t, int16(260), pInfo.CellTemps[1])
	assert.Equal(t, int16(270), pInfo.CellTemps[2])
	assert.Equal(t, int16(280), pInfo.CellTemps[3])

	assert.Equal(t, int16(300), pInfo.MOSFETTemp)
	assert.Equal(t, int16(200), pInfo.EnvTemp)

	client.AssertExpectations(t)
}

func TestPaceBMS_ReadExtraInfo(t *testing.T) {
	client := new(mockModbusClient)
	bms := bms_pkg.NewPace()

	expectedExtraInfo := bms_pkg.PaceModbusExtraBatteryInfo{
		Version: [20]byte{'V', '1', '.', '0', '0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
		ModelSN: [20]byte{'M', 'O', 'D', 'E', 'L', '1', '2', '3', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
		PackSN:  [20]byte{'P', 'A', 'C', 'K', 'S', 'N', '4', '5', '6', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, &expectedExtraInfo)
	assert.NoError(t, err)
	response := buf.Bytes()

	client.On("ReadHoldingRegisters", uint8(0x01), uint16(150), uint8(30)).Return(response, nil)

	extraInfo, err := bms.ReadExtraInfo(t.Context(), client, 0x01, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, extraInfo)

	pExtraInfo, ok := extraInfo.(*bms_pkg.PaceModbusExtraBatteryInfo)
	assert.True(t, ok)
	assert.NotNil(t, pExtraInfo)

	assert.Equal(t, expectedExtraInfo.Version, pExtraInfo.Version)
	assert.Equal(t, expectedExtraInfo.ModelSN, pExtraInfo.ModelSN)
	assert.Equal(t, expectedExtraInfo.PackSN, pExtraInfo.PackSN)

	client.AssertExpectations(t)
}

func TestPaceBMS_DefaultProtocol(t *testing.T) {
	bms := bms_pkg.NewPace()
	assert.Equal(t, modbus.TCPProtocol, bms.DefaultProtocol("tcp"))
	assert.Equal(t, modbus.RTUProtocol, bms.DefaultProtocol("serial"))
	assert.Equal(t, modbus.RTUProtocol, bms.DefaultProtocol("other"))
}

func TestPaceBMS_ReadInfo_Errors(t *testing.T) {
	bms := bms_pkg.NewPace()

	t.Run("read failure", func(t *testing.T) {
		client := new(mockModbusClient)
		client.On("ReadHoldingRegisters", uint8(0x01), uint16(0x0000), uint8(37)).Return([]byte(nil), errors.New("read error"))
		info, err := bms.ReadInfo(t.Context(), client, 0x01, 1*time.Second)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "read error")
		client.AssertExpectations(t)
	})

	t.Run("unexpected data length", func(t *testing.T) {
		client := new(mockModbusClient)
		longResponse := make([]byte, 80) // want 74
		client.On("ReadHoldingRegisters", uint8(0x01), uint16(0x0000), uint8(37)).Return(longResponse, nil)
		info, err := bms.ReadInfo(t.Context(), client, 0x01, 1*time.Second)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "unexpected data length: got 80, want 74")
		client.AssertExpectations(t)
	})
}

func TestPaceBMS_ReadExtraInfo_Errors(t *testing.T) {
	bms := bms_pkg.NewPace()

	t.Run("read failure", func(t *testing.T) {
		client := new(mockModbusClient)
		client.On("ReadHoldingRegisters", uint8(0x01), uint16(150), uint8(30)).Return([]byte(nil), errors.New("read error"))
		info, err := bms.ReadExtraInfo(t.Context(), client, 0x01, 1*time.Second)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "read error")
		client.AssertExpectations(t)
	})

	t.Run("unexpected data length", func(t *testing.T) {
		client := new(mockModbusClient)
		longResponse := make([]byte, 70) // want 60
		client.On("ReadHoldingRegisters", uint8(0x01), uint16(150), uint8(30)).Return(longResponse, nil)
		info, err := bms.ReadExtraInfo(t.Context(), client, 0x01, 1*time.Second)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "unexpected data length: got 70, want 60")
		client.AssertExpectations(t)
	})
}
