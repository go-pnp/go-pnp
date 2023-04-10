package pnpzap

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-pnp/go-pnp/logging"
)

type LoggingDelegate struct {
	zapLogger *zap.Logger
}

var _ logging.Delegate = (*LoggingDelegate)(nil)

func (l LoggingDelegate) Info(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintf(msg, args...))
}

func (l LoggingDelegate) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintf(msg, args...))
}

func (l LoggingDelegate) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintf(msg, args...))
}

func (l LoggingDelegate) Error(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Error(fmt.Sprintf(msg, args...))
}

func (l LoggingDelegate) WithFields(fields map[string]interface{}) logging.Delegate {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &LoggingDelegate{
		zapLogger: l.zapLogger.With(zapFields...),
	}
}

func (l LoggingDelegate) WithField(key string, value interface{}) logging.Delegate {
	return &LoggingDelegate{
		zapLogger: l.zapLogger.With(zap.Any(key, value)),
	}
}

func (l LoggingDelegate) WithError(err error) logging.Delegate {
	return &LoggingDelegate{
		zapLogger: l.zapLogger.With(zap.Error(err)),
	}
}

func (l LoggingDelegate) Named(component string) logging.Delegate {
	return &LoggingDelegate{
		zapLogger: l.zapLogger.Named(component),
	}
}

func (l LoggingDelegate) SkipCallers(count int) logging.Delegate {
	return &LoggingDelegate{
		zapLogger: l.zapLogger.WithOptions(zap.AddCallerSkip(count)),
	}
}

func (l LoggingDelegate) withContext(ctx context.Context) *zap.Logger {
	return l.zapLogger.With(zap.Field{
		Key:       "context",
		Type:      zapcore.SkipType,
		String:    "context",
		Interface: ctx,
	})
}

func NewLoggingLogger(logger *zap.Logger) *logging.Logger {
	return &logging.Logger{
		Delegate: &LoggingDelegate{
			zapLogger: logger,
		},
	}
}
