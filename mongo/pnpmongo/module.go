package pnpmongo

import (
	"context"

	"github.com/go-pnp/go-pnp/config/configutil"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"

	mongoopts "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	moduleOptions := newOptions(opts)

	builder := fxutil.OptionsBuilder{}
	builder.Provide(moduleOptions)
	builder.ProvideIf(!moduleOptions.configFromContainer, configutil.NewPrefixedConfigProvider[Config](moduleOptions.configPrefix))
	builder.PublicProvideIf(!moduleOptions.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](moduleOptions.configPrefix))
	builder.Provide(fx.Annotate(NewMongoClient, fx.OnStop(DisconnectClient)))

	return builder.Build()
}

func NewMongoClient(config *Config) (*mongo.Client, error) {
	return mongo.Connect(mongoopts.Client().ApplyURI(config.DSN))
}

func DisconnectClient(ctx context.Context, client *mongo.Client) error {
	return client.Disconnect(ctx)
}
