package pnpsqlx

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
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

	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: "DB_",
	}))
	builder.Provide(NewSqlxDBProvider(driver))

	return builder.Build()
}

func NewSqlxDBProvider(driver string) func(config *Config) (*sqlx.DB, *sql.DB, error) {
	return func(config *Config) (*sqlx.DB, *sql.DB, error) {
		conn, err := sqlx.Open(driver, config.DSN)
		if err != nil {
			return nil, nil, err
		}

		return conn, conn.DB, nil
	}
}
