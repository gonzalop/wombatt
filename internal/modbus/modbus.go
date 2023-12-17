package modbus

import (
	"fmt"

	"wombatt/internal/common"
)

const (
	RTUProtocol        = "ModbusRTU"
	TCPProtocol        = "ModbusTCP"
	Lifepower4Protocol = "lifepower4"
)

type RegisterReader interface {
	ReadRegisters(id uint8, start uint16, count uint8) ([]byte, error)
}

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
