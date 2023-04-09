package pnpprometheus

import (
	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: "METRICS_SERVER_",
	}))
	builder.Provide(NewServer)
	builder.Provide(NewPrometheusRegistry)
	builder.InvokeIf(options.start, RegisterServerStartHooks)

	return builder.Build()
}
