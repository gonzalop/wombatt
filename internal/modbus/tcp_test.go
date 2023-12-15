package modbus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"wombatt/internal/common"
)

func TestTCPReadRegisters(t *testing.T) {
	tests := []struct {
		resp   string
		errstr string
		id     uint8
		fcode  RTUFunction
	}{
		{
			resp:  "00010000001001032000670000006314d3ff10001f09c49ab09c400204000000060000000015e00000",
			id:    1,
			fcode: 3,
		},
		{
			resp:   "00010000001001032000670000006314d3ff10001f09c49ab09c400204000000060000000015e00000",
			id:     1,
			fcode:  3,
			errstr: "unexpected transaction ID: got 0x0001; want 0x0002",
		},
		{
			resp:   "000303040506",
			errstr: "EOF", // from the io package directly
		},
		{
			resp:   "00040304050607",
			errstr: "unexpected EOF",
		},
	}

	tid.Store(0) // Reset the transaction counter in tcp.go so we get predictable TIDs
	for _, tt := range tests {
		resp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed response string in test: %s", tt.resp)
		}
		port := common.NewTestPort(bytes.NewReader(resp), io.Discard, 0)
		tcp, _ := Reader(port, TCPProtocol, "")
		data, err := tcp.ReadRegisters(1, 16, 1)
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
		frame := NewRTUFrame(data)
		if frame.ID() != tt.id {
			t.Errorf("wrong ID in response(%s): got %02d; want %02d", tt.resp, frame.ID(), tt.id)
		}
		if frame.Function() != tt.fcode {
			t.Errorf("wrong function code in response(%s): got %02d; want %02d", tt.resp, frame.Function(), tt.fcode)
		}
	}
}
