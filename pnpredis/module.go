package pnpredis

import (
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

func NewRedisOptions(config *Config) (*redis.Options, error) {
	tls, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	return &redis.Options{
		Addr:      config.Address,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: tls,
	}, nil
}

func NewRedisClient(options *redis.Options) (*redis.Client, error) {
	return redis.NewClient(options), nil
}

func CloseClient(client *redis.Client) error {
	return client.Close()
}
