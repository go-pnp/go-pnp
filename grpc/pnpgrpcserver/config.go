package pnpgrpcserver

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	// prefixed tls config
	Addr string                  `env:"LISTEN_ADDR" envDefault:"127.0.0.1:50051"`
	TLS  tlsutil.ServerTLSConfig `envPrefix:"TLS_"`
}
