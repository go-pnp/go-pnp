package pnphttpserverh2c

import (
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type PanicHandler func(w http.ResponseWriter, panicValue any)

type options struct {
	configPrefix        string
	configFromContainer bool
	fxPrivate           bool
	order               int
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		configPrefix: "HTTP_SERVER_H2C_",
	}, opts...)
}

func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
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
