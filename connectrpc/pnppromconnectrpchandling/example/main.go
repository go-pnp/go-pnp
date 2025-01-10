package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/http/pnphttpserverrecovery"
	"github.com/go-pnp/go-pnp/http/pnppromhttp"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/connectrpc/pnppromconnectrpchandling"
	"github.com/go-pnp/go-pnp/connectrpc/pnppromconnectrpchandling/example/gen"
	"github.com/go-pnp/go-pnp/connectrpc/pnppromconnectrpchandling/example/gen/genconnect"
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

func NewHandler() *Handler {
	return &Handler{}
}

func main() {
	fx.New(
		pnphttpserver.Module(),
		pnphttpserverrecovery.Module(),
		pnpprometheus.Module(),
		pnppromhttp.Module(),
		pnpconnectrpchandling.Module(),
		fx.Provide(
			// Providing our handler
			fx.Annotate(NewHandler, fx.As(new(genconnect.TestServiceHandler))),
			// Providing ConnectRPC handler constructor
			pnpconnectrpchandling.ConnectHandlerConstructorProvider(
				genconnect.NewTestServiceHandler,
				// Here we can specify options for TestHandler
			),
		),
		pnppromconnectrpchandling.Module(),
	).Run()
}
