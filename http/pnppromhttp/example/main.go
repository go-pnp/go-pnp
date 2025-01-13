package main

import (
	"net/http"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnppromhttp"
)

func main() {
	fx.New(
		pnpprometheus.Module(),
		pnphttpserver.Module(),
		pnppromhttp.Module(),
		fx.Provide(
			pnphttpserver.MuxHandlerRegistrarProvider(func() pnphttpserver.MuxHandlerRegistrarFunc {
				return func(mux *mux.Router) {
					router := mux.PathPrefix("/hello").Subrouter()
					router.Path("/hello/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					})
				}
			}),
		),
	).Run()
}
