package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"wombatt/internal/batteries"
	"wombatt/internal/common"
	"wombatt/internal/mqttha"
	"wombatt/internal/web"

	"go.bug.st/serial"
)

type MonitorBatteriesCmd struct {
	MQTTFlags `embed:""`

	SerialPort string `short:"p" required:"" help:"Serial port attached to all batteries, except maybe battery ID 1"`
	BaudRate   uint   `short:"B" default:"9600" help:"Baud rate for serial ports"`

	IDs []uint `short:"i" required:"" name:"battery-ids" help:"IDs of the batteries to monitor"`

	PollInterval time.Duration `short:"P" default:"10s" help:"Time to wait between polling cycles"`
	ReadTimeout  time.Duration `short:"t" default:"500ms" help:"Timeout when reading from serial ports"`

	BatteryType BatteryType `default:"EG4LLv2" help:"Battery type" enum:"${battery_types}"`
	MQTTPrefix  string      `default:"eg4" help:"MQTT prefix for the fields published"`

	WebServerAddress string `short:"w" help:"Address to use for serving HTTP. <IP>:<Port>, i.e., 127.0.0.1:8080"`
}

type batteryInfo struct {
	ID   uint8
	Info any
}

func (mb *MonitorBatteriesCmd) Run(globals *Globals) error {
	for _, id := range mb.IDs {
		if id == 0 || id >= 247 {
			log.Fatalf("id out of range: %d", id)
		}
	}
	var webServer *web.Server
	if len(mb.WebServerAddress) > 0 {
		webServer = web.NewServer(mb.WebServerAddress, "/battery/")
		if err := webServer.Start(); err != nil {
			log.Fatalf("%v", err)
		}
	}
	battery := batteries.Instance(string(mb.BatteryType))
	var mqttChannel chan *batteryInfo
	if mb.MQTTBroker != "" {
		mqttChannel = make(chan *batteryInfo, len(mb.IDs))
		client, err := mqttha.Connect(mb.MQTTBroker, mb.MQTTUser, mb.MQTTPassword)
		if err != nil {
			log.Fatalf("error connecting to MQTT broker at %s: %v\n", mb.MQTTBroker, err)
		}
		defer client.Disconnect(250)
		go mqttPublish(client, mqttChannel, mb, battery.InfoInstance())
	}
	if webServer == nil && mqttChannel == nil {
		log.Fatalf("need at least MQTT or web server argument to publish info to.\n")
	}
	ch := make(chan *batteryInfo, len(mb.IDs))
	go func() {
		for bi := range ch {
			if mqttChannel != nil {
				mqttChannel <- bi
			}
			if webServer != nil {
				webServer.Publish(fmt.Sprintf("%d", bi.ID), bi.Info)
			}
		}
	}()
	portOptions := &common.PortOptions{
		Name: mb.SerialPort,
		Mode: &serial.Mode{BaudRate: int(mb.BaudRate)},
	}
	port := common.OpenPortOrFatal(portOptions)
	monitorBatteries(ch, port, mb, battery)
	return nil
}

func monitorBatteries(ch chan *batteryInfo, port common.Port, opts *MonitorBatteriesCmd, battery batteries.Battery) {
	for {
		log.Printf("Fetching info from batteries\n")
		success := []uint{}
		for _, id := range opts.IDs {
			info, err := battery.ReadInfo(port, uint8(id), opts.ReadTimeout)
			if err != nil {
				if err := port.ReopenWithBackoff(); err != nil {
					log.Printf("error reopening: %v\n", err)
				}
				continue
			}
			if ch != nil {
				ch <- &batteryInfo{uint8(id), info}
			} else {
				fmt.Printf("Battery #%d\n===========\n", id)
				writeBatteryInfo(info)
				fmt.Println()
			}
			success = append(success, id)
		}
		log.Printf("Published info for %v\n", success)
		time.Sleep(opts.PollInterval)
	}
}

func mqttPublish(client mqttha.Client, ch chan *batteryInfo, opts *MonitorBatteriesCmd, emptyInfo any) {
	createDiscoveryConfig(client, opts, emptyInfo)
	for bi := range ch {
		config := make(map[string]interface{})
		f := func(info map[string]string, value any) {
			config[info["name"]] = value
		}
		common.TraverseStruct(bi.Info, f)
		topic := fmt.Sprintf("%s/sensor/%s_battery%d_info/state", opts.MQTTTopicPrefix, opts.MQTTPrefix, bi.ID)
		if err := client.PublishMap(topic, config); err != nil {
			log.Printf("[mqtt] error publishing to %s: %v\n", opts.MQTTBroker, err)
		}
	}
}

func createDiscoveryConfig(client mqttha.Client, opts *MonitorBatteriesCmd, emptyInfo any) {
	for _, id := range opts.IDs {
		addDiscoveryConfig(client, opts, id, emptyInfo)
	}
}

func addDiscoveryConfig(client mqttha.Client, opts *MonitorBatteriesCmd, id uint, st any) {
	f := func(info map[string]string, value any) {
		name := info["name"]
		config := map[string]interface{}{
			// "expire_after":?
			// "force_update":   true,
			"state_topic":    fmt.Sprintf("%s/sensor/%s_battery%d_info/state", opts.MQTTTopicPrefix, opts.MQTTPrefix, id),
			"name":           fmt.Sprintf("Battery %d %s", id, strings.ReplaceAll(name, "_", " ")),
			"object_id":      fmt.Sprintf("%s_battery_%d_%s", opts.MQTTPrefix, id, name),
			"value_template": fmt.Sprintf("{{ value_json.%s }}", name),
		}
		config["unique_id"] = config["object_id"]
		dclass := info["dclass"]
		unit := info["unit"]
		icon := info["icon"]
		if dclass != "" {
			config["device_class"] = dclass
		}
		if unit != "" {
			config["unit_of_measurement"] = unit
			config["state_class"] = "measurement"
		}
		if icon != "" {
			config["icon"] = icon
		}

		topic := fmt.Sprintf("%s/sensor/%s_battery%d_%s/config", opts.MQTTTopicPrefix, opts.MQTTPrefix, id, name)
		if err := client.PublishMap(topic, config); err != nil {
			log.Printf("[mqtt] error publishing to %s: %v\n", opts.MQTTBroker, err)
		}
	}
	common.TraverseStruct(st, f)
}
