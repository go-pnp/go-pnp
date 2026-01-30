package pnpjobber

import (
	"context"
	"fmt"

	"github.com/go-pnp/jobber"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/logging"
)

type invokeJobberRunnerParams struct {
	fx.In
	Lc         fx.Lifecycle
	Shutdowner fx.Shutdowner
	Job        jobber.Job
	Tracer     trace.Tracer
	Logger     *logging.Logger `optional:"true"`
}

// Module runs worker provided via jobProvider,
// One of the result of jobProvider should be jobber.Job
func Module(workerName string, workerProvider any, workerOptions ...jobber.OptionFunc) fx.Option {
	return fx.Module(
		workerName,
		fx.Provide(
			workerProvider,
			fx.Private,
		),
		logging.DecorateNamed(fmt.Sprintf("%s_worker", workerName)),
		fx.Invoke(func(params invokeJobberRunnerParams) {
			opts := []jobber.OptionFunc{
				jobber.WithMiddlewares(traceMiddleware(params.Tracer, workerName)),
			}
			opts = append(opts, workerOptions...)
			workerRunner := jobber.NewRunner(params.Job, opts...)
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

func traceMiddleware(tracer trace.Tracer, workerName string) jobber.Middleware {
	return func(next jobber.HandleFunc) jobber.HandleFunc {
		return func(ctx context.Context) error {
			ctx, span := tracer.Start(ctx, "job "+workerName)
			defer span.End()

			span.SetAttributes(attribute.String("job.name", workerName))

			return next(ctx)
		}
	}
}
