package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/mqttha"
	"wombatt/internal/pi30"
	"wombatt/internal/web"

	"go.bug.st/serial"
)

type MonitorInvertersCmd struct {
	MQTTFlags `embed:""`

	BaudRate     uint          `short:"B" default:"2400" help:"Baud rate for serial ports"`
	PollInterval time.Duration `short:"P" default:"10s" help:"Time to wait between polling cycles"`
	ReadTimeout  time.Duration `short:"t" default:"5s" help:"Timeout when reading from devices"`

	Monitors []string `arg:"" required:"" help:"<device>,<command1[:command2:command3...]>[,<mqtt_prefix>]. E.g. /dev/ttyS0,QPIRI:QPGS1,eg4_1"`

	WebServerAddress string `short:"w" help:"Address to use for serving HTTP. <IP>:<Port>, i.e., 127.0.0.1:8080"`

	DeviceType string `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
}

func (cmd *MonitorInvertersCmd) Run(globals *Globals) error {
	if len(cmd.Monitors) == 0 {
		return fmt.Errorf("missing inverter ports")
	}
	monitors, err := getMonitors(cmd.Monitors)
	if err != nil {
		log.Fatal(err)
	}
	var client mqttha.Client
	if cmd.MQTTBroker != "" {
		var err error
		client, err = mqttha.Connect(cmd.MQTTBroker, cmd.MQTTUser, cmd.MQTTPassword)
		if err != nil {
			log.Fatalf("error connecting to MQTT broker at %s: %v\n", cmd.MQTTBroker, err)
		}
	}
	var webServer *web.Server
	if len(cmd.WebServerAddress) > 0 {
		webServer = web.NewServer(cmd.WebServerAddress, "/inverter/")
		if err := webServer.Start(); err != nil {
			log.Fatalf("%v", err)
		}
	}
	for _, m := range monitors {
		m.client = client
		m.webServer = webServer
	}
	return runInverterMonitor(cmd, monitors)
}

type inverterMonitor struct {
	Device   string
	Commands []string
	MQTTTag  string

	client    mqttha.Client
	webServer *web.Server
}

func runInverterMonitor(cmd *MonitorInvertersCmd, monitors []*inverterMonitor) error {
	var wg sync.WaitGroup
	if monitors[0].client != nil {
		invertersDiscoveryConfig(cmd.MQTTTopicPrefix, monitors)
	}
	ctx := context.Background()
	for {
		responses := make([]*cmdResponse, len(monitors))
		wg.Add(len(monitors))
		for i, m := range monitors {
			go func(i int, m *inverterMonitor) {
				defer wg.Done()
				portOptions := &common.PortOptions{
					Address: m.Device,
					Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
					Type:    common.DeviceTypeFromString[cmd.DeviceType],
				}
				port, err := common.OpenPort(portOptions)
				if err != nil {
					slog.Error("error opening device", "device", m.Device, "error", err)
					responses[i] = &cmdResponse{nil, []error{err}, m}
					return
				}
				defer port.Close()
				ctx_to, cancel := context.WithTimeout(ctx, cmd.ReadTimeout)
				defer cancel()
				slog.Info("fetching info from inverter", "inverter-name", m.Device, "commands", m.Commands)
				results, errors := pi30.RunCommands(ctx_to, port, m.Commands)
				responses[i] = &cmdResponse{results, errors, m}
				okCommands := []string{}
				for k := range errors {
					if errors[k] != nil {
						continue
					}
					okCommands = append(okCommands, m.Commands[k])
				}
				slog.Info("publishing info from inverter", "inverter-name", m.Device, "commands", okCommands)
			}(i, m)
		}
		wg.Wait()
		for i, r := range responses {
			r.ValidateResponses()
			r.Publish(cmd.MQTTTopicPrefix, i)
		}
		responses = nil
		time.Sleep(cmd.PollInterval)
	}
}

type cmdResponse struct {
	Responses []any
	Errors    []error
	monitor   *inverterMonitor
}

func (r *cmdResponse) ValidateResponses() {
	for i, resp := range r.Responses {
		v, ok := resp.(pi30.ResponseChecker)
		if !ok || v.Valid() {
			continue
		}
		r.Responses[i] = nil
		if r.Errors[i] == nil {
			r.Errors[i] = fmt.Errorf("invalid response for %v", r.monitor.Commands[i])
		}
	}
}

func (r *cmdResponse) Publish(topicPrefix string, cmdIndex int) {
	m := r.monitor
	for ic, ir := range r.Responses {
		if r.Errors[ic] != nil {
			slog.Error("error running command", "command", m.Commands[ic], "device", m.Device, "error", r.Errors[ic])
			continue
		}
		if m.webServer != nil {
			m.webServer.Publish(fmt.Sprintf("%d/%s", cmdIndex+1, m.Commands[ic]), ir)
		}
	}

	if m.client != nil {
		m.publishToMQTT(topicPrefix, r.Responses, r.Errors)
	}

	if m.client == nil && m.webServer == nil {
		publishToStdout(m, r.Responses)
	}
}

func publishToStdout(im *inverterMonitor, results []any) {
	for i, r := range results {
		if r == nil {
			continue
		}
		fmt.Printf("%s -> %s\n=======================\n", im.Device, im.Commands[i])
		pi30.WriteTo(os.Stdout, r)
		fmt.Println()
	}
}

func invertersDiscoveryConfig(mqttTopicPrefix string, monitors []*inverterMonitor) {
	for _, m := range monitors {
		for _, c := range m.Commands {
			st := pi30.StructForCommand(c)
			switch st.(type) {
			case *pi30.EmptyResponse:
				continue
			}
			addStructDiscoveryConfig(m.client, st, mqttTopicPrefix, m.MQTTTag)
		}
	}
}

func addStructDiscoveryConfig(client mqttha.Client, st any, topicPrefix, tag string) {
	f := func(info map[string]string, value any) {
		name := info["name"]
		config := map[string]any{
			// "expire_after":?
			// "force_update":   true,
			"state_topic":    fmt.Sprintf("%s/sensor/%s_info/state", topicPrefix, tag),
			"name":           fmt.Sprintf("Inverter %s %s", strings.TrimSpace(strings.ReplaceAll(tag, "_", " ")), name),
			"object_id":      fmt.Sprintf("%s_%s", tag, name),
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

		topic := fmt.Sprintf("%s/sensor/%s_%s/config", topicPrefix, tag, name)
		if err := client.PublishMap(topic, true, config); err != nil {
			slog.Error("mqtt error publishing", "error", err)
		}
	}
	common.TraverseStruct(st, f)
}

func (im *inverterMonitor) publishToMQTT(mqttTopicPrefix string, results []any, errors []error) {
	config := make(map[string]any)
	f := func(info map[string]string, value any) {
		config[info["name"]] = value
	}
	for i, st := range results {
		if errors[i] != nil {
			slog.Error("error running command", "commandIndex", i, "error", errors[i])
			continue
		}
		common.TraverseStruct(st, f)
	}
	if len(config) == 0 {
		return
	}
	topic := fmt.Sprintf("%s/sensor/%s_info/state", mqttTopicPrefix, im.MQTTTag)
	if err := im.client.PublishMap(topic, false, config); err != nil {
		slog.Error("mqtt error publishing", "error", err)
	}
}

func getMonitors(args []string) ([]*inverterMonitor, error) {
	var monitors []*inverterMonitor
	for _, arg := range args {
		p := strings.SplitN(arg, ",", 3)
		if len(p) < 2 {
			return nil, fmt.Errorf("invalid inverter argument: '%s'", arg)
		}
		dev := p[0]
		var cmds []string
		for c := range strings.SplitSeq(p[1], ":") {
			cmd := strings.TrimSpace(c)
			if cmd == "" {
				continue
			}
			cmds = append(cmds, cmd)
		}
		if len(cmds) == 0 {
			return nil, fmt.Errorf("no inverter commands in '%s'", arg)
		}
		prefix := ""
		if len(p) > 2 {
			prefix = p[2]
		}
		monitors = append(monitors, &inverterMonitor{dev, cmds, prefix, nil, nil})
	}
	return monitors, nil
}
