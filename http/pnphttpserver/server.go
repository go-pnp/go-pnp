package pnphttpserver

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
)

type NewServerParams struct {
	fx.In

	Config  *Config
	Handler http.Handler
	CORS    *cors.Cors `optional:"true"`
}

func NewServer(params NewServerParams) (*http.Server, error) {
	tlsConfig, err := params.Config.TLS.TLSConfig()
	if err != nil {
		return nil, err
	}

	handler := params.Handler
	if params.CORS != nil {
		handler = params.CORS.Handler(handler)
	}

	return &http.Server{
		Addr:              params.Config.Addr,
		ReadTimeout:       params.Config.ReadTimeout,
		ReadHeaderTimeout: params.Config.ReadHeaderTimeout,
		WriteTimeout:      params.Config.WriteTimeout,
		IdleTimeout:       params.Config.IdleTimeout,
		Handler:           handler,
		TLSConfig:         tlsConfig,
	}, nil
}

type MuxHandlerRegistrar interface {
	Register(mux *mux.Router)
}

type MuxHandlerRegistrarFunc func(mux *mux.Router)

func (f MuxHandlerRegistrarFunc) Register(mux *mux.Router) {
	f(mux)
}

func MuxHandlerRegistrarProvider(target any) any {
	return fxutil.GroupProvider[MuxHandlerRegistrar](
		"pnp_http_server.mux_handler_registrars",
		target,
	)
}
func MuxMiddlewareFuncProvider(target any) any {
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
		handlerRegistrar.Register(result)
	}

	return result
}

type RegisterStartHooksParams struct {
	fx.In
	Shutdowner fx.Shutdowner
	Lc         fx.Lifecycle
	Logger     *logging.Logger `optional:"true"`
	Config     *Config
	Server     *http.Server
}

func RegisterStartHooks(params RegisterStartHooksParams) {
	logger := params.Logger.WithFields(map[string]interface{}{
		"addr":        params.Config.Addr,
		"tls_enabled": params.Config.TLS.Enabled,
	})

	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info(ctx, "Starting HTTP server...")
				var err error
				if params.Server.TLSConfig != nil {
					err = params.Server.ListenAndServeTLS(params.Config.TLS.CertPath, params.Config.TLS.KeyPath)
				} else {
					err = params.Server.ListenAndServe()
				}
				if err != nil && err != http.ErrServerClosed {
					logger.WithError(err).Error(ctx, "Error starting HTTP server")
					params.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info(ctx, "Stopping HTTP server...")
			if err := params.Server.Shutdown(ctx); err != nil {
				logger.WithError(err).Error(ctx, "Error stopping HTTP server")
				return err
			}
			logger.Info(ctx, "HTTP server stopped")
			return nil
		},
	})
}
