package main

import (
	"log"
	"log/slog"
	"os"

	"wombatt/cmd"

	"github.com/alecthomas/kong"
)

func logSetup(levelName string) {
	var level slog.Level

	switch levelName {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		log.Fatalf("invalid log level %q", levelName)
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

func main() {
	cli := cmd.CLI{
		Globals: cmd.Globals{},
	}
	ctx := kong.Parse(&cli,
		kong.Name("wombatt"),
		kong.Description("A wanna-be Swiss army knife for inverter and battery monitoring."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			//			Compact: true,
		}),
		kong.Vars{
			"version":      "0.0.15",
			"bms_types":    "EG4LLv2,lifepower4,lifepowerv2,pacemodbus",
			"device_types": "serial,hidraw,tcp",
			"protocols":    "auto,ModbusRTU,ModbusTCP,lifepower4",
		})
	logSetup(cli.Globals.LogLevel)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
