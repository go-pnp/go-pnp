package pnpgrpcclient

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"google.golang.org/grpc"
)

type options struct {
	fxPrivate          bool
	dialOptions        []grpc.DialOption
	unaryInterceptors  []ordering.OrderedItem[grpc.UnaryClientInterceptor]
	streamInterceptors []ordering.OrderedItem[grpc.StreamClientInterceptor]
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		fxPrivate: false,
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithDialOptions(dialOptions ...grpc.DialOption) optionutil.Option[options] {
	return func(o *options) {
		o.dialOptions = append(o.dialOptions, dialOptions...)
	}
}

func WithUnaryInterceptor(order int, interceptor grpc.UnaryClientInterceptor) optionutil.Option[options] {
	return func(o *options) {
		o.unaryInterceptors = append(o.unaryInterceptors, ordering.OrderedItem[grpc.UnaryClientInterceptor]{
			Order: order,
			Value: interceptor,
		})
	}
}

func WithStreamInterceptor(order int, interceptor grpc.StreamClientInterceptor) optionutil.Option[options] {
	return func(o *options) {
		o.streamInterceptors = append(o.streamInterceptors, ordering.OrderedItem[grpc.StreamClientInterceptor]{
			Order: order,
			Value: interceptor,
		})
	}
}
