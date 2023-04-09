package logging

import "context"

type Delegate interface {
	Info(ctx context.Context, msg string, args ...interface{})
	Warn(ctx context.Context, msg string, args ...interface{})
	Debug(ctx context.Context, msg string, args ...interface{})
	Error(ctx context.Context, msg string, args ...interface{})

	WithFields(fields map[string]interface{}) Delegate
	WithField(key string, value interface{}) Delegate
	WithError(err error) Delegate
	Named(component string) Delegate
	SkipCallers(count int) Delegate
}

// Logger is a unified logging interface that can be used to reduce binding to a specific logging implementation.
// It can be used even if any logging module was not used by consuming it as optional.
type Logger struct {
	Delegate Delegate
}

func (o *Logger) Info(ctx context.Context, msg string, args ...interface{}) {
	if o == nil {
		return
	}
	o.Delegate.Info(ctx, msg, args...)
}

func (o *Logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if o == nil {
		return
	}
	o.Delegate.Warn(ctx, msg, args...)
}

func (o *Logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	if o == nil {
		return
	}
	o.Delegate.Debug(ctx, msg, args...)
}

func (o *Logger) Error(ctx context.Context, msg string, args ...interface{}) {
	if o == nil {
		return
	}
	o.Delegate.Error(ctx, msg, args...)
}

func (o *Logger) WithFields(fields map[string]interface{}) *Logger {
	if o == nil {
		return nil
	}
	return &Logger{Delegate: o.Delegate.WithFields(fields)}
}

func (o *Logger) WithField(key string, value interface{}) *Logger {
	if o == nil {
		return nil
	}
	return &Logger{Delegate: o.Delegate.WithField(key, value)}
}

func (o *Logger) Named(component string) *Logger {
	if o == nil {
		return nil
	}
	return &Logger{Delegate: o.Delegate.Named(component)}
}

func (o *Logger) SkipCallers(count int) *Logger {
	if o == nil {
		return nil
	}
	return &Logger{Delegate: o.Delegate.SkipCallers(count)}
}

func (o *Logger) WithError(err error) *Logger {
	if o == nil {
		return nil
	}

	return &Logger{Delegate: o.Delegate.WithError(err)}
}
