package pnphttpserver

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	start      bool
	provideMux bool
	fxPrivate  bool
	config     *Config
}

func WithConfig(config *Config) optionutil.Option[options] {
	return func(o *options) {
		o.config = config
	}
}

func DoNotProvideMux() optionutil.Option[options] {
	return func(o *options) {
		o.provideMux = false
	}
}

func DoNotStart() optionutil.Option[options] {
	return func(o *options) {
		o.start = false
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}
