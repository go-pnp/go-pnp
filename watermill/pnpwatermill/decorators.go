package pnpwatermill

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
)

type SubscriberDecorator func(subscriber message.Subscriber) message.Subscriber

func SubscriberDecoratorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[SubscriberDecorator]]("pnpwatermill.subscriber_decorators", target)
}

type PublisherDecorator func(publisher message.Publisher) message.Publisher

func PublisherDecoratorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[PublisherDecorator]]("pnpwatermill.publisher_decorators", target)
}
