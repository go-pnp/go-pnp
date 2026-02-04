package pnpredis

import (
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
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
