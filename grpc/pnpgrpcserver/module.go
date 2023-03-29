package pnpgrpcserver

import (
	"context"
	"net"

	"github.com/caarlos0/env/v6"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"

	"github.com/go-pnp/go-pnp/fxutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		start: true,
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		configutil.NewConfigProvider[Config](env.Options{
			Prefix: "GRPC_",
		}),
		NewGRPCServer,
	)

	if len(options.serverOptions) > 0 {
		for _, serverOption := range options.serverOptions {
			fxutil.OptionsBuilderGroupSupply(builder, "pnpgrpcserver.server_options", serverOption)
		}

	}

	builder.InvokeIf(options.start, RegisterStartHooks)

	return builder.Build()

}

type ServiceRegistrar func(server *grpc.Server)

func ServiceRegistrarProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[ServiceRegistrar]("pnpgrpcserver.service_registrars", target)
}
func UnaryInterceptorProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[grpc.UnaryServerInterceptor]("pnpgrpcserver.unary_interceptors", target)
}
func StreamInterceptorProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[grpc.StreamServerInterceptor]("pnpgrpcserver.stream_interceptors", target)
}
func ServerOptionProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[grpc.ServerOption]("pnpgrpcserver.server_options", target)
}

type NewGRPCServerParams struct {
	fx.In
	ServiceRegistrars  []ServiceRegistrar             `group:"pnpgrpcserver.service_registrars"`
	UnaryInterceptors  []grpc.UnaryServerInterceptor  `group:"pnpgrpcserver.unary_interceptors"`
	StreamInterceptors []grpc.StreamServerInterceptor `group:"pnpgrpcserver.stream_interceptors"`
	ServerOptions      []grpc.ServerOption            `group:"pnpgrpcserver.server_options"`
	Config             *Config
}

func NewGRPCServer(params NewGRPCServerParams) (*grpc.Server, error) {
	tlsConfig, err := params.Config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	grpcCredentials := credentials.NewTLS(tlsConfig)
	if tlsConfig == nil {
		grpcCredentials = insecure.NewCredentials()
	}

	serverOptions := append([]grpc.ServerOption{grpc.Creds(grpcCredentials)}, params.ServerOptions...)

	if len(params.UnaryInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(params.UnaryInterceptors...))
	}

	if len(params.StreamInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(params.StreamInterceptors...))
	}

	server := grpc.NewServer(serverOptions...)
	for _, reg := range params.ServiceRegistrars {
		reg(server)
	}

	return server, nil
}

func RegisterStartHooks(
	lc fx.Lifecycle,
	runtimeErr chan error,
	server *grpc.Server,
	config *Config,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", config.Addr)
			if err != nil {
				return err
			}

			go func() {
				if err := server.Serve(listener); err != nil {
					runtimeErr <- err
				}
			}()

			return nil
		},
	})
}
