package pnpzap

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ContextFieldResolver func(context.Context) map[string]interface{}

type ZapCore struct {
	Delegate              zapcore.Core
	ContextFieldResolvers []ContextFieldResolver
}

func (z ZapCore) Enabled(level zapcore.Level) bool {
	return z.Delegate.Enabled(level)
}

func (z ZapCore) With(fields []zapcore.Field) zapcore.Core {
	newFields := make([]zapcore.Field, len(fields))
	copy(newFields, fields)

	for _, field := range fields {
		if field.Key != "context" {
			continue
		}
		ctx, ok := field.Interface.(context.Context)
		if !ok {
			continue
		}

		for _, resolver := range z.ContextFieldResolvers {
			for k, v := range resolver(ctx) {
				newFields = append(fields, zap.Any(k, v))
			}
		}
	}
	return ZapCore{
		Delegate:              z.Delegate.With(newFields),
		ContextFieldResolvers: z.ContextFieldResolvers,
	}
}

func (z ZapCore) Check(entry zapcore.Entry, entry2 *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return z.Delegate.Check(entry, entry2)
}

func (z ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return z.Delegate.Write(entry, fields)
}

func (z ZapCore) Sync() error {
	return z.Delegate.Sync()
}
