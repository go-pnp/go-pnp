package pnphttpserver

import (
	"github.com/caarlos0/env/v6"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		start:      true,
		provideMux: true,
	}, opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewServer)
	moduleBuilder.ProvideIf(options.provideMux, NewMux)
	moduleBuilder.InvokeIf(options.start, RegisterStartHooks)

	if options.config == nil {
		moduleBuilder.Provide(configutil.NewConfigProvider[Config](
			env.Options{Prefix: "HTTP_SERVER_"},
		))
	} else {
		fxutil.OptionsBuilderSupply(moduleBuilder, options.config)
	}

	return moduleBuilder.Build()
}
