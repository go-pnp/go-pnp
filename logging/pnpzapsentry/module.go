package pnpzapsentry

import (
	"errors"

	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	builder := fxutil.OptionsBuilder{}
	builder.ProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigProvider[Config](options.configPrefix))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix))
	builder.Provide(pnpzap.ZapOptionProvider(NewZapCoreWrapperOption))

	return builder.Build()
}

type NewZapCoreWrapperOptionParams struct {
	fx.In

	SentryClient *sentry.Client
	Config       *Config
}

func NewZapCoreWrapperOption(params NewZapCoreWrapperOptionParams) (zap.Option, error) {
	reportLevelValue, ok := sentryLevelValues[params.Config.ReportLevel]
	if !ok {
		return nil, errors.New("invalid sentry report level. valid values are: debug, info, warning, error, fatal")
	}
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return NewZapCore(core, params.SentryClient, params.Config.FlushTimeout, reportLevelValue)
	}), nil
}
