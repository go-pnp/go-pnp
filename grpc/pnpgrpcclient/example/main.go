package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-pnp/go-pnp/grpc/pnpgrpcclient"
)

func main() {
	fx.New(
		pnpgrpcclient.Module(
			pnpgrpcclient.WithDialOptions(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			),
			pnpgrpcclient.WithUnaryInterceptor(1, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				fmt.Println("interceptor from option")
				return invoker(ctx, method, req, reply, cc, opts...)
			}),
		),
		fx.Invoke(func(dialer *pnpgrpcclient.Dialer) {
			conn, err := dialer.Dial("localhost:50051")
			if err != nil {
				fmt.Println(err)
			}
			client := NewTestServiceClient(conn)
			fmt.Println(client.Test(context.Background(), &TestRequest{}))
		}),
	).Run()
}
