package batteries

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"reflect"
	"testing"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/modbus"
)

func TestBatteryInfo(t *testing.T) {
	tests := []struct {
		resp        string
		isExtra     bool
		protocol    string
		batteryType string
		data        string // hex encoded struct data:
		value       any
	}{
		{
			resp:        "02034e14f600780d190d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1a0d1a0d1b0d1a0d1a0019001b0018006100640064006100010000000000000000000b15752a00181818180000001003e800003dea",
			protocol:    modbus.RTUProtocol,
			batteryType: "EG4LLv2",
			data:        "14f600780d190d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1a0d1a0d1b0d1a0d1a0019001b0018006100640064006100010000000000000000000b000186a0181818180000001003e800000d1b0d190d1a0d1a",
			value: &EG4BatteryInfo{
				EG4ModbusBatteryInfo: EG4ModbusBatteryInfo{
					Voltage:            5366,
					Current:            120,
					CellVoltages:       [16]uint16{3353, 3355, 3354, 3355, 3354, 3355, 3354, 3355, 3354, 3355, 3354, 3354, 3354, 3355, 3354, 3354},
					PCBTemp:            25,
					MaxTemp:            27,
					AvgTemp:            24,
					CapRemaining:       97,
					MaxChargingCurrent: 100,
					SOH:                100,
					SOC:                97,
					Status:             1,
					Warning:            0,
					Protection:         0,
					ErrorCode:          0,
					CycleCounts:        11,
					FullCapacity:       100000,
					Temp1:              24,
					Temp2:              24,
					Temp3:              24,
					Temp4:              24,
					Temp5:              0,
					Temp6:              0,
					CellNum:            16,
					DesignedCapacity:   1000,
					CellBalanceStatus:  0,
				},
				MaxVoltage:    3355,
				MinVoltage:    3353,
				MeanVoltage:   3354,
				MedianVoltage: 3354,
			},
		},
		{
			resp:        "02032e4c46502d35312e325631303041682d56312e3000000000005a3032543034323032322d31302d32360000000000004818",
			isExtra:     true,
			protocol:    modbus.RTUProtocol,
			batteryType: "EG4LLv2",
			data:        "4c46502d35312e325631303041682d56312e3000000000005a3032543034323032322d31302d3236000000000000",
			value: &EG4ModbusExtraBatteryInfo{
				Model:           [24]byte{76, 70, 80, 45, 53, 49, 46, 50, 86, 49, 48, 48, 65, 104, 45, 86, 49, 46, 48, 0, 0, 0, 0, 0},
				FirmwareVersion: [6]byte{90, 48, 50, 84, 48, 52},
				Serial:          [16]byte{50, 48, 50, 50, 45, 49, 48, 45, 50, 54, 0, 0, 0, 0, 0, 0},
			},
		},
	}

	for tid, tt := range tests {
		rawResp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed raw response string in test %d: %s", tid, tt.resp)
		}
		port := common.NewTestPort(bytes.NewReader(rawResp), io.Discard)
		reader, err := modbus.ReaderFromProtocol(port, tt.protocol)
		if reader == nil {
			t.Fatalf("no available reader: %v", err)
		}
		llv2 := Instance(tt.batteryType)
		if llv2 == nil {
			t.Fatalf("invalid battery type with no instance")
		}
		var bat any
		if !tt.isExtra {
			bat, err = llv2.ReadInfo(reader, 2, 1*time.Second)
		} else {
			bat, err = llv2.ReadExtraInfo(reader, 2, 1*time.Second)
		}
		if err != nil {
			t.Fatalf("error reading battery info: %v", err)
		}
		var b bytes.Buffer
		if err := binary.Write(&b, binary.BigEndian, bat); err != nil {
			t.Fatalf("error writing to buffer: %v", err)
		}
		if !reflect.DeepEqual(bat, tt.value) {
			t.Errorf("got %v; want %v", bat, tt.value)
		}
	}
}
