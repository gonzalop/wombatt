package mqttha

import (
	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MqttClient has additional methods to the regular mqtt.Client ones.
type Client interface {
	mqtt.Client

	// PublishMap will publish a JSON encoded map to the given topic.
	PublishMap(topic string, config map[string]interface{}) error
}

type haClient struct {
	mqtt.Client
}

// Connect connects the the given MQTT instance.
func Connect(host, user, password string) (Client, error) {
	opts := mqtt.NewClientOptions()
	opts.Order = false
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.AddBroker(host)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return haClient{client}, nil
}

// PublishMap will publish a JSON encoded map to the given topic.
func (c haClient) PublishMap(topic string, data map[string]interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	token := c.Publish(topic, 0, true, j)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
