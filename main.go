package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"wombatt/cmd"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
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

	// Create a context that is canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	kctx := kong.Parse(&cli,
		kong.Name("wombatt"),
		kong.Description("A wanna-be Swiss army knife for inverter and battery monitoring."),
		kong.UsageOnError(),
		kong.Configuration(kongyaml.Loader, "/etc/wombatt.yaml", "~/.wombatt.yaml"),
		kong.ConfigureHelp(kong.HelpOptions{
			//			Compact: true,
		}),
		kong.Vars{
			"bms_types":    "EG4LLv2,lifepower4,lifepowerv2,pacemodbus",
			"device_types": "serial,hidraw,tcp",
			"protocols":    "auto,ModbusRTU,ModbusTCP,lifepower4",
		})
	logSetup(cli.Globals.LogLevel)
	err := kctx.Run(&cli.Globals, ctx)
	kctx.FatalIfErrorf(err)
}
