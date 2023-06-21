package pnpgorm

import (
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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
	builder.Provide(NewGormDB)
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

func NewGormDB(dialector gorm.Dialector) (*gorm.DB, error) {
	return gorm.Open(dialector, &gorm.Config{})
}
