package cmd

import (
	"regexp"
	"testing"

	"wombatt/internal/bms"
	"wombatt/internal/common"
	"wombatt/internal/eg4_18kpv"
	"wombatt/internal/eg4_6000xp"
	"wombatt/internal/mqttha"
	"wombatt/internal/pi30"
	"wombatt/internal/solark"
)

// Home Assistant only accepts these characters in discovery object ids.
var validObjectID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func TestMQTTKeysAreValid(t *testing.T) {
	structs := map[string]any{
		"pi30 Q1":                    pi30.StructForCommand("Q1"),
		"pi30 QPIRI":                 pi30.StructForCommand("QPIRI"),
		"pi30 QPIGS":                 pi30.StructForCommand("QPIGS"),
		"pi30 QPIGS2":                pi30.StructForCommand("QPIGS2"),
		"pi30 QPGS1":                 pi30.StructForCommand("QPGS1"),
		"solark RealtimeData":        &solark.RealtimeData{},
		"solark IntrinsicAttributes": &solark.IntrinsicAttributes{},
		"eg4_18kpv RealtimeData":     &eg4_18kpv.RealtimeData{},
		"eg4_6000xp RealtimeData":    &eg4_6000xp.RealtimeData{},
	}
	for _, bmsType := range []string{bms.EG4LLv2BMS, bms.Lifepower4BMS, bms.PaceBMS} {
		b, err := bms.Instance(bmsType)
		if err != nil {
			t.Fatal(err)
		}
		structs["bms "+bmsType] = b.InfoInstance()
	}

	for desc, st := range structs {
		seen := map[string]string{}
		common.TraverseStruct(st, func(info map[string]string, _ any) {
			key := mqttha.Key(info["name"])
			if !validObjectID.MatchString(key) {
				t.Errorf("%s: field %q gives invalid MQTT key %q", desc, info["name"], key)
			}
			if prev, ok := seen[key]; ok {
				t.Errorf("%s: fields %q and %q share MQTT key %q", desc, prev, info["name"], key)
			}
			seen[key] = info["name"]
		})
	}
}

func TestGetMonitors(t *testing.T) {
	monitors, err := getMonitors([]string{"/dev/ttyS0,QPIGS: QPGS1:,eg4_1", "/dev/ttyUSB0,RealtimeData,sk,solark"})
	if err != nil {
		t.Fatal(err)
	}
	if len(monitors) != 2 {
		t.Fatalf("got %d monitors; want 2", len(monitors))
	}
	m := monitors[0]
	if m.Device != "/dev/ttyS0" || m.MQTTTag != "eg4_1" || m.InverterType != "pi30" {
		t.Errorf("got %+v", m)
	}
	if len(m.Commands) != 2 || m.Commands[0] != "QPIGS" || m.Commands[1] != "QPGS1" {
		t.Errorf("got commands %q; want [QPIGS QPGS1]", m.Commands)
	}
	if monitors[1].InverterType != "solark" {
		t.Errorf("got type %q; want solark", monitors[1].InverterType)
	}

	for _, bad := range []string{"/dev/ttyS0,QPIGS", "/dev/ttyS0,:,x", "/dev/ttyS0,QPIGS,x,unknown"} {
		if _, err := getMonitors([]string{bad}); err == nil {
			t.Errorf("getMonitors(%q) succeeded; want error", bad)
		}
	}
}
