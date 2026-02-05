package pnpjobber

import (
	"context"

	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/jobber"
	"go.uber.org/fx"
)

type invokeJobberRunnerParams struct {
	fx.In

	Lc            fx.Lifecycle
	Shutdowner    fx.Shutdowner
	Job           jobber.Job
	JobDecorators ordering.OrderedItems[JobDecorator] `group:"pnp_jobber.job_decorators"`
	Logger        *logging.Logger                     `optional:"true"`
}

// Module runs worker provided via jobProvider,
// One of the result of jobProvider should be jobber.Job
func Module(workerProvider any, workerOptions ...jobber.OptionFunc) fx.Option {
	return fx.Module(
		"job",
		fx.Provide(
			workerProvider,
			fx.Private,
		),
		decorateLoggerWithJobName(),

		fx.Invoke(func(params invokeJobberRunnerParams) {
			job := params.Job
			for _, decorator := range params.JobDecorators.Get() {
				job = decorator(job)
			}

			workerRunner := jobber.NewRunner(job, workerOptions...)

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
