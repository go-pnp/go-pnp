//go:build example
// +build example

package pnphttpserver

import (
	"net/http"
	"testing"

	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type Handler struct {
}

func (h Handler) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("World1"))
}

func (h Handler) Register(mux *mux.Router) {
	mux.Path("/hello").HandlerFunc(h.Hello)
}

func TestApp(t *testing.T) {
	fx.New(
		Module(
			WithFxPrivate(),
			Start(true),
		),

		// Register our application endpoints
		fx.Provide(
			MuxHandlerRegistrarProvider(func() MuxHandlerRegistrar {
				return Handler{}
			}),
			MuxMiddlewareFuncProvider(func() ordering.OrderedItem[mux.MiddlewareFunc] {
				return ordering.OrderedItem[mux.MiddlewareFunc]{
					Order: 10,
					Value: func(mux http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Write([]byte("Hello from middleware\n"))

							mux.ServeHTTP(w, r)
						})
					},
				}
			}),
			fx.Private,
		),

		// Register middleware
	).Run()
}
