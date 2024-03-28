package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/pkg/ordering"

	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
)

type Handler struct {
	UnimplementedTestServiceServer
}

func (h *Handler) Test(context.Context, *TestRequest) (*TestResponse, error) {
	return &TestResponse{Message: "Hello, world!"}, nil
}

func (h *Handler) Register(server *grpc.Server) {
	RegisterTestServiceServer(server, h)
}

func NewHandler() *Handler {
	return &Handler{}
}

func main() {
	os.Setenv("GRPC_LISTEN_ADDR", "localhost:50051")
	/*
		os.Setenv("GRPC_TLS_ENABLED", "true")
		os.Setenv("GRPC_TLS_CERT_PATH", "cert.pem")
		os.Setenv("GRPC_TLS_KEY_PATH", "key.pem")
		os.Setenv("GRPC_TLS_CLIENT_AUTH", "require_and_verify_client_cert")
		os.Setenv("GRPC_TLS_CLIENT_CA_PATHS", "ca.pem")
	*/
	fx.New(
		pnpgrpcserver.Module(
			// Add fx.Private to all module provides
			pnpgrpcserver.WithFxPrivate(),
			// pass options to gRPC server
			pnpgrpcserver.WithServerOptions(grpc.ConnectionTimeout(time.Second)),
			pnpgrpcserver.Start(true),
		),
		fx.Provide(
			pnpgrpcserver.ServiceRegistrarProvider(NewHandler),
			pnpgrpcserver.UnaryInterceptorProvider(func() ordering.OrderedItem[grpc.UnaryServerInterceptor] {
				return ordering.Ordered[grpc.UnaryServerInterceptor](
					0,
					func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
						// Add your gRPC interceptor here
						fmt.Println("Hello from unary interceptor")

						return handler(ctx, req)
					},
				)
			}),
			pnpgrpcserver.StreamInterceptorProvider(func() ordering.OrderedItem[grpc.StreamServerInterceptor] {
				return ordering.Ordered[grpc.StreamServerInterceptor](
					0,
					func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
						// Add your gRPC interceptor here
						fmt.Println("Hello from stream interceptor")

						return handler(srv, ss)
					})
			}),
			pnpgrpcserver.ServerOptionProvider(func() grpc.ServerOption {
				return grpc.MaxConcurrentStreams(1000)
			}),
			fx.Private,
		),
	).Run()
}
