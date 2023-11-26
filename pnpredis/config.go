package pnpredis

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	Address  string                  `env:"ADDRESS"`
	Username string                  `env:"USERNAME"`
	Password string                  `env:"PASSWORD"`
	DB       int                     `env:"DB"`
	TLS      tlsutil.ClientTLSConfig `envPrefix:"TLS_"`
}
