package pnpfiberhealthcheck

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-pnp/go-pnp/http/pnpfiber"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/healthcheck/pnphealthcheck"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	fxutil.OptionsBuilderSupply(moduleBuilder, options)
	moduleBuilder.Provide(NewHealthcheckHandler)
	moduleBuilder.Provide(pnpfiber.EndpointRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

func WriteResponse(alive bool, checks map[string]error, ctx *fiber.Ctx) {
	ctx.Response().Header.Set("Content-Type", "application/json")

	if !alive {
		ctx.Response().SetStatusCode(http.StatusServiceUnavailable)
	} else {
		ctx.Response().SetStatusCode(http.StatusOK)
	}

	_ = json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]interface{}{
		"alive":       alive,
		"checkErrors": checks,
	})
}

type HealthCheckHandler fiber.Handler

func NewHealthcheckHandler(
	options *options,
	healthResolver *pnphealthcheck.HealthResolver,
) HealthCheckHandler {
	return func(ctx *fiber.Ctx) error {
		checkResults, alive := healthResolver.Resolve(ctx.Context())
		if options.responseWriter != nil {
			options.responseWriter(alive, checkResults, ctx)

			return nil
		}

		if alive {
			ctx.Response().SetStatusCode(http.StatusOK)
		} else {
			ctx.Response().SetStatusCode(http.StatusServiceUnavailable)
		}

		return nil
	}
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Options *options
	Handler HealthCheckHandler
	Logger  *logging.Logger `optional:"true"`
}

func NewMuxHandlerRegistrar(params NewMuxHandlerRegistrarParams) pnpfiber.EndpointRegistrar {
	return func(app *fiber.App) {
		params.Logger.Named("fiber-healthchecks").Debug(context.Background(), "Registering healthcheck handler")
		app.Add(params.Options.method, params.Options.path, params.Handler)
	}
}
