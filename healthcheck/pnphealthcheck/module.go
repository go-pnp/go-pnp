package pnphealthcheck

import (
	"context"
	"sync"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(NewHealthResolver)

	return moduleBuilder.Build()
}

type HealthResolver struct {
	HealthCheckers []HealthChecker
}

func (h *HealthResolver) Resolve(ctx context.Context) (map[string]error, bool) {
	results := make(map[string]error)
	var resultsMu sync.Mutex
	wg := &sync.WaitGroup{}
	hasError := false

	for _, checker := range h.HealthCheckers {
		wg.Add(1)
		go func(checker HealthChecker) {
			defer wg.Done()
			err := checker.check(ctx)
			if err != nil {
				hasError = true
			}
			resultsMu.Lock()
			results[checker.Name] = err
			resultsMu.Unlock()
		}(checker)
	}
	wg.Wait()

	return results, !hasError
}

func HealthCheckerProvider(target any) any {
	return fxutil.GroupProvider[HealthChecker]("pnphttphealthcheck.health_checkers", target)
}

type NewHealthResolverParams struct {
	fx.In
	Options        *options
	HealthCheckers []HealthChecker `group:"pnphttphealthcheck.health_checkers"`
}

func NewHealthResolver(params NewHealthResolverParams) *HealthResolver {
	for i := range params.HealthCheckers {
		if params.HealthCheckers[i].Timeout == 0 {
			params.HealthCheckers[i].Timeout = params.Options.defaultTimeout
		}
	}
	return &HealthResolver{
		HealthCheckers: params.HealthCheckers,
	}
}
