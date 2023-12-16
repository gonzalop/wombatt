package modbus

import (
	"fmt"
	"io"
	"testing"

	"wombatt/internal/common"
)

func TestReader(t *testing.T) {
	tests := []struct {
		readerTypeName string
		protocol       string
		batteryType    string
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
			batteryType:    "lifepower4",
			deviceType:     common.SerialDevice,
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol:       "auto",
			batteryType:    "lifepower4",
			deviceType:     common.TCPDevice,
			readerTypeName: "*modbus.LFP4",
		},
		{
			protocol:       "auto",
			batteryType:    "lifepower4",
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
		r, err := Reader(port, tt.protocol, tt.batteryType)
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
