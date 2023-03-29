package pnphttpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-pnp/go-pnp/tls/tlsutil"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

func NewServer(
	config *Config,
	handler http.Handler,
) (*http.Server, error) {
	tlsConfig, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              config.Addr,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		Handler:           handler,
		TLSConfig:         tlsConfig,
	}, nil
}

type MuxHandlerRegistrar func(mux *mux.Router)

func MuxHandlerRegistrarProvider(target any) fx.Annotated {
	return fxutil.GroupProvider[MuxHandlerRegistrar](
		"pnp_http_server.mux_handler_registrars",
		target,
	)
}
func MuxMiddlewareFuncProvider(target any) fx.Annotated {
	return fxutil.GroupProvider[mux.MiddlewareFunc](
		"pnp_http_server.mux_middleware_funcs",
		target,
	)
}

type NewMuxParams struct {
	fx.In
	Middlewares       []mux.MiddlewareFunc  `group:"pnp_http_server.mux_middleware_funcs"`
	HandlerRegistrars []MuxHandlerRegistrar `group:"pnp_http_server.mux_handler_registrars"`
}

func NewMux(params NewMuxParams) http.Handler {
	result := mux.NewRouter()
	result.Use(params.Middlewares...)

	for _, handlerRegistrar := range params.HandlerRegistrars {
		handlerRegistrar(result)
	}

	return result
}

type RegisterStartHooksParams struct {
	fx.In
	RuntimeErrors chan error
	Lc            fx.Lifecycle
	//Logger        *logfx.Logger
	Config *Config
	Server *http.Server
}

func RegisterStartHooks(params RegisterStartHooksParams) {
	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				var err error
				if params.Server.TLSConfig != nil {
					err = params.Server.ListenAndServeTLS(params.Config.TLS.TLSCertPath, params.Config.TLS.TLSKeyPath)
				} else {
					err = params.Server.ListenAndServe()
				}
				if err != nil && err != http.ErrServerClosed {
					params.RuntimeErrors <- err
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return params.Server.Shutdown(ctx)
		},
	})
}
