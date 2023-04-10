package pnpprometheus

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging"
)

type Server struct {
	srv *http.Server
}

func NewServer(registry *prometheus.Registry, config *Config) *Server {
	mux := http.NewServeMux()
	srv := &http.Server{
		Handler: mux,
	}

	mux.Handle(config.Path, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	return &Server{
		srv: srv,
	}
}

func (s *Server) Start(listenAddr string) error {
	metricsListener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	return s.srv.Serve(metricsListener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

type RegisterServerStartHooksParams struct {
	fx.In
	Lc         fx.Lifecycle
	RuntimeErr chan<- error
	Logger     *logging.Logger `optional:"true"`
	Server     *Server
	Config     *Config
}

func RegisterServerStartHooks(params RegisterServerStartHooksParams) {
	logger := params.Logger.Named("prometheus.metrics_exporter").WithField("addr", params.Config.Addr)
	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info(ctx, "Starting metrics server")

			go func() {
				if err := params.Server.Start(params.Config.Addr); err != nil && err != http.ErrServerClosed {
					logger.WithError(err).Error(ctx, "Error starting metrics server")
					params.RuntimeErr <- err
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info(ctx, "Stopping metrics server...")

			if err := params.Server.Shutdown(ctx); err != nil {
				params.Logger.WithError(err).Error(ctx, "Error shutting down metrics server")

				return err
			}

			params.Logger.Info(ctx, "Metrics server stopped")
			return nil
		},
	})
}
