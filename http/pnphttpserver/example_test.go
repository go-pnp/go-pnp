package pnphttpserver

import (
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
	"net/http"
	"testing"
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
		Module(),
		fx.Supply(Handler{}),

		// Register our application endpoints
		ProvideMuxHandlerRegistrar(func(handler Handler) MuxHandlerRegistrar {
			return handler.RegisterEndpoints
		}),

		// Register middleware
		ProvideMuxMiddlewareFunc(func() mux.MiddlewareFunc {
			return func(mux http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello from middleware\n"))

					mux.ServeHTTP(w, r)
				})
			}
		}),
	)
}
