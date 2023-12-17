// Package modbus provides types and functions to read and write RTU messages.
package modbus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"wombatt/internal/common"
)

// RTUFunction represents RTU function codes.
type RTUFunction uint8

// RTUProtocolError error is used for RTU response protocol errors.
type RTUProtocolError uint8

const (
	// ReadCoils is the RTU function code for reading coils.
	ReadCoils RTUFunction = 0x01
	// ReadDiscreteInputs is the RTU function code for reading discrete inputs.
	ReadDiscreteInputs RTUFunction = 0x02
	// ReadHoldingRegisters is the RTU function code for reading holding registers.
	ReadHoldingRegisters RTUFunction = 0x03
	// ReadInputRegisters is the RTU function code for reading input registers.
	ReadInputRegisters RTUFunction = 0x04
	// WriteSingleCoil is the RTU function code for writing one coil.
	WriteSingleCoil RTUFunction = 0x05
	// WriteSingleRegister is the RTU function code for writing one register.
	WriteSingleRegister RTUFunction = 0x06
	// WriteMultipleCoil is the RTU function code for writing multiple coils.
	WriteMultipleCoil RTUFunction = 0x0f
	// WriteMultipleRegisters is the RTU function code for writing multiple registers.
	WriteMultipleRegisters RTUFunction = 0x10

	// IllegalFunction is returned in an RTU error response when an illegal function is found.
	IllegalFunction RTUProtocolError = 0x01
	// IllegalDataAddress is returned in an RTU error response when an illegal address found.
	IllegalDataAddress RTUProtocolError = 0x02
	// IllegalDataValue is returned in an RTU error response when an illegal value is found.
	IllegalDataValue RTUProtocolError = 0x03
	// ServerDeviceFailure is returned in an RTU error response for a device failure.
	ServerDeviceFailure RTUProtocolError = 0x04
	// Acknowledge is returned in an RTU error response when the server got the request but coul not do anything else.
	Acknowledge RTUProtocolError = 0x05
	// ServerDeviceBusy is returned in an RTU error response when the server is busy.
	ServerDeviceBusy RTUProtocolError = 0x06
	// MemoryParityError is returned in an RTU error response when a memory error is detected.
	MemoryParityError RTUProtocolError = 0x08
	// GWPathUnavailable is returned in an RTU error response when a gateway is not found.
	GWPathUnavailable RTUProtocolError = 0x0a
	// GWTargetFailedToRespond is returned in an RTU error response when a gateway does not respond.
	GWTargetFailedToRespond RTUProtocolError = 0x0b

	// MaxRTUFrameLength is the maximum length of an RTU frame.
	MaxRTUFrameLength = 256
)

var protocolErrorMap = map[RTUProtocolError]string{
	IllegalFunction:         "illegal function",
	IllegalDataAddress:      "illegal data address",
	IllegalDataValue:        "illegal data value",
	ServerDeviceFailure:     "server device failure",
	Acknowledge:             "acknowledge",
	ServerDeviceBusy:        "server device busy",
	MemoryParityError:       "memory parity error",
	GWPathUnavailable:       "gateway path unavailable",
	GWTargetFailedToRespond: "gateway target failed to respond",
}

// RTUFrame contains the data of an RTU message.
type RTUFrame struct {
	rawData []byte
}

// NewRTUFrame creates a new RTUFrame from a byte array.
func NewRTUFrame(rawData []byte) *RTUFrame {
	return &RTUFrame{rawData: rawData}
}

func buildReadRequestRTUFrame(id uint8, function RTUFunction, address uint16, length uint16) []byte {
	var b bytes.Buffer
	b.WriteByte(id)
	b.WriteByte(byte(function))
	b.WriteByte(uint8((address & uint16(0xff00) >> 8)))
	b.WriteByte(uint8(address & uint16(0xff)))
	b.WriteByte(uint8((length & uint16(0xff00) >> 8)))
	b.WriteByte(uint8(length & uint16(0xff)))
	checksum := CRC(b.Bytes())
	b.WriteByte(uint8(checksum & uint16(0xff)))
	b.WriteByte(uint8((checksum & uint16(0xff00) >> 8)))
	return b.Bytes()
}

// ID returns the client ID of the RTUFrame.
func (f *RTUFrame) ID() uint8 {
	return f.rawData[0]
}

// Function returns the function code of the RTUFrame.
func (f *RTUFrame) Function() RTUFunction {
	return RTUFunction(f.rawData[1])
}

// Data returns the data after the RTU function code up to the CRC.
func (f *RTUFrame) Data() []byte {
	return f.rawData[3 : len(f.rawData)-2]
}

// RawData returns the entire RTU buffer, including ID, function, and CRC.
func (f *RTUFrame) RawData() []byte {
	return f.rawData
}

// CRC returns the CRC of the RTUFrame.
func (f *RTUFrame) CRC() uint16 {
	return uint16(f.rawData[len(f.rawData)-2]) + (uint16(f.rawData[len(f.rawData)-1]) * uint16(256))
}

// readRTUResponse reads an entire RTUFrame.
//
// The frame returned will be nil in case of an error, except for protocol and/or CRC errors.
func readRTUResponse(port io.Reader) (*RTUFrame, error) {
	b := make([]byte, MaxRTUFrameLength)
	if n, err := io.ReadFull(port, b[0:3]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame: got %d, want at least 3 bytes", n)
		}
		return nil, err
	}
	pending := expectedResponseLength(RTUFunction(b[1]), b[2])
	if pending == -1 {
		return nil, fmt.Errorf("invalid function code: %02x", b[1])
	}
	pending += 2 // Add the 2 CRC bytes.
	if (3 + pending) >= MaxRTUFrameLength {
		return nil, fmt.Errorf("invalid data reading frame (out of bounds): %s", hex.EncodeToString(b[0:8]))
	}
	if n, err := io.ReadFull(port, b[3:3+pending]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame data: got %d, want %d bytes", n, 3+pending)
		}
		return nil, fmt.Errorf("error reading frame data: %w", err)
	}
	checksum := CRC(b[0 : 3+pending-2])
	frame := NewRTUFrame(b[0 : 3+pending])

	// CRC error or protocol error also return the frame.
	var err error
	if (b[1] & 0x80) == 0x80 {
		err = protocolError(b[1])
	}
	if checksum != frame.CRC() {
		if err == nil {
			err = fmt.Errorf("invalid crc: got %x, want %x", frame.CRC(), checksum)
		} else {
			err = fmt.Errorf("%w (in addition, invalid crc: got %x, want %x)", err, frame.CRC(), checksum)
		}
	}
	return frame, err
}

// CRC returns the CRC16 for the given bytes.
func CRC(data []byte) uint16 {
	var crc16 uint16 = 0xffff
	l := len(data)
	for i := 0; i < l; i++ {
		crc16 ^= uint16(data[i])
		for j := 0; j < 8; j++ {
			if crc16&1 > 0 {
				crc16 = (crc16 >> 1) ^ 0xa001
			} else {
				crc16 >>= 1
			}
		}
	}
	return crc16
}

// readRTURequest reads an entire RTUFrame for a request.
func readRTURequest(port io.Reader) (*RTUFrame, error) {
	b := make([]byte, MaxRTUFrameLength)
	// Reading 8 works for all request types.
	if n, err := io.ReadFull(port, b[0:8]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame: got %d, want at least 8 bytes", n)
		}
		return nil, err
	}
	pending := expectedRequestLength(RTUFunction(b[1]), uint16(b[5])*256+uint16(b[6]))
	if pending == -1 {
		return nil, fmt.Errorf("invalid function code: %02x\n", b[1])
	}
	pending += 2 // Add 2 CRC bytes
	pending -= 6 // Subtract 6 bytes of the response already read (not including ID and function).
	if n, err := io.ReadFull(port, b[8:8+pending]); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("short frame: got %d, want at least 8 bytes", n)
		}
		return nil, fmt.Errorf("error reading frame data: %w", err)
	}
	checksum := CRC(b[0 : 8+pending-2])
	frame := NewRTUFrame(b[0 : 8+pending])
	if checksum != frame.CRC() {
		return frame, fmt.Errorf("invalid crc: got %x, want %x", frame.CRC(), checksum)
	}
	return frame, nil
}

type RTU struct {
	port common.Port
}

func NewRTU(port common.Port) RegisterReader {
	return &RTU{port: port}
}

// ReadRegisters requests 'count' holding registers from unit 'id' from the 'start' memory address.
// and reads the response back.
func (r *RTU) ReadRegisters(id uint8, start uint16, count uint8) ([]byte, error) {
	f := buildReadRequestRTUFrame(id, ReadHoldingRegisters, start, uint16(count))
	if _, err := r.port.Write(f); err != nil {
		return nil, err
	}
	frame, err := readRTUResponse(r.port)
	if err != nil {
		return nil, err
	}
	return frame.Data(), nil
}

func expectedResponseLength(functionCode RTUFunction, receivedLength uint8) int {
	switch functionCode {
	case ReadCoils, ReadInputRegisters, ReadHoldingRegisters, ReadDiscreteInputs:
		return int(receivedLength)
	case WriteSingleCoil, WriteSingleRegister, WriteMultipleCoil, WriteMultipleRegisters:
		return 3
	default:
		if (functionCode & 0x80) == 0x80 {
			// Error response
			f := functionCode & 0x7f
			switch f {
			case ReadCoils, ReadInputRegisters, ReadHoldingRegisters, ReadDiscreteInputs, WriteSingleCoil,
				WriteSingleRegister, WriteMultipleCoil, WriteMultipleRegisters:
				return 1 // 1 byte in the response data contains the error code.
			}
		}
		return -1
	}
}

func expectedRequestLength(functionCode RTUFunction, count uint16) int {
	switch functionCode {
	case ReadCoils, ReadInputRegisters, ReadHoldingRegisters, ReadDiscreteInputs, WriteSingleCoil, WriteSingleRegister:
		return 4
	case WriteMultipleCoil, WriteMultipleRegisters:
		// count is at byte 6 in the request. ID+function code excluded, that results in 5 bytes in the response data already.
		return 5 + int(count)
	}
	return -1
}

func protocolError(code uint8) error {
	if s, ok := protocolErrorMap[RTUProtocolError(0x7f&code)]; ok {
		return fmt.Errorf("protocol error: %s", s)
	}
	return fmt.Errorf("protocol error: unknown error code %02x", code)
}
