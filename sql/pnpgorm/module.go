package pnpgorm

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(driver string, opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{
		configPrefix: "DB_",
	}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	switch driver {
	case "mysql":
		builder.Provide(NewMySQLDialector)
	case "postgres":
		builder.Provide(NewPostgresDialector)
	case "sqlite":
		builder.Provide(NewSQLiteDialector)
	default:
		panic("unsupported driver")
	}
	builder.Provide(NewGormDBProvider(options))
	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{
		Prefix: options.configPrefix,
	}))

	return builder.Build()
}

func NewMySQLDialector(config *Config) gorm.Dialector {
	return mysql.Open(config.DSN)
}
func NewPostgresDialector(config *Config) gorm.Dialector {
	return postgres.Open(config.DSN)
}
func NewSQLiteDialector(config *Config) gorm.Dialector {
	return sqlite.Open(config.SQLiteDB)
}
func PluginProvider(target any) any {
	return fxutil.GroupProvider[gorm.Plugin]("pnpgorm.gorm_plugins", target)
}

type NewGormDBParams struct {
	fx.In
	Lc        fx.Lifecycle
	Dialector gorm.Dialector
	Plugins   []gorm.Plugin   `group:"pnpgorm.gorm_plugins"`
	Logger    *logging.Logger `optional:"true"`
}

func NewGormDBProvider(opts *options) func(params NewGormDBParams) (_ *gorm.DB, rerr error) {
	return func(params NewGormDBParams) (_ *gorm.DB, rerr error) {
		config := &gorm.Config{
			Logger: logger.Discard,
		}
		if opts.enableLogger {
			config.Logger = &Logger{Delegate: params.Logger}
		}
		db, err := gorm.Open(params.Dialector, config)
		if err != nil {
			return nil, err
		}

		for _, plugin := range params.Plugins {
			if err := db.Use(plugin); err != nil {
				params.Logger.WithError(err).Error(context.Background(), "gorm plugin enable failed")
			}
		}

		params.Lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}

				params.Logger.Info(ctx, "closing database connection")

				return sqlDB.Close()
			},
		})
		return db, nil
	}
}
