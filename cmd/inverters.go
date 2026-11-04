package cmd

import (
	"context"
	"maps"
	"slices"
	"strings"

	"wombatt/internal/common"
	"wombatt/internal/eg4_18kpv"
	"wombatt/internal/eg4_6000xp"
	"wombatt/internal/pi30"
	"wombatt/internal/solark"
)

// inverterDriver is everything the commands need to know about one inverter type.
type inverterDriver struct {
	// run sends the commands to the inverter. protocol and id are ignored by pi30.
	run func(ctx context.Context, port common.Port, protocol string, id uint8, commands []string) ([]any, []error)
	// discoveryStruct returns an empty result struct for command, or nil if
	// the command has no Home Assistant sensors.
	discoveryStruct func(command string) any
}

// inverterDrivers is the one list of supported inverter types.
var inverterDrivers = map[string]inverterDriver{
	"pi30": {
		run: func(ctx context.Context, port common.Port, _ string, _ uint8, commands []string) ([]any, []error) {
			return pi30.RunCommands(ctx, port, commands)
		},
		discoveryStruct: func(c string) any {
			st := pi30.StructForCommand(c)
			if _, ok := st.(*pi30.EmptyResponse); ok {
				return nil
			}
			return st
		},
	},
	"solark": {
		run: solark.RunCommands,
		discoveryStruct: func(c string) any {
			switch c {
			case "RealtimeData":
				return &solark.RealtimeData{}
			case "IntrinsicAttributes":
				return &solark.IntrinsicAttributes{}
			}
			return nil
		},
	},
	"eg4_18kpv": {
		run: eg4_18kpv.RunCommands,
		discoveryStruct: func(c string) any {
			if c == "RealtimeData" {
				return &eg4_18kpv.RealtimeData{}
			}
			return nil
		},
	},
	"eg4_6000xp": {
		run: eg4_6000xp.RunCommands,
		discoveryStruct: func(c string) any {
			if c == "RealtimeData" {
				return &eg4_6000xp.RealtimeData{}
			}
			return nil
		},
	},
}

// InverterTypes lists the supported inverter types, comma-separated, for
// command-line enums and error messages.
func InverterTypes() string {
	return strings.Join(slices.Sorted(maps.Keys(inverterDrivers)), ",")
}
