package pnpprometheus

import (
	"context"
	"fmt"
	"github.com/go-pnp/go-pnp/logging"
	"go.uber.org/fx"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func (s *Server) Start(listenPort int) error {
	metricsListener, err := net.Listen("tcp", ":"+strconv.Itoa(listenPort))
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
	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info(ctx, fmt.Sprintf("Running metrics server on 0.0.0.0:%d", params.Config.Port))

			go func() {
				if err := params.Server.Start(params.Config.Port); err != nil && err != http.ErrServerClosed {
					params.RuntimeErr <- err
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info(ctx, "Metrics server shutting down...")

			return params.Server.Shutdown(ctx)
		},
	})
}
