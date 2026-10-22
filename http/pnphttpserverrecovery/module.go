package pnphttpserverrecovery

import (
	"net/http"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/go-pnp/pkg/panicutil"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.SupplyIf(!options.panicHandlerFromContainer && options.panicHandler != nil, options.panicHandler)
	moduleBuilder.ProvideIf(!options.panicHandlerFromContainer && options.panicHandler == nil, newDefaultPanicHandler)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

type NewDefaultPanicHandlerParams struct {
	fx.In

	Logger *logging.Logger `optional:"true"`
}

// newDefaultPanicHandler logs the panic with its stack (http.ErrAbortHandler is not logged) and responds with 500.
func newDefaultPanicHandler(params NewDefaultPanicHandlerParams) PanicHandler {
	return func(w http.ResponseWriter, r *http.Request, panicValue any) {
		if panicValue != http.ErrAbortHandler { //nolint:errorlint // net/http compares ErrAbortHandler with ==
			params.Logger.WithError(panicutil.NewError(panicValue)).Error(r.Context(), "http handler panicked")
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
					if panicValue := recover(); panicValue != nil {
						params.PanicHandler(w, r, panicValue)
					}
				}()

				handler.ServeHTTP(w, r)
			})
		},
		Order: params.Options.order,
	}
}
