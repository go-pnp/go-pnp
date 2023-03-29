//go:build example
// +build example

package pnpgrpcserver

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/fxutil"
)

func TestApp(t *testing.T) {
	os.Setenv("GRPC_LISTEN_ADDR", "localhost:50051")
	/*
		os.Setenv("GRPC_TLS_ENABLED", "true")
		os.Setenv("GRPC_TLS_CERT_PATH", "cert.pem")
		os.Setenv("GRPC_TLS_KEY_PATH", "key.pem")
		os.Setenv("GRPC_TLS_CLIENT_AUTH", "require_and_verify_client_cert")
		os.Setenv("GRPC_TLS_CLIENT_CA_PATHS", "ca.pem")
	*/
	fxutil.StartApp(
		Module(
			// Add fx.Private to all module provides
			WithFxPrivate(),
			// pass options to gRPC server
			WithServerOptions(grpc.ConnectionTimeout(time.Second)),
			// Do not register server start hook
			Start(false),
		),
		fx.Provide(
			ServiceRegistrarProvider(func() ServiceRegistrar {
				return func(server *grpc.Server) {
					// Register your gRPC services here
				}
			}),
			UnaryInterceptorProvider(func() grpc.UnaryServerInterceptor {
				return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
					// Add your gRPC interceptor here
					fmt.Println("Hello from unary interceptor")

					return handler(ctx, req)
				}
			}),
			StreamInterceptorProvider(func() grpc.StreamServerInterceptor {
				return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
					// Add your gRPC interceptor here
					fmt.Println("Hello from stream interceptor")

					return handler(srv, stream)
				}
			}),
			ServerOptionProvider(func() grpc.ServerOption {
				return grpc.MaxConcurrentStreams(1000)
			}),
			fx.Private,
		),
	)
}
