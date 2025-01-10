package pnpconnectrpchandling

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

// Module provides an interface to easily register connectrpc handlers in a http server mux.
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlersRegistrar))

	return builder.Build()
}

func ConnectHandlerConstructorProvider[T any](
	originalConstructor func(handler T, opts ...connect.HandlerOption) (string, http.Handler),
	handlerOpts ...connect.HandlerOption,
) any {
	return fxutil.GroupProvider[ConnectHandlerConstructor]("pnpconnectrpchandling.handler_constructors", func(handler T) ConnectHandlerConstructor {
		return func(opts ...connect.HandlerOption) (string, http.Handler) {
			allOpts := make([]connect.HandlerOption, 0, len(opts)+len(handlerOpts))
			allOpts = append(allOpts, opts...)
			allOpts = append(allOpts, handlerOpts...)
			return originalConstructor(handler, allOpts...)
		}
	})
}

type ConnectHandlerConstructor func(opt ...connect.HandlerOption) (string, http.Handler)

func InterceptorProvider(target any) any {
	return fxutil.GroupProvider[connect.Interceptor]("pnpconnectrpchandling.interceptors", target)
}

func HandlerOptionProvider(target any) any {
	return fxutil.GroupProvider[connect.HandlerOption]("pnpconnectrpchandling.handler_options", target)
}

type NewMuxHandlersRegistrarParams struct {
	fx.In
	Interceptors        ordering.OrderedItems[connect.Interceptor] `group:"pnpconnectrpchandling.interceptors"`
	Options             []connect.HandlerOption                    `group:"pnpconnectrpchandling.handler_options"`
	HandlerConstructors []ConnectHandlerConstructor                `group:"pnpconnectrpchandling.handler_constructors"`
}

func NewMuxHandlersRegistrar(params NewMuxHandlersRegistrarParams) pnphttpserver.MuxHandlerRegistrar {
	return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
		for _, constructor := range params.HandlerConstructors {
			handlerOpts := make([]connect.HandlerOption, 0, len(params.Options))
			handlerOpts = append(handlerOpts, params.Options...)
			handlerOpts = append(handlerOpts, connect.WithInterceptors(params.Interceptors.Get()...))

			path, handler := constructor(handlerOpts...)
			mux.PathPrefix(path).Handler(handler)
		}
	})
}
