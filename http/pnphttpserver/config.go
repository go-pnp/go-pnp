package pnphttpserver

import (
	"time"

	"github.com/go-pnp/go-pnp/tls/tlsutil"
)

type Config struct {
	Addr              string        `env:"ADDR" envDefault:"0.0.0.0:8080"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" envDefault:"0s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"0s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" envDefault:"0s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" envDefault:"0s"`

	TLS tlsutil.TLSConfig `envPrefix:"TLS_"`
}
