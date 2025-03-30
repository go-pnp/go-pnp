package pnppromconnectrpchandling

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/prometheus/client_golang/prometheus"
)

type options struct {
	fxPrivate   bool
	namespace   string
	subsystem   string
	order       int
	constLabels prometheus.Labels
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}

func WithConstLabels(constLabels prometheus.Labels) optionutil.Option[options] {
	return func(o *options) {
		o.constLabels = constLabels
	}
}
