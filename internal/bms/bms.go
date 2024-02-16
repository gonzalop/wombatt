package bms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"wombatt/internal/modbus"

	"golang.org/x/exp/slices"
)

const (
	EG4LLv2BMS    = "EG4LLv2"
	Lifepower4BMS = "lifepower4"
	PaceBMS       = "pacemodbus"
)

type BMS interface {
	ReadInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	ReadExtraInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	InfoInstance() any
	DefaultProtocol(deviceType string) string
}

func Instance(bmsType string) BMS {
	switch bmsType {
	case EG4LLv2BMS:
		return NewEG4LLv2()
	case Lifepower4BMS:
		return NewLFP4()
	case PaceBMS:
		return NewPace()
	default:
		log.Fatalf("Unsupported BMS type: %v", bmsType)
	}
	return nil
}

type VoltageStats struct {
	MaxVoltage    uint16 `name:"max_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MinVoltage    uint16 `name:"min_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MeanVoltage   uint16 `name:"mean_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
	MedianVoltage uint16 `name:"median_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001"`
}

func updateVoltageStats(cellVoltage [16]uint16, vs *VoltageStats) {
	var sum uint
	voltages := make([]uint16, 16)
	for i := 0; i < 16; i++ {
		mv := cellVoltage[i]
		voltages[i] = mv
		if i == 0 {
			vs.MinVoltage = mv
			vs.MaxVoltage = mv
		}
		sum += uint(mv)
		if vs.MinVoltage > mv {
			vs.MinVoltage = mv
		}
		if vs.MaxVoltage < mv {
			vs.MaxVoltage = mv
		}
	}
	vs.MeanVoltage = uint16(sum / 16)
	slices.Sort(voltages)
	vs.MedianVoltage = (voltages[7] + voltages[8]) / 2
}

func readIntoStruct(result any, reader modbus.RegisterReader, timeout time.Duration, id uint8, address uint16, count uint8) ([]byte, error) {
	data, err := readWithTimeout(reader, timeout, id, address, count)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		return nil, err
	}
	return data, nil
}

func readWithTimeout(reader modbus.RegisterReader, timeout time.Duration, id uint8, start uint16, count uint8) ([]byte, error) {
	var data []byte
	var err error
	result := make(chan struct{}, 1)
	go func() {
		data, err = reader.ReadHoldingRegisters(id, start, count)
		result <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	case <-result:
		return data, err
	}
}
