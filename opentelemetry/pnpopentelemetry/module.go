package pnpopentelemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
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

func TracerProviderOptionProvider(target any) any {
	return fxutil.GroupProvider[sdkTrace.TracerProviderOption]("pnopentelemetry.tracer_provider_options", target)
}

type NewTracerProviderParams struct {
	fx.In
	Options  []sdkTrace.TracerProviderOption `group:"pnpopentelemetry.tracer_provider_options"`
	Resource *resource.Resource              `optional:"true"`
}

func NewTracerProvider(params NewTracerProviderParams) (*sdkTrace.TracerProvider, error) {
	sdkTrace.NewTracerProvider()
}

func ResourceOptionProvider(target any) any {
	return fxutil.GroupProvider[sdkTrace.TracerProviderOption]("pnopentelemetry.resource_options", target)
}

type NewResourceParams struct {
	fx.In
	Options    []resource.Option    `group:"pnpopentelemetry.resource_options"`
	Attributes []attribute.KeyValue `group:"pnpopentelemetry.resource_attributes"`
}

func NewResource(params NewResourceParams) (*resource.Resource, error) {
	resourceOptions := append([]resource.Option{
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),
		resource.WithAttributes(params.Attributes...),
	}, params.Options...)
	res, err := resource.New(
		context.Background(),
		resourceOptions...,
	)
	if err != nil && !errors.Is(err, resource.ErrPartialResource) && !errors.Is(err, resource.ErrSchemaURLConflict) {
		return nil, err
	}

	return res, nil
}

type NewResourceFromSchemaURLParams struct {
	fx.In
	Options    []resource.Option    `group:"pnpopentelemetry.resource_options"`
	Attributes []attribute.KeyValue `group:"pnpopentelemetry.resource_attributes"`
}

func NewResourceFromSchemaURL(params NewResourceFromSchemaURLParams) (*resource.Resource, error) {
	resourceOptions := append([]resource.Option{
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),
		resource.WithAttributes(params.Attributes...),
	}, params.Options...)
	res, err := resource.NewWithAttributes(

		resourceOptions...,
	)
	if err != nil && !errors.Is(err, resource.ErrPartialResource) && !errors.Is(err, resource.ErrSchemaURLConflict) {
		return nil, err
	}

	return res, nil
}
