package configutil

import (
	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
)

func NewConfigProvider[T any](opts ...env.Options) func() (*T, error) {
	return func() (*T, error) {
		c := new(T)

		if err := env.Parse(c, opts...); err != nil {
			return nil, errors.WithStack(err)
		}

		return c, nil
	}
}
