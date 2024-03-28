package configutil

import (
	"fmt"
	"reflect"

	"github.com/caarlos0/env/v10"
	"go.uber.org/fx"
)

type ConfigInfo struct {
	ConfigType reflect.Type
	Fields     []env.FieldParams
}

type ConfigsInfoIn struct {
	fx.In
	ConfigsInfo []ConfigInfo `group:"configs_info"`
}

type ConfigInfoResult struct {
	fx.Out
	ConfigInfo ConfigInfo `group:"configs_info"`
}

// NewPrefixedConfigInfoProvider similar to NewConfigInfoProvider, but with prefix option.
func NewPrefixedConfigInfoProvider[T any](prefix string) func() (ConfigInfoResult, error) {
	return NewConfigInfoProvider[T](env.Options{
		Prefix: prefix,
	})
}

// NewConfigInfoProvider returns a function that provides a slice of ConfigField.
// This can be helpful when you want to get all the config fields to dump them in a file.
func NewConfigInfoProvider[T any](opts Options) func() (ConfigInfoResult, error) {
	return func() (ConfigInfoResult, error) {
		c := new(T)

		fields, err := env.GetFieldParamsWithOptions(c, opts)
		if err != nil {
			return ConfigInfoResult{}, fmt.Errorf("get config field params: %w", err)
		}

		return ConfigInfoResult{
			ConfigInfo: ConfigInfo{
				ConfigType: reflect.TypeOf(c),
				Fields:     fields,
			},
		}, nil
	}
}
