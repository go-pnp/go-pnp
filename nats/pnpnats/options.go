package pnpnats

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	fxPrivate           bool
	configFromContainer bool

	jetStream          bool
	jetStreamSubscribe bool
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

func WithJetstream() optionutil.Option[options] {
	return func(o *options) {
		o.jetStream = true
	}
}

// JetStreamSubscribe if used, module will collect all subsciptions from uber fx and subscribe to them.
func JetStreamSubscribe(subscribe bool) optionutil.Option[options] {
	return func(o *options) {
		o.jetStreamSubscribe = subscribe
	}
}
