package pnppromhttp

import "github.com/go-pnp/go-pnp/pkg/optionutil"

type endpoint struct {
	path   string
	method string
}
type options struct {
	registerInMux bool
	fxPrivate     bool
	endpoint      endpoint
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		registerInMux: true,
		endpoint: endpoint{
			method: "GET",
			path:   "/metrics",
		},
	}, opts...)
}

func RegisterInMux(registerInMux bool) optionutil.Option[options] {
	return func(o *options) {
		o.registerInMux = registerInMux
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithEndpoint(method, path string) optionutil.Option[options] {
	return func(o *options) {
		o.endpoint = endpoint{
			path:   path,
			method: method,
		}
	}
}
