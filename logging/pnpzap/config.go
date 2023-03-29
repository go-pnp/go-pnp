package pnpzap

import (
	"strings"

	"go.uber.org/zap"
)

type Config struct {
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
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

func (c *Config) EnvironmentConfig() zap.Config {
	switch strings.TrimSpace(strings.ToLower(c.Environment)) {
	case "d", "dev", "development":
		config := zap.NewDevelopmentConfig()
		config.Level = c.ZapAtomicLevel()

		return config
	case "p", "prod", "production":
		config := zap.NewProductionConfig()
		config.Level = c.ZapAtomicLevel()

		return config

	}

	return zap.NewProductionConfig()
}
