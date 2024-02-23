package pnppgx

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm/logger"

	"github.com/go-pnp/go-pnp/logging"
)

type Logger struct {
	Delegate *logging.Logger
	LevelMu  sync.RWMutex
	Level    logger.LogLevel
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.LevelMu.Lock()
	l.Level = level
	l.LevelMu.Unlock()

	return l
}

func (l *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.Level < logger.Info {
		return
	}
	l.Delegate.Info(ctx, s, i...)
}

func (l *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.Level < logger.Warn {
		return
	}
	l.Delegate.Warn(ctx, s, i...)
}

func (l *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.Level < logger.Error {
		return
	}
	l.Delegate.Error(ctx, s, i...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)

	l.Delegate.WithFields(map[string]interface{}{
		"sql":           sql,
		"rows_affected": rows,
		"elapsed":       elapsed,
	}).Debug(ctx, "sql query")
}

var _ logger.Interface = &Logger{}
