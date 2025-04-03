package pnpwatermill

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
)

type originalPubSubMessageContextKey struct{}

func OriginalPubSubMessageFromContext(ctx context.Context) (*pubsub.Message, bool) {
	result, ok := ctx.Value(originalPubSubMessageContextKey{}).(*pubsub.Message)

	return result, ok
}
func contextWithOriginalPubSubMessage(ctx context.Context, pubsubMsg *pubsub.Message) context.Context {
	return context.WithValue(ctx, originalPubSubMessageContextKey{}, pubsubMsg)
}

type GCloudPubSubUnmarshaler struct {
	defaultUnmarshaler googlecloud.Unmarshaler
}

func NewGCloudPubSubUnmarshaler(defaultUnmarshaler googlecloud.Unmarshaler) GCloudPubSubUnmarshaler {
	return GCloudPubSubUnmarshaler{defaultUnmarshaler}
}

func (g GCloudPubSubUnmarshaler) Unmarshal(pubsubMsg *pubsub.Message) (*message.Message, error) {
	result, err := g.defaultUnmarshaler.Unmarshal(pubsubMsg)
	if err != nil {
		return nil, err
	}

	result.SetContext(contextWithOriginalPubSubMessage(result.Context(), pubsubMsg))

	return result, nil
}
