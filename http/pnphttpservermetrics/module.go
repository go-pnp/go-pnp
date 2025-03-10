package pnphttpservermetrics

import (
	"io"
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
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
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(NewMetricsCollector)
	moduleBuilder.PublicProvide(pnpprometheus.MetricsCollectorProvider(newPrometheusCollector))
	moduleBuilder.Provide(pnphttpserver.MuxMiddlewareFuncProvider(NewMiddleware))

	return moduleBuilder.Build()
}

func newPrometheusCollector(metricsCollector *MetricsCollector) prometheus.Collector {
	return metricsCollector
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	MetricsCollector *MetricsCollector

	Logger  *logging.Logger `optional:"true"`
	Options *options
}

func NewMiddleware(params NewMuxHandlerRegistrarParams) ordering.OrderedItem[mux.MiddlewareFunc] {
	return ordering.OrderedItem[mux.MiddlewareFunc]{
		Value: func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				currentRoute := mux.CurrentRoute(request)
				path := "<unknown>"
				if currentRoute != nil {
					if pathTemplate, err := currentRoute.GetPathTemplate(); err == nil {
						path = pathTemplate
					}
				}
				requestObserver := params.MetricsCollector.trackRequest(request.Method, path)

				responseWriter := &httpResponseWriterTracker{ResponseWriter: writer}
				bodySizeTracker := &requestBodyReaderTracker{ReadCloser: request.Body}
				request.Body = bodySizeTracker
				handler.ServeHTTP(responseWriter, request)

				requestObserver.Observe(bodySizeTracker.size, responseWriter.bodySize, responseWriter.status)

			})
		},
		Order: params.Options.order,
	}
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
