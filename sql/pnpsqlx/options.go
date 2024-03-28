package pnpsqlx

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	driver              string
	wrapSQLDB           bool
	configPrefix        string
	fxPrivate           bool
	configFromContainer bool
}

func newOptions(driver string, opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		driver:       driver,
		configPrefix: "DB_",
	}, opts...)
}

// WrapSQLDB changes behavior of module to wrap sql.DB with sqlx.DB instead of creating new sql.DB with sqlx.DB
func WrapSQLDB() optionutil.Option[options] {
	return func(o *options) {
		o.wrapSQLDB = true
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

// WithConfigFromContainer if used, module will not provide config, but will use config already provided to fx di container.
func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}
