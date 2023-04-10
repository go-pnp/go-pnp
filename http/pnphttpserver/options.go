package pnphttpserver

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	start               bool
	provideMux          bool
	fxPrivate           bool
	configFromContainer bool
}

func DoNotProvideMux() optionutil.Option[options] {
	return func(o *options) {
		o.provideMux = false
	}
}

func Start(start bool) optionutil.Option[options] {
	return func(o *options) {
		o.start = start
	}
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
