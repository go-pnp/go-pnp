package pnphttpserver

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"go.uber.org/fx"
)

func Module(options ...ServerOption) fx.Option {
	srvOptions := &serverOptions{
		start:      true,
		provideMux: true,
	}
	for _, opt := range options {
		opt(srvOptions)
	}

	moduleBuilder := fxutil.ModuleBuilder{
		ModuleName: "httpserver",
		Options: []fx.Option{
			fx.Provide(NewServer),
		},
	}
	moduleBuilder.InvokeIf(srvOptions.start, RegisterStartHooks)
	moduleBuilder.ProvideIf(srvOptions.provideMux, NewMux)

	if srvOptions.config == nil {
		moduleBuilder.Provide(configutil.NewConfigProvider[Config](
			env.Options{Prefix: "HTTP_SERVER_"},
		))
	} else {
		moduleBuilder.Supply(srvOptions.config)
	}

	return moduleBuilder.Build()
}
