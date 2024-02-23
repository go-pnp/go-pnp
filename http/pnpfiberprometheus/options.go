package pnpfiberprometheus

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	// all provides are private
	fxPrivate bool

	// httpPath for prometheus metrics
	httpPath string

	// order of endpoint registration
	order int
}

func newOptions() *options {
	return &options{
		fxPrivate: false,
		httpPath:  "/metrics",
	}
}

func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithHTTPPath(path string) optionutil.Option[options] {
	return func(o *options) {
		o.httpPath = path
	}
}
