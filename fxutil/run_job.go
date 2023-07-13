package fxutil

import (
	"context"
	"fmt"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type JobResult error

// RunJob creates and starts application and waits for JobResult. It's useful when you want to run a job like db migrations apply.
func RunJob(options ...fx.Option) error {
	systemLogger := logrus.New()
	systemLogger.SetFormatter(&logrus.JSONFormatter{})
	systemLogger.SetLevel(logrus.DebugLevel)

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
		systemLogger.WithError(err).Error("failed to start application. stopping...")
		stopApp(systemLogger, app)

		return errors.WithStack(err)
	}

	select {
	case signal := <-app.Done():
		systemLogger.Infof("received %s signal. stopping...", signal)
		stopApp(systemLogger, app)

		return nil
	case err := <-runtimeErrors:
		systemLogger.WithError(err).Error("failed to start application. stopping...")
		stopApp(systemLogger, app)

		return err
	case err := <-jobResult:

		stopApp(systemLogger, app)

		return err
	}
}

func RunInvokes(options ...fx.Option) error {
	systemLogger := logrus.New()
	systemLogger.SetFormatter(&logrus.JSONFormatter{})
	systemLogger.SetLevel(logrus.DebugLevel)

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
		systemLogger.WithError(app.Err()).Error("failed to start application. stopping...")

		return errors.WithStack(app.Err())
	}

	return nil
}
