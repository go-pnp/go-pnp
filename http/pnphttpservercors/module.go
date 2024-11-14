package pnphttpservercors

import (
	"context"
	"net/url"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/gobwas/glob"
	"github.com/rs/cors"
	"go.uber.org/fx"
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
	moduleBuilder.Provide(newCORS)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

type newCORSParams struct {
	fx.In
	Options *options
	Config  *Config
	Logger  *logging.Logger `optional:"true"`
}

func newCORS(params newCORSParams) (*cors.Cors, error) {
	if params.Config.AllowAll {
		return cors.AllowAll(), nil
	}
	logger := params.Logger.Named("cors")

	validOrigins := make(map[string]struct{}, len(params.Config.AllowedOrigins))
	for _, origin := range params.Config.AllowedOrigins {
		validOrigins[origin] = struct{}{}
	}
	globs := make([]glob.Glob, 0, len(params.Config.AllowedOriginGlobs))
	for _, globStr := range params.Config.AllowedOriginGlobs {
		g, err := glob.Compile(globStr, '.')
		if err != nil {
			return nil, err
		}
		globs = append(globs, g)
	}

	return cors.New(cors.Options{
		AllowedHeaders: params.Config.AllowedHeaders,
		AllowOriginFunc: func(origin string) bool {
			if params.Config.AllowAll {
				return true
			}

			if origin == "" {
				return true
			}

			originURL, err := url.Parse(origin)
			if err != nil {
				logger.WithField("origin", origin).Error(context.Background(), "CORS: can't parse origin")

				return false
			}

			_, ok := validOrigins[originURL.Hostname()]
			if ok {
				return true
			}

			for _, g := range globs {
				if g.Match(originURL.Hostname()) {
					return true
				}
			}

			if !params.Options.disableWarningLogs {
				params.Logger.WithField("request_origin_url", originURL).Warn(context.Background(), "received request from not allowed origin")
			}

			return ok
		},
	}), nil
}

func newMiddleware(options *options, cors *cors.Cors) ordering.OrderedItem[pnphttpserver.HandlerMiddleware] {
	return ordering.OrderedItem[pnphttpserver.HandlerMiddleware]{
		Value: cors.Handler,
		Order: options.order,
	}
}
