package modbus

import (
	"fmt"

	"wombatt/internal/common"
)

const (
	RTUProtocol = "ModbusRTU"
	TCPProtocol = "ModbusTCP"
	Lifepower4  = "lifepower4"
)

type RegisterReader interface {
	ReadRegisters(id uint8, start uint16, count uint8) (*RTUFrame, error)
}

func ReaderFromProtocol(port common.Port, protocol string) (RegisterReader, error) {
	switch protocol {
	case "auto":
		switch port.Type() {
		case common.SerialDevice, common.HidRawDevice:
			return NewRTU(port), nil
		case common.TCPDevice:
			return NewTCP(port), nil
		default:
			return nil, fmt.Errorf("unable to guess a protocol")
		}
	case RTUProtocol:
		return NewRTU(port), nil
	case TCPProtocol:
		return NewTCP(port), nil
	case Lifepower4:
		return NewLFP4(port), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %v", protocol)
	}
}
