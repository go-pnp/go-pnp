package pnpnats

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		jetStreamSubscribe: true,
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	builder.Provide(NewNatsClient)

	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: "NATS_",
	}))
	builder.ProvideIf(options.jetStream, NewJetstream)
	builder.InvokeIf(options.jetStreamSubscribe, RegisterJetStreamSubscriptionHooks)

	return builder.Build()

}

func ClientOptionProvider(target any) any {
	return fxutil.GroupProvider[nats.Option](
		"pnpnats.client_options",
		target,
	)
}

type NewNatsClientParams struct {
	fx.In
	Config        *Config
	ClientOptions []nats.Option `group:"pnpnats.client_options"`
	Lc            fx.Lifecycle
	Shutdowner    fx.Shutdowner
}

func NewNatsClient(params NewNatsClientParams) (*nats.Conn, error) {
	connectOpts := make([]nats.Option, 0, len(params.ClientOptions)+2)
	copy(connectOpts, params.ClientOptions)

	reconnectOpts := params.Config.getReconnectOptions()
	connectOpts = append(connectOpts, reconnectOpts...)

	tlsOpts, err := params.Config.getTLSOptions()
	if err != nil {
		return nil, err
	}
	connectOpts = append(connectOpts, tlsOpts...)

	conn, err := nats.Connect(params.Config.Addr, connectOpts...)
	if err != nil {
		return nil, err
	}

	params.Lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			conn.Close()
			return nil
		},
	})

	return conn, nil
}

func JetstreamOptionProvider(target any) any {
	return fxutil.GroupProvider[nats.Option](
		"pnpnats.jetstream_options",
		target,
	)
}

type NewJetstreamParams struct {
	fx.In
	Conn             *nats.Conn
	JetstreamOptions []nats.JSOpt `group:"pnpnats.jetstream_options"`
}

func NewJetstream(params NewJetstreamParams) (nats.JetStreamContext, error) {
	return params.Conn.JetStream(params.JetstreamOptions...)
}
