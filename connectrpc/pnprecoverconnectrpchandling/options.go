package pnprecoverconnectrpchandling

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/pkg/errors"
)

// PanicHandler converts a recovered panic into the error the interceptor returns to the client.
type PanicHandler func(ctx context.Context, spec connect.Spec, panicValue any) error

type options struct {
	fxPrivate    bool
	order        int
	panicHandler PanicHandler
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		panicHandler: func(context.Context, connect.Spec, any) error {
			return connect.NewError(connect.CodeInternal, errors.New("internal error"))
		},
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
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

// WithPanicHandler is an option to set custom panic handler. By default a panic is turned into
// a connect.CodeInternal error with an "internal error" message, which discards the panic value.
func WithPanicHandler(panicHandler PanicHandler) optionutil.Option[options] {
	return func(o *options) {
		o.panicHandler = panicHandler
	}
}
