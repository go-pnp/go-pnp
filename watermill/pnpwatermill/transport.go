package pnpwatermill

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/v2/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	wsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SubscriberConfig struct {
	GcloudPubSubHandlerSubscriberConfigOption []optionutil.Option[googlecloud.SubscriberConfig]
	RedisHandlerSubscriberConfigOption        []optionutil.Option[redisstream.SubscriberConfig]
	SQLHandlerSubscriberConfigOption          []optionutil.Option[wsql.SubscriberConfig]
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
	Options     *options
	Config      *Config
	RedisClient redis.UniversalClient `optional:"true"`
	SQLDB       *sql.DB               `optional:"true"`

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
	case TransportChannel:
		publisher, subscriberFactory, err = newChannelTransport()
	case TransportGCloudPubSub:
		publisher, subscriberFactory, err = newGCloudPubSubTransport(newGCloudPubSubTransportParams{
			Config:                 params.Config,
			PublisherConfigOptions: params.GCloudPubSubPublisherConfigOptions,
		})
	case TransportRedis:
		publisher, subscriberFactory, err = newRedisTransport(params.RedisClient, params.Config)
	case TransportSQL:
		publisher, subscriberFactory, err = newSQLTransport(params.SQLDB, params.Config, params.Options)
	default:
		return nil, nil, fmt.Errorf("unsupported watermill transport: '%d'", params.Config.Transport)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("create transport: %w", err)
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
			GenerateSubscriptionName: func(topic string) string {
				return params.Config.GCloudPubSub.SubscriptionNamePrefix + topic + "_" + handler
			},
			GenerateSubscription: func(params googlecloud.GenerateSubscriptionParams) *pubsubpb.Subscription {
				return &pubsubpb.Subscription{
					EnableMessageOrdering: true,
					RetryPolicy: &pubsubpb.RetryPolicy{
						MinimumBackoff: durationpb.New(time.Second),
						MaximumBackoff: durationpb.New(time.Minute),
					},
				}
			},
			Unmarshaler: NewGCloudPubSubUnmarshaler(googlecloud.DefaultMarshalerUnmarshaler{}),
		}, subConfig.GcloudPubSubHandlerSubscriberConfigOption...)

		subscriber, err := googlecloud.NewSubscriber(*subscriberConfig, watermill.NopLogger{})
		if err != nil {
			return nil, fmt.Errorf("create subscriber: %w", err)
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

func newSQLTransport(db *sql.DB, config *Config, opts *options) (message.Publisher, subscriberFactory, error) {
	if db == nil {
		return nil, nil, fmt.Errorf("sql.DB not provided, please add pnpsql.Module() or pnppgx.Module() to your fx application")
	}

	if config.SQL.ConsumerGroup == "" {
		return nil, nil, fmt.Errorf("%sSQL_CONSUMER_GROUP can't be empty when using 'sql' transport", opts.configPrefix)
	}

	schemaAdapter, offsetsAdapter, err := sqlAdaptersForDriver(config.SQL.Driver)
	if err != nil {
		return nil, nil, err
	}

	publisher, err := wsql.NewPublisher(
		wsql.BeginnerFromStdSQL(db),
		wsql.PublisherConfig{
			SchemaAdapter:        schemaAdapter,
			AutoInitializeSchema: true,
		},
		watermill.NopLogger{},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create sql publisher: %w", err)
	}

	return publisher, subscriberFactoryFn(func(handler string, subscriberConfig *SubscriberConfig) (message.Subscriber, error) {
		consumerGroup := config.SQL.ConsumerGroup + "_" + handler

		wsqlSubscriberConfig := optionutil.ApplyOptions(&wsql.SubscriberConfig{
			ConsumerGroup:    consumerGroup,
			SchemaAdapter:    schemaAdapter,
			OffsetsAdapter:   offsetsAdapter,
			InitializeSchema: true,
		}, subscriberConfig.SQLHandlerSubscriberConfigOption...)

		subscriber, err := wsql.NewSubscriber(
			wsql.BeginnerFromStdSQL(db),
			*wsqlSubscriberConfig,
			watermill.NopLogger{},
		)
		if err != nil {
			return nil, fmt.Errorf("create sql subscriber: %w", err)
		}

		return subscriber, nil
	}), nil
}

func sqlAdaptersForDriver(driver string) (wsql.SchemaAdapter, wsql.OffsetsAdapter, error) {
	switch driver {
	case "postgres":
		return wsql.DefaultPostgreSQLSchema{}, wsql.DefaultPostgreSQLOffsetsAdapter{}, nil
	case "mysql":
		return wsql.DefaultMySQLSchema{}, wsql.DefaultMySQLOffsetsAdapter{}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported SQL driver for watermill: '%s', supported: postgres, mysql", driver)
	}
}
