package logging

import "context"

// Logger is unified interface for logging, which should be used instead of concrete logger
// to simplify switching between logger implementations.
type Logger interface {
	Info(ctx context.Context, msg string, args ...interface{})
	Debug(ctx context.Context, msg string, args ...interface{})
	Error(ctx context.Context, msg string, args ...interface{})

	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	Named(component string) Logger
}
