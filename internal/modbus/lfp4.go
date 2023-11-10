package modbus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"wombatt/internal/common"
)

// LFP4 comms is described in https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
// It is NOT Modbus.
type LFP4 struct {
	port common.Port
}

func NewLFP4(port common.Port) RegisterReader {
	return &LFP4{port: port}
}

func buildReadRequestLFP4Frame(id uint8, cid2 uint8) *RTUFrame {
	var b bytes.Buffer
	b.WriteByte(0x7e)                                           // SOI
	b.WriteString("20")                                         // VER
	b.WriteString(fmt.Sprintf("%02X", id))                      // ADR
	b.WriteString("4A")                                         // CID1 = BMS/LiFePO4 battery BMS
	b.WriteString(fmt.Sprintf("%02X", cid2))                    // CID2
	b.WriteString("0000")                                       // LENGTH
	b.WriteString(fmt.Sprintf("%04X", lfp4Checksum(b.Bytes()))) // CHKSUM
	b.WriteByte(0x0d)                                           // EOI
	return &RTUFrame{rawData: b.Bytes()}
}

// ReadRegisters sends the cid2 command to unit id and returns the response.
func (t *LFP4) ReadRegisters(id uint8, _ uint16, cid2 uint8) (*RTUFrame, error) {
	f := buildReadRequestLFP4Frame(id, cid2)
	if _, err := t.port.Write(f.RawData()); err != nil {
		return nil, err
	}
	rframe, err := t.ReadResponse(id)
	if err != nil {
		return nil, err
	}
	// Remove header info, crc and EOI.
	ascii := rframe.RawData()
	data, err := hex.DecodeString(string(ascii[13 : len(ascii)-3]))
	if err != nil {
		return nil, fmt.Errorf("error decoding ascii data: %w", err)
	}
	return NewRTUFrame(data), nil
}

func (t *LFP4) ReadResponse(id uint8) (*RTUFrame, error) {
	header := make([]byte, 13) // 1 byte for SOI; 2 for each of VER, ADR, CIR1, and RTN; 4 for LENGTH
	if _, err := io.ReadFull(t.port, header); err != nil {
		return nil, err
	}
	// Check RTN
	ret, err := asciiToBin(header[7:9])
	if err != nil {
		return nil, err
	}
	if ret != 0 {
		return nil, fmt.Errorf("received error code %d", ret)
	}
	// Check LCHKSUM
	bin, err := asciiToBin(header[9:13])
	if err != nil {
		return nil, err
	}
	length := uint16(bin)
	if err := checkLengthChecksum(length); err != nil {
		return nil, err
	}
	length &= 0x0fff                   // clear out LCHKSUM and leave just the actual length
	ascii := make([]byte, 13+length+5) // 13 for the header, 3 for CHKSUM + EOI
	copy(ascii[0:13], header)
	n, err := io.ReadFull(t.port, ascii[13:])
	if err != nil {
		return nil, err
	}
	if n != (len(ascii) - 13) {
		return nil, fmt.Errorf("short response. got %d, want %d", n, len(ascii))
	}
	// Check CHKSUM
	err = verifyChecksum(ascii) // if there's an error here, it's sent back with the data.
	if ascii[len(ascii)-1] != 0xd {
		log.Printf("warning: EOI missing in response")
	}
	return NewRTUFrame(ascii), err
}

func (t *LFP4) Close() {
	t.port.Close()
}

func verifyChecksum(b []byte) error {
	targetCRC, err := asciiToBin(b[len(b)-5 : len(b)-1])
	if err != nil {
		return fmt.Errorf("error getting target CRC: %w", err)
	}
	sum := lfp4Checksum(b[:len(b)-5])
	if sum != uint16(targetCRC) {
		return fmt.Errorf("CHKSUM error: got %X, want %X", sum, targetCRC)
	}
	return nil
}

func lfp4Checksum(b []byte) uint16 {
	var sum uint32
	for _, c := range b[1:] {
		sum += uint32(c)
	}
	sum = sum & 65535
	sum = (^sum + 1) & 65535
	return uint16(sum)
}

func checkLengthChecksum(length uint16) error {
	chksum := length >> 12
	d11 := (length & 0x0f00) >> 8
	d7 := (length & 0x00f0) >> 4
	d3 := (length & 0x000f)
	sum := (d11 + d7 + d3) & 0xf
	sum = (^sum + 1) & 0xf
	if chksum != sum {
		return fmt.Errorf("LCHKSUM error")
	}
	return nil
}

func asciiToBin(ascii []byte) (int, error) {
	b := make([]byte, hex.DecodedLen(len(ascii)))
	if _, err := hex.Decode(b, ascii); err != nil {
		return 0, fmt.Errorf("error decoding ascii string '%s'", ascii)
	}
	var result int
	for i := 0; i < len(b); i++ {
		result = (result << 8) + int(b[i])
	}
	return result, nil
}
