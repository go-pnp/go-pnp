package pnpzapsentry

import (
	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	fxPrivate           bool
	configFromContainer bool
	configPrefix        string
	sentryClientOptions []optionutil.Option[sentry.ClientOptions]
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		configPrefix: "SENTRY_",
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}
