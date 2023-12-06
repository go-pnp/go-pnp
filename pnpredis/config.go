package pnpredis

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	Address  string                  `env:"ADDRESS,notEmpty"`
	Username string                  `env:"USERNAME"`
	Password string                  `env:"PASSWORD"`
	DB       int                     `env:"DB" envDefault:"0"`
	TLS      tlsutil.ClientTLSConfig `envPrefix:"TLS_"`
}
