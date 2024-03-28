package pnpgorm

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	fxPrivate           bool
	configFromContainer bool
	configPrefix        string
	enableLogger        bool
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		configPrefix: "DB_",
	}, opts...)
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

// WithConfigPrefix is an option to set config prefix for config provider.
func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

func WithLogger() optionutil.Option[options] {
	return func(o *options) {
		o.enableLogger = true
	}
}
