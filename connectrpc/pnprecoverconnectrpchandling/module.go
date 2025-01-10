package pnprecoverconnectrpchandling

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Supply(options)
	builder.Provide(NewInterceptor)
	builder.Provide(pnpconnectrpchandling.InterceptorProvider(NewInterceptor))

	return builder.Build()
}

type Interceptor struct {
}

func NewInterceptor(options *options) ordering.OrderedItem[connect.Interceptor] {
	return ordering.OrderedItem[connect.Interceptor]{
		Value: &Interceptor{},
		Order: options.order,
	}
}

func (i Interceptor) WrapUnary(unaryFunc connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (_ connect.AnyResponse, rErr error) {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				rErr = connect.NewError(connect.CodeInternal, errors.New("internal error"))
			}
		}()

		return unaryFunc(ctx, request)
	}
}

func (i Interceptor) WrapStreamingClient(clientFunc connect.StreamingClientFunc) connect.StreamingClientFunc {
	return clientFunc
}

func (i Interceptor) WrapStreamingHandler(handlerFunc connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) (rErr error) {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				rErr = connect.NewError(connect.CodeInternal, errors.New("internal error"))
			}
		}()

		return handlerFunc(ctx, conn)
	}
}
