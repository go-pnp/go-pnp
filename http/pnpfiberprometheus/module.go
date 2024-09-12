package pnpfiberprometheus

import (
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnpfiber"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(newOptions(), opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(pnpfiber.EndpointRegistrarProvider(NewPrometheusEndpointRegistrarProvider))

	return moduleBuilder.Build()
}

func NewPrometheusEndpointRegistrarProvider(options *options, registry *prometheus.Registry) pnpfiber.EndpointRegistrar {
	handler := promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	).ServeHTTP

	return func(app *fiber.App) {
		app.Get(options.httpPath, adaptor.HTTPHandlerFunc(handler))
	}
}
