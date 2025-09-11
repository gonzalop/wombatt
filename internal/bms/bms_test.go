package bms

import (
	"bytes"
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
		resp     string // hex-encoded response
		isExtra  bool
		protocol string
		bmsType  string
		value    any
	}{
		{
			resp:     "02034e14f600780d1a0d190d1b0d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1a0d1a0d1b0d1a0d1a0019001b0018006100640064006100010000000000000000000b15752a00181818180000001003e800004e71",
			protocol: modbus.RTUProtocol,
			bmsType:  "EG4LLv2",
			value: &EG4BatteryInfo{
				EG4ModbusBatteryInfo: EG4ModbusBatteryInfo{
					Voltage:            5366,
					Current:            120,
					CellVoltages:       [16]uint16{3354, 3353, 3355, 3355, 3354, 3355, 3354, 3355, 3354, 3355, 3354, 3354, 3354, 3355, 3354, 3354},
					PCBTemp:            25,
					MaxTemp:            27,
					AvgTemp:            24,
					CapRemaining:       97,
					MaxChargingCurrent: 100,
					SOH:                100,
					SOC:                97,
					Status:             1,
					CycleCounts:        11,
					FullCapacity:       100000,
					Temp1:              24,
					Temp2:              24,
					Temp3:              24,
					Temp4:              24,
					CellNum:            16,
					DesignedCapacity:   1000,
				},
				VoltageStats: VoltageStats{
					MaxVoltage:    3355,
					MinVoltage:    3353,
					MeanVoltage:   3354,
					MedianVoltage: 3354,
				},
			},
		},
		{
			resp:     "02032e4c46502d35312e325631303041682d56312e3000000000005a3032543034323032322d31302d32360000000000004818",
			isExtra:  true,
			protocol: modbus.RTUProtocol,
			bmsType:  "EG4LLv2",
			value: &EG4ModbusExtraBatteryInfo{
				// "LFP-51.2V100Ah-V1.0"
				// "Z02T04"
				// "2022-10-26"
				Model:           [24]byte{76, 70, 80, 45, 53, 49, 46, 50, 86, 49, 48, 48, 65, 104, 45, 86, 49, 46, 48, 0, 0, 0, 0, 0},
				FirmwareVersion: [6]byte{90, 48, 50, 84, 48, 52},
				Serial:          [16]byte{50, 48, 50, 50, 45, 49, 48, 45, 50, 54, 0, 0, 0, 0, 0, 0},
			},
		},
		// lifepowerv2 tests are the same as the EG4LLv2 tests except for their bmsType
		{
			resp:     "02034e14f600780d1a0d190d1b0d1b0d1a0d1b0d1a0d1b0d1a0d1b0d1a0d1a0d1a0d1b0d1a0d1a0019001b0018006100640064006100010000000000000000000b15752a00181818180000001003e800004e71",
			protocol: modbus.RTUProtocol,
			bmsType:  "lifepowerv2",
			value: &EG4BatteryInfo{
				EG4ModbusBatteryInfo: EG4ModbusBatteryInfo{
					Voltage:            5366,
					Current:            120,
					CellVoltages:       [16]uint16{3354, 3353, 3355, 3355, 3354, 3355, 3354, 3355, 3354, 3355, 3354, 3354, 3354, 3355, 3354, 3354},
					PCBTemp:            25,
					MaxTemp:            27,
					AvgTemp:            24,
					CapRemaining:       97,
					MaxChargingCurrent: 100,
					SOH:                100,
					SOC:                97,
					Status:             1,
					CycleCounts:        11,
					FullCapacity:       100000,
					Temp1:              24,
					Temp2:              24,
					Temp3:              24,
					Temp4:              24,
					CellNum:            16,
					DesignedCapacity:   1000,
				},
				VoltageStats: VoltageStats{
					MaxVoltage:    3355,
					MinVoltage:    3353,
					MeanVoltage:   3354,
					MedianVoltage: 3354,
				},
			},
		},
		{
			resp:     "02032e4c46502d35312e325631303041682d56312e3000000000005a3032543034323032322d31302d32360000000000004818",
			isExtra:  true,
			protocol: modbus.RTUProtocol,
			bmsType:  "lifepowerv2",
			value: &EG4ModbusExtraBatteryInfo{
				// "LFP-51.2V100Ah-V1.0"
				// "Z02T04"
				// "2022-10-26"
				Model:           [24]byte{76, 70, 80, 45, 53, 49, 46, 50, 86, 49, 48, 48, 65, 104, 45, 86, 49, 46, 48, 0, 0, 0, 0, 0},
				FirmwareVersion: [6]byte{90, 48, 50, 84, 48, 52},
				Serial:          [16]byte{50, 48, 50, 50, 45, 49, 48, 45, 50, 54, 0, 0, 0, 0, 0, 0},
			},
		},
		{
			resp:     "7e32303031344130304130434130313031313030433534304338313043383130433832304338313043383130433831304338313043383230433832304338323043383230433832304338323043383230433745303430424344304243443042434430424344304244373042443730303030313346443030303032373130303030303046303030303030363430433832304335343030324530424344304243443030303030303135303030303030334330303030303030413030303030303144303030303030303030303030303030303030303230303144443330300d",
			protocol: modbus.Lifepower4Protocol,
			bmsType:  "lifepower4",
			value: &LFP4AnalogValueBatteryInfo{
				DataFlag:          1,
				NumberOfCells:     16,
				CellVoltages:      [16]uint16{3156, 3201, 3201, 3202, 3201, 3201, 3201, 3201, 3202, 3202, 3202, 3202, 3202, 3202, 3202, 3198},
				CellTemps:         [4]uint16{3021, 3021, 3021, 3021},
				EnvTemp:           3031,
				MOSFETTemp:        3031,
				PackVoltage:       5117,
				FullCapacity:      10000,
				UserDefined:       15,
				SOH:               100,
				MaxCellVoltage:    3202,
				MinCellVoltage:    3156,
				CellVoltageDiff:   46,
				MaxCellTemp:       3021,
				MinCellTemp:       3021,
				CumChargingCap:    21,
				CumDischargeCap:   60,
				CumChargingPower:  10,
				CumDischargePower: 29,
				CumChargingTimes:  2,
				CumDischargeTimes: 29,
			},
		},
		{
			resp:     "7e323030313441303037303534303130313130303030303030303030303030303030303030303030303030303030303030303030343030303030303030303030303030303030393030303030303030303030313033303030303030303030303030454443340d",
			isExtra:  true,
			protocol: modbus.Lifepower4Protocol,
			bmsType:  "lifepower4",
			value: &LFP4AlarmInfo{
				DataFlag:               1,
				NumberOfCells:          16,
				UserDefined:            9,
				RemainingCapacityAlarm: 1,
				FETStatusCode:          3,
			},
		},
	}

	for tid, tt := range tests {
		rawResp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed raw response string in test %d: %s", tid, tt.resp)
		}
		port := common.NewTestPort(bytes.NewReader(rawResp), io.Discard, 0)
		reader, err := modbus.Reader(port, tt.protocol, "")
		if reader == nil {
			t.Fatalf("no available reader: %v", err)
		}
		inst, err := Instance(tt.bmsType)
		if err != nil {
			t.Fatalf("error creating BMS instance: %v", err)
		}
		if inst == nil {
			t.Fatalf("invalid battery type with no instance")
		}
		if inst.DefaultProtocol("") != tt.protocol {
			t.Fatalf("wrong protocol (%d). got %v; want %v", tid, inst.DefaultProtocol(""), tt.protocol)
		}
		var bat any
		if !tt.isExtra {
			if reflect.TypeOf(inst.InfoInstance()) != reflect.TypeOf(tt.value) {
				t.Errorf("wrong instance type (%d). want %v; got %v", tid, reflect.TypeOf(bat), reflect.TypeOf(tt.value))
			}
			bat, err = inst.ReadInfo(reader, 2, 1*time.Second)
		} else {
			bat, err = inst.ReadExtraInfo(reader, 2, 1*time.Second)
		}
		if err != nil {
			t.Fatalf("error reading battery info (%d): %v", tid, err)
		}

		if !reflect.DeepEqual(bat, tt.value) {
			t.Errorf("structs not equal (%d). got %v; want %v", tid, bat, tt.value)
		}
	}
}
