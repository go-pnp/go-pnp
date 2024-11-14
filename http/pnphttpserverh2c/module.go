package pnphttpserverh2c

import (
	"net/http"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.ProvideIf(
		!options.configFromContainer,
		configutil.NewPrefixedConfigProvider[Config](options.configPrefix),
		configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix),
	)
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

type NewMiddlewareParams struct {
	fx.In
	Options *options
	Config  *Config
}

func newMiddleware(params NewMiddlewareParams) ordering.OrderedItem[pnphttpserver.HandlerMiddleware] {
	return ordering.OrderedItem[pnphttpserver.HandlerMiddleware]{
		Value: func(handler http.Handler) http.Handler {
			return h2c.NewHandler(handler, &http2.Server{
				MaxHandlers:                  params.Config.MaxHandlers,
				MaxConcurrentStreams:         params.Config.MaxConcurrentStreams,
				MaxDecoderHeaderTableSize:    params.Config.MaxDecoderHeaderTableSize,
				MaxEncoderHeaderTableSize:    params.Config.MaxEncoderHeaderTableSize,
				MaxReadFrameSize:             params.Config.MaxReadFrameSize,
				PermitProhibitedCipherSuites: params.Config.PermitProhibitedCipherSuites,
				IdleTimeout:                  params.Config.IdleTimeout,
				ReadIdleTimeout:              params.Config.ReadIdleTimeout,
				PingTimeout:                  params.Config.PingTimeout,
				WriteByteTimeout:             params.Config.WriteByteTimeout,
				MaxUploadBufferPerConnection: params.Config.MaxUploadBufferPerConnection,
				MaxUploadBufferPerStream:     params.Config.MaxUploadBufferPerStream,
			})
		},
		Order: params.Options.order,
	}
}
