package pnpwatermill

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	moduleBuilder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	moduleBuilder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))

	moduleBuilder.Supply(options)
	moduleBuilder.Provide(NewTransport)
	moduleBuilder.Provide(newRouter)
	moduleBuilder.InvokeIf(options.startRouter, fx.Annotate(runRouter, fx.OnStop(closeRouter)))

	return moduleBuilder.Build()
}

func HandlerMiddlewareProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[message.HandlerMiddleware]]("pnpwatermill.handler_middlewares", target)
}

type SubscriberConfigProvider interface {
	Config() *SubscriberConfig
}

type newRouterParams struct {
	fx.In
	Lc                fx.Lifecycle
	Logger            *logging.Logger                                  `optional:"true"`
	HandleMiddlewares ordering.OrderedItems[message.HandlerMiddleware] `group:"pnpwatermill.handler_middlewares"`
	Handlers          []Handler                                        `group:"pnpwatermill.handlers"`
	SubscriberFactory subscriberFactory
}

func newRouter(params newRouterParams) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})
	if err != nil {
		return nil, err
	}

	for _, handler := range params.Handlers {
		subscriberConfig := new(SubscriberConfig)

		if handler, ok := handler.(SubscriberConfigProvider); ok {
			subscriberConfig = handler.Config()
		}

		params.Logger.Info(context.Background(), "creating subscriber for '%s' handler", handler.Name())
		subscriber, err := params.SubscriberFactory.NewSubscriber(handler.Name(), subscriberConfig)
		if err != nil {
			return nil, fmt.Errorf("create subscriber: %w", err)
		}

		router.AddNoPublisherHandler(handler.Name(), handler.Topic(), subscriber, handler.Handle)
	}

	for _, middleware := range params.HandleMiddlewares.Get() {
		router.AddMiddleware(middleware)
	}

	return router, nil
}

type runRouterParams struct {
	fx.In
	Shutdowner fx.Shutdowner
	Logger     *logging.Logger
	Router     *message.Router
}

func runRouter(params runRouterParams) {
	go func() {
		params.Logger.Info(context.Background(), "starting watermill router")
		if err := params.Router.Run(context.Background()); err != nil {
			params.Logger.WithError(err).Error(context.Background(), "watermill router start error, shutting down")
		} else {
			params.Logger.Info(context.Background(), "watermill router finished running")
		}

		if err := params.Shutdowner.Shutdown(); err != nil {
			params.Logger.WithError(err).Error(context.Background(), "failed to shutdown")
		}
	}()
}

func closeRouter(router *message.Router) error {
	return router.Close()
}
