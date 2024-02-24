package pnphealthcheck

import (
	"context"
	"time"
)

type HealthChecker struct {
	Name    string
	Timeout time.Duration
	Check   func(ctx context.Context) error
}

func (c *HealthChecker) check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	return c.check(ctx)
}
