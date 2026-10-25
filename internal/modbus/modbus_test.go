package modbus

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"wombatt/internal/common"
)

func TestReader(t *testing.T) {
	tests := []struct {
		readerTypeName string
		protocol       string
		bmsType        string
		deviceType     common.DeviceType
		mustFail       bool
	}{
		{
			protocol:       "ModbusRTU",
			readerTypeName: "*modbus.RTU",
		},
		{
			protocol:       "ModbusTCP",
			readerTypeName: "*modbus.TCP",
		},
		{
			protocol:       "lifepower4",
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol:       "auto",
			deviceType:     common.SerialDevice,
			readerTypeName: "*modbus.RTU",
		},
		{
			protocol:       "auto",
			deviceType:     common.HidRawDevice,
			readerTypeName: "*modbus.RTU",
		},
		{
			protocol:       "auto",
			deviceType:     common.TCPDevice,
			readerTypeName: "*modbus.TCP",
		},
		{
			protocol:       "auto",
			bmsType:        "lifepower4",
			deviceType:     common.SerialDevice,
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol:       "auto",
			bmsType:        "lifepower4",
			deviceType:     common.TCPDevice,
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol:       "auto",
			bmsType:        "lifepower4",
			deviceType:     common.HidRawDevice,
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol: "auto",
			mustFail: true,
		},
		{
			protocol: "whatever",
			mustFail: true,
		},
	}
	for tid, tt := range tests {
		port := common.NewTestPort(nil, io.Discard, tt.deviceType)
		r, err := Reader(port, tt.protocol, tt.bmsType)
		if err == nil && tt.mustFail {
			t.Errorf("error (#%d): got no error, expecting an error", tid)
			continue
		}
		if err != nil && !tt.mustFail {
			t.Errorf("error (#%d): got error %q, expecting no error", tid, err)
			continue
		}
		if r == nil {
			// tt.mustFail = true, and got a nil
			continue
		}
		rtype := fmt.Sprintf("%T", r)
		if rtype != tt.readerTypeName {
			t.Errorf("error(#%d): got %v; want %v", tid, rtype, tt.readerTypeName)
		}
	}
}

func TestRunCommands_ContextCancellation(t *testing.T) {
	port := common.NewTestPort(nil, io.Discard, common.TCPDevice)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	results, errs := RunCommands(ctx, port, "ModbusTCP", 1, []string{"cmd1", "cmd2"}, func(reader RegisterReader, id uint8, cmd string) (any, error) {
		time.Sleep(80 * time.Millisecond)
		return "ok", nil
	})
	elapsed := time.Since(start)

	if len(results) != 2 || len(errs) != 2 {
		t.Fatalf("expected 2 results and 2 errs, got %d and %d", len(results), len(errs))
	}
	if errs[0] == nil || !errors.Is(errs[0], context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded for cmd1, got: %v", errs[0])
	}
	if errs[1] == nil || !errors.Is(errs[1], context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded for cmd2, got: %v", errs[1])
	}
	if elapsed < 70*time.Millisecond {
		t.Errorf("expected RunCommands to wait for in-flight goroutine, elapsed: %v", elapsed)
	}
}
