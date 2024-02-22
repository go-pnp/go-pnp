package pnpfiber

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	Addr string                  `env:"ADDR" envDefault:":8080"`
	TLS  tlsutil.ServerTLSConfig `env:"TLS_"`
}
