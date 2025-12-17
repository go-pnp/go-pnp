package pnpzapsanitize

import (
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Module(opts ...optionutil.Option[Options]) fx.Option {
	options := newOptions(opts)

	builder := fxutil.OptionsBuilder{}
	builder.Supply(options)
	builder.Provide(pnpzap.ZapOptionProvider(NewZapOption))

	return builder.Build()
}

func NewZapOption(options *Options) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &fieldHidingCore{
			Core:     core,
			regex:    options.regex,
			redacted: options.redacted,
		}
	})
}
