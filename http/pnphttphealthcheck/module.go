package pnphttphealthcheck

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type HealthCheckHandler http.HandlerFunc

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		registerInMux: true,
	}, opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewHealthcheckHandler)
	moduleBuilder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](
		configutil.Options{Prefix: "HTTP_HEALTHCHECK_"},
	))
	moduleBuilder.ProvideIf(options.registerInMux, pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

// TODO: add possibility to add checkers
func NewHealthcheckHandler() HealthCheckHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Config  *Config
	Handler HealthCheckHandler
	Logger  *logging.Logger `optional:"true"`
}

func NewMuxHandlerRegistrar(params NewMuxHandlerRegistrarParams) pnphttpserver.MuxHandlerRegistrar {
	return func(mux *mux.Router) {
		params.Logger.Named("http healthcheck module").Info(context.Background(), "Registering healthcheck handler")
		mux.Methods(params.Config.Method).Path(params.Config.Path).HandlerFunc(params.Handler)
	}
}
