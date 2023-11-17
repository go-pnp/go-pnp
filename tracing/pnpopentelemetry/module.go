package pnpopentelemetry

import (
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

// Module To find opentelemetry configuration,
// check URL https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	builder.Provide(sdkTrace.NewTracerProvider)
	builder.Provide(func(provider *sdkTrace.TracerProvider) trace.TracerProvider {
		return provider
	})

	return builder.Build()
}
