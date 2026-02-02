package bms

// Package bms provides interfaces and implementations for interacting with various Battery Management Systems (BMS).
// It defines common operations like reading battery information and provides a factory for creating BMS instances.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"time"

	"wombatt/internal/modbus"
)

const (
	EG4LLv2BMS     = "EG4LLv2"
	Lifepower4BMS  = "lifepower4"
	Lifepowerv2BMS = "lifepowerv2" // Protocol switches: 1-off, 2 through 6-on
	PaceBMS        = "pacemodbus"

	NumCells = 16 // Standard number of cells in a 48V battery pack for voltage stats
)

// BMS defines the interface for interacting with different Battery Management Systems.
type BMS interface {
	// ReadInfo reads primary battery information from the BMS.
	// It takes a modbus.RegisterReader, battery ID, and a timeout.
	// It returns a struct containing the battery information or an error.
	ReadInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	// ReadExtraInfo reads additional, less frequently accessed battery information from the BMS.
	// It takes a modbus.RegisterReader, battery ID, and a timeout.
	// It returns a struct containing the extra battery information or an error.
	ReadExtraInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	// InfoInstance returns an empty instance of the primary battery info struct
	// that this BMS implementation uses. This is useful for reflection-based operations.
	InfoInstance() any
	// DefaultProtocol returns the default communication protocol (e.g., RTU, TCP)
	// for a given device type (e.g., "serial", "network") for this BMS.
	DefaultProtocol(deviceType string) string
}

// Instance creates and returns a new BMS instance based on the provided BMS type string.
// It returns an error if the BMS type is unsupported.
func Instance(bmsType string) (BMS, error) {
	switch bmsType {
	case EG4LLv2BMS:
		return NewEG4LLv2(), nil
	case Lifepower4BMS:
		return NewLFP4(), nil
	case Lifepowerv2BMS:
		return NewEG4LLv2(), nil // Same protocol as EG4LLv2 BMS.
	case PaceBMS:
		return NewPace(), nil
	default:
		return nil, fmt.Errorf("unsupported BMS type: %v", bmsType)
	}
}

// VoltageStats holds statistics about cell voltages within a battery pack.
type VoltageStats struct {
	MaxVoltage    uint16 `name:"max_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
	MinVoltage    uint16 `name:"min_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
	MeanVoltage   uint16 `name:"mean_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
	MedianVoltage uint16 `name:"median_cell_voltage" dclass:"voltage" unit:"V" multiplier:"0.001" precision:"3"`
}

// updateVoltageStats calculates and updates voltage statistics (min, max, mean, median)
// for a given array of cell voltages.
func updateVoltageStats(cellVoltage [16]uint16, vs *VoltageStats) {
	voltages := make([]uint16, NumCells)

	// Initialize min/max with the first cell's voltage
	vs.MinVoltage = cellVoltage[0]
	vs.MaxVoltage = cellVoltage[0]
	voltages[0] = cellVoltage[0]
	sum := uint(cellVoltage[0])

	for i := 1; i < NumCells; i++ { // Start from the second cell
		mv := cellVoltage[i]
		voltages[i] = mv
		sum += uint(mv)
		if vs.MinVoltage > mv {
			vs.MinVoltage = mv
		}
		if vs.MaxVoltage < mv {
			vs.MaxVoltage = mv
		}
	}
	vs.MeanVoltage = uint16(sum / NumCells)
	slices.Sort(voltages)
	vs.MedianVoltage = (voltages[NumCells/2-1] + voltages[NumCells/2]) / 2
}

// readIntoStruct reads data from the Modbus device into the provided struct.
// The `quantityOrCommand` parameter serves a dual purpose:
// - For standard Modbus protocols (RTU, TCP), it represents the number of registers to read.
// - For the Lifepower4 protocol, it represents a command code.
func readIntoStruct(result any, reader modbus.RegisterReader, timeout time.Duration, id uint8, address uint16, quantityOrCommand uint8) ([]byte, error) {
	data, err := readWithTimeout(reader, timeout, id, address, quantityOrCommand)
	if err != nil {
		return nil, err
	}
	slog.Debug("reading into struct", "data", hex.EncodeToString(data), "struct-type", reflect.TypeOf(result).String())
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		return nil, err
	}
	return data, nil
}

func readWithTimeout(reader modbus.RegisterReader, timeout time.Duration, id uint8, start uint16, quantityOrCommand uint8) ([]byte, error) {
	var data []byte
	var err error
	result := make(chan struct{}, 1)
	go func() {
		data, err = reader.ReadHoldingRegisters(id, start, quantityOrCommand)
		result <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	case <-result:
		return data, err
	}
}
