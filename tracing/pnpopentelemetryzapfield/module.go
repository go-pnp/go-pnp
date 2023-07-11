package pnpopentelemetryzapfield

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	builder.Provide(pnpzap.ZapContextFieldResolverProvider(NewZapContextFieldResolver))

	return builder.Build()
}

func NewZapContextFieldResolver() pnpzap.ContextFieldResolver {
	return func(ctx context.Context) map[string]interface{} {
		traceID := trace.SpanFromContext(ctx).SpanContext().TraceID()
		if !traceID.IsValid() {
			return nil
		}

		return map[string]interface{}{
			"trace_id": traceID.String(),
		}
	}
}
