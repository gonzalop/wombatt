package mqttha

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gonzalop/mq"
)

const (
	NoRetain     = false
	Retain       = true
	NoTopicAlias = false
	TopicAlias   = true
)

type Client struct {
	client *mq.Client
}

// Connect connects to the given MQTT instance.
func Connect(host, user, password string) (*Client, error) {
	client, err := mq.Dial(host,
		mq.WithCredentials(user, password),
		mq.WithProtocolVersion(mq.ProtocolV50),
		mq.WithKeepAlive(30*time.Second),
		mq.WithTopicAliasMaximum(400),
		mq.WithLogger(slog.Default()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

// haAliases contains the short abbreviations for Home Assistant discovery keys.
// This map only includes the keys that are actually used by Wombatt.
var haAliases = map[string]string{
	"default_entity_id":           "def_ent_id",
	"device":                      "dev",
	"device_class":                "dev_cla",
	"icon":                        "ic",
	"identifiers":                 "ids",
	"model":                       "mdl",
	"state_class":                 "stat_cla",
	"state_topic":                 "stat_t",
	"suggested_display_precision": "sug_dsp_prc",
	"unique_id":                   "uniq_id",
	"unit_of_measurement":         "unit_of_meas",
	"value_template":              "val_tpl",
}

// compactMap recursively replaces long-form HA keys with their abbreviations.
func compactMap(data map[string]any) map[string]any {
	compact := make(map[string]any, len(data))

	for k, v := range data {
		key := k
		if alias, ok := haAliases[k]; ok {
			key = alias
		}

		switch val := v.(type) {
		case map[string]any:
			// Recurse into nested maps (like the "device" object)
			compact[key] = compactMap(val)
		default:
			compact[key] = val
		}
	}

	return compact
}

// PublishMap will publish a JSON encoded map to the given topic.
func (c *Client) PublishMap(topic string, data map[string]any, retain bool, useTopicAlias bool) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	opts := []mq.PublishOption{
		mq.WithRetain(retain),
	}
	if useTopicAlias {
		opts = append(opts, mq.WithAlias())
	}

	token := c.client.Publish(topic, j, opts...)
	return token.Wait(context.Background())
}

// PublishDiscovery publishes a JSON encoded map to the given topic, compacting the keys.
func (c *Client) PublishDiscovery(topic string, data map[string]any) error {
	compactData := compactMap(data)
	return c.PublishMap(topic, compactData, Retain, NoTopicAlias)
}

func (c *Client) Disconnect(ms time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), ms)
	defer cancel()
	_ = c.client.Disconnect(ctx)
}
