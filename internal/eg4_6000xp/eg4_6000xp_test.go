package eg4_6000xp

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	data map[uint16][]byte
}

func newMockReader() *mockReader {
	return &mockReader{
		data: make(map[uint16][]byte),
	}
}

func (m *mockReader) setRegisterData(addr uint16, value []byte) {
	m.data[addr] = value
}

func (m *mockReader) ReadInputRegisters(id uint8, addr uint16, count uint8) ([]byte, error) {
	result := make([]byte, count*2)
	for i := uint16(0); i < uint16(count); i++ {
		if val, ok := m.data[addr+i]; ok {
			copy(result[i*2:(i*2)+uint16(len(val))], val)
		} else {
			return nil, fmt.Errorf("no data for register %d", addr+i)
		}
	}
	return result, nil
}

func (m *mockReader) ReadHoldingRegisters(id uint8, addr uint16, count uint8) ([]byte, error) {
	return nil, fmt.Errorf("ReadHoldingRegisters not implemented for mockReader")
}

// Helper function to convert uint16 to little-endian byte slice
func u16ToBytes(val uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, val)
	return buf
}

// Helper function to convert uint32 to little-endian byte slice (two uint16 registers)
func u32ToBytes(val uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, val)
	return buf
}

// Helper function to convert int16 to little-endian byte slice
func i16ToBytes(val int16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(val))
	return buf
}

func TestReadRealtimeData(t *testing.T) {
	reader := newMockReader()

	expectedData := RealtimeData{
		State:                 1,
		Vpv1:                  2,
		Vpv2:                  3,
		Vpv3:                  4,
		Vbat:                  5,
		SOC:                   6,
		SOH:                   7,
		Ppv1:                  8,
		Ppv2:                  9,
		Ppv3:                  10,
		Pcharge:               11,
		Pdischarge:            12,
		VacR:                  13,
		VacS:                  14,
		VacT:                  15,
		Fac:                   16,
		Pinv:                  17,
		Prec:                  18,
		LinvRMS:               19,
		PF:                    20,
		VepsR:                 21,
		VepsS:                 22,
		VepsT:                 23,
		Feps:                  24,
		Peps:                  25,
		Seps:                  26,
		Ptogrid:               27,
		Ptouser:               28,
		Epv1Day:               29,
		Epv2Day:               30,
		Epv3Day:               31,
		EinvDay:               32,
		ErecDay:               33,
		EchgDay:               34,
		EdischgDay:            35,
		EepsDay:               36,
		EtogridDay:            37,
		EtouserDay:            38,
		Vbus1:                 39,
		Vbus2:                 40,
		Epv1All:               uint32(1000000000),
		Epv2All:               uint32(1100000000),
		Epv3All:               uint32(1200000000),
		EinvAll:               uint32(1300000000),
		ErecAll:               uint32(1400000000),
		EchgAll:               uint32(1500000000),
		EdischgAll:            uint32(1600000000),
		EepsAll:               uint32(1700000000),
		EtogridAll:            uint32(1800000000),
		EtouserAll:            uint32(1900000000),
		FaultCode:             51,
		WarningCode:           52,
		Tinner:                53,
		Tradiator1:            54,
		Tradiator2:            55,
		Tbat:                  56,
		RunningTime:           uint32(2000000000),
		AutoTestInfo:          58,
		WAutoTestLimit:        59,
		UwAutoTestDefaultTime: 60,
		UwAutoTestTripValue:   61,
		UwAutoTestTripTime:    62,
		ACInputType:           63,
		BatTypeAndBrand:       64,
		MaxChgCurr:            65,
		MaxDischgCurr:         66,
		ChargeVoltRef:         67,
		DischgCutVolt:         68,
		BatStatus0BMS:         69,
		BatStatus1BMS:         70,
		BatStatus2BMS:         71,
		BatStatus3BMS:         72,
		BatStatus4BMS:         73,
		BatStatus5BMS:         74,
		BatStatus6BMS:         75,
		BatStatus7BMS:         76,
		BatStatus8BMS:         77,
		BatStatus9BMS:         78,
		BatStatusINV:          79,
		BatParallelNum:        80,
		BatCapacity:           81,
		BatCurrentBMS:         -82,
		FaultCodeBMS:          83,
		WarningCodeBMS:        84,
		MaxCellVoltBMS:        85,
		MinCellVoltBMS:        86,
		MaxCellTempBMS:        -87,
		MinCellTempBMS:        -88,
		BMSFWUpdateState:      89,
		CycleCntBMS:           90,
		BatVoltSampleINV:      91,
		ParallelInfo:          92,
		OnGridLoadPower:       93,
		VBusP:                 94,
		GenVolt:               95,
		GenFreq:               96,
		GenPower:              97,
		EgenDay:               98,
		EgenAll:               uint32(1000000000),
		EPSVoltL1N:            100,
		EPSVoltL2N:            101,
		PepsL1N:               102,
		PepsL2N:               103,
		SepsL1N:               104,
		SepsL2N:               105,
		EepsL1NDay:            106,
		EepsL2NDay:            107,
		EepsL1NAll:            uint32(1100000000),
		EepsL2NAll:            uint32(1200000000),
		AFCICurrCH1:           110,
		AFCICurrCH2:           111,
		AFCICurrCH3:           112,
		AFCICurrCH4:           113,
		AFCIFlag:              114,
		AFCIArcCH1:            115,
		AFCIArcCH2:            116,
		AFCIArcCH3:            117,
		AFCIArcCH4:            118,
		AFCIMaxArcCH1:         119,
		AFCIMaxArcCH2:         120,
		AFCIMaxArcCH3:         121,
		AFCIMaxArcCH4:         122,
		ACCouplePower:         123,
		Pload:                 124,
		EloadDay:              125,
		EloadAll:              uint32(1300000000),
		SwitchState:           127,
		PinvS:                 128,
		PinvT:                 129,
		PrecS:                 130,
		PrecT:                 131,
		PtogridS:              132,
		PtogridT:              133,
		PtouserS:              134,
		PtouserT:              135,
		GenPowerS:             136,
		GenPowerT:             137,
		LinvRMSS:              138,
		LinvRMST:              139,
		PFS:                   140,
		PFT:                   141,
	}

	// Populate mock reader with data from expectedData
	v := reflect.ValueOf(expectedData)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typeOfS.Field(i)
		modbusTag := field.Tag.Get("modbus")
		if modbusTag == "" {
			continue
		}

		addr := uint16(0)
		_, err := fmt.Sscanf(modbusTag, "%d", &addr)
		assert.NoError(t, err)

		fieldValue := v.Field(i)

		switch fieldValue.Kind() {
		case reflect.Uint16:
			reader.setRegisterData(addr, u16ToBytes(uint16(fieldValue.Uint())))
		case reflect.Uint32:
			// For uint32, we need to set two registers (low word and high word)
			bytes := u32ToBytes(uint32(fieldValue.Uint()))
			reader.setRegisterData(addr, bytes[0:2])   // Low word
			reader.setRegisterData(addr+1, bytes[2:4]) // High word
		case reflect.Int16:
			reader.setRegisterData(addr, i16ToBytes(int16(fieldValue.Int())))
		}
	}

	actualData, err := ReadRealtimeData(reader, 1)
	assert.NoError(t, err)
	assert.NotNil(t, actualData)

	// Assert that all fields match
	assert.Equal(t, expectedData, *actualData)
}

func TestReadRealtimeDataError(t *testing.T) {
	reader := newMockReader() // Reset mock reader
	// Do not set any data, so ReadInputRegisters will return an error for the first block
	_, err := ReadRealtimeData(reader, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data for register 0")
}

func TestReadRealtimeDataPartialError(t *testing.T) {
	reader := newMockReader()

	// Populate some data, but leave a gap for register 45
	for i := uint16(40); i < 45; i++ {
		reader.setRegisterData(i, u16ToBytes(i))
	}
	for i := uint16(46); i < 80; i++ {
		reader.setRegisterData(i, u16ToBytes(i))
	}

	// Also populate data for other blocks to ensure the error comes from the intended block
	for i := range uint16(40) {
		reader.setRegisterData(i, u16ToBytes(i))
	}
	for i := uint16(80); i < 120; i++ {
		reader.setRegisterData(i, u16ToBytes(i))
	}
	for i := uint16(120); i < 160; i++ {
		reader.setRegisterData(i, u16ToBytes(i))
	}
	for i := uint16(160); i < 200; i++ {
		reader.setRegisterData(i, u16ToBytes(i))
	}

	_, err := ReadRealtimeData(reader, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read input registers 40-79: no data for register 45")
}
