package pnpprometheus

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

func MetricsCollectorProvider(target any) any {
	return fxutil.GroupProvider[prometheus.Collector](
		"pnpprometheus.metric_collectors",
		target,
	)
}

type NewPrometheusRegistryParams struct {
	fx.In

	Collectors []prometheus.Collector `group:"pnpprometheus.metric_collectors"`
}

func NewPrometheusRegistry(params NewPrometheusRegistryParams) (*prometheus.Registry, error) {
	result := prometheus.NewRegistry()
	if err := result.Register(collectors.NewGoCollector()); err != nil {
		return nil, errors.WithStack(err)
	}

	for _, collector := range params.Collectors {
		if err := result.Register(collector); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return result, nil
}
