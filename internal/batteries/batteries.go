package batteries

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"wombatt/internal/modbus"
)

type Battery interface {
	ReadInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	ReadExtraInfo(modbus.RegisterReader, uint8, time.Duration) (any, error)
	InfoInstance() any
	DefaultProtocol() string
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

func readIntoStruct2(result any, reader modbus.RegisterReader, timeout time.Duration, id uint8, address uint16, count uint8) error {
	frame, err := readWithTimeout(reader, timeout, id, address, count)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(frame.RawData())
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		return err
	}
	return nil
}

func readIntoStruct(result any, reader modbus.RegisterReader, timeout time.Duration, id uint8, address uint16, count uint8) error {
	frame, err := readWithTimeout(reader, timeout, id, address, count)
	if err != nil {
		return err
	}

	data := frame.Data()
	if len(data) != (int(count) * 2) {
		log.Printf("%s\n", hex.EncodeToString(data))
		return fmt.Errorf("unexpected data length: got %d, want 78", len(data))
	}
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, result); err != nil {
		return err
	}
	return nil
}

func readWithTimeout(reader modbus.RegisterReader, timeout time.Duration, id uint8, start uint16, count uint8) (*modbus.RTUFrame, error) {
	var frame *modbus.RTUFrame
	var err error
	result := make(chan struct{}, 1)
	go func() {
		frame, err = reader.ReadRegisters(id, start, count)
		result <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out")
	case <-result:
		return frame, err
	}
}
