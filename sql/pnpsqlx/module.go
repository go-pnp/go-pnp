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
	options := newOptions(driver, opts)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Supply(options)
	if options.wrapSQLDB {
		builder.Provide(NewSQLxDBWrapper)
	} else {
		builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
		builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))
		builder.Provide(NewSqlxDBProvider)
	}

	return builder.Build()
}

func NewSqlxDBProvider(options *options, config *Config) (*sqlx.DB, *sql.DB, error) {
	conn, err := sqlx.Open(options.driver, config.DSN)
	if err != nil {
		return nil, nil, err
	}

	return conn, conn.DB, nil
}

func NewSQLxDBWrapper(options *options, db *sql.DB) (*sqlx.DB, error) {
	return sqlx.NewDb(db, options.driver), nil
}
