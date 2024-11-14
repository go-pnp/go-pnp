package main

import (
	"net/http"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnphttpserverh2c"
)

func main() {
	fx.New(
		pnphttpserver.Module(),
		pnphttpserverh2c.Module(), // This adds h2c support to the server
		fx.Provide(
			pnphttpserver.MuxHandlerRegistrarProvider(func() pnphttpserver.MuxHandlerRegistrar {
				return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
					mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("Hello, World!"))
					})
				})
			}),
		),
	).Run()
}
