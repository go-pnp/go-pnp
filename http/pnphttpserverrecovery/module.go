package pnphttpserverrecovery

import (
	"context"
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
	moduleBuilder.SupplyIf(!options.panicHandlerFromContainer, options.panicHandler)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

type NewMiddlewareParams struct {
	fx.In

	Options      *options
	PanicHandler PanicHandler
	Logger       *logging.Logger `optional:"true"`
}

func newMiddleware(params NewMiddlewareParams) ordering.OrderedItem[pnphttpserver.HandlerMiddleware] {
	return ordering.OrderedItem[pnphttpserver.HandlerMiddleware]{
		Value: func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func(ctx context.Context) {
					if panicValue := recover(); panicValue != nil {
						if panicValue != http.ErrAbortHandler { //nolint:errorlint // net/http compares ErrAbortHandler with ==
							params.Logger.WithError(panicutil.NewError(panicValue)).Error(ctx, "http handler panicked")
						}

						params.PanicHandler(w, panicValue)
					}
				}(r.Context())

				handler.ServeHTTP(w, r)
			})
		},
		Order: params.Options.order,
	}
}
