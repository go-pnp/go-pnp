## pnpgrpcserver

This package provides a gRPC server module that can be easily integrated into your application. 

### Getting started
Install pnpgrpcserver package:
```shell
go get github.com/go-pnp/go-pnp/grpc/pnpgrpcserver
````

Add pnpgrpc.Module() to your fx application options and provide your gRPC service registrar to fx container:
```go
package main
import "github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
import "go.uber.org/fx"
func main() {
	fx.New( 
		//< your options > 
		pnpgrpcserver.Module(),
		fx.Provide(NewMyGRPCServiceHandler), // NewMyGRPCServiceHandler should implement method Register(server *grpc.Server)
	)
}
```

### Config 
By default, module will create config from environment variables. 
You can override this behavior by providing your own config.
Default env variables prefix is `GRPC_`, you can change it by providing custom prefix. 
```shell
GRPC_LISTEN_ADDR # default - 127.0.0.1:50051
GRPC_TLS_ENABLED # default - false
GRPC_TLS_CERT_PATH
GRPC_TLS_KEY_PATH
GRPC_TLS_CLIENT_AUTH
GRPC_TLS_CLIENT_CA_PATH
GRPC_TLS_APPEND_SYSTEM_CAS_TO_CLIENT # default - false
```

### Module Options
- `pnpgrpcserver.Start(start bool)` - Start gRPC server on application start. Default - true.
- `pnpgrpcserver.WithFxPrivate` - All module provisions will be private. This allows you to to call the module multiple times in your application
- `pnpgrpcserver.WithConfigFromContainer()` - The module will use the pnpgrpcserver.Config from your provider
- `pnpgrpcserver.WithConfigPrefix(prefix string)` - The module will use a custom environment variable prefix to load pnpgrpcserver.Config
- `pnpgrpcserver.WithReflection()` - Registers the reflection service on the gRPC server
- `pnpgrpcserver.WithServerOptions(opts ...grpc.ServerOption)` - Provides custom gRPC server options

### Middlewares provisioning
You can add middlewares to your gRPC server by providing them to the fx container. 
```go
package main
import (
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
)
func main() {
	fx.New( 
		//< your options > 
		pnpgrpcserver.Module(),
		fx.Provide(
			// Unary server interceptor
			pnpgrpcserver.UnaryInterceptorProvider(func() ordering.OrderedItem[grpc.UnaryServerInterceptor] {
				return ordering.Ordered[grpc.UnaryServerInterceptor](
					0, // Order of execution
					func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
						fmt.Println("Hello from unary interceptor")
	
						return handler(ctx, req)
					},
				)
			}),
			// Stream server interceptor
			pnpgrpcserver.StreamInterceptorProvider(func() ordering.OrderedItem[grpc.StreamServerInterceptor] {
				return ordering.Ordered[grpc.StreamServerInterceptor](
					0, // Order of execution
					func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
						fmt.Println("Hello from stream interceptor")

						return handler(srv, ss)
					})
			}),
		),
	)
}
```
