package pnpgrpcrecovery

import (
	"context"

	intLogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		pnpgrpcserver.UnaryInterceptorProvider(NewRecoveryUnaryInterceptorProvider(options)),
		pnpgrpcserver.StreamInterceptorProvider(NewRecoveryStreamInterceptorProvider(options)),
	)

	return builder.Build()
}

type NewLoggerInterceptorParams struct {
	fx.In
	Logger  *logging.Logger     `optional:"true"`
	Options []intLogging.Option `group:"pnpgrpclogging.option"`
}

func NewRecoveryUnaryInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
		return ordering.OrderedItem[grpc.UnaryServerInterceptor]{
			Order: opts.order,
			Value: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				defer func() {
					if r := recover(); r != nil {
						params.Logger.WithField("panic", r).Error(ctx, "recovered from panic")
						err = status.Error(codes.Internal, "internal server error")
					}
				}()

				return handler(ctx, req)
			},
		}
	}
}

func NewRecoveryStreamInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
		return ordering.OrderedItem[grpc.StreamServerInterceptor]{
			Order: opts.order,
			Value: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				defer func() {
					if r := recover(); r != nil {
						params.Logger.WithField("panic", r).Error(ss.Context(), "recovered from panic")
						err = status.Error(codes.Internal, "internal server error")
					}
				}()

				return handler(srv, ss)
			},
		}
	}
}
