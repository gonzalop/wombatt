package cmd

import (
	"context"
	"fmt"
	"log"
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
	ReadTimeout  time.Duration `short:"t" default:"5s" help:"Timeout when reading from serial ports"`

	Monitors []string `arg:"" required:"" help:"<device>,<command1[:command2:command3...]>[,<mqtt_prefix>]. E.g. /dev/ttyS0,QPIRI:QPGS1,eg4_1"`

	WebServerAddress string `short:"w" help:"Address to use for serving HTTP. <IP>:<Port>, i.e., 127.0.0.1:8080"`

	DeviceType string `short:"T" default:"serial" enum:"${device_types}" help:"Device type"`
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
					log.Printf("error opening %s: %v\n", m.Device, err)
					responses[i] = &cmdResponse{nil, []error{err}, m}
					return
				}
				defer port.Close()
				ctx_to, cancel := context.WithTimeout(ctx, cmd.ReadTimeout)
				defer cancel()
				results, errors := pi30.RunCommands(ctx_to, port, m.Commands)
				responses[i] = &cmdResponse{results, errors, m}
			}(i, m)
		}
		wg.Wait()
		for i, r := range responses {
			m := r.monitor
			m.Publish(cmd.MQTTTopicPrefix, r.Responses, r.Errors)
			for ic, ir := range r.Responses {
				if r.Errors[ic] != nil {
					continue
				}
				m.webServer.Publish(fmt.Sprintf("%d/%s", i+1, r.monitor.Commands[ic]), ir)
			}
		}
		time.Sleep(cmd.PollInterval)
	}
}

type cmdResponse struct {
	Responses []any
	Errors    []error
	monitor   *inverterMonitor
}

func publishToStdout(im *inverterMonitor, results []any, errors []error) {
	for i, err := range errors {
		if err == nil {
			continue
		}
		log.Printf("error running %s on %s: %v\n", im.Commands[i], im.Device, err)
	}
	for i, r := range results {
		if r == nil {
			continue
		}
		fmt.Printf("%s -> %s\n=======================\n", im.Device, im.Commands[i])
		pi30.WriteTo(os.Stdout, r)
		fmt.Println()
	}
}

func (im *inverterMonitor) Publish(mqttTopicPrefix string, results []any, errors []error) {
	if im.client == nil && im.webServer == nil {
		publishToStdout(im, results, errors)
		return
	}
	if im.client != nil {
		im.publishToMQTT(mqttTopicPrefix, results, errors)
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
		config := map[string]interface{}{
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
		if err := client.PublishMap(topic, config); err != nil {
			log.Printf("[mqtt] error publishing: %v\n", err)
		}
	}
	common.TraverseStruct(st, f)
}

func (im *inverterMonitor) publishToMQTT(mqttTopicPrefix string, results []any, errors []error) {
	config := make(map[string]interface{})
	for i, st := range results {
		f := func(info map[string]string, value any) {
			config[info["name"]] = value
		}
		if errors[i] == nil {
			common.TraverseStruct(st, f)
		} else {
			log.Printf("%v\n", errors[i])
		}
	}
	if len(config) == 0 {
		return
	}
	topic := fmt.Sprintf("%s/sensor/%s_info/state", mqttTopicPrefix, im.MQTTTag)
	if err := im.client.PublishMap(topic, config); err != nil {
		log.Printf("[mqtt] error publishing: %v\n", err)
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
		for _, c := range strings.Split(p[1], ":") {
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
