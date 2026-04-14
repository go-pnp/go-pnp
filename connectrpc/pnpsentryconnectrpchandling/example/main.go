package main

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling/example/gen"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling/example/gen/genconnect"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/http/pnphttpserversentry"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/logging/pnpzapsentry"
	"github.com/go-pnp/go-pnp/pnpenv"
	"github.com/go-pnp/go-pnp/pnpsentry"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/connectrpc/pnpsentryconnectrpchandling"
)

type Handler struct {
	logger *logging.Logger
}

var _ genconnect.TestServiceHandler = (*Handler)(nil)

func (h *Handler) Test(ctx context.Context, req *connect.Request[gen.TestRequest]) (*connect.Response[gen.TestResponse], error) {
	h.logger.Error(ctx, "Hello from ConnectRPC handler")
	return connect.NewResponse(&gen.TestResponse{
		Message: "Hello, " + req.Msg.Name,
	}), nil
}

func (h *Handler) WithPanic(ctx context.Context, req *connect.Request[gen.TestRequest]) (*connect.Response[gen.TestResponse], error) {
	panic("test panic from ConnectRPC handler")
}

func main() {
	os.Setenv("SENTRY_DSN", "") // REPLACE WITH YOUR SENTRY DSN
	os.Setenv("SENTRY_ENVIRONMENT", "development")
	os.Setenv("SENTRY_ENABLE_TRACING", "true")
	os.Setenv("SENTRY_TRACES_SAMPLE_RATE", "1.0")

	fx.New(
		pnpenv.Module(),
		pnpzap.Module(),
		pnpzapsentry.Module(),
		pnpsentry.Module(),
		pnphttpserver.Module(pnphttpserver.Start(true)),
		pnphttpserversentry.Module(),
		pnpconnectrpchandling.Module(),
		pnpsentryconnectrpchandling.Module(),
		fx.Provide(
			fx.Annotate(func(logger *logging.Logger) *Handler { return &Handler{logger: logger} }, fx.As(new(genconnect.TestServiceHandler))),
			pnpconnectrpchandling.ConnectHandlerConstructorProvider(
				genconnect.NewTestServiceHandler,
			),
		),
	).Run()
}
