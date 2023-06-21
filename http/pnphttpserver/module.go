package pnphttpserver

import (
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		start:        true,
		provideMux:   true,
		configPrefix: "HTTP_SERVER_",
	}, opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewServer)
	moduleBuilder.ProvideIf(options.provideMux, NewMux)
	moduleBuilder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](
		configutil.Options{Prefix: options.configPrefix},
	))

	moduleBuilder.InvokeIf(options.start, RegisterStartHooks)

	return moduleBuilder.Build()
}
