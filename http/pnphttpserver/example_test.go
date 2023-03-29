//go:build example
// +build example

package pnphttpserver

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

type Handler struct {
}

func (h Handler) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("World1"))
}

func (h Handler) RegisterEndpoints(mux *mux.Router) {
	mux.Path("/hello").HandlerFunc(h.Hello)
}

func TestApp(t *testing.T) {
	fxutil.StartApp(
		Module(
			WithFxPrivate(),
		),
		fx.Supply(Handler{}),

		// Register our application endpoints
		fx.Provide(
			MuxHandlerRegistrarProvider(func(handler Handler) MuxHandlerRegistrar {
				return handler.RegisterEndpoints
			}),
			MuxMiddlewareFuncProvider(func() mux.MiddlewareFunc {
				return func(mux http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("Hello from middleware\n"))

						mux.ServeHTTP(w, r)
					})
				}
			}),
			fx.Private,
		),

		// Register middleware
	)
}
