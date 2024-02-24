package pnpfiber

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	Addr string                  `env:"ADDR,notEmpty" envDefault:":8080"`
	TLS  tlsutil.ServerTLSConfig `envPrefix:"TLS_"`
}
