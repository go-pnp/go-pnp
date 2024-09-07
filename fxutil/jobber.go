package fxutil

import (
	"context"
	"fmt"

	"github.com/go-pnp/jobber"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging"
)

// InvokeJob returns jobber.Job start/stop hooks registration options for fx
func InvokeJob[T jobber.Job](jobName string, jobProvider any, jobberOptions ...jobber.OptionFunc) fx.Option {
	type invokeJobParams[T jobber.Job] struct {
		fx.In
		Lc         fx.Lifecycle
		Shutdowner fx.Shutdowner
		Job        T
		Logger     *logging.Logger `optional:"true"`
	}

	return fx.Module(
		jobName,
		fx.Provide(
			jobProvider,
			fx.Private,
		),
		logging.DecorateNamed(fmt.Sprintf("%s_job", jobName)),
		fx.Invoke(func(params invokeJobParams[T]) {
			workerRunner := jobber.NewRunner(params.Job, jobberOptions...)
			params.Lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					params.Logger.Info(ctx, "Starting worker")
					go func() {
						if err := workerRunner.Start(ctx); err != nil {
							params.Logger.WithError(err).Error(ctx, "Start worker error")
							params.Shutdowner.Shutdown()
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					params.Logger.Info(ctx, "Stopping worker")
					err := workerRunner.Close()
					if err != nil {
						params.Logger.WithError(err).Error(ctx, "Close worker error")
					}

					return err
				},
			})
		}),
	)
}

// InvokeJobIf is the same as InvokeJob but with condition
func InvokeJobIf[T jobber.Job](condition bool, workerName string, workerProvider any, workerOptions ...jobber.OptionFunc) fx.Option {
	if !condition {
		return fx.Options()
	}

	return InvokeJob[T](workerName, workerProvider, workerOptions...)
}
