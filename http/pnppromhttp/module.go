package pnppromhttp

import (
	"context"
	"io"
	"net/http"

	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
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
	moduleBuilder.Provide(NewHealthcheckHandler)
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(NewMetricsCollector)
	moduleBuilder.Provide(pnpprometheus.MetricsCollectorProvider(newPrometheusCollector))
	moduleBuilder.ProvideIf(options.registerInMux, pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

func newPrometheusCollector(metricsCollector *MetricsCollector) prometheus.Collector {
	return metricsCollector
}

func NewHealthcheckHandler(registry *prometheus.Registry) MetricsHandler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	).ServeHTTP
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Handler          MetricsHandler
	MetricsCollector *MetricsCollector

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
		router.Use(func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				path, err := mux.CurrentRoute(request).GetPathTemplate()
				if err != nil {
					path = "<unknown>"
				}
				requestObserver := params.MetricsCollector.trackRequest(request.Method, path)

				responseWriter := &httpResponseWriterTracker{ResponseWriter: writer}
				bodySizeTracker := &requestBodyReaderTracker{}
				request.Body = bodySizeTracker
				handler.ServeHTTP(responseWriter, request)

				requestObserver.Observe(bodySizeTracker.size, responseWriter.bodySize, responseWriter.status)

			})
		})
	})
}

type httpResponseWriterTracker struct {
	http.ResponseWriter
	status   int
	bodySize int
}

func (w *httpResponseWriterTracker) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *httpResponseWriterTracker) Write(b []byte) (int, error) {
	w.bodySize += len(b)
	return w.ResponseWriter.Write(b)
}

type requestBodyReaderTracker struct {
	io.ReadCloser
	size int
}

func (n requestBodyReaderTracker) Read(p []byte) (int, error) {
	size, err := n.ReadCloser.Read(p)
	n.size += size

	return size, err
}
