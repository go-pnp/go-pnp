package pnpgrpcclient

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/fxutil"
)

// Module provides *grpc.Server to fx container.
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	for _, dialOption := range options.dialOptions {
		fxutil.OptionsBuilderGroupSupply(builder, "pnpgrpcclient.dial_options", dialOption)
	}

	for _, interceptor := range options.unaryInterceptors {
		fxutil.OptionsBuilderGroupSupply(builder, "pnpgrpcclient.unary_interceptors", interceptor)
	}

	for _, interceptor := range options.streamInterceptors {
		fxutil.OptionsBuilderGroupSupply(builder, "pnpgrpcclient.stream_interceptors", interceptor)
	}

	for _, interceptor := range options.streamInterceptors {
		fxutil.OptionsBuilderSupply(builder, interceptor)
	}

	builder.Provide(NewDialer)

	return builder.Build()
}

type Dialer struct {
	Options []grpc.DialOption
}

func (d *Dialer) Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := append([]grpc.DialOption{}, d.Options...)
	options = append(options, opts...)

	return grpc.Dial(addr, options...)
}

func DialOptionProvider(target any) any {
	return fxutil.GroupProvider[grpc.DialOption](
		"pnpgrpcclient.dial_options",
		target,
	)
}

func UnaryClientInterceptorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[grpc.UnaryClientInterceptor]](
		"pnpgrpcclient.unary_interceptors",
		target,
	)
}
func StreamClientInterceptorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[grpc.StreamClientInterceptor]](
		"pnpgrpcclient.stream_interceptors",
		target,
	)
}

type NewGRPCDialerParams struct {
	fx.In
	Options            []grpc.DialOption                                   `group:"pnpgrpcclient.dial_options"`
	UnaryInterceptors  ordering.OrderedItems[grpc.UnaryClientInterceptor]  `group:"pnpgrpcclient.unary_interceptors"`
	StreamInterceptors ordering.OrderedItems[grpc.StreamClientInterceptor] `group:"pnpgrpcclient.stream_interceptors"`
}

func NewDialer(params NewGRPCDialerParams) *Dialer {
	options := append([]grpc.DialOption{}, params.Options...)
	options = append(options, grpc.WithChainUnaryInterceptor(params.UnaryInterceptors.Get()...))
	options = append(options, grpc.WithChainStreamInterceptor(params.StreamInterceptors.Get()...))

	return &Dialer{
		Options: options,
	}
}
