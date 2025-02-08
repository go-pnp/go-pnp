package pnphttpserversentry

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	order     int
	fxPrivate bool
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		order:     0,
		fxPrivate: false,
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithOrder is an option to set order of recovery middleware.
func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}
