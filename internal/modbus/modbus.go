package modbus

// Package modbus provides Modbus communication interfaces and implementations.
// It supports different Modbus protocols (RTU, TCP, Lifepower4) and provides a factory
// function to create appropriate Modbus readers.

import (
	"fmt"

	"wombatt/internal/common"
)

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
