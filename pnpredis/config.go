package pnpredis

import "github.com/go-pnp/go-pnp/tls/tlsutil"

type Config struct {
	Sentinel struct {
		Enabled  bool     `env:"ENABLED" envDefault:"false"`
		Addrs    []string `env:"ADDRESSES"`
		Master   string   `env:"MASTER_NAME"`
		Username string   `env:"USERNAME"`
		Password string   `env:"PASSWORD"`
	} `envPrefix:"SENTINEL_"`
	Address  string                  `env:"ADDRESS,notEmpty"`
	Username string                  `env:"USERNAME"`
	Password string                  `env:"PASSWORD"`
	DB       int                     `env:"DB" envDefault:"0"`
	TLS      tlsutil.ClientTLSConfig `envPrefix:"TLS_"`
}
