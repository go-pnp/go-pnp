package pnpgrpcopentelemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		pnpgrpcserver.UnaryInterceptorProvider(NewOpenTelemetryUnaryServerInterceptorProvider(options)),
		pnpgrpcserver.StreamInterceptorProvider(NewOpenTelemetryStreamServerInterceptorProvider(options)),

		//NewOpenTelemetryUnaryClientInterceptorProvider,
		//NewOpenTelemetryStreamClientInterceptorProvider,
	)

	return builder.Build()
}

type NewTracerInterceptorParams struct {
	fx.In
	TraceProvider *trace.TracerProvider
}

func NewOpenTelemetryUnaryServerInterceptorProvider(opts *options) func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
	return func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
		return ordering.OrderedItem[grpc.UnaryServerInterceptor]{
			Order: opts.getServerOrder(),
			Value: otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(params.TraceProvider)),
		}
	}
}
func NewOpenTelemetryStreamServerInterceptorProvider(opts *options) func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
	return func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
		return ordering.OrderedItem[grpc.StreamServerInterceptor]{
			Order: opts.getServerOrder(),
			Value: otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(params.TraceProvider)),
		}
	}
}

func NewOpenTelemetryUnaryClientInterceptorProvider(opts *options) func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
	return func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {

		return ordering.OrderedItem[grpc.StreamClientInterceptor]{
			Order: opts.getClientOrder(),
			Value: otelgrpc.StreamClientInterceptor(otelgrpc.WithTracerProvider(params.TraceProvider)),
		}
	}
}

func NewOpenTelemetryStreamClientInterceptorProvider(opts *options) func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
	return func(params NewTracerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
		return ordering.OrderedItem[grpc.StreamClientInterceptor]{
			Order: opts.getClientOrder(),
			Value: otelgrpc.StreamClientInterceptor(otelgrpc.WithTracerProvider(params.TraceProvider)),
		}
	}
}
