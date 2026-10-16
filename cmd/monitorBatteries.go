package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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

func (cmd *MonitorBatteriesCmd) Run(globals *Globals, ctx context.Context) error {
	var webServer *web.Server
	if len(cmd.WebServerAddress) > 0 {
		webServer = web.NewServer(cmd.WebServerAddress, "/battery/")
		if err := webServer.Start(); err != nil {
			log.Fatalf("%v", err)
		}
		defer func() {
			slog.Info("shutting down web server...")
			sdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = webServer.Shutdown(sdCtx)
		}()
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
		go mqttPublish(ctx, client, mqttChannel, cmd, battery.InfoInstance())
	}
	if webServer == nil && mqttChannel == nil {
		log.Fatalf("need at least MQTT or web server argument to publish info to.\n")
	}
	if mqttChannel != nil {
		defer close(mqttChannel)
	}
	portOptions := &common.PortOptions{
		Address:     cmd.Address,
		Mode:        &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:        common.DeviceTypeFromString[cmd.DeviceType],
		ReadTimeout: cmd.ReadTimeout,
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			port, err := common.OpenPort(portOptions)
			if err != nil {
				slog.Error("failed to open port", "address", cmd.Address, "error", err)
			} else {
				monitorBatteries(ctx, mqttChannel, webServer, port, cmd, battery)
				port.Close()
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(cmd.PollInterval):
			}
		}
	}
}

func monitorBatteries(ctx context.Context, mqttChannel chan *batteryInfo, webServer *web.Server, port common.Port, cmd *MonitorBatteriesCmd, battery bms.BMS) {
	reader, err := modbus.Reader(port, cmd.Protocol, string(cmd.BMSType))
	if err != nil {
		slog.Error("error creating modbus reader", "error", err)
		return
	}
	slog.Info("fetching info from batteries", "battery-id", cmd.ID)
	success := []uint{}
	for _, id := range cmd.ID {
		if ctx.Err() != nil {
			return
		}
		info, err := battery.ReadInfo(ctx, reader, uint8(id), cmd.ReadTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("error reading battery", "battery-id", id, "error", err)
			if webServer != nil {
				webServer.Publish(fmt.Sprintf("%d", id), nil) // Don't keep serving stale data.
			}
			if err := port.ReopenWithBackoff(ctx); err != nil {
				slog.Error("error reopening", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
			continue
		}
		if ctx.Err() != nil {
			return
		}
		if mqttChannel != nil {
			mqttChannel <- &batteryInfo{uint8(id), info}
		}
		if webServer != nil {
			webServer.Publish(fmt.Sprintf("%d", id), info)
		}
		success = append(success, id)
		select {
		case <-ctx.Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
	if ctx.Err() != nil {
		return
	}
	slog.Info("published info for batteries", "battery-id", success)
}

func mqttPublish(ctx context.Context, client *mqttha.Client, ch chan *batteryInfo, cmd *MonitorBatteriesCmd, emptyInfo any) {
	createDiscoveryConfig(ctx, client, cmd, emptyInfo)
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
		if err := client.PublishMap(ctx, topic, config, mqttha.NoRetain, mqttha.TopicAlias); err != nil {
			slog.Error("mqtt error publishing", "server", cmd.MQTTBroker, "error", err)
		}
	}
}

func createDiscoveryConfig(ctx context.Context, client *mqttha.Client, cmd *MonitorBatteriesCmd, emptyInfo any) {
	for _, id := range cmd.ID {
		addDiscoveryConfig(ctx, client, cmd, id, emptyInfo)
	}
}

func addDiscoveryConfig(ctx context.Context, client *mqttha.Client, cmd *MonitorBatteriesCmd, id uint, st any) {
	stateTopic := fmt.Sprintf("%s/sensor/%s_battery%d_info/state", cmd.MQTTTopicPrefix, cmd.MQTTPrefix, id)
	device := map[string]any{
		"identifiers": []string{fmt.Sprintf("%s_battery%d", cmd.MQTTPrefix, id)},
		"name":        fmt.Sprintf("Battery %d", id),
		"model":       cmd.BMSType,
	}
	f := func(info map[string]string, _ any) {
		name := mqttha.Key(info["name"])
		uniqueID := fmt.Sprintf("%s_battery_%d_%s", cmd.MQTTPrefix, id, name)
		config := mqttha.DiscoveryConfig(info, stateTopic, uniqueID, device)
		topic := fmt.Sprintf("%s/sensor/%s_battery%d_%s/config", cmd.MQTTTopicPrefix, cmd.MQTTPrefix, id, name)
		if err := client.PublishDiscovery(ctx, topic, config); err != nil {
			slog.Error("mqtt error publishing", "server", cmd.MQTTBroker, "error", err)
		}
	}
	common.TraverseStruct(st, f)
}
