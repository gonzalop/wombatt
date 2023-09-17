package main

import (
	"log"

	"wombatt/cmd"

	"github.com/alecthomas/kong"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
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
			"version":       "0.0.1",
			"battery_types": "EG4LLv2",
			"device_types":  "serial,hidraw",
		})
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
