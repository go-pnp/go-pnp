package pnppgx

import (
	"time"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	// fxPrivate
	fxPrivate bool

	// if configFromContainer is true, module will not provide his config into di container
	configFromContainer bool

	configPrefix string

	// if stdDB == true, module will provide *sql.DB into provider instead of *pgx.Conn
	stdDB bool

	initialPingTimeout time.Duration
}

func newOptions(opts []optionutil.Option[options]) *options {
	result := &options{
		fxPrivate:           false,
		configFromContainer: false,
		configPrefix:        "DB_",
		stdDB:               false,
		initialPingTimeout:  time.Second * 2,
	}

	return optionutil.ApplyOptions(result, opts...)
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

// WithEnvConfigPrefix is an option to set config env prefix for config provider.
func WithEnvConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

func WithStdDB() optionutil.Option[options] {
	return func(o *options) {
		o.stdDB = true
	}
}

func WithInitialPingTimeout(t time.Duration) optionutil.Option[options] {
	return func(o *options) {
		o.initialPingTimeout = t
	}
}
