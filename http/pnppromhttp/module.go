package pnppromhttp

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type MetricsHandler http.HandlerFunc

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewMetricsHandler)
	moduleBuilder.Supply(options)
	moduleBuilder.ProvideIf(options.registerInMux, pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

func NewMetricsHandler(registry *prometheus.Registry) MetricsHandler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	).ServeHTTP
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Handler MetricsHandler

	Logger  *logging.Logger `optional:"true"`
	Options *options
}

func NewMuxHandlerRegistrar(params NewMuxHandlerRegistrarParams) pnphttpserver.MuxHandlerRegistrar {
	endpoint := params.Options.endpoint
	return pnphttpserver.MuxHandlerRegistrarFunc(func(router *mux.Router) {
		params.Logger.Named("promhttp").WithFields(map[string]interface{}{
			"path":   endpoint.path,
			"method": endpoint.method,
		}).Debug(context.Background(), "Registering metrics handler")
		router.Methods(endpoint.method).Path(endpoint.path).HandlerFunc(params.Handler)
	})
}
