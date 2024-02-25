package fxutil

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/fx"
)

// StartApp creates and starts a new Go PnP application using the provided options.
// It returns an error if the application fails to start or if an error occurs during runtime.
// The runtime errors channel is used to capture any errors that occur during runtime and propagate them
// back to the main thread.
func StartApp(options ...fx.Option) error {
	systemLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	runtimeErrors := make(chan error)

	options = append([]fx.Option{
		fx.Supply((chan<- error)(runtimeErrors))},
		options...,
	)

	app := fx.New(
		options...,
	)

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFn()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println(fx.VisualizeError(err))
		systemLogger.Error("failed to start application. stopping...", "error", err)
		stopApp(systemLogger, app)

		return errors.WithStack(err)
	}

	select {
	case signal := <-app.Done():
		systemLogger.Info(fmt.Sprintf("received %s signal. stopping...", signal))
		stopApp(systemLogger, app)

		return nil
	case err := <-runtimeErrors:
		systemLogger.Error("failed to start application. stopping...", "error", err)
		stopApp(systemLogger, app)

		return err
	}
}

func stopApp(logger *slog.Logger, app *fx.App) {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFn()

	if err := app.Stop(ctx); err != nil {
		logger.Error("failed to stop application.", "error", err)
	}
}
