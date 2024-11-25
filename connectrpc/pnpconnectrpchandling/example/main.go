package main

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling/example/gen"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling/example/gen/genconnect"
)

type Handler struct {
}

func (h *Handler) WithPanic(ctx context.Context, c *connect.Request[gen.TestRequest]) (*connect.Response[gen.TestResponse], error) {
	panic("Hey, I'm panicking")
}

var _ genconnect.TestServiceHandler = (*Handler)(nil)

func (h *Handler) Test(ctx context.Context, c *connect.Request[gen.TestRequest]) (*connect.Response[gen.TestResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented yet"))
}

type CustomInterceptor struct {
}

func (c CustomInterceptor) WrapUnary(unaryFunc connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		fmt.Println("Hello from unary interceptor")
		return unaryFunc(ctx, request)
	}
}

func (c CustomInterceptor) WrapStreamingClient(clientFunc connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		fmt.Println("Hello from streaming client interceptor")
		return clientFunc(ctx, spec)
	}
}

func (c CustomInterceptor) WrapStreamingHandler(handlerFunc connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		fmt.Println("Hello from streaming handler interceptor")
		return handlerFunc(ctx, conn)
	}
}

func NewHandler() *Handler {
	return &Handler{}
}

func main() {
	fx.New(
		pnphttpserver.Module(),
		pnpconnectrpchandling.Module(),
		fx.Provide(
			// Providing our handler
			fx.Annotate(NewHandler, fx.As(new(genconnect.TestServiceHandler))),
			// Providing ConnectRPC handler constructor
			pnpconnectrpchandling.ConnectHandlerConstructorProvider(
				genconnect.NewTestServiceHandler,
				// Here we can specify options for TestHandler
			),
			// Providing Interceptor for all handlers
			pnpconnectrpchandling.InterceptorProvider(func() *CustomInterceptor {
				return &CustomInterceptor{}
			}),
			// Providing Option for all handlers
			pnpconnectrpchandling.HandlerOptionProvider(func() connect.HandlerOption {
				return connect.WithRecover(func(ctx context.Context, spec connect.Spec, header http.Header, a any) error {
					return errors.New("recovered from panic")
				})
			}),
		),
	).Run()
}
