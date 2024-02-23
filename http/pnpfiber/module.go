package pnpfiber

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(newOptions(), opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Provide(NewFiber)
	moduleBuilder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	moduleBuilder.InvokeIf(options.startServer, RegisterStartHooks)
	if options.fiberConfig != nil {
		moduleBuilder.Option(fx.Supply(&options.fiberConfig))
	}
	return moduleBuilder.Build()
}

type EndpointRegistrar func(app *fiber.App)

func EndpointRegistrarProvider(target any) any {
	return fxutil.GroupProvider[EndpointRegistrar](
		"pnp_fiber.endpoint_registrars",
		target,
	)
}

type NewFiberParams struct {
	fx.In
	FiberConfig        *fiber.Config
	EndpointsRegistrar EndpointRegistrar `group:"pnp_fiber.endpoint_registrars"`
}

func NewFiber(params NewFiberParams) (*fiber.App, error) {
	var configs []fiber.Config
	if params.FiberConfig != nil {
		configs = append(configs, *params.FiberConfig)
	}

	app := fiber.New(configs...)

	for params.EndpointsRegistrar != nil {
		params.EndpointsRegistrar(app)
	}

	return app, nil
}

type RegisterStartHooksParams struct {
	fx.In
	RuntimeErrors chan<- error
	Lc            fx.Lifecycle
	Logger        *logging.Logger `optional:"true"`
	Config        *Config
	App           *fiber.App
}

func RegisterStartHooks(params RegisterStartHooksParams) {
	logger := params.Logger.WithFields(map[string]interface{}{
		"addr":        params.Config.Addr,
		"tls_enabled": params.Config.TLS.Enabled,
	})

	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			tlsConfig, err := params.Config.TLS.TLSConfig()
			if err != nil {
				return err
			}

			go func() {
				logger.Info(ctx, "Starting Fiber HTTP server...")
				var err error
				if tlsConfig != nil {
					httpListener, err := net.Listen("tcp", params.Config.Addr)
					if err != nil {
						logger.WithError(err).Error(ctx, "Error creating TCP listener")
						params.RuntimeErrors <- err
						return
					}

					tlsListener := tls.NewListener(httpListener, tlsConfig)
					err = params.App.Listener(tlsListener)
				} else {
					err = params.App.Listen(params.Config.Addr)
				}
				if err != nil {
					logger.WithError(err).Error(ctx, "Error starting Fiber HTTP server")
					params.RuntimeErrors <- err
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info(ctx, "Stopping HTTP server...")
			if err := params.App.ShutdownWithContext(ctx); err != nil {
				logger.WithError(err).Error(ctx, "Error stopping Fiber HTTP server")
				return err
			}
			logger.Info(ctx, "Fiber HTTP server stopped")
			return nil
		},
	})
}
