package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type Globals struct {
	Config   kong.ConfigFlag `help:"Location of client config files" type:"path"`
	LogLevel string          `short:"l" enum:"debug,info,warn,error" help:"Set the logging level (debug|info|warn|error)" default:"info"`
	Version  VersionFlag     `short:"v" name:"version" help:"Print version information and quit"`
}

type VersionFlag bool

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	Globals

	BatteryInfo      BatteryInfoCmd      `cmd:"" help:"Displays battery information"`
	Forward          ForwardCmd          `cmd:"" help:"Forwards commands between a two devices"`
	InverterQuery    InverterQueryCmd    `cmd:"" help:"Sends PI30 protocol commands to inverters"`
	ModbusRead       ModbusReadCmd       `cmd:"" help:"Reads Modbus holding registers\n"`
	MonitorBatteries MonitorBatteriesCmd `cmd:"" help:"Monitors batteries state, MQTT publishing optional"`
	MonitorInverters MonitorInvertersCmd `cmd:"" help:"Monitors inverters state, MQTT publishing optional"`
}

// MQTTFlags are embedded in multiple commands.
type MQTTFlags struct {
	MQTTBroker      string `group:"MQTT" env:"MQTT_BROKER" help:"The MQTT server to publish battery data. E.g. tcp://127.0.0.1:1883"`
	MQTTPassword    string `group:"MQTT" env:"MQTT_PASSWORD" help:"Password for the MQTT connection"`
	MQTTTopicPrefix string `group:"MQTT" env:"MQTT_TOPIC_PREFIX" default:"homeassistant" help:"Prefix for all topics published to MQTT"`
	MQTTUser        string `group:"MQTT" env:"MQTT_USER" help:"User for the MQTT connection"`
}

// Validate checks MQTT flags for multiple commands.
func (f *MQTTFlags) Validate() error {
	broker := f.MQTTBroker != ""
	user := f.MQTTUser != ""
	password := f.MQTTPassword != ""
	if user && !broker {
		return fmt.Errorf("MQTT user provided but no MQTT broker")
	}
	if (user && !password) || (!user && password) {
		return fmt.Errorf("both MQTT user and password are needed")
	}
	return nil
}
