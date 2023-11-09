package modbus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"wombatt/internal/common"
)

func TestReadRTURequest(t *testing.T) {
	tests := []struct {
		req    string
		errstr string
		id     uint8
		fcode  RTUFunction
		crc    uint16
	}{
		{ // Request from EG4 6500-EX to battery ID#1
			req:   "010300130010b5c3",
			id:    1,
			fcode: ReadHoldingRegisters,
			crc:   uint16(0xc3b5),
		},
		{
			req:    "",
			errstr: "EOF",
		},
		{
			req:    "018300130010b5c3",
			errstr: "invalid function code", // 0x83
		},
		{ // TODO: add one for write coils?
			req:   "010300131010b803",
			id:    1,
			fcode: ReadHoldingRegisters,
			crc:   uint16(0x3b8),
		},
		{
			req:    "110300130010b5c3",
			errstr: "invalid crc",
		},
		{
			req:    "01",
			errstr: "short frame",
		},
		{
			req:    "010300130010",
			errstr: "short frame",
		},
	}

	for _, tt := range tests {
		req, err := hex.DecodeString(tt.req)
		if err != nil {
			t.Fatalf("malformed request string in test: %s", tt.req)
		}
		r := bytes.NewReader(req)
		frame, err := readRTURequest(r)
		if err != nil && tt.errstr == "" {
			t.Errorf("read request failed(%s): got %v; want no error", tt.req, err)
			continue
		} else if err == nil && tt.errstr != "" {
			t.Errorf("read request succeded, but it should fail(%s): got no error; want %v", tt.req, err)
			continue
		}
		if err != nil {
			s := fmt.Sprintf("%v", err)
			if !strings.Contains(s, tt.errstr) {
				t.Errorf("unkown error(%s): got '%s'; want error with '%s'", tt.req, s, tt.errstr)
			}
			continue
		}
		if frame.ID() != tt.id {
			t.Errorf("wrong ID in request(%s): got %02d; want %02d", tt.req, frame.ID(), tt.id)
		}
		if frame.Function() != tt.fcode {
			t.Errorf("wrong function code in request(%s): got %02d; want %02d", tt.req, frame.Function(), tt.fcode)
		}
		if frame.CRC() != tt.crc {
			t.Errorf("wrong CRC in request(%s): got %04x; want %04x", tt.req, frame.CRC(), tt.crc)
		}
	}
}

func TestReadRTUResponse(t *testing.T) {
	tests := []struct {
		resp   string
		errstr string
		id     uint8
		fcode  RTUFunction
		crc    uint16
	}{
		{
			resp:   "",
			errstr: "EOF",
		},
		{
			resp:   "01",
			errstr: "short frame",
		},
		{ // Response from EG4LLv2 #1 to EG4 6500-EX.
			resp:  "01032000670000006314d3ff10001f09c49ab09c400204000000060000000015e0000070c0",
			id:    1,
			fcode: ReadHoldingRegisters,
			crc:   0xc070,
		},
		{
			resp:   "01f320",
			errstr: "invalid function code",
		},
		{
			resp:   "01032000670000006314d3ff10001f09c49ab09c400204000000060000000015e0000070c1",
			errstr: "invalid crc",
		},
		{
			resp:   "0103fd00670000006314d3ff10001f09c49ab09c400204000000060000000015e0000070c0",
			errstr: "out of bounds", // 0xFD
		},
		{
			resp:   "010320",
			errstr: "EOF",
		},
		{
			resp:   "018320012830",
			errstr: "illegal data value", // error 3
		},
		{
			resp:   "018320012831",
			errstr: "in addition, invalid crc", // illegal data value + crc error
		},
	}

	for _, tt := range tests {
		resp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed response string in test: %s", tt.resp)
		}
		r := bytes.NewReader(resp)
		frame, err := readRTUResponse(r)
		if err != nil && tt.errstr == "" {
			t.Errorf("read response failed(%s): got %v; want no error", tt.resp, err)
			continue
		} else if err == nil && tt.errstr != "" {
			t.Errorf("read response succeded, but it should fail(%s): got no error; want %v", tt.resp, tt.errstr)
			continue
		}
		if err != nil {
			s := fmt.Sprintf("%v", err)
			if !strings.Contains(s, tt.errstr) {
				t.Errorf("unkown error(%s): got '%s'; want error with '%s'", tt.resp, s, tt.errstr)
			}
			continue
		}
		if frame.ID() != tt.id {
			t.Errorf("wrong ID in response(%s): got %02d; want %02d", tt.resp, frame.ID(), tt.id)
		}
		if frame.Function() != tt.fcode {
			t.Errorf("wrong function code in response(%s): got %02d; want %02d", tt.resp, frame.Function(), tt.fcode)
		}
		if frame.CRC() != tt.crc {
			t.Errorf("wrong CRC in response(%s): got %04x; want %04x", tt.resp, frame.CRC(), tt.crc)
		}
	}
}

func TestReadRegisters(t *testing.T) {
	tests := []struct {
		resp   string
		errstr string
		id     uint8
		fcode  RTUFunction
		crc    uint16
	}{
		{
			resp:   "0103",
			errstr: "invalid crc",
		},
		{
			resp:  "01032000670000006314d3ff10001f09c49ab09c400204000000060000000015e0000070c0",
			id:    1,
			fcode: 3,
			crc:   0xc070,
		},
	}

	for _, tt := range tests {
		resp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed response string in test: %s", tt.resp)
		}
		port, _ := common.OpenPort(&common.PortOptions{Type: common.TestByteDevice, Address: string(resp)})
		rtu, _ := ReaderFromProtocol(port, "RTU")
		defer rtu.Close()
		frame, err := rtu.ReadRegisters(1, 16, 1)
		if err != nil && tt.errstr == "" {
			t.Errorf("read response failed(%s): got %v; want no error", tt.resp, err)
			continue
		} else if err == nil && tt.errstr != "" {
			t.Errorf("read response succeded, but it should fail(%s): got no error; want %v", tt.resp, tt.errstr)
			continue
		}
		if err != nil {
			s := fmt.Sprintf("%v", err)
			if !strings.Contains(s, tt.errstr) {
				t.Errorf("unkown error(%s): got '%s'; want error with '%s'", tt.resp, s, tt.errstr)
			}
			continue
		}
		if frame.ID() != tt.id {
			t.Errorf("wrong ID in response(%s): got %02d; want %02d", tt.resp, frame.ID(), tt.id)
		}
		if frame.Function() != tt.fcode {
			t.Errorf("wrong function code in response(%s): got %02d; want %02d", tt.resp, frame.Function(), tt.fcode)
		}
		if frame.CRC() != tt.crc {
			t.Errorf("wrong CRC in response(%s): got %04x; want %04x", tt.resp, frame.CRC(), tt.crc)
		}
	}
}
