package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"

	"wombatt/internal/common"
)

type TCPRTUHeader struct {
	TID    uint16
	PID    uint16
	Length uint16
	UnitID uint8
}

var tid atomic.Uint32

type TCP struct {
	port common.Port
}

func NewTCP(port common.Port) RegisterReader {
	return &TCP{port: port}
}

// ReadHoldingRegisters requests 'count' holding registers from unit 'id' from the 'start' memory address.
// and reads the response back.
func (t *TCP) ReadHoldingRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	return t.readRegisters(id, ReadHoldingRegisters, start, count)
}

// ReadInputRegisters requests 'count' input registers from unit 'id' from the 'start' memory address.
// and reads the response back.
func (t *TCP) ReadInputRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	return t.readRegisters(id, ReadInputRegisters, start, count)
}

func (t *TCP) readRegisters(id uint8, functionCode RTUFunction, start uint16, count uint8) ([]byte, error) {
	raw := buildReadRequestRTUFrame(id, functionCode, start, uint16(count))
	tf := &TCPRTUHeader{
		TID:    uint16(tid.Add(1) & 0x0ffff),
		Length: uint16(len(raw)) - 2, // -2 for CRC
		UnitID: id,
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, tf); err != nil {
		return nil, err
	}

	buf.Write(raw[1 : len(raw)-2]) // Exclude slave ID and CRC
	if _, err := t.port.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	raw, err := t.ReadTCPResponse(tf.TID, id)
	if err != nil {
		return nil, err
	}
	return NewRTUFrame(raw).Data(), nil
}

func (t *TCP) ReadTCPResponse(tid uint16, unitID uint8) ([]byte, error) {
	mbap := make([]byte, 7)
	// The UnitID is not read at this moment
	if n, err := io.ReadFull(t.port, mbap[0:6]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame: read %d, want at least 6 bytes", n)
		}
		return nil, err
	}
	var header TCPRTUHeader
	if err := binary.Read(bytes.NewReader(mbap), binary.BigEndian, &header); err != nil {
		return nil, err
	}
	if header.TID != tid {
		return nil, fmt.Errorf("unexpected transaction ID: got 0x%04x; want 0x%04x", header.TID, tid)
	}
	rtu := make([]byte, header.Length+2) // Add 2 more bytes because RTUFrame expects a CRC there.
	if n, err := io.ReadFull(t.port, rtu[0:len(rtu)-2]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame: read %d, want at least %d bytes", n, header.Length)
		}
		return nil, err
	}
	return rtu, nil
}
