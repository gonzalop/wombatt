package mqttha

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gonzalop/mq"
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

// PublishMap will publish a JSON encoded map to the given topic.
func (c *Client) PublishMap(topic string, retain bool, data map[string]any) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	token := c.client.Publish(topic, j, mq.WithRetain(retain), mq.WithAlias())
	err = token.Wait(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Disconnect(ms time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), ms)
	defer cancel()
	_ = c.client.Disconnect(ctx)
}
