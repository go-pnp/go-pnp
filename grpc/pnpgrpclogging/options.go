package pnpgrpclogging

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type options struct {
	fxPrivate   bool
	order       int
	serverOrder *int
	clientOrder *int
}

func (o options) getServerOrder() int {
	if o.serverOrder != nil {
		return *o.serverOrder
	}

	return o.order
}

func (o options) getClientOrder() int {
	if o.clientOrder != nil {
		return *o.clientOrder
	}

	return o.order
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithInterceptorsOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}
func WithServerInterceptorsOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.serverOrder = &order
	}
}

func WithClientInterceptorsOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.clientOrder = &order
	}
}
