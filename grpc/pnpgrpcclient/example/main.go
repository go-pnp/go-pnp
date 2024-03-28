package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-pnp/go-pnp/pkg/ordering"

	"github.com/go-pnp/go-pnp/grpc/pnpgrpcclient"
)

type Some interface {
	Do()
}

type SomeImpl struct{}

func (s *SomeImpl) Do() {
	fmt.Println("do something")
}

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
		fx.Invoke(fx.Annotate(func(some []Some) {
			fmt.Println(some)
		}, fx.ParamTags(`group:some`))),
		fx.Provide(
			fx.Annotated{
				Target: fx.Annotate(func() SomeImpl {
					return SomeImpl{}
				}, fx.As(new(Some))),
				Group: "some",
			},
			pnpgrpcclient.DialOptionProvider(func() grpc.DialOption {
				return grpc.WithBlock()
			}),
			pnpgrpcclient.UnaryClientInterceptorProvider(func() ordering.OrderedItem[grpc.UnaryClientInterceptor] {
				return ordering.Ordered[grpc.UnaryClientInterceptor](
					0,
					func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						fmt.Println("unary interceptor from provider")
						return invoker(ctx, method, req, reply, cc, opts...)
					},
				)
			}),
			pnpgrpcclient.StreamClientInterceptorProvider(func() ordering.OrderedItem[grpc.StreamClientInterceptor] {
				return ordering.Ordered[grpc.StreamClientInterceptor](
					0,
					func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
						fmt.Println("stream interceptor from provider")
						return streamer(ctx, desc, cc, method, opts...)
					},
				)
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
