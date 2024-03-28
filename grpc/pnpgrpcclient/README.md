## pnpgrpcclient

This package provides a gRPC client dialer which will automatically add options and interceptors from the fx container.

### Getting started

Install pnpgrpcclient package:

```shell
go get github.com/go-pnp/go-pnp/grpc/pnpgrpcclient
````

Add pnpgrpcclient.Module() to your fx application options and you are ready to use it by requesting *
pnpgrpcclient.Dialer from the fx container:

```go
package main

import "github.com/go-pnp/go-pnp/grpc/pnpgrpcclient"
import "go.uber.org/fx"

func main() {
	fx.New(
		//< your options > 
		pnpgrpcclient.Module(),
		// other modules, for example pnpgrrpcprometheus
		fx.Provide(func(dialer *pnpgrpcclient.Dialer) (SomeClient, error) {
			conn, _ := dialer.Dial("127.0.0.1:50051") // Here all provided options and interceptors will be added
			return NewSomeClient(conn), nil
		}),
	)
}
```

### Module Options

- `pnpgrpcclient.WithFxPrivate()` - All module provisions will be private. This allows you to to call the module
  multiple
  times in your application
- `WithDialOptions(dialOptions ...grpc.DialOption)` - Provides custom gRPC dial options during module initialization
- `WithUnaryInterceptor(order int, interceptor grpc.UnaryClientInterceptor)` - Provides custom unary client interceptor
  during module initialization
- `WithStreamInterceptor(order int, interceptor grpc.StreamClientInterceptor)` - Provides custom stream client
  interceptor during module initialization

### Extending dialer

You can add middlewares or dial options to Dialer by providing them to the fx container.

```go
package main

import (
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcclient"
	"github.com/go-pnp/go-pnp/grpc/pnpgrpcserver"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		//< your options > 
		pnpgrpcclient.Module(),
		fx.Provide(
			// Common dial options
			pnpgrpcclient.NewDialOptionProvider(func() grpc.DialOption {
				return grpc.WithBlock()
			}),
			// Common unary client interceptor
			pnpgrpcclient.NewUnaryClientInterceptorProvider(func() ordering.OrderedItem[grpc.UnaryClientInterceptor] {
				return ordering.Ordered[grpc.UnaryClientInterceptor](
					0,
					func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						fmt.Println("interceptor from provider")
						return invoker(ctx, method, req, reply, cc, opts...)
					},
				)
			}),
			// Common stream client interceptor
			pnpgrpcclient.NewStreamClientInterceptorProvider(func() ordering.OrderedItem[grpc.UnaryClientInterceptor] {
				return ordering.Ordered[grpc.UnaryClientInterceptor](
					0,
					func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						fmt.Println("interceptor from provider")
						return invoker(ctx, method, req, reply, cc, opts...)
					},
				)
			}),
		),
	)
}
```
