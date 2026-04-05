package pnpwatermill

import "errors"

const (
	TransportChannel Transport = iota + 1
	TransportRedis
	TransportGCloudPubSub
	TransportSQL
)

type Transport byte

func (t *Transport) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.New("transport cannot be empty")
	}

	switch string(text) {
	case "channel":
		*t = TransportChannel
	case "gcloudpubsub":
		*t = TransportGCloudPubSub
	case "redis":
		*t = TransportRedis
	case "sql":
		*t = TransportSQL
	default:
		return errors.New("unsupported transport: " + string(text))
	}

	return nil
}

type Config struct {
	Transport Transport `env:"TRANSPORT,notEmpty"`
	Redis     struct {
		// When empty, fan-out mode will be used.
		ConsumerGroup string `env:"CONSUMER_GROUP"`
	} `envPrefix:"REDIS_"`
	GCloudPubSub struct {
		ProjectID              string `env:"PROJECT_ID"`
		SubscriptionNamePrefix string `env:"SUBSCRIPTION_NAME_PREFIX"`
	} `envPrefix:"GCLOUD_PUB_SUB_"`
	SQL struct {
		// Driver is the SQL driver to determine schema adapter. Supported: "postgres", "mysql".
		Driver string `env:"DRIVER" envDefault:"postgres"`
		// ConsumerGroup is the consumer group name for the subscriber.
		ConsumerGroup string `env:"CONSUMER_GROUP"`
	} `envPrefix:"SQL_"`
	TopicPrefix string `env:"TOPICS_PREFIX"`
}
