package pnpredis

import (
	"context"

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
			NewRedisClient,
		),
	)
}

func NewRedisClient(lc fx.Lifecycle, config *Config) (*redis.Client, error) {
	tls, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:      config.Address,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: tls,
	})

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}
