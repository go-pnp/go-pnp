package pnpzap

import (
	"strings"

	"go.uber.org/zap"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

func (c *Config) ZapAtomicLevel() zap.AtomicLevel {
	switch strings.TrimSpace(strings.ToLower(c.LogLevel)) {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	return zap.NewAtomicLevelAt(zap.InfoLevel)
}
