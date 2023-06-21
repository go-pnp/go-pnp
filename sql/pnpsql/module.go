package pnpsql

import (
	"database/sql"
	"fmt"

	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(driver string, opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		configPrefix: "DB_",
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(NewSqlDBProvider(driver, options.configPrefix))
	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: options.configPrefix,
	}))

	return builder.Build()
}

func NewSqlDBProvider(driver, configPrefix string) func(config *Config) (*sql.DB, error) {
	return func(config *Config) (*sql.DB, error) {
		if config.DSN == "" {
			return nil, fmt.Errorf("please, provide %s database source name to %sDSN env variable", driver, configPrefix)
		}
		return sql.Open(driver, config.DSN)
	}
}
