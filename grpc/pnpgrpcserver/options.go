package pnpgrpcserver

import (
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	start         bool
	serverOptions []grpc.ServerOption
	fxPrivate     bool
}

func Start(start bool) optionutil.Option[options] {
	return func(o *options) {
		o.start = start
	}
}

func WithServerOptions(serverOptions ...grpc.ServerOption) optionutil.Option[options] {
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, serverOptions...)
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}
