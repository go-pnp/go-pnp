package pnpwatermill

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type SubscriberConfig struct {
	GcloudPubSubHandlerSubscriberConfigOption []optionutil.Option[googlecloud.SubscriberConfig]
	RedisHandlerSubscriberConfigOption        []optionutil.Option[redisstream.SubscriberConfig]
}

type subscriberFactory interface {
	NewSubscriber(handler string, config *SubscriberConfig) (message.Subscriber, error)
}
type subscriberFactoryFn func(handler string, config *SubscriberConfig) (message.Subscriber, error)

func (f subscriberFactoryFn) NewSubscriber(handler string, config *SubscriberConfig) (message.Subscriber, error) {
	return f(handler, config)
}

type GCloudPubSubPublisherConfigOption = optionutil.Option[googlecloud.PublisherConfig]

func GCloudPubSubPublisherConfigOptionProvider(target any) any {
	return fxutil.GroupProvider[GCloudPubSubPublisherConfigOption]("pnpwatermill.gcloud_pub_sub_publisher_config_options", target)
}

type newTransportParams struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Config      *Config
	RedisClient redis.UniversalClient `optional:"true"`

	SubscriberDecorators               ordering.OrderedItems[SubscriberDecorator] `group:"pnpwatermill.subscriber_decorators"`
	PublisherDecorators                ordering.OrderedItems[PublisherDecorator]  `group:"pnpwatermill.publisher_decorators"`
	GCloudPubSubPublisherConfigOptions []GCloudPubSubPublisherConfigOption        `group:"pnpwatermill.gcloud_pub_sub_publisher_config_options"`
}

// Waiting for fx.Evaluate https://github.com/uber-go/fx/issues/1132 :(
func NewTransport(params newTransportParams) (message.Publisher, subscriberFactory, error) {
	publisherTopicPrefixDecorator := func(pub message.Publisher) message.Publisher {
		return topicPrefixPublisherDecorator{
			publisher: pub,
			prefix:    params.Config.TopicPrefix,
		}
	}
	subscriberTopicPrefixDecorator := func(sub message.Subscriber) message.Subscriber {
		return topicPrefixSubscriberDecorator{
			subscriber: sub,
			prefix:     params.Config.TopicPrefix,
		}
	}
	decorateSubscriber := func(sub message.Subscriber) message.Subscriber {
		// reversing as we're wrapping and order is reversed, first wrapper should be on top
		for _, decorator := range lo.Reverse(params.SubscriberDecorators.Get()) {
			sub = decorator(sub)
		}

		return subscriberTopicPrefixDecorator(sub)
	}
	decoratePublisher := func(publisher message.Publisher) message.Publisher {
		// reversing as we're wrapping and order is reversed, first wrapper should be on top
		for _, decorator := range lo.Reverse(params.PublisherDecorators.Get()) {
			publisher = decorator(publisher)
		}

		return publisherTopicPrefixDecorator(publisher)
	}
	var publisher message.Publisher
	var subscriberFactory subscriberFactory
	var err error

	switch params.Config.Transport {
	case "channel":
		publisher, subscriberFactory, err = newChannelTransport()
	case "gcloudpubsub":
		publisher, subscriberFactory, err = newGCloudPubSubTransport(newGCloudPubSubTransportParams{
			Config:                 params.Config,
			PublisherConfigOptions: params.GCloudPubSubPublisherConfigOptions,
		})
	case "redis":
		publisher, subscriberFactory, err = newRedisTransport(params.RedisClient, params.Config)
	default:
		return nil, nil, fmt.Errorf("unsupported watermill transport: '%s'", params.Config.Transport)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "create transport")
	}

	params.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return publisher.Close()
		},
	})

	return decoratePublisher(publisher), subscriberFactoryFn(func(handler string, config *SubscriberConfig) (message.Subscriber, error) {
		subscriber, err := subscriberFactory.NewSubscriber(handler, config)
		if err != nil {
			return nil, err
		}

		params.Lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return subscriber.Close()
			},
		})

		return decorateSubscriber(subscriber), nil
	}), nil
}

func newChannelTransport() (message.Publisher, subscriberFactory, error) {
	result := gochannel.NewGoChannel(
		gochannel.Config{
			Persistent: true,
		},
		watermill.NopLogger{},
	)

	return result, subscriberFactoryFn(func(handler string, config *SubscriberConfig) (message.Subscriber, error) {
		return result, nil
	}), nil
}

type newGCloudPubSubTransportParams struct {
	Config *Config

	PublisherConfigOptions []optionutil.Option[googlecloud.PublisherConfig]
}

func newGCloudPubSubTransport(params newGCloudPubSubTransportParams) (message.Publisher, subscriberFactory, error) {
	publisherConfig := optionutil.ApplyOptions(&googlecloud.PublisherConfig{
		ProjectID: params.Config.GCloudPubSub.ProjectID,
	}, params.PublisherConfigOptions...)
	result, err := googlecloud.NewPublisher(*publisherConfig, watermill.NopLogger{})
	if err != nil {
		return nil, nil, fmt.Errorf("create gcloud pubsub publisher: %w", err)
	}

	return result, subscriberFactoryFn(func(handler string, subConfig *SubscriberConfig) (message.Subscriber, error) {
		subscriberConfig := optionutil.ApplyOptions(&googlecloud.SubscriberConfig{
			ProjectID: params.Config.GCloudPubSub.ProjectID,
			SubscriptionConfig: pubsub.SubscriptionConfig{
				EnableMessageOrdering: true,
				RetryPolicy: &pubsub.RetryPolicy{
					MinimumBackoff: time.Second,
					MaximumBackoff: time.Minute,
				},
			},
			GenerateSubscriptionName: func(topic string) string {
				return params.Config.GCloudPubSub.SubscriptionNamePrefix + topic + "_" + handler
			},
		}, subConfig.GcloudPubSubHandlerSubscriberConfigOption...)

		subscriber, err := googlecloud.NewSubscriber(*subscriberConfig, watermill.NopLogger{})
		if err != nil {
			return nil, fmt.Errorf("create subscriber")
		}

		return subscriber, nil
	}), nil
}

func newRedisTransport(client redis.UniversalClient, config *Config) (message.Publisher, subscriberFactory, error) {
	if client == nil {
		return nil, nil, fmt.Errorf("redis client not provided, please add pnpredis.Module() to your fx application")
	}

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: client,
		},
		watermill.NopLogger{},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create redis publisher: %w", err)
	}

	return publisher, subscriberFactoryFn(func(handler string, subscriberConfig *SubscriberConfig) (message.Subscriber, error) {
		redisSubscriberConfig := optionutil.ApplyOptions(&redisstream.SubscriberConfig{
			Client:        client,
			Consumer:      uuid.New().String(),
			ConsumerGroup: config.Redis.ConsumerGroup,
		}, subscriberConfig.RedisHandlerSubscriberConfigOption...)
		subscriber, err := redisstream.NewSubscriber(
			*redisSubscriberConfig,
			watermill.NopLogger{},
		)
		if err != nil {
			return nil, fmt.Errorf("create redis subscriber: %w", err)
		}

		return subscriber, nil
	}), nil
}
