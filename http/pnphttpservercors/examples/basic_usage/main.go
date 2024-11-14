package main

import (
	"net/http"
	"os"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnphttpservercors"
)

func main() {
	os.Setenv("HTTP_SERVER_CORS_ALLOW_ALL_ORIGINS", "false")
	os.Setenv("HTTP_SERVER_CORS_ALLOWED_HEADERS", "*")
	os.Setenv("HTTP_SERVER_CORS_ALLOWED_ORIGIN_GLOBS", "*.example.com")
	os.Setenv("HTTP_SERVER_CORS_ALLOWED_ORIGINS", "some-domain.com")

	fx.New(
		pnphttpserver.Module(),
		pnphttpservercors.Module(
			pnphttpservercors.WithOrder(1),              // middleware order
			pnphttpservercors.WithDisabledWarningLogs(), // disable logs if someone tries to request from not allowed origin
		),
		fx.Provide(
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
