package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"wombatt/internal/bms"
	"wombatt/internal/common"
	"wombatt/internal/modbus"
	"wombatt/internal/mqttha"
	"wombatt/internal/web"

	"go.bug.st/serial"
)

type MonitorBatteriesCmd struct {
	MQTTFlags `embed:""`

	Address  string `short:"p" required:"" help:"Serial port attached to the batteries"`
	BaudRate uint   `short:"B" default:"9600" help:"Baud rate for serial ports"`

	ID []uint `short:"i" required:"" name:"battery-id" help:"IDs of the batteries to monitor"`

	PollInterval time.Duration `short:"P" default:"10s" help:"Time to wait between polling cycles"`
	ReadTimeout  time.Duration `short:"t" default:"500ms" help:"Timeout when reading from devices"`

	BMSType    string `default:"EG4LLv2" help:"One of ${bms_types}" enum:"${bms_types}"`
	MQTTPrefix string `default:"eg4" help:"MQTT prefix for the fields published"`

	WebServerAddress string `short:"w" help:"Address to use for serving HTTP. <IP>:<Port>, i.e., 127.0.0.1:8080"`

	Protocol   string `default:"auto" enum:"${protocols}" help:"One of ${protocols}"`
	DeviceType string `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
}

type batteryInfo struct {
	ID   uint8
	Info any
}

func (cmd *MonitorBatteriesCmd) Run(globals *Globals) error {
	var webServer *web.Server
	if len(cmd.WebServerAddress) > 0 {
		webServer = web.NewServer(cmd.WebServerAddress, "/battery/")
		if err := webServer.Start(); err != nil {
			log.Fatalf("%v", err)
		}
	}
	battery, err := bms.Instance(string(cmd.BMSType))
	if err != nil {
		return fmt.Errorf("failed to create BMS instance: %w", err)
	}
	if cmd.Protocol == "auto" {
		cmd.Protocol = battery.DefaultProtocol(cmd.DeviceType)
	}
	var mqttChannel chan *batteryInfo
	if cmd.MQTTBroker != "" {
		mqttChannel = make(chan *batteryInfo, len(cmd.ID))
		client, err := mqttha.Connect(cmd.MQTTBroker, cmd.MQTTUser, cmd.MQTTPassword)
		if err != nil {
			log.Fatalf("error connecting to MQTT broker at %s: %v\n", cmd.MQTTBroker, err)
		}
		defer client.Disconnect(250)
		go mqttPublish(client, mqttChannel, cmd, battery.InfoInstance())
	}
	if webServer == nil && mqttChannel == nil {
		log.Fatalf("need at least MQTT or web server argument to publish info to.\n")
	}
	ch := make(chan *batteryInfo, len(cmd.ID))
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
		Address: cmd.Address,
		Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:    common.DeviceTypeFromString[cmd.DeviceType],
	}
	port, err := common.OpenPort(portOptions)
	if err != nil {
		return fmt.Errorf("failed to open port: %w", err)
	}
	monitorBatteries(ch, port, cmd, battery)
	return nil
}

func monitorBatteries(ch chan *batteryInfo, port common.Port, cmd *MonitorBatteriesCmd, battery bms.BMS) {
	reader, err := modbus.Reader(port, cmd.Protocol, string(cmd.BMSType))
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		slog.Info("fetching info from batteries", "battery-id", cmd.ID)
		success := []uint{}
		for _, id := range cmd.ID {
			info, err := battery.ReadInfo(reader, uint8(id), cmd.ReadTimeout)
			if err != nil {
				if err := port.ReopenWithBackoff(); err != nil {
					slog.Error("error reopening", "error", err)
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
		slog.Info("published info for batteries", "battery-id", success)
		time.Sleep(cmd.PollInterval)
	}
}

func mqttPublish(client *mqttha.Client, ch chan *batteryInfo, cmd *MonitorBatteriesCmd, emptyInfo any) {
	createDiscoveryConfig(client, cmd, emptyInfo)
	for bi := range ch {
		config := make(map[string]any)
		f := func(info map[string]string, value any) {
			config[info["name"]] = value
		}
		common.TraverseStruct(bi.Info, f)
		config["device"] = map[string]string{
			"identifiers": fmt.Sprintf("%s_battery%d", cmd.MQTTPrefix, bi.ID),
		}
		topic := fmt.Sprintf("%s/sensor/%s_battery%d_info/state", cmd.MQTTTopicPrefix, cmd.MQTTPrefix, bi.ID)
		if err := client.PublishMap(topic, false, config); err != nil {
			slog.Error("mqtt error publishing", "server", cmd.MQTTBroker, "error", err)
		}
	}
}

func createDiscoveryConfig(client *mqttha.Client, cmd *MonitorBatteriesCmd, emptyInfo any) {
	for _, id := range cmd.ID {
		addDiscoveryConfig(client, cmd, id, emptyInfo)
	}
}

func addDiscoveryConfig(client *mqttha.Client, cmd *MonitorBatteriesCmd, id uint, st any) {
	f := func(info map[string]string, value any) {
		name := info["name"]
		config := map[string]any{
			// "expire_after":?
			// "force_update":   true,
			"state_topic":       fmt.Sprintf("%s/sensor/%s_battery%d_info/state", cmd.MQTTTopicPrefix, cmd.MQTTPrefix, id),
			"name":              fmt.Sprintf("Battery %d %s", id, strings.ReplaceAll(name, "_", " ")),
			"default_entity_id": fmt.Sprintf("%s_battery_%d_%s", cmd.MQTTPrefix, id, name),
			"value_template":    fmt.Sprintf("{{ value_json.%s }}", name),
			"device": map[string]any{
				"identifiers": []string{fmt.Sprintf("%s_battery%d", cmd.MQTTPrefix, id)},
				"name":        fmt.Sprintf("Battery %d", id),
				"model":       cmd.BMSType,
			},
		}
		config["unique_id"] = config["default_entity_id"]
		dclass := info["dclass"]
		if dclass != "" {
			config["device_class"] = dclass
		}
		unit := info["unit"]
		if unit != "" {
			config["unit_of_measurement"] = unit
			config["state_class"] = "measurement"
		}
		icon := info["icon"]
		if icon != "" {
			config["icon"] = icon
		}
		precision := info["precision"]
		if precision != "" {
			num, err := strconv.Atoi(precision)
			if err == nil {
				config["suggested_display_precision"] = num
			}
		}

		topic := fmt.Sprintf("%s/sensor/%s_battery%d_%s/config", cmd.MQTTTopicPrefix, cmd.MQTTPrefix, id, name)
		if err := client.PublishMap(topic, true, config); err != nil {
			slog.Error("mqtt error publishing", "server", cmd.MQTTBroker, "error", err)
		}
	}
	common.TraverseStruct(st, f)
}
