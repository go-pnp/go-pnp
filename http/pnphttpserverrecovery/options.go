package pnphttpserverrecovery

import (
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

// request for reusing its context
type PanicHandler func(w http.ResponseWriter, r *http.Request, panicValue any)

type options struct {
	panicHandler              PanicHandler
	panicHandlerFromContainer bool
	fxPrivate                 bool
	order                     int
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithOrder is an option to set order of middleware.
func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}

// WithPanicHandler is an option to set custom panic handler. It replaces the default one, which logs the panic
// with its stack through *logging.Logger when one is provided (http.ErrAbortHandler is not logged) and responds with 500.
func WithPanicHandler(panicHandler PanicHandler) optionutil.Option[options] {
	return func(o *options) {
		o.panicHandler = panicHandler
	}
}

// WithPanicHandlerFromContainer if used, module will not provide panic handler, but will use panic handler already provided to fx di container.
func WithPanicHandlerFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.panicHandlerFromContainer = true
	}
}
