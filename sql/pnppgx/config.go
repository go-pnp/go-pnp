package pnppgx

import "time"

type Config struct {
	DSN string `env:"DSN,notEmpty,expand"`

	MaxOpenConnections    int           `env:"MAX_OPEN_CONNECTIONS" envDefault:"5"`
	MaxIdleConnections    int           `env:"MAX_IDLE_CONNECTIONS" envDefault:"5"`
	MaxConnectionLifetime time.Duration `env:"MAX_CONNECTION_LIFE_TIME" envDefault:"1h"`
	MaxConnectionIdleTime time.Duration `env:"MAX_CONNECTION_IDLE_TIME" envDefault:"30m"`
}
