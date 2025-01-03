package pnpredis

import (
	"errors"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	builder := fxutil.OptionsBuilder{}
	builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))
	builder.Provide(fx.Annotate(NewRedisClient, fx.OnStop(CloseClient)))

	return builder.Build()
}

func NewRedisClient(config *Config) (redis.UniversalClient, error) {
	tls, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	if config.Sentinel.Enabled {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       config.Sentinel.Master,
			SentinelAddrs:    config.Sentinel.Addrs,
			SentinelUsername: config.Sentinel.Username,
			SentinelPassword: config.Sentinel.Password,
			Username:         config.Username,
			Password:         config.Password,
			DB:               config.DB,
		}), nil
	}

	return redis.NewClient(&redis.Options{
		Addr:      config.Address,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: tls,
	}), nil
}

func CloseClient(client redis.UniversalClient) error {
	if err := client.Close(); err != nil && !errors.Is(redis.ErrClosed, err) {
		return err
	}

	return nil
}
