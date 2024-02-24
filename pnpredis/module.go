package pnpredis

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/logging"
)

func Module() fx.Option {
	return fx.Module(
		"redis",
		logging.DecorateNamed("redis_client"),
		fx.Provide(
			configutil.NewPrefixedConfigProvider[Config]("REDIS_"),
			fx.Annotate(
				NewRedisClient,
				fx.OnStop(CloseClient),
			),
		),
	)
}

func NewRedisClient(config *Config) (*redis.Client, error) {
	tls, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	return redis.NewClient(&redis.Options{
		Addr:      config.Address,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: tls,
	}), nil
}

func CloseClient(client *redis.Client) error {
	return client.Close()
}
