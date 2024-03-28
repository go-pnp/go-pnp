package pnpgrpcopentelemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/grpc/pnpgrpcclient"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	fxutil.OptionsBuilderSupply(builder, options)
	builder.Provide(
		pnpgrpcserver.ServerOptionProvider(NewOpenTelemetryStatsHandlerServerOption),
		pnpgrpcclient.DialOptionProvider(NewOpenTelemetryStatsHandlerClientOption),
	)

	return builder.Build()
}

func ServerHandlerOptionsProvider(target any) any {
	return fxutil.GroupProvider[otelgrpc.Option](
		"pnpgrpcopentelemetry.server_handler_options",
		target,
	)
}
func ClientHandlerOptionsProvider(target any) any {
	return fxutil.GroupProvider[otelgrpc.Option](
		"pnpgrpcopentelemetry.server_handler_options",
		target,
	)
}

type NewOpenTelemetryStatsHandlerServerOptionParams struct {
	fx.In
	ServerHandlerOptions []otelgrpc.Option `group:"pnpgrpcopentelemetry.server_handler_options"`
	Options              *options
}

func NewOpenTelemetryStatsHandlerServerOption(params NewOpenTelemetryStatsHandlerServerOptionParams) grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler(
		params.ServerHandlerOptions...,
	))
}

type NewOpenTelemetryStatsHandlerClientOptionParams struct {
	fx.In
	ClientHandlerOptions []otelgrpc.Option `group:"pnpgrpcopentelemetry.client_handler_options"`
	Options              *options
}

func NewOpenTelemetryStatsHandlerClientOption(params NewOpenTelemetryStatsHandlerClientOptionParams) grpc.DialOption {
	return grpc.WithStatsHandler(otelgrpc.NewClientHandler(
		params.ClientHandlerOptions...,
	))
}
