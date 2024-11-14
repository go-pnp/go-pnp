package pnphttpserverrecovery

import (
	"net/http"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.SupplyIf(!options.panicHandlerFromContainer, options.panicHandler)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

type NewMiddlewareParams struct {
	fx.In
	Options      *options
	PanicHandler PanicHandler
}

func newMiddleware(params NewMiddlewareParams) ordering.OrderedItem[pnphttpserver.HandlerMiddleware] {
	return ordering.OrderedItem[pnphttpserver.HandlerMiddleware]{
		Value: func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if r := recover(); r != nil {
						params.PanicHandler(w, r)
					}
				}()

				handler.ServeHTTP(w, r)
			})
		},
		Order: params.Options.order,
	}
}
