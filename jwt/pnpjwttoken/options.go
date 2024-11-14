package pnpjwttoken

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	fxPrivate           bool
	configPrefix        string
	configFromContainer bool
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		fxPrivate:    false,
		configPrefix: "JWT_",
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
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
