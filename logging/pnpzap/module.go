package pnpzap

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	builder.Provide(
		configutil.NewConfigProvider[Config](),
		NewLogger,
		NewLoggingLogger,
	)
	if options.zapConfig != nil {
		fxutil.OptionsBuilderSupply(builder, *options.zapConfig)
	} else {
		builder.Provide(NewZapLoggerConfig)
	}

	for _, option := range options.zapOptions {
		fxutil.OptionsBuilderGroupSupply(builder, "pnpzap.zap_options", option)
	}

	return builder.Build()
}

func NewZapLoggerConfig(config *Config) zap.Config {
	return config.EnvironmentConfig()
}

func ZapHookProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[func(zapcore.Entry) error](
		"pnpzap.hooks",
		target,
	)
}
func ZapOptionProvider(target interface{}) fx.Annotated {
	return fxutil.GroupProvider[zap.Option](
		"pnpzap.zap_options",
		target,
	)
}

type NewLoggerParams struct {
	fx.In
	ZapConfig zap.Config
	Hooks     []func(zapcore.Entry) error `group:"pnpzap.hooks"`
	Options   []zap.Option                `group:"pnpzap.zap_options"`
}

func (n NewLoggerParams) BuildOptions() []zap.Option {
	var result []zap.Option
	if len(n.Hooks) > 0 {
		result = append(result, zap.Hooks(n.Hooks...))
	}
	return append(result, n.Options...)
}

func NewLogger(params NewLoggerParams) (*zap.Logger, error) {
	return params.ZapConfig.Build(params.BuildOptions()...)
}
