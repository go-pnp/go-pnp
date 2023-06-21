package pnpgrpcserver

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	start               bool
	fxPrivate           bool
	configFromContainer bool
	configPrefix        string
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

func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}
