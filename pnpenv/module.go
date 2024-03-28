package pnpenv

import (
	"strings"

	"github.com/caarlos0/env/v10"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
)

type Environment string

func (e Environment) IsOneOfCI(values ...string) bool {
	elc := strings.TrimSpace(strings.ToLower(string(e)))
	for _, val := range values {
		if strings.TrimSpace(strings.ToLower(val)) == elc {
			return true
		}
	}

	return false
}
func (e Environment) IsDev() bool {
	return e.IsOneOfCI("dev", "deveopment", "d")
}

func (e Environment) IsProd() bool {
	return e.IsOneOfCI("prod", "production", "p", "prd")
}

func (e Environment) IsTest() bool {
	return e.IsOneOfCI("test", "t", "tst")
}

type Config struct {
	Environment Environment `env:"ENVIRONMENT,notEmpty" envDefault:"development"`
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			configutil.NewConfigProvider[Config](env.Options{}),
			configutil.NewConfigInfoProvider[Config](env.Options{}),
			NewEnvironment,
		),
	)
}

func NewEnvironment(config *Config) Environment {
	return config.Environment
}
