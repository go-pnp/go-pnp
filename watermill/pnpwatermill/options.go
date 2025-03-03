package pnpwatermill

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	configFromContainer bool
	configPrefix        string
	fxPrivate           bool
	startRouter         bool
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		configFromContainer: false,
		configPrefix:        "WATERMILL_",
		fxPrivate:           false,
		startRouter:         true,
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithConfigFromContainer is an option to use config from provider instead of providing it.
func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}

// WithConfigPrefix is an option to set prefix for environment variables.
func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

// WithStart if true - module adds invoke method to fx container to start router.
func WithStart(start bool) optionutil.Option[options] {
	return func(o *options) {
		o.startRouter = start
	}
}
