package pnpzap

import (
	"go.uber.org/zap"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	zapConfig  *zap.Config
	zapOptions []zap.Option
	fxPrivate  bool
}

func WithZapConfig(config zap.Config) optionutil.Option[options] {
	return func(o *options) {
		o.zapConfig = &config
	}
}

func WithZapOptions(zapOptions ...zap.Option) optionutil.Option[options] {
	return func(o *options) {
		o.zapOptions = append(o.zapOptions, zapOptions...)
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}
