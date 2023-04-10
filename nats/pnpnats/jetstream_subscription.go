package pnpnats

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

type JetStreamSubscription struct {
	Subject string
	Handler MessageHandler
	// Middlewares are applied to this subscription
	Middlewares []Middleware
	// Options are applied to this subscription
	Options []nats.SubOpt
}

type MessageHandler interface {
	Handle(context.Context, *nats.Msg)
}

type MessageHandlerFunc func(context.Context, *nats.Msg)

func (f MessageHandlerFunc) Handle(ctx context.Context, natsMsg *nats.Msg) {
	f(ctx, natsMsg)
}

type RegisterJetStreamSubscriptionHooksParams struct {
	Lc            fx.Lifecycle
	JetStream     nats.JetStreamContext
	Subscriptions []JetStreamSubscription `group:"pnpnats.jetstream_subscriptions"`
	// Middlewares are applied to all subscriptions
	Middlewares []Middleware `group:"pnpnats.middlewares"`
	// Options are applied to all subscriptions
	Options []nats.SubOpt `group:"pnpnats.sub_opts"`
}

func RegisterJetStreamSubscriptionHooks(params RegisterJetStreamSubscriptionHooksParams) {
	for _, sub := range params.Subscriptions {
		middlewares := make([]Middleware, 0, len(params.Middlewares)+len(sub.Middlewares))
		middlewares = append(middlewares, params.Middlewares...)
		middlewares = append(middlewares, sub.Middlewares...)
		subOpts := make([]nats.SubOpt, 0, len(params.Options)+len(sub.Options))
		subOpts = append(subOpts, params.Options...)
		subOpts = append(subOpts, sub.Options...)
		RegisterJetStreamSubscriptionHook(RegisterJetStreamSubscriptionHookParams{
			Lc:          params.Lc,
			Js:          params.JetStream,
			Middlewares: middlewares,
			Options:     subOpts,
		})
	}

}

type RegisterJetStreamSubscriptionHookParams struct {
	Lc           fx.Lifecycle
	Js           nats.JetStreamContext
	Middlewares  []Middleware
	Options      []nats.SubOpt
	Subscription JetStreamSubscription
}

func RegisterJetStreamSubscriptionHook(params RegisterJetStreamSubscriptionHookParams) {
	var subscription *nats.Subscription
	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sub, err := params.Js.Subscribe(params.Subscription.Subject, func(msg *nats.Msg) {
				ctx := context.Background()
				runMiddlewares(ctx, msg, params.Subscription.Handler, params.Middlewares, 0)
			}, params.Options...)
			if err != nil {
				return err
			}
			subscription = sub

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if subscription != nil {
				subscription.Unsubscribe()
			}
			return nil
		},
	})
}

type Middleware func(ctx context.Context, msg *nats.Msg, next MessageHandler)

func runMiddlewares(ctx context.Context, msg *nats.Msg, handler MessageHandler, middlewares []Middleware, index int) {
	if index >= len(middlewares) {
		handler.Handle(ctx, msg)
		return
	}

	middlewares[index](ctx, msg, MessageHandlerFunc(func(nextCtx context.Context, msg *nats.Msg) {
		runMiddlewares(nextCtx, msg, handler, middlewares, index+1)
	}))
}
