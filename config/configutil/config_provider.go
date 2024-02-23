package configutil

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
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
	ConfigParams []env.FieldParams `group:"config_fields,flatten"`
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

		configParams, err := env.GetFieldParamsWithOptions(c, opts)
		if err != nil {
			return ConfigProviderResult[T]{}, fmt.Errorf("get config params: %w", err)
		}

		return ConfigProviderResult[T]{
			Config:       c,
			ConfigParams: configParams,
		}, nil
	}
}

type DumpConfigInDotEnvFormatParams struct {
	ConfigParams []env.FieldParams `group:"config_fields"`
}

func DumpConfigsInDotEnvFormat(params DumpConfigInDotEnvFormatParams) {
	for _, fieldParams := range params.ConfigParams {
		var comments []string
		if fieldParams.Required {
			comments = append(comments, "required")
		}
		if fieldParams.NotEmpty {
			comments = append(comments, "not empty")
		}
		if fieldParams.Expand {
			comments = append(comments, "expands")
		}
		if fieldParams.LoadFile {
			comments = append(comments, "loaded from file")
		}

		fmt.Printf("%s=\"%s\" #%s\n", fieldParams.Key, fieldParams.DefaultValue, strings.Join(comments, ","))
	}
}
