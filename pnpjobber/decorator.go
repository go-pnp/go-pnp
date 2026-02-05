package pnpjobber

import (
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/go-pnp/jobber"
	"go.uber.org/fx"
)

type JobDecorator func(job jobber.Job) jobber.Job

func JobDecoratorProvider(target any) any {
	return fxutil.GroupProvider[ordering.OrderedItem[JobDecorator]](
		"pnp_jobber.job_decorators",
		target,
	)
}

func decorateLoggerWithJobName() fx.Option {
	return fx.Decorate(func(params invokeJobberRunnerParams) *logging.Logger {
		if params.Logger == nil {
			return nil
		}

		return params.Logger.Named(params.Job.Name() + "_worker")
	})
}
