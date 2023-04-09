package pnpzap

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	zapConfigFromContainer bool
	configFromContainer    bool
	fxPrivate              bool
}

func WithZapConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.zapConfigFromContainer = true
	}
}

func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}
