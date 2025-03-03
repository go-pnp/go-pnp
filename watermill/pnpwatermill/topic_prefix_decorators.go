package pnpwatermill

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type topicPrefixPublisherDecorator struct {
	prefix    string
	publisher message.Publisher
}

func (t topicPrefixPublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	return t.publisher.Publish(t.prefix+topic, messages...)
}

func (t topicPrefixPublisherDecorator) Close() error {
	return t.publisher.Close()
}

type topicPrefixSubscriberDecorator struct {
	prefix     string
	subscriber message.Subscriber
}

func (t topicPrefixSubscriberDecorator) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return t.subscriber.Subscribe(ctx, t.prefix+topic)
}

func (t topicPrefixSubscriberDecorator) Close() error {
	return t.subscriber.Close()
}
