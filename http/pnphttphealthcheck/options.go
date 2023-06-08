package pnphttphealthcheck

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	registerInMux       bool
	fxPrivate           bool
	configFromContainer bool
}

func RegisterInMux(registerInMux bool) optionutil.Option[options] {
	return func(o *options) {
		o.registerInMux = registerInMux
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
