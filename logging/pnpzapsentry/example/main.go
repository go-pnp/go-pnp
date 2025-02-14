package main

import (
	"context"
	"os"

	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/pnpenv"
	"github.com/go-pnp/go-pnp/pnpsentry"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging/pnpzapsentry"
)

func main() {
	os.Setenv("SENTRY_DSN", "YOUR_DSN_HERE")
	os.Setenv("ENVIRONMENT", "local")
	fx.New(
		pnpenv.Module(),
		pnpzap.Module(),
		pnpsentry.Module(),
		pnpzapsentry.Module(),
		fx.Invoke(func(logger *logging.Logger, shutdowner fx.Shutdowner) {
			logger.WithError(someFunc()).Named("named").Error(context.Background(), "Hello from test logger")
			shutdowner.Shutdown()
		}),
	).Run()
}

func someFunc() error {
	return errors.Wrap(someOtherFunc(), "from func")
}

func someOtherFunc() error {
	return errors.New("some error")
}
