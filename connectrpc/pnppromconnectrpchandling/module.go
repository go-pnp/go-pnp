package pnppromconnectrpchandling

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

// Module provides an interface to easily register connectrpc handlers in a http server mux.
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		subsystem: "connectrpc",
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Supply(options)
	builder.Provide(NewInterceptor)
	builder.PublicProvide(pnpprometheus.MetricsCollectorProvider(newPrometheusCollector))
	builder.Provide(pnpconnectrpchandling.InterceptorProvider(newConnectInterceptor))

	return builder.Build()
}

func newPrometheusCollector(interceptor *Interceptor) prometheus.Collector {
	return interceptor
}

func newConnectInterceptor(interceptor *Interceptor, options *options) ordering.OrderedItem[connect.Interceptor] {
	return ordering.OrderedItem[connect.Interceptor]{
		Value: interceptor,
		Order: options.order,
	}
}

type Interceptor struct {
	unaryRequestsPanicsCount *prometheus.CounterVec
	unaryRequestsCount       *prometheus.CounterVec
	unaryRequestsDuration    *prometheus.HistogramVec
	streamHandlesPanicsCount *prometheus.CounterVec
	streamHandlesCount       *prometheus.CounterVec
	streamHandlesDuration    *prometheus.HistogramVec
}

func NewInterceptor(options *options) *Interceptor {
	return &Interceptor{
		unaryRequestsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "unary_requests_count",
				Help:        "Number of unary requests.",
				ConstLabels: options.constLabels,
			}, []string{"procedure", "result"},
		),
		unaryRequestsPanicsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "unary_requests_panics_count",
				Help:        "Number of unary requests that panicked.",
				ConstLabels: options.constLabels,
			}, []string{"procedure"},
		),
		unaryRequestsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "unary_requests_duration",
				Help:        "Duration of unary requests.",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: options.constLabels,
			}, []string{"procedure"},
		),

		streamHandlesCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "stream_handles_count",
				Help:        "Number of stream handles.",
				ConstLabels: options.constLabels,
			}, []string{"procedure", "result"},
		),
		streamHandlesPanicsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "stream_handles_panics_count",
				Help:        "Number of stream handles that panicked.",
				ConstLabels: options.constLabels,
			}, []string{"procedure"},
		),
		streamHandlesDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "stream_handles_duration",
				Help:        "Duration of stream handles.",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: options.constLabels,
			}, []string{"procedure"},
		),
	}
}

func (i Interceptor) Describe(descs chan<- *prometheus.Desc) {
	i.unaryRequestsPanicsCount.Describe(descs)
	i.unaryRequestsCount.Describe(descs)
	i.unaryRequestsDuration.Describe(descs)

	i.streamHandlesPanicsCount.Describe(descs)
	i.streamHandlesCount.Describe(descs)
	i.streamHandlesDuration.Describe(descs)
}

func (i Interceptor) Collect(metrics chan<- prometheus.Metric) {
	i.unaryRequestsPanicsCount.Collect(metrics)
	i.unaryRequestsCount.Collect(metrics)
	i.unaryRequestsDuration.Collect(metrics)

	i.streamHandlesPanicsCount.Collect(metrics)
	i.streamHandlesCount.Collect(metrics)
	i.streamHandlesDuration.Collect(metrics)
}

func (i Interceptor) WrapUnary(unaryFunc connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				i.unaryRequestsPanicsCount.WithLabelValues(request.Spec().Procedure).Inc()
				panic(panicValue)
			}
		}()

		procedure := request.Spec().Procedure
		timer := prometheus.NewTimer(
			i.unaryRequestsDuration.WithLabelValues(procedure),
		)

		resp, err := unaryFunc(ctx, request)
		if err != nil {
			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				i.unaryRequestsCount.WithLabelValues(procedure, connectErr.Code().String()).Inc()
			} else {
				i.unaryRequestsCount.WithLabelValues(procedure, "internal_error").Inc()
			}
			return nil, err
		} else {
			i.unaryRequestsCount.WithLabelValues(procedure, "ok").Inc()
		}

		timer.ObserveDuration()

		return resp, nil
	}
}

func (i Interceptor) WrapStreamingClient(clientFunc connect.StreamingClientFunc) connect.StreamingClientFunc {
	return clientFunc
}

func (i Interceptor) WrapStreamingHandler(handlerFunc connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		procedure := conn.Spec().Procedure

		defer func() {
			if panicValue := recover(); panicValue != nil {
				i.streamHandlesPanicsCount.WithLabelValues(procedure).Inc()
				panic(panicValue)
			}
		}()

		timer := prometheus.NewTimer(
			i.streamHandlesDuration.WithLabelValues(procedure),
		)

		err := handlerFunc(ctx, conn)
		if err != nil {
			var connectErr connect.Error
			if errors.As(err, &connectErr) {
				i.streamHandlesCount.WithLabelValues(procedure, connectErr.Code().String()).Inc()
			} else {
				i.streamHandlesCount.WithLabelValues(procedure, "internal_error").Inc()
			}
			return err
		} else {
			i.streamHandlesCount.WithLabelValues(procedure, "ok").Inc()
		}

		timer.ObserveDuration()

		return nil
	}
}
