package modbus

// Package modbus provides Modbus communication interfaces and implementations.
// It supports different Modbus protocols (RTU, TCP, Lifepower4) and provides a factory
// function to create appropriate Modbus readers.

import (
	"context"
	"encoding/binary"
	"fmt"

	"wombatt/internal/common"
)

// RunCommands is a generic runner for Modbus register-reading commands across inverters.
func RunCommands(ctx context.Context, port common.Port, protocol string, id uint8, commands []string, runCmd func(reader RegisterReader, id uint8, cmd string) (any, error)) ([]any, []error) {
	reader, err := Reader(port, protocol, "")
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
		// Don't start more commands on the shared port once cancelled.
		if err := ctx.Err(); err != nil {
			results = append(results, nil)
			errors = append(errors, err)
			continue
		}
		// The port's read timeout bounds this call; a result that arrives
		// after ctx is done is dropped.
		res, err := runCmd(reader, id, cmd)
		if ctxErr := ctx.Err(); ctxErr != nil {
			res, err = nil, ctxErr
		} else if err != nil {
			// runCmd may return a typed nil pointer (e.g. (*RealtimeData)(nil)),
			// which is non-nil as an `any`. Drop it so callers can test `== nil`.
			res = nil
		}
		results = append(results, res)
		errors = append(errors, err)
	}
	return results, errors
}

// Uint32LowWordFirst decodes a 32-bit value sent as two big-endian registers,
// low word first.
func Uint32LowWordFirst(b []byte) uint32 {
	return uint32(binary.BigEndian.Uint16(b[2:4]))<<16 | uint32(binary.BigEndian.Uint16(b[0:2]))
}

const (
	RTUProtocol        = "ModbusRTU"
	TCPProtocol        = "ModbusTCP"
	Lifepower4Protocol = "lifepower4"
)

// RegisterReader defines the interface for reading Modbus registers.
type RegisterReader interface {
	// ReadHoldingRegisters reads a block of holding registers from a Modbus device.
	// It takes the device ID, starting address, and number of registers to read.
	ReadHoldingRegisters(id uint8, start uint16, count uint8) ([]byte, error)
	// ReadInputRegisters reads a block of input registers from a Modbus device.
	// It takes the device ID, starting address, and number of registers to read.
	ReadInputRegisters(id uint8, start uint16, count uint8) ([]byte, error)
}

// Reader creates and returns a new Modbus RegisterReader based on the specified protocol and BMS type.
// It attempts to auto-detect the protocol if "auto" is provided.
func Reader(port common.Port, protocol, bmsType string) (RegisterReader, error) {
	switch protocol {
	case "auto":
		if bmsType == "lifepower4" {
			return NewLFP4(port), nil
		}
		switch port.Type() {
		case common.SerialDevice, common.HidRawDevice:
			return NewRTU(port), nil
		case common.TCPDevice:
			return NewTCP(port), nil
		default:
			return nil, fmt.Errorf("unable to guess a protocol for %v/%v - %v", protocol, bmsType, port.Type())
		}
	case RTUProtocol:
		return NewRTU(port), nil
	case TCPProtocol:
		return NewTCP(port), nil
	case Lifepower4Protocol:
		return NewLFP4(port), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %v", protocol)
	}
}
