package pnpwatermilllogging

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
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
	moduleBuilder.Provide(pnpwatermill.HandlerMiddlewareProvider(NewLoggerMiddleware))

	return moduleBuilder.Build()
}

func NewLoggerMiddleware(logger *logging.Logger, opts *options) ordering.OrderedItem[message.HandlerMiddleware] {
	return ordering.OrderedItem[message.HandlerMiddleware]{
		Order: opts.order,
		Value: func(h message.HandlerFunc) message.HandlerFunc {
			return func(msg *message.Message) ([]*message.Message, error) {
				topic := message.SubscribeTopicFromCtx(msg.Context())
				subscriber := message.SubscriberNameFromCtx(msg.Context())
				handlerName := message.HandlerNameFromCtx(msg.Context())

				messageLogger := logger.WithFields(map[string]interface{}{
					"topic":      topic,
					"subscriber": subscriber,
					"handler":    handlerName,
					"metadata":   msg.Metadata,
				})
				messages, err := h(msg)
				if err != nil {
					messageLogger.WithField("error", err).WithField("payload", string(msg.Payload)).Error(msg.Context(), "message handling error")
				} else {
					messageLogger.WithField("payload", msg.Payload).Debug(msg.Context(), "message handled")
				}

				return messages, err
			}
		},
	}
}
