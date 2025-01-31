package pnpzap

import (
	"github.com/go-pnp/go-pnp/pnpenv"
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
	builder.Provide(NewLogger, NewLoggingLogger)

	builder.ProvideIf(!options.zapConfigFromContainer, NewZapLoggerConfig)
	builder.ProvideIf(!options.configFromContainer, configutil.NewConfigProvider[Config](configutil.Options{}))
	builder.PublicProvideIf(!options.configFromContainer, configutil.NewConfigInfoProvider[Config](configutil.Options{}))

	return builder.Build()
}

type NewZapLoggerConfigParams struct {
	fx.In

	Config *Config
	Env    pnpenv.Environment `optional:"true"`
}

func NewZapLoggerConfig(params NewZapLoggerConfigParams) zap.Config {
	atomicLevel := params.Config.ZapAtomicLevel()
	switch {
	case params.Env.IsDev():
		config := zap.NewDevelopmentConfig()
		config.Level = atomicLevel

		return config
	default:
		config := zap.NewProductionConfig()
		config.Level = atomicLevel

		return config
	}
}

func ZapHookProvider(target any) any {
	return fxutil.GroupProvider[func(zapcore.Entry) error](
		"pnpzap.hooks",
		target,
	)
}

func ZapOptionProvider(target any) any {
	return fxutil.GroupProvider[zap.Option](
		"pnpzap.zap_options",
		target,
	)
}

func ZapContextFieldResolverProvider(target any) any {
	return fxutil.GroupProvider[ContextFieldResolver](
		"pnpzap.context_fields_resolver",
		target,
	)
}

type NewLoggerParams struct {
	fx.In
	ZapConfig             zap.Config
	Hooks                 []func(zapcore.Entry) error `group:"pnpzap.hooks"`
	Options               []zap.Option                `group:"pnpzap.zap_options"`
	ContextFieldResolvers []ContextFieldResolver      `group:"pnpzap.context_fields_resolver"`
}

func (n NewLoggerParams) BuildOptions() []zap.Option {
	result := []zap.Option{
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return ZapCore{
				Delegate:              core,
				ContextFieldResolvers: n.ContextFieldResolvers,
			}
		}),
	}
	if len(n.Hooks) > 0 {
		result = append(result, zap.Hooks(n.Hooks...))
	}

	return append(result, n.Options...)
}

func NewLogger(params NewLoggerParams) (*zap.Logger, error) {
	return params.ZapConfig.Build(params.BuildOptions()...)
}
