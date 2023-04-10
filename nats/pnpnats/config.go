package pnpnats

import (
	"time"

	"github.com/nats-io/nats.go"

	"github.com/go-pnp/go-pnp/tls/tlsutil"
)

type Config struct {
	Addr       string                  `env:"ADDR" envDefault:"127.0.0.1:443"`
	TLS        tlsutil.ClientTLSConfig `envPrefix:"TLS_"`
	Reconnects struct {
		Max   int           `env:"MAX_COUNT" envDefault:"-1"`
		Allow bool          `env:"ALLOW" envDefault:"true"`
		Wait  time.Duration `env:"WAIT" envDefault:"500ms"`
	} `envPrefix:"RECONNECTS_"`
}

func (c *Config) getReconnectOptions() []nats.Option {
	if c.Reconnects.Allow {
		return []nats.Option{
			nats.ReconnectWait(c.Reconnects.Wait),
			nats.MaxReconnects(c.Reconnects.Max),
		}
	}
	return []nats.Option{nats.NoReconnect()}
}

func (c *Config) getTLSOptions() ([]nats.Option, error) {
	if c.TLS.Enabled {
		tlsConfig, err := c.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}
		return []nats.Option{nats.Secure(tlsConfig)}, nil
	}

	return nil, nil
}
