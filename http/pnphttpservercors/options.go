package pnphttpservercors

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	order               int
	fxPrivate           bool
	configPrefix        string
	configFromContainer bool
	disableWarningLogs  bool
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		order:        0,
		fxPrivate:    false,
		configPrefix: "HTTP_SERVER_CORS_",
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithOrder is an option to set order of recovery middleware.
func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}

// WithConfigPrefix is an option to set env config prefix for config provider.
func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

// WithConfigFromContainer if used, module will not provide config, but will use config already provided to fx DI container.
func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}

func WithDisabledWarningLogs() optionutil.Option[options] {
	return func(o *options) {
		o.disableWarningLogs = true
	}
}
