package pnpzap

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-pnp/go-pnp/logging"
)

type LoggingLogger struct {
	zapLogger *zap.Logger
}

var _ logging.Delegate = (*LoggingLogger)(nil)

func (l LoggingLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintf(msg, args...))
}

func (l LoggingLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintf(msg, args...))
}

func (l LoggingLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintf(msg, args...))
}

func (l LoggingLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.withContext(ctx).Error(fmt.Sprintf(msg, args...))
}

func (l LoggingLogger) WithFields(fields map[string]interface{}) logging.Delegate {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &LoggingLogger{
		zapLogger: l.zapLogger.With(zapFields...),
	}
}

func (l LoggingLogger) WithField(key string, value interface{}) logging.Delegate {
	return &LoggingLogger{
		zapLogger: l.zapLogger.With(zap.Any(key, value)),
	}
}

func (l LoggingLogger) WithError(err error) logging.Delegate {
	return &LoggingLogger{
		zapLogger: l.zapLogger.With(zap.Error(err)),
	}
}

func (l LoggingLogger) Named(component string) logging.Delegate {
	return &LoggingLogger{
		zapLogger: l.zapLogger.Named(component),
	}
}

func (l LoggingLogger) SkipCallers(count int) logging.Delegate {
	return &LoggingLogger{
		zapLogger: l.zapLogger.WithOptions(zap.AddCallerSkip(count)),
	}
}

func (l LoggingLogger) withContext(ctx context.Context) *zap.Logger {
	return l.zapLogger.With(zap.Field{
		Key:       "context",
		Type:      zapcore.SkipType,
		String:    "context",
		Interface: ctx,
	})
}

func NewLoggingLogger(logger *zap.Logger) logging.Delegate {
	return &LoggingLogger{
		zapLogger: logger,
	}
}
