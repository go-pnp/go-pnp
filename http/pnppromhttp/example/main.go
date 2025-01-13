package main

import (
	"bytes"
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
		pnppromhttp.Module(
			pnppromhttp.WithEndpoint("GET", "/metrics"), // not required
			pnppromhttp.RegisterInMux(true),             // not required
		),
		fx.Provide(
			pnphttpserver.MuxHandlerRegistrarProvider(func() pnphttpserver.MuxHandlerRegistrarFunc {
				return func(mux *mux.Router) {
					router := mux.PathPrefix("/hello").Subrouter()
					router.Path("/hello/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write(bytes.Repeat([]byte{1, 2, 3}, 200))
					})
				}
			}),
		),
	).Run()
}
