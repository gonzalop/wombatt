package batteries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"wombatt/internal/modbus"
)

type Battery interface {
	ReadInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	ReadExtraInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	InfoInstance() any
	DefaultProtocol(deviceType string) string
}

func Instance(batteryType string) Battery {
	switch batteryType {
	case "EG4LLv2":
		return NewEG4LLv2()
	case "lifepower4":
		return NewLFP4()
	default:
		log.Fatalf("Unsupported battery type: %v", batteryType)
	}
	return nil
}

func readIntoStruct(result any, reader modbus.RegisterReader, timeout time.Duration, id uint8, address uint16, count uint8) error {
	data, err := readWithTimeout(reader, timeout, id, address, count)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		return err
	}
	return nil
}

func readWithTimeout(reader modbus.RegisterReader, timeout time.Duration, id uint8, start uint16, count uint8) ([]byte, error) {
	var data []byte
	var err error
	result := make(chan struct{}, 1)
	go func() {
		data, err = reader.ReadRegisters(id, start, count)
		result <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	case <-result:
		return data, err
	}
}
