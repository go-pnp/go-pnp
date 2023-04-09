package pnpprometheus

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	fxPrivate           bool
	configFromContainer bool

	start bool
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithConfigFromContainer if used, module will not provide config, but will use config already provided to fx di container.
func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}

func Start(start bool) optionutil.Option[options] {
	return func(o *options) {
		o.start = start
	}
}
