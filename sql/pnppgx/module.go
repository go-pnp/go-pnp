package pnppgx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Supply(options)
	builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))

	builder.ProvideIf(options.stdDB, fx.Annotate(
		NewPgxStdConnection,
		fx.OnStop(CloseStdConnection),
	))
	builder.ProvideIf(!options.stdDB, fx.Annotate(
		NewPgxPool,
		fx.OnStop(ClosePool),
	))

	return builder.Build()
}

type NewPgxStdConnectionParams struct {
	fx.In
	Lc      fx.Lifecycle
	Config  *Config
	Logger  *logging.Logger `optional:"true"`
	Options *options
}

func NewPgxStdConnection(params NewPgxStdConnectionParams) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(params.Config.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	db := stdlib.OpenDB(*connConfig)
	db.SetMaxOpenConns(params.Config.MaxOpenConnections)
	db.SetMaxIdleConns(params.Config.MaxIdleConnections)
	db.SetConnMaxLifetime(params.Config.MaxConnectionLifetime)
	db.SetConnMaxIdleTime(params.Config.MaxConnectionIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), params.Options.initialPingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("initial ping: %w", err)
	}

	return db, nil
}

func CloseStdConnection(db *sql.DB) error {
	return db.Close()
}

type NewPgxConnectionParams struct {
	fx.In
	Lc      fx.Lifecycle
	Config  *Config
	Logger  *logging.Logger `optional:"true"`
	Options *options
}

func NewPgxPool(params NewPgxConnectionParams) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(params.Config.DSN)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(params.Config.MaxOpenConnections)
	poolConfig.MaxConnLifetime = params.Config.MaxConnectionLifetime
	poolConfig.MaxConnIdleTime = params.Config.MaxConnectionIdleTime

	pool, err := pgxpool.NewWithConfig(
		context.Background(),
		poolConfig,
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), params.Options.initialPingTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("initial ping: %w", err)
	}

	return pool, nil
}

func ClosePool(db *pgxpool.Pool) {
	db.Close()
}
