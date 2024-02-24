package fxutil

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

type JobResult error

// RunJob creates and starts application and waits for JobResult. It's useful when you want to run a job like db migrations apply.
func RunJob(options ...fx.Option) error {
	systemLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	runtimeErrors := make(chan error)
	jobResult := make(chan JobResult)

	options = append([]fx.Option{
		fx.Supply((chan<- error)(runtimeErrors)),
		fx.Supply((chan<- JobResult)(jobResult)),
	},
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
	case err := <-jobResult:

		stopApp(systemLogger, app)

		return err
	}
}

func RunInvokes(options ...fx.Option) error {
	systemLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	runtimeErrors := make(chan error)
	jobResult := make(chan JobResult)

	options = append([]fx.Option{
		fx.Supply(runtimeErrors),
		fx.Supply(jobResult),
	},
		options...,
	)

	app := fx.New(
		options...,
	)

	if app.Err() != nil {
		fmt.Println(fx.VisualizeError(app.Err()))
		systemLogger.Error("failed to start application. stopping...", "error", app.Err())

		return errors.WithStack(app.Err())
	}

	return nil
}
