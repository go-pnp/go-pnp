package fxutil

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type App struct {
	options        []fx.Option
	startCtxCancel context.CancelFunc
	quitCh         chan chan struct{}
}

func (a *App) Start() error {
	systemLogger := logrus.New()
	systemLogger.SetFormatter(&logrus.JSONFormatter{})
	systemLogger.SetLevel(logrus.DebugLevel)

	runtimeErrors := make(chan error)

	options := append([]fx.Option{
		fx.Supply((chan<- error)(runtimeErrors))},
		a.options...,
	)

	app := fx.New(
		options...,
	)

	ctx, cancel := context.WithCancel(context.Background())
	a.startCtxCancel = cancel

	ctx, cancelFn := context.WithTimeout(ctx, time.Second*10)
	defer cancelFn()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println(fx.VisualizeError(err))
		systemLogger.WithError(err).Error("failed to start application. stopping...")
		stopApp(systemLogger, app)

		return errors.WithStack(err)
	}

	select {
	case <-a.quitCh:
		systemLogger.Infof("stopping application...")
		stopApp(systemLogger, app)

		return nil
	case signal := <-app.Done():
		systemLogger.Infof("received %s signal. stopping...", signal)
		stopApp(systemLogger, app)

		return nil
	case err := <-runtimeErrors:
		systemLogger.WithError(err).Error("failed to start application. stopping...")
		stopApp(systemLogger, app)

		return err
	}
}

func (a *App) Close() error {
	if a.startCtxCancel != nil {
		a.startCtxCancel()
	}
	resultCh := make(chan struct{})
	a.quitCh <- resultCh
	select {
	case <-resultCh:
		return nil
	case <-time.After(time.Second * 10):
		return errors.New("stopping application time out")
	}

	return nil
}

func NewApp(options ...fx.Option) *App {
	return &App{
		options: options,
		quitCh:  make(chan chan struct{}),
	}
}
