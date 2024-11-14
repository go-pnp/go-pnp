package pnphttpserverrecovery

import (
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type PanicHandler func(w http.ResponseWriter, panicValue any)

type options struct {
	panicHandler              PanicHandler
	panicHandlerFromContainer bool
	fxPrivate                 bool
	order                     int
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		panicHandler: func(w http.ResponseWriter, panicValue any) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	}, opts...)
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

// WithPanicHandler is an option to set custom panic handler.
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
