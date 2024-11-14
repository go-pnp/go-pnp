package main

import (
	"net/http"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnphttpservercors"
)

func main() {
	fx.New(
		pnphttpserver.Module(),
		pnphttpservercors.Module(
			pnphttpservercors.WithOrder(1),              // middleware order
			pnphttpservercors.WithDisabledWarningLogs(), // disable logs if someone tries to request from not allowed origin
			pnphttpservercors.WithConfigFromContainer(),
		),
		fx.Provide(
			func() *pnphttpservercors.Config {
				return &pnphttpservercors.Config{
					AllowAll: true,
				}
			},
			pnphttpserver.MuxHandlerRegistrarProvider(func() pnphttpserver.MuxHandlerRegistrar {
				return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
					mux.Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("Hello"))
					}))
				})
			}),
		),
	).Run()
}
