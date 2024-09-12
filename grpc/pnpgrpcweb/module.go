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
	if options.serveGRPC && options.useMux && options.muxPrefix != "/" {
		panic("pnpgrpcweb: WithServeGRPC and WithMuxHandler must not be used together")
	}

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(WrapGRPCServer)

	if options.useMux {
		builder.ProvideIf(options.muxPrefix != "/", GRPCWebOptionProvider(func() grpcweb.Option {
			return grpcweb.WithAllowNonRootResource(true)
		}))
		builder.Provide(pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrarProvider(options)))
	} else {
		builder.Provide(NewHTTPHandlerProvider(options))
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
	return grpcweb.WrapServer(params.Server, params.Options...)
}

func NewMuxHandlerRegistrarProvider(options *options) func(*grpc.Server, *grpcweb.WrappedGrpcServer) pnphttpserver.MuxHandlerRegistrar {
	return func(server *grpc.Server, wrappedServer *grpcweb.WrappedGrpcServer) pnphttpserver.MuxHandlerRegistrar {
		handler := NewHTTPHandler(server, wrappedServer, options.serveGRPC)

		return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
			mux.PathPrefix(options.muxPrefix).Handler(handler)
		})
	}
}

func NewHTTPHandlerProvider(options *options) func(*grpc.Server, *grpcweb.WrappedGrpcServer) http.Handler {
	return func(server *grpc.Server, wrappedServer *grpcweb.WrappedGrpcServer) http.Handler {
		return NewHTTPHandler(server, wrappedServer, options.serveGRPC)
	}
}

func NewHTTPHandler(server *grpc.Server, wrappedServer *grpcweb.WrappedGrpcServer, serveGRPC bool) http.Handler {
	if serveGRPC {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wrappedServer.IsGrpcWebRequest(r) {
				wrappedServer.ServeHTTP(w, r)
			} else {
				server.ServeHTTP(w, r)
			}
		})
	}

	return wrappedServer
}
