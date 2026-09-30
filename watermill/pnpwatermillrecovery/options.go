package pnpwatermillrecovery

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	order     int
	fxPrivate bool
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{}, opts...)
}

func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}
