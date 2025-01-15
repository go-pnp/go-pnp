package main

import (
	"context"
	"errors"

	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pnpenv"
	"github.com/go-pnp/go-pnp/pnpsentry"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging/pnpzapsentry"
)

func main() {
	fx.New(
		pnpenv.Module(),
		pnpzap.Module(),
		pnpsentry.Module(),
		pnpzapsentry.Module(),
		fx.Invoke(func(logger *logging.Logger) {
			logger.WithError(errors.New("hello")).Error(context.Background(), "Hello from test logger")
		}),
	).Run()
}
