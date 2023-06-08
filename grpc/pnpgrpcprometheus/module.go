package pnpgrpcprometheus

import (
	"math"

	intPrometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		intPrometheus.NewServerMetrics,
		intPrometheus.NewClientMetrics,
		pnpprometheus.MetricsCollectorProvider(NewServerPrometheusCollector),
		pnpprometheus.MetricsCollectorProvider(NewClientPrometheusCollector),

		pnpgrpcserver.UnaryInterceptorProvider(NewLoggerUnaryServerInterceptorProvider),
		pnpgrpcserver.StreamInterceptorProvider(NewLoggerStreamServerInterceptorProvider),
		//NewLoggerUnaryClientInterceptorProvider,
		//NewLoggerStreamClientInterceptorProvider,
	)

	return builder.Build()
}

func NewServerPrometheusCollector(metrics *intPrometheus.ServerMetrics) prometheus.Collector {
	return metrics
}

func NewClientPrometheusCollector(metrics *intPrometheus.ClientMetrics) prometheus.Collector {
	return metrics
}

type NewServerPrometheusInterceptorParams struct {
	fx.In
	ServerMetrics *intPrometheus.ServerMetrics
}

func NewLoggerUnaryServerInterceptorProvider(params NewServerPrometheusInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
	return ordering.OrderedItem[grpc.UnaryServerInterceptor]{
		Order: math.MaxInt,
		Value: params.ServerMetrics.UnaryServerInterceptor(),
	}
}

func NewLoggerStreamServerInterceptorProvider(params NewServerPrometheusInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
	return ordering.OrderedItem[grpc.StreamServerInterceptor]{
		Order: math.MaxInt,
		Value: params.ServerMetrics.StreamServerInterceptor(),
	}
}

type NewClientPrometheusInterceptorParams struct {
	fx.In
	ClientMetrics *intPrometheus.ClientMetrics
}

func NewLoggerUnaryClientInterceptorProvider(params NewClientPrometheusInterceptorParams) ordering.OrderedItem[grpc.UnaryClientInterceptor] {
	return ordering.OrderedItem[grpc.UnaryClientInterceptor]{
		Order: math.MaxInt,
		Value: params.ClientMetrics.UnaryClientInterceptor(),
	}
}
func NewLoggerStreamClientInterceptorProvider(params NewClientPrometheusInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
	return ordering.OrderedItem[grpc.StreamClientInterceptor]{
		Order: math.MaxInt,
		Value: params.ClientMetrics.StreamClientInterceptor(),
	}
}
