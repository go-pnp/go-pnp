package configutil

import (
	"fmt"

	"github.com/caarlos0/env/v8"
	"go.uber.org/fx"
)

type ConfigParams struct {
	EnvVariable  string
	DefaultValue string
	Required     bool
}

type ConfigProviderResult[T any] struct {
	fx.Out
	Config       *T
	ConfigParams []ConfigParams `group:"config_fields,flatten"`
}

type Options = env.Options

func NewPrefixedConfigProvider[T any](prefix string) func() (ConfigProviderResult[T], error) {
	return NewConfigProvider[T](env.Options{
		Prefix: prefix,
	})
}

func NewConfigProvider[T any](opts Options) func() (ConfigProviderResult[T], error) {
	return func() (ConfigProviderResult[T], error) {
		c := new(T)
		if err := env.ParseWithOptions(c, opts); err != nil {
			return ConfigProviderResult[T]{}, fmt.Errorf("parse config from env: %w", err)
		}

		// https://github.com/caarlos0/env/issues/260
		//configParams, err := env.GetFieldParams(c, opts)
		//if err != nil {
		//	return ConfigProviderResult[T]{}, fmt.Errorf("get config params: %w", err)
		//}

		return ConfigProviderResult[T]{
			Config: c,
			//ConfigParams: configParams,
		}, nil
	}
}
