package pnpgormprometheus

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type DBStats struct {
	MaxOpenConnections prometheus.Gauge // Maximum number of open connections to the database.

	// Pool status
	OpenConnections prometheus.Gauge // The number of established connections both in use and idle.
	InUse           prometheus.Gauge // The number of connections currently in use.
	Idle            prometheus.Gauge // The number of idle connections.

	// Counters
	WaitCount         prometheus.Gauge // The total number of connections waited for.
	WaitDuration      prometheus.Gauge // The total time blocked waiting for a new connection.
	MaxIdleClosed     prometheus.Gauge // The total number of connections closed due to SetMaxIdleConns.
	MaxLifetimeClosed prometheus.Gauge // The total number of connections closed due to SetConnMaxLifetime.
	MaxIdleTimeClosed prometheus.Gauge // The total number of connections closed due to SetConnMaxIdleTime.
}

func NewDBStats() *DBStats {
	return &DBStats{
		MaxOpenConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_max_open_connections",
			Help: "Maximum number of open connections to the database.",
		}),
		OpenConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_open_connections",
			Help: "The number of established connections both in use and idle.",
		}),
		InUse: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_in_use",
			Help: "The number of connections currently in use.",
		}),
		Idle: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_idle",
			Help: "The number of idle connections.",
		}),
		WaitCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_wait_count",
			Help: "The total number of connections waited for.",
		}),
		WaitDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_wait_duration",
			Help: "The total time blocked waiting for a new connection.",
		}),
		MaxIdleClosed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_max_idle_closed",
			Help: "The total number of connections closed due to SetMaxIdleConns.",
		}),
		MaxLifetimeClosed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_max_lifetime_closed",
			Help: "The total number of connections closed due to SetConnMaxLifetime.",
		}),
		MaxIdleTimeClosed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gorm_dbstats_max_idletime_closed",
			Help: "The total number of connections closed due to SetConnMaxIdleTime.",
		}),
	}
}

func (d *DBStats) Describe(descs chan<- *prometheus.Desc) {
	descs <- d.MaxOpenConnections.Desc()
	descs <- d.OpenConnections.Desc()
	descs <- d.InUse.Desc()
	descs <- d.Idle.Desc()
	descs <- d.WaitCount.Desc()
	descs <- d.WaitDuration.Desc()
	descs <- d.MaxIdleClosed.Desc()
	descs <- d.MaxLifetimeClosed.Desc()
	descs <- d.MaxIdleTimeClosed.Desc()

}

func (d *DBStats) Collect(metrics chan<- prometheus.Metric) {
	d.MaxOpenConnections.Collect(metrics)
	d.OpenConnections.Collect(metrics)
	d.InUse.Collect(metrics)
	d.Idle.Collect(metrics)
	d.WaitCount.Collect(metrics)
	d.WaitDuration.Collect(metrics)
	d.MaxIdleClosed.Collect(metrics)
	d.MaxLifetimeClosed.Collect(metrics)
	d.MaxIdleTimeClosed.Collect(metrics)
}

func (d *DBStats) Set(dbStats sql.DBStats) {
	d.MaxOpenConnections.Set(float64(dbStats.MaxOpenConnections))
	d.OpenConnections.Set(float64(dbStats.OpenConnections))
	d.InUse.Set(float64(dbStats.InUse))
	d.Idle.Set(float64(dbStats.Idle))
	d.WaitCount.Set(float64(dbStats.WaitCount))
	d.WaitDuration.Set(float64(dbStats.WaitDuration))
	d.MaxIdleClosed.Set(float64(dbStats.MaxIdleClosed))
	d.MaxLifetimeClosed.Set(float64(dbStats.MaxLifetimeClosed))
	d.MaxIdleTimeClosed.Set(float64(dbStats.MaxIdleTimeClosed))
}
