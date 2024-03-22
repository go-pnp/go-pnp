package pnpgormprometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"gorm.io/gorm"

	gormPrometheus "gorm.io/plugin/prometheus"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"github.com/go-pnp/go-pnp/sql/pnpgorm"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(
		NewDBStats,
		NewDBStatsPlugin,
		pnpgorm.PluginProvider(func(p *DBStatsPlugin) gorm.Plugin {
			return p
		}),
		pnpprometheus.MetricsCollectorProvider(NewDBStatsPrometheusCollectors),
	)
	builder.Invoke(RunDBStatsPlugin)

	return builder.Build()
}

func NewGormPrometheus() *gormPrometheus.Prometheus {
	return gormPrometheus.New(gormPrometheus.Config{})
}

func NewDBStatsPrometheusCollectors(dbStats *DBStats) prometheus.Collector {
	return dbStats
}

func RunDBStatsPlugin(shutdowner fx.Shutdowner, logger *logging.Logger, lc fx.Lifecycle, dbStatsPlugin *DBStatsPlugin) {
	logger = logger.Named("gorm_db_stats_plugin")
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Debug(context.Background(), "starting db stats plugin")
				if err := dbStatsPlugin.Run(); err != nil {
					logger.Error(context.Background(), "db stats plugin error", err)
					shutdowner.Shutdown()
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return dbStatsPlugin.Close()
		},
	})
}
