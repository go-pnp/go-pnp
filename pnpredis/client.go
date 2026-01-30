package pnpredis

import (
	"errors"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type ClientDecorator func(redis.UniversalClient) redis.UniversalClient

func ClientDecoratorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[ClientDecorator]](
		"pnp_redis.client_decorators",
		target,
	)
}

type NewRedisClientParams struct {
	fx.In

	Config           *Config
	ClientDecorators ordering.OrderedItems[ClientDecorator] `group:"pnp_redis.client_decorators"`
}

func NewRedisClient(params NewRedisClientParams) (redis.UniversalClient, error) {
	client, err := newRedisClient(params)
	if err != nil {
		return nil, err
	}

	for _, decorator := range params.ClientDecorators.Get() {
		client = decorator(client)
	}

	return client, nil
}

func CloseClient(client redis.UniversalClient) error {
	if err := client.Close(); err != nil && !errors.Is(redis.ErrClosed, err) {
		return err
	}

	return nil
}

func newRedisClient(params NewRedisClientParams) (redis.UniversalClient, error) {
	tls, err := params.Config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	if params.Config.Sentinel.Enabled {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       params.Config.Sentinel.Master,
			SentinelAddrs:    params.Config.Sentinel.Addrs,
			SentinelUsername: params.Config.Sentinel.Username,
			SentinelPassword: params.Config.Sentinel.Password,
			Username:         params.Config.Username,
			Password:         params.Config.Password,
			DB:               params.Config.DB,
		}), nil
	}

	return redis.NewClient(&redis.Options{
		Addr:      params.Config.Address,
		Password:  params.Config.Password,
		DB:        params.Config.DB,
		TLSConfig: tls,
	}), nil
}
