package configutil

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Options = env.Options

func NewPrefixedConfigProvider[T any](prefix string) func() (*T, error) {
	return NewConfigProvider[T](env.Options{
		Prefix: prefix,
	})
}

func NewConfigProvider[T any](opts Options) func() (*T, error) {
	return func() (*T, error) {
		c := new(T)
		if err := env.ParseWithOptions(c, opts); err != nil {
			return nil, fmt.Errorf("parse config from env: %w", err)
		}

		return c, nil
	}
}
