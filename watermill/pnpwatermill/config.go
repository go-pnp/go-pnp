package pnpwatermill

type Config struct {
	Transport string `env:"TRANSPORT,notEmpty"`
	Redis     struct {
		// When empty, fan-out mode will be used.
		ConsumerGroup string `env:"CONSUMER_GROUP"`
	} `envPrefix:"REDIS_"`
	GCloudPubSub struct {
		ProjectID              string `env:"PROJECT_ID"`
		SubscriptionNamePrefix string `env:"SUBSCRIPTION_NAME_PREFIX"`
	} `envPrefix:"GCLOUD_PUB_SUB_"`
	TopicPrefix string `env:"TOPICS_PREFIX"`
}
