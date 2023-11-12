package modbus

import (
	"fmt"

	"wombatt/internal/common"
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
	case "RTU":
		return NewRTU(port), nil
	case "TCP":
		return NewTCP(port), nil
	case "lifepower4":
		return NewLFP4(port), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %v", protocol)
	}
}
