package pnphealthcheckpgx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/healthcheck/pnphealthcheck"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	fxutil.OptionsBuilderSupply(moduleBuilder, options)
	moduleBuilder.Provide(pnphealthcheck.HealthCheckerProvider(NewHealthChecker))

	return moduleBuilder.Build()
}

func NewHealthChecker(
	pool *pgxpool.Pool,
	options *options,
) pnphealthcheck.HealthChecker {
	return pnphealthcheck.HealthChecker{
		Name:    options.name,
		Timeout: options.timeout,
		Check: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
	}
}
