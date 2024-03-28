package pnpgrpcserver

import (
	"context"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/ordering"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"

	"github.com/go-pnp/go-pnp/fxutil"
)

// Module provides *grpc.Server to fx container.
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		start:        true,
		configPrefix: "GRPC_",
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	for _, opt := range options.serverOptions {
		option := opt

		builder.Provide(ServerOptionProvider(func() grpc.ServerOption {
			return option
		}))
	}

	builder.Provide(NewGRPCServer)
	builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))
	builder.ProvideIf(options.reflection, ServiceRegistrarProvider(func() ServiceRegistrarFunc {
		return func(server *grpc.Server) {
			reflection.Register(server)
		}
	}))

	builder.InvokeIf(options.start, RegisterStartHooks)

	return builder.Build()
}

type ServiceRegistrar interface {
	Register(server *grpc.Server)
}

type ServiceRegistrarFunc func(server *grpc.Server)

func (f ServiceRegistrarFunc) Register(server *grpc.Server) {
	f(server)
}

func ServiceRegistrarProvider(target any) any {
	return fxutil.GroupProvider[ServiceRegistrar]("pnpgrpcserver.service_registrars", target)
}
func UnaryInterceptorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[grpc.UnaryServerInterceptor]]("pnpgrpcserver.unary_interceptors", target)
}
func StreamInterceptorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[grpc.StreamServerInterceptor]]("pnpgrpcserver.stream_interceptors", target)
}

func ServerOptionProvider(target any) any {
	return fxutil.GroupProvider[grpc.ServerOption]("pnpgrpcserver.server_options", target)
}

type NewGRPCServerParams struct {
	fx.In
	Logger             *logging.Logger                                     `optional:"true"`
	ServiceRegistrars  []ServiceRegistrar                                  `group:"pnpgrpcserver.service_registrars"`
	UnaryInterceptors  ordering.OrderedItems[grpc.UnaryServerInterceptor]  `group:"pnpgrpcserver.unary_interceptors"`
	StreamInterceptors ordering.OrderedItems[grpc.StreamServerInterceptor] `group:"pnpgrpcserver.stream_interceptors"`
	ServerOptions      []grpc.ServerOption                                 `group:"pnpgrpcserver.server_options"`
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
	if len(params.ServerOptions) > 0 {
		params.Logger.Debug(context.Background(), "Registered %d server options...", len(params.ServerOptions))
	}

	if len(params.UnaryInterceptors) > 0 {
		params.Logger.Debug(context.Background(), "Registered %d unary interceptors...", len(params.UnaryInterceptors))
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(params.UnaryInterceptors.Get()...))
	}

	if len(params.StreamInterceptors) > 0 {
		params.Logger.Debug(context.Background(), "Registered %d stream interceptors...", len(params.StreamInterceptors))
		serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(params.StreamInterceptors.Get()...))
	}

	server := grpc.NewServer(serverOptions...)

	params.Logger.Debug(context.Background(), "Calling %d service registrars...", len(params.ServiceRegistrars))
	for _, reg := range params.ServiceRegistrars {
		if reg != nil {
			reg.Register(server)
		}
	}

	return server, nil
}

type RegisterStartHooksParams struct {
	fx.In
	Shutdowner fx.Shutdowner
	Server     *grpc.Server
	Config     *Config
	Lc         fx.Lifecycle
	Logger     *logging.Logger `optional:"true"`
}

func RegisterStartHooks(params RegisterStartHooksParams) {
	logger := params.Logger.WithField("addr", params.Config.Addr)
	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info(ctx, "Starting gRPC server")
			listener, err := net.Listen("tcp", params.Config.Addr)
			if err != nil {
				return err
			}

			go func() {
				if err := params.Server.Serve(listener); err != nil {
					params.Logger.WithError(err).Error(ctx, "gRPC server serve error")
					params.Shutdowner.Shutdown()
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info(ctx, "Stopping gRPC server...")
			params.Server.GracefulStop()
			logger.Info(ctx, "Stopped gRPC server")
			return nil
		},
	})
}
