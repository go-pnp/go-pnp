package pnpsql

import (
	"database/sql"

	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(driver string, opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Provide(NewSqlDBProvider(driver))
	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: "DB_",
	}))

	return builder.Build()
}

func NewSqlDBProvider(driver string) func(config *Config) (*sql.DB, error) {
	return func(config *Config) (*sql.DB, error) {
		return sql.Open(driver, config.DSN)
	}
}
