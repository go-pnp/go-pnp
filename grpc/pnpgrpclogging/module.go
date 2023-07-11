package pnpgrpclogging

import (
	"context"

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
		pnpgrpcserver.UnaryInterceptorProvider(NewLoggerUnaryServerInterceptorProvider(options)),
		pnpgrpcserver.StreamInterceptorProvider(NewLoggerStreamServerInterceptorProvider(options)),
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
	for iter.Next() {
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

func LoggingOptionProvider(target any) any {
	return fxutil.GroupProvider[intLogging.Option](
		"pnpgrpclogging.option",
		target,
	)
}

type NewLoggerInterceptorParams struct {
	fx.In
	Logger  *logging.Logger     `optional:"true"`
	Options []intLogging.Option `group:"pnpgrpclogging.option"`
}

func NewLoggerUnaryServerInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryServerInterceptor] {
		return ordering.OrderedItem[grpc.UnaryServerInterceptor]{
			Order: opts.getServerOrder(),
			Value: intLogging.UnaryServerInterceptor(InterceptorLogger{params.Logger}, params.Options...),
		}
	}
}

func NewLoggerStreamServerInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamServerInterceptor] {
		return ordering.OrderedItem[grpc.StreamServerInterceptor]{
			Order: opts.getServerOrder(),
			Value: intLogging.StreamServerInterceptor(InterceptorLogger{params.Logger}, params.Options...),
		}
	}
}

func NewLoggerUnaryClientInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryClientInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.UnaryClientInterceptor] {
		return ordering.OrderedItem[grpc.UnaryClientInterceptor]{
			Order: opts.getClientOrder(),
			Value: intLogging.UnaryClientInterceptor(InterceptorLogger{params.Logger}, params.Options...),
		}
	}
}
func NewLoggerStreamClientInterceptorProvider(opts *options) func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
	return func(params NewLoggerInterceptorParams) ordering.OrderedItem[grpc.StreamClientInterceptor] {
		return ordering.OrderedItem[grpc.StreamClientInterceptor]{
			Order: opts.getClientOrder(),
			Value: intLogging.StreamClientInterceptor(InterceptorLogger{params.Logger}, params.Options...),
		}
	}
}
