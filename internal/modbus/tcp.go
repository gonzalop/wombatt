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

type TCPRTUFrame struct {
	TCPRTUHeader
	RTUFrame
}

var tid atomic.Uint32

type TCP struct {
	port common.Port
}

func NewTCP(port common.Port) RegisterReader {
	return &TCP{port: port}
}

// ReadRegisters requests 'count' holding registers from unit 'id' from the 'start' memory address.
// and reads the response back.
func (t *TCP) ReadRegisters(id uint8, start uint16, count uint8) (*RTUFrame, error) {
	f := buildReadRequestRTUFrame(id, ReadHoldingRegisters, start, uint16(count))
	tf := &TCPRTUFrame{RTUFrame: *f, TCPRTUHeader: TCPRTUHeader{PID: 0}}
	tf.TID = uint16(tid.Add(1) & 0x0ffff)
	tf.Length = uint16(len(f.RawData())) - 2 // -2 for CRC
	tf.UnitID = id

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, tf.TCPRTUHeader); err != nil {
		return nil, err
	}

	raw := tf.RTUFrame.RawData()
	buf.Write(raw[1 : len(raw)-2]) // Exclude slave ID and CRC
	if _, err := t.port.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	return t.ReadTCPResponse(tf.TID, id)
}

func (t *TCP) ReadTCPResponse(tid uint16, unitID uint8) (*RTUFrame, error) {
	mbap := make([]byte, 7)
	if _, err := io.ReadFull(t.port, mbap[0:6]); err != nil {
		return nil, err
	}
	var header TCPRTUHeader
	if err := binary.Read(bytes.NewReader(mbap), binary.BigEndian, &header); err != nil {
		return nil, err
	}
	rtu := make([]byte, header.Length+2) // Add 2 more bytes because RTUFrame expects a CRC there.
	n, err := io.ReadFull(t.port, rtu[0:len(rtu)-2])
	if err != nil {
		return nil, err
	}
	if n != int((header.Length)) {
		return nil, fmt.Errorf("short response. got %d, want %d", n, header.Length)
	}
	return NewRTUFrame(rtu), nil
}
