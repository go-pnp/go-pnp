package main

import (
	"net/http"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnphttpserverrecovery"
)

func main() {
	fx.New(
		pnphttpserver.Module(),
		pnphttpserverrecovery.Module(),
		fx.Provide(
			pnphttpserver.MuxHandlerRegistrarProvider(func() pnphttpserver.MuxHandlerRegistrar {
				return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
					mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						panic("Panic!")
					})
				})
			}),
		),
	).Run()
}
