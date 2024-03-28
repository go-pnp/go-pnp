package pnpgrpcopentelemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	fxutil.OptionsBuilderSupply(builder, options)
	builder.Provide(
		pnpgrpcopentelemetry.ServerHandlerOptionsProvider(NewOpenTelemetryStatsHandlerServerOption),
		pnpgrpcopentelemetry.ClientHandlerOptionsProvider(NewOpenTelemetryStatsHandlerClientOption),
	)

	return builder.Build()
}

type NewTracerInterceptorParams struct {
	fx.In
	TraceProvider trace.TracerProvider
	Options       *options
}

func NewOpenTelemetryStatsHandlerServerOption(params NewTracerInterceptorParams) grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(params.TraceProvider),
	))
}

func NewOpenTelemetryStatsHandlerClientOption(params NewTracerInterceptorParams) grpc.DialOption {
	return grpc.WithStatsHandler(otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(params.TraceProvider),
	))
}
