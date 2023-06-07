package pnpgrpclogging

import (
	"context"
	"math"

	intLogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/fx"
	"google.golang.org/grpc"

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
		pnpgrpcserver.UnaryInterceptorProvider(NewLoggerUnaryServerInterceptorProvider),
		pnpgrpcserver.StreamInterceptorProvider(NewLoggerStreamServerInterceptorProvider),
		//NewLoggerUnaryClientInterceptorProvider,
		//NewLoggerStreamClientInterceptorProvider,
	)

	return builder.Build()
}

type InterceptorLogger struct {
	delegate *logging.Logger
}

func (i InterceptorLogger) Log(ctx context.Context, level intLogging.Level, msg string, fields ...any) {
	delegateFields := make(map[string]interface{})
	iter := intLogging.Fields(fields).Iterator()
	if iter.Next() {
		k, v := iter.At()
		delegateFields[k] = v
	}
	logger := i.delegate.WithFields(delegateFields).SkipCallers(1)

	switch level {
	case intLogging.LevelDebug:
		logger.Debug(ctx, msg)
	case intLogging.LevelInfo:
		logger.Info(ctx, msg)
	case intLogging.LevelWarn:
		logger.Warn(ctx, msg)
	case intLogging.LevelError:
		logger.Error(ctx, msg)
	default:
		logger.Error(ctx, "unknown logging level")
		logger.Info(ctx, msg)
	}
}

type NewLoggerInterceptorParams struct {
	fx.In
	Logger *logging.Logger `optional:"true"`
}

func NewLoggerUnaryServerInterceptorProvider(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
	return ordering.OrderedItem[grpc.UnaryServerInterceptor]{
		Order: math.MaxInt,
		Value: intLogging.UnaryServerInterceptor(InterceptorLogger{params.Logger}),
	}
}

func NewLoggerStreamServerInterceptorProvider(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
	return ordering.OrderedItem[grpc.StreamServerInterceptor]{
		Order: math.MaxInt,
		Value: intLogging.StreamServerInterceptor(InterceptorLogger{params.Logger}),
	}
}

func NewLoggerUnaryClientInterceptorProvider(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryClientInterceptor] {
	return ordering.OrderedItem[grpc.UnaryClientInterceptor]{
		Order: math.MaxInt,
		Value: intLogging.UnaryClientInterceptor(InterceptorLogger{params.Logger}),
	}
}
func NewLoggerStreamClientInterceptorProvider(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
	return ordering.OrderedItem[grpc.StreamClientInterceptor]{
		Order: math.MaxInt,
		Value: intLogging.StreamClientInterceptor(InterceptorLogger{params.Logger}),
	}
}
