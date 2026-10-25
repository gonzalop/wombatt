package modbus

import (
	"bytes"
	"cmp"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"wombatt/internal/common"
)

func TestTCPReadHoldingRegisters(t *testing.T) {
	tests := []struct {
		resp       string
		nregisters uint8
		errstr     string
		id         uint8
		fcode      RTUFunction
	}{
		{
			resp:       "00010000002303034e14d800000d060d080d060d080d070d080d070d080d070d080d070d080d060d070d060d070019001b0018006400640064006400000000000000000000000a15752a00181817170000001003e800000000",
			id:         3,
			fcode:      3,
			nregisters: 16,
		},
		{
			resp:   "00010000001001032000670000006314d3ff10001f09",
			id:     1,
			fcode:  3,
			errstr: "unexpected transaction ID: got 0x0001; want 0x0002",
		},
		{
			resp:   "000303040506", // no bytes after length
			errstr: "EOF",
		},
		{
			resp:   "00040304050607",
			errstr: "short frame: read 1, want at least 1286 bytes",
		},
		{
			resp:   "0005", // incomplete header
			errstr: "short frame: read 2, want at least 6 bytes",
		},
		{
			resp:   "",
			errstr: "EOF",
		},
		{
			resp:       "00070000003103032e4c46502d35312e325631303041682d56312e3000000000005a3032543034323032322d31302d3236000000000000",
			id:         3,
			fcode:      3,
			nregisters: 23,
		},
		{
			resp:   "000800000000", // length 0
			errstr: "invalid frame length: 0",
		},
		{
			resp:   "000900000003018302", // exception: illegal data address
			errstr: "protocol error: illegal data address",
		},
	}

	tid.Store(0) // Reset the transaction counter in tcp.go so we get predictable TIDs
	for tid, tt := range tests {
		resp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed response string in test: %s", tt.resp)
		}
		port := common.NewTestPort(bytes.NewReader(resp), io.Discard, 0)
		tcp, _ := Reader(port, TCPProtocol, "")
		reqID := cmp.Or(tt.id, 1)
		data, err := tcp.ReadHoldingRegisters(reqID, 0, tt.nregisters)
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
		if len(data) != (int(tt.nregisters) * 2) {
			t.Errorf("%d wrong data length: got %d; want %d", tid, len(data), int(tt.nregisters)*2)
		}
	}
}

func TestTCP_UnitIDMismatch(t *testing.T) {
	// TID: 0x0001, Protocol: 0x0000, Length: 0x0005 (UnitID + Function + ByteCount + 2 bytes data)
	// UnitID in response: 0x02, but requested: 0x01
	resp, _ := hex.DecodeString("000100000005020302002a")
	port := common.NewTestPort(bytes.NewReader(resp), io.Discard, common.TCPDevice)
	tcp, _ := Reader(port, TCPProtocol, "")
	tid.Store(0)
	_, err := tcp.ReadHoldingRegisters(1, 0, 1)
	if err == nil {
		t.Fatal("expected unit ID mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected unit ID") {
		t.Errorf("expected error containing 'unexpected unit ID', got: %v", err)
	}
}
