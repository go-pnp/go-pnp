package pnppromhttp

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type MetricsHandler http.HandlerFunc

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		registerInMux: true,
	}, opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewHealthcheckHandler)
	moduleBuilder.ProvideIf(
		!options.configFromContainer,
		configutil.NewPrefixedConfigProvider[Config]("PROMETHEUS_HANDLER_"),
	)
	moduleBuilder.ProvideIf(options.registerInMux, pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

func NewHealthcheckHandler(registry *prometheus.Registry) MetricsHandler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	).ServeHTTP
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Config  *Config
	Handler MetricsHandler
	Logger  *logging.Logger `optional:"true"`
}

func NewMuxHandlerRegistrar(params NewMuxHandlerRegistrarParams) pnphttpserver.MuxHandlerRegistrar {
	return func(mux *mux.Router) {
		params.Logger.Named("promhttp module").Info(context.Background(), "Registering metrics")
		mux.Methods(params.Config.Method).Path(params.Config.Path).HandlerFunc(params.Handler)
	}
}
