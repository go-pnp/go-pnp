package pnphttpservermetrics

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/prometheus/client_golang/prometheus"
)

type options struct {
	fxPrivate   bool
	namespace   string
	subsystem   string
	constLabels prometheus.Labels
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		subsystem: "http",
	}, opts...)
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithNamespace(namespace string) optionutil.Option[options] {
	return func(o *options) {
		o.namespace = namespace
	}
}

func WithSubsystem(subsystem string) optionutil.Option[options] {
	return func(o *options) {
		o.subsystem = subsystem
	}
}

func WithConstLabels(constLabels prometheus.Labels) optionutil.Option[options] {
	return func(o *options) {
		o.constLabels = constLabels
	}
}
