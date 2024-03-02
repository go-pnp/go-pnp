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
	Lc            fx.Lifecycle
	RuntimeErrors chan<- error
	Job           jobber.Job
	Logger        *logging.Logger `optional:"true"`
}

// InvokeJob runs job provided via jobProvider,
// One of the result of jobProvider should be jobber.Job
func InvokeJob(jobName string, jobProvider any, jobOptions ...jobber.OptionFunc) fx.Option {
	return fx.Module(
		jobName,
		fx.Provide(
			jobProvider,
			fx.Private,
		),
		logging.DecorateNamed(fmt.Sprintf("%s_job", jobName)),
		fx.Invoke(func(params invokeParams) {
			jobRunner := jobber.NewRunner(params.Job, jobOptions...)
			params.Lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					params.Logger.Info(ctx, "Starting job")
					go func() {
						if err := jobRunner.Start(ctx); err != nil {
							params.Logger.WithError(err).Error(ctx, "Start job error")
							params.RuntimeErrors <- fmt.Errorf("start job %s: %w", jobName, err)
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					params.Logger.Info(ctx, "Stopping job")
					err := jobRunner.Close()
					if err != nil {
						params.Logger.WithError(err).Error(ctx, "Close job error")
					}

					return err
				},
			})
		}),
	)
}

func InvokeJobIf(condition bool, jobName string, jobProvider any, jobOptions ...jobber.OptionFunc) fx.Option {
	if !condition {
		return fx.Options()
	}

	return InvokeJob(jobName, jobProvider, jobOptions...)
}
