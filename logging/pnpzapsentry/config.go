package pnpzapsentry

import (
	"time"

	"github.com/getsentry/sentry-go"
)

type Config struct {
	FlushTimeout time.Duration `env:"FLUSH_TIMEOUT" envDefault:"2s"`
	ReportLevel  sentry.Level  `env:"REPORT_LEVEL" envDefault:"error"`
}
