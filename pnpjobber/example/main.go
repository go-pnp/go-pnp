package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pnp/jobber"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/pnpjobber"
)

func main() {
	fx.New(
		pnpjobber.Module("example", func() jobber.Job {
			return jobber.NewIntervalJob(true, time.Second, func(ctx context.Context) error {
				fmt.Println("Hello, world!")
				return nil
			})
		}),

		pnpjobber.Module("example", func() (jobber.Job, error) {
			return jobber.NewCronJob(true, "* * * * *", func(ctx context.Context) error {
				fmt.Println("Hello from cron job")
				return nil
			})
		}),
	).Run()
}
