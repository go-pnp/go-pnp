package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pnpenv"
	"github.com/go-pnp/go-pnp/pnpsentry"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnphttpservercors"
)

func main() {
	os.Setenv("SENTRY_DSN", "YOUR_DSN_HERE")
	os.Setenv("SENTRY_ENVIRONMENT", "local")
	os.Setenv("SENTRY_ENABLE_TRACING", "true")
	fx.New(
		pnpenv.Module(),
		pnpzap.Module(),
		pnpsentry.Module(),
		pnphttpserver.Module(pnphttpserver.Start(true)),
		pnphttpserversentry.Module(
			pnphttpserversentry.WithOrder(1), // middleware order
		),
		fx.Provide(
			pnphttpserver.MuxHandlerRegistrarProvider(func(sentryClient *sentry.Client) pnphttpserver.MuxHandlerRegistrar {
				return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
					mux.Methods("GET").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						hub := sentry.GetHubFromContext(r.Context())
						hub.CaptureException(errors.New("smth bad happened"))
						w.Write([]byte("Hello"))
					}))
				})
			}),
		),
	).Run()
}
