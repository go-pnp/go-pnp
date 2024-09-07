package fxutil

import (
	"context"

	"go.uber.org/fx"
)

// RunJob1 runs some job function inside fx.Application with one argument and returns job result
func RunJob1[T any](jobFn func(context.Context, T) error, options ...fx.Option) error {
	var jobErr error
	app := fx.New(
		fx.Options(options...),
		fx.Invoke(func(lc fx.Lifecycle, val T, shutdowner fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						jobErr = jobFn(ctx, val)
						shutdowner.Shutdown()
					}()

					return nil
				},
			})
		}),
	)
	if err := app.Err(); err != nil {
		return err
	}
	app.Run()

	return jobErr
}

// RunJob2 runs some job function inside fx.Application with two arguments and returns job result
func RunJob2[T, K any](jobFn func(context.Context, T, K) error, options ...fx.Option) error {
	var jobErr error
	app := fx.New(
		fx.Options(options...),
		fx.Invoke(func(lc fx.Lifecycle, val T, val2 K, shutdowner fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						jobErr = jobFn(ctx, val, val2)
						shutdowner.Shutdown()
					}()

					return nil
				},
			})
		}),
	)
	if err := app.Err(); err != nil {
		return err
	}
	app.Run()

	return jobErr
}
