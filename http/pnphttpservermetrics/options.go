package pnphttpservermetrics

import (
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/prometheus/client_golang/prometheus"
)

// requestLabel is a metric label whose value is derived from the incoming request.
type requestLabel struct {
	name  string
	value func(*http.Request) string
}

type options struct {
	fxPrivate     bool
	namespace     string
	subsystem     string
	order         int
	constLabels   prometheus.Labels
	requestLabels []requestLabel
}

func newOptions(opts ...optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		subsystem: "http",
		order:     1,
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

func WithOrder(order int) optionutil.Option[options] {
	return func(o *options) {
		o.order = order
	}
}

// WithRequestLabel adds a label to every metric of the module whose value is computed from the request.
// The value function must return a bounded set of values: label cardinality is the caller's responsibility.
func WithRequestLabel(name string, value func(*http.Request) string) optionutil.Option[options] {
	return func(o *options) {
		o.requestLabels = append(o.requestLabels, requestLabel{name: name, value: value})
	}
}
