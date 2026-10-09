package pnpwatermillrecovery

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/go-pnp/pkg/panicutil"
	"github.com/go-pnp/go-pnp/watermill/pnpwatermill"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(pnpwatermill.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

func newMiddleware(options *options) ordering.OrderedItem[message.HandlerMiddleware] {
	return ordering.OrderedItem[message.HandlerMiddleware]{
		Value: recoverHandler,
		Order: options.order,
	}
}

func recoverHandler(handler message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (produced []*message.Message, err error) {
		defer panicutil.Recover(func(panicError error) {
			err = panicError
		})

		return handler(msg)
	}
}
