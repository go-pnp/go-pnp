package fxutil

import (
	"context"
	"fmt"

	"github.com/go-pnp/jobber"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging"
)

type invokeParams struct {
	fx.In
	Lc         fx.Lifecycle
	Shutdowner fx.Shutdowner
	Job        jobber.Job
	Logger     *logging.Logger `optional:"true"`
}

// InvokeWorker runs worker provided via jobProvider,
// One of the result of jobProvider should be jobber.Job
func InvokeWorker(workerName string, workerProvider any, workerOptions ...jobber.OptionFunc) fx.Option {
	return fx.Module(
		workerName,
		fx.Provide(
			workerProvider,
			fx.Private,
		),
		logging.DecorateNamed(fmt.Sprintf("%s_worker", workerName)),
		fx.Invoke(func(params invokeParams) {
			workerRunner := jobber.NewRunner(params.Job, workerOptions...)
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

func InvokeWorkerIf(condition bool, workerName string, workerProvider any, workerOptions ...jobber.OptionFunc) fx.Option {
	if !condition {
		return fx.Options()
	}

	return InvokeWorker(workerName, workerProvider, workerOptions...)
}
