package pnpwatermillsentry

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/getsentry/sentry-go"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
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

func newMiddleware(options *options, client *sentry.Client) ordering.OrderedItem[message.HandlerMiddleware] {
	return ordering.OrderedItem[message.HandlerMiddleware]{
		Value: func(h message.HandlerFunc) message.HandlerFunc {
			return func(msg *message.Message) ([]*message.Message, error) {
				hub := sentry.GetHubFromContext(msg.Context())
				var scope *sentry.Scope
				if hub == nil {
					scope = sentry.NewScope()
					hub = sentry.NewHub(client, scope)
					msg.SetContext(sentry.SetHubOnContext(msg.Context(), hub))
				} else {
					scope = hub.PushScope()
					defer hub.PopScope()
				}

				topic := message.SubscribeTopicFromCtx(msg.Context())
				handlerName := message.HandlerNameFromCtx(msg.Context())

				span := sentry.StartSpan(msg.Context(), handlerName)
				defer span.Finish()

				msg.SetContext(span.Context())

				scope.SetSpan(span)
				scope.SetTag("topic", topic)
				scope.SetTag("handler", handlerName)

				return h(msg)
			}
		},
		Order: options.order,
	}
}
