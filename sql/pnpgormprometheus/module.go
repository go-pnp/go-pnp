package pnpgormprometheus

import (
	"context"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/go-pnp/go-pnp/sql/pnpgorm"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		NewDBStats,
		NewDBStatsPlugin,
		pnpgorm.PluginProvider(func(lc fx.Lifecycle, logger *logging.Logger, shutdowner fx.Shutdowner, p *DBStatsPlugin) gorm.Plugin {
			// Hook registered here to prevent starting the plugin if no plugin used(If it will be in Invoke - it will always start)
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						logger.Debug(context.Background(), "starting db stats plugin")
						if err := p.Run(); err != nil {
							logger.Error(context.Background(), "db stats plugin error", err)
							shutdowner.Shutdown()
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					return p.Close()
				},
			})
			return p
		}),
		pnpprometheus.MetricsCollectorProvider(NewDBStatsPrometheusCollectors),
	)

	return builder.Build()
}

func NewDBStatsPrometheusCollectors(dbStats *DBStats) prometheus.Collector {
	return dbStats
}
