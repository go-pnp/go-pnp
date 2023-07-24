package pnpgrpcweb

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"

	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"github.com/go-pnp/go-pnp/fxutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	if options.useMux {
		builder.Provide(WrapGRPCServer)
		builder.ProvideIf(options.muxPrefix != "/", GRPCWebOptionProvider(func() grpcweb.Option {
			return grpcweb.WithAllowNonRootResource(true)
		}))
		builder.Provide(pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrarProvider(options)))
	} else {
		builder.Provide(fx.Annotate(WrapGRPCServer, fx.As(new(http.Handler))))
	}

	return builder.Build()
}

func GRPCWebOptionProvider(target any) any {
	return fxutil.GroupProvider[grpcweb.Option]("pnpgrpcweb.options", target)
}

type WrapGRPCServerParams struct {
	fx.In
	Server  *grpc.Server
	Options []grpcweb.Option `group:"pnpgrpcweb.options"`
}

func WrapGRPCServer(params WrapGRPCServerParams) *grpcweb.WrappedGrpcServer {
	return grpcweb.WrapHandler(params.Server, params.Options...)
}

func NewMuxHandlerRegistrarProvider(options *options) func(*grpcweb.WrappedGrpcServer) pnphttpserver.MuxHandlerRegistrar {
	return func(wrapper *grpcweb.WrappedGrpcServer) pnphttpserver.MuxHandlerRegistrar {
		return func(mux *mux.Router) {
			mux.PathPrefix(options.muxPrefix).Handler(wrapper)
		}
	}
}
