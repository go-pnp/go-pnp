//go:build example
// +build example

package pnpzap

import (
	"context"
	"fmt"
	"os"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
)

func TestSimpleUsage(t *testing.T) {
	os.Setenv("ENVIRONMENT", "dev")
	fxutil.StartApp(
		Module(),
		fx.Invoke(func(logger logging.Logger) {
			logger.Named("test").Info(context.Background(), "Hello from test logger")
		}),
	)
}

func TestConfiguration(t *testing.T) {
	fxutil.StartApp(
		Module(
			WithZapConfig(zap.NewDevelopmentConfig()),
			WithZapOptions(zap.WithCaller(false)),
		),
		fx.Provide(
			ZapHookProvider(func() func(entry zapcore.Entry) error {
				return func(entry zapcore.Entry) error {
					fmt.Println("Hook called")
					return nil
				}
			}),
		),
		fx.Invoke(func(logger logging.Logger) {
			logger.Named("test").WithField("hello", "world").Info(context.Background(), "Hello from test logger")
		}),
	)
}
