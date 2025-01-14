package pnpsentry

import (
	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	builder := fxutil.OptionsBuilder{}
	builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))
	builder.Provide(fx.Annotate(NewSentryClient, fx.OnStop(CloseClient)))

	return builder.Build()
}

func NewSentryClient(config *Config, options *options) (*sentry.Client, error) {
	clientOptions := optionutil.ApplyOptions(&sentry.ClientOptions{
		Dsn:              config.DSN,
		Debug:            config.Debug,
		AttachStacktrace: config.AttachStacktrace,
		SampleRate:       config.SampleRate,
		EnableTracing:    config.EnableTracing,
		TracesSampleRate: config.TracesSampleRate,
		ServerName:       config.ServerName,
		Release:          config.Release,
		Dist:             config.Dist,
		Environment:      config.Environment,
		MaxBreadcrumbs:   config.MaxBreadcrumbs,
		MaxSpans:         config.MaxSpans,
	}, options.sentryClientOptions...)

	return sentry.NewClient(*clientOptions)
}

func CloseClient(client *sentry.Client) {
	client.Close()
}
