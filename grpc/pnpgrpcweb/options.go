package pnpgrpcweb

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	fxPrivate bool
	useMux    bool
	muxPrefix string
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithMuxHandler(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.useMux = true
		o.muxPrefix = prefix
	}
}
