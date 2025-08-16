package cmd

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kong"

	"wombatt/internal/common"
	"wombatt/internal/modbus"
	"wombatt/internal/solark"
)

type Globals struct {
	//	Config   string      `help:"Location of client config files" default:"~/.eg4.yaml" type:"path"`
	LogLevel string      `short:"l" enum:"debug,info,warn,error" help:"Set the logging level (debug|info|warn|error)" default:"info"`
	Version  VersionFlag `short:"v" name:"version" help:"Print version information and quit"`
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
	MonitorInverters MonitorInvertersCmd `cmd:"" help:"Monitors inverters using PI30 protocol, MQTT publishing optional"`
	SolarkQuery      SolarkQueryCmd      `cmd:"" help:"Query Solark inverter via Modbus RTU/TCP"`
}

// MQTTFlags are embedded in multiple commands.
type MQTTFlags struct {
	MQTTBroker      string `group:"MQTT" env:"MQTT_BROKER" help:"The MQTT server to publish battery data. E.g. tcp://127.00.0.1:1883"`
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

type SolarkQueryCmd struct {
	PortName string `short:"p" help:"Port name (e.g., /dev/ttyUSB0, COM1, localhost:502)" required:"true"`
	BaudRate int    `short:"b" help:"Baud rate for serial port" default:"9600"`
	DataBits int    `short:"D" help:"Data bits for serial port" default:"8"`
	StopBits int    `short:"S" help:"Stop bits for serial port" default:"1"`
	Parity   string `short:"P" help:"Parity for serial port (N, E, O)" default:"N"`
	Timeout  int    `short:"t" help:"Timeout in seconds" default:"5"`
	Protocol string `short:"R" help:"Modbus protocol (auto, ModbusRTU, ModbusTCP)" default:"auto"`
	ModbusID int    `short:"i" help:"Modbus slave ID" default:"1"`
}

func (c *SolarkQueryCmd) Run() error {
	port, err := common.NewPort(c.PortName, c.BaudRate, c.DataBits, c.StopBits, c.Parity, c.Timeout)
	if err != nil {
		slog.Error("failed to create port", "error", err)
		return err
	}
	defer port.Close()

	reader, err := modbus.Reader(port, c.Protocol, "")
	if err != nil {
		slog.Error("failed to create modbus reader", "error", err)
		return err
	}

	rtd, err := solark.ReadRealtimeData(reader, uint8(c.ModbusID))
	if err != nil {
		slog.Error("failed to read realtime data", "error", err)
		return err
	}

	fmt.Println("--- Solark Real-time Data ---")
	common.TraverseStruct(rtd, func(tags map[string]string, value any) {
		fmt.Printf("%s: %v %s\n", tags["name"], value, tags["unit"])
	})

	ia, err := solark.ReadIntrinsicAttributes(reader, uint8(c.ModbusID))
	if err != nil {
		slog.Error("failed to read intrinsic attributes", "error", err)
		return err
	}

	fmt.Println("\n--- Solark Intrinsic Attributes ---")
	fmt.Printf("Serial Number: %s\n", ia.SerialNumber())

	return nil
}
