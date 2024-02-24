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

type App struct {
	options        []fx.Option
	startCtxCancel context.CancelFunc
	quitCh         chan chan struct{}
}

func (a *App) Start() error {
	systemLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

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
		systemLogger.Error("failed to start application. stopping...", "error", err)
		stopApp(systemLogger, app)

		return errors.WithStack(err)
	}

	select {
	case res := <-a.quitCh:
		systemLogger.Info("stopping application...")
		stopApp(systemLogger, app)
		res <- struct{}{}

		return nil
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
