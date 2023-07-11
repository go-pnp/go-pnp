package pnpgormprometheus

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/go-pnp/go-pnp/logging"
)

type DBStatsPlugin struct {
	logger  *logging.Logger
	db      *gorm.DB
	dbStats *DBStats
	quitCh  chan struct{}
}

func NewDBStatsPlugin(logger *logging.Logger, dbStats *DBStats) *DBStatsPlugin {
	return &DBStatsPlugin{
		logger:  logger,
		dbStats: dbStats,
		quitCh:  make(chan struct{}),
	}
}

func (d *DBStatsPlugin) Name() string {
	return "pnpgormprometheus"
}

func (d *DBStatsPlugin) Initialize(db *gorm.DB) error {
	d.db = db
	return nil
}

func (d *DBStatsPlugin) Run() error {
	if d.db == nil {
		return errors.Errorf("db is nil")
	}
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			db, err := d.db.DB()
			if err != nil {
				return fmt.Errorf("can't get underlying db: %w", err)
			}

			d.dbStats.Set(db.Stats())
		case <-d.quitCh:
			return nil
		}
	}
}
func (d *DBStatsPlugin) Close() error {
	select {
	case d.quitCh <- struct{}{}:
	case <-time.After(time.Second * 3):
		return errors.New("can't close db stats plugin: timeout")
	}

	return nil
}
