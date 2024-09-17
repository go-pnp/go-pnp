package pnphealthcheckgorm

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/healthcheck/pnphealthcheck"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(pnphealthcheck.HealthCheckerProvider(NewHealthChecker))

	return moduleBuilder.Build()
}

func NewHealthChecker(
	db *gorm.DB,
	options *options,
) pnphealthcheck.HealthChecker {
	return pnphealthcheck.HealthChecker{
		Name:    options.name,
		Timeout: options.timeout,
		Check: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}

			return sqlDB.PingContext(ctx)
		},
	}
}
