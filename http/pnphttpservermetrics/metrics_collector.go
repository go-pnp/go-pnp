package pnphttpservermetrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector struct {
	durationHistogramVec  *prometheus.HistogramVec
	callsTotal            *prometheus.CounterVec
	requestBodySizeBytes  *prometheus.CounterVec
	responseBodySizeBytes *prometheus.CounterVec
	requestLabels         []requestLabel
}

// NewMetricsCollector returns a collector of request count, duration and body size metrics.
// Every metric carries the request labels configured by WithRequestLabel after its own labels.
func NewMetricsCollector(options *options) *MetricsCollector {
	requestLabelNames := make([]string, 0, len(options.requestLabels))
	for _, label := range options.requestLabels {
		requestLabelNames = append(requestLabelNames, label.name)
	}

	labelNames := func(base ...string) []string {
		return append(base, requestLabelNames...)
	}

	return &MetricsCollector{
		requestLabels: options.requestLabels,
		callsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "requests_total",
				ConstLabels: options.constLabels,
			},
			labelNames("method", "path", "code"),
		),
		durationHistogramVec: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "request_duration_seconds",
				Buckets:     []float64{0.05, .1, .25, .5, 1, 2.5, 5, 10},
				ConstLabels: options.constLabels,
			},
			labelNames("method", "path", "status_class"),
		),
		requestBodySizeBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "request_body_size_bytes",
				ConstLabels: options.constLabels,
			},
			labelNames("method", "path"),
		),
		responseBodySizeBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   options.namespace,
				Subsystem:   options.subsystem,
				Name:        "response_body_size_bytes",
				ConstLabels: options.constLabels,
			},
			labelNames("method", "path"),
		),
	}
}

func (m *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	m.durationHistogramVec.Collect(ch)
	m.callsTotal.Collect(ch)
	m.requestBodySizeBytes.Collect(ch)
	m.responseBodySizeBytes.Collect(ch)
}

func (m *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	m.durationHistogramVec.Describe(ch)
	m.callsTotal.Describe(ch)
	m.requestBodySizeBytes.Describe(ch)
	m.responseBodySizeBytes.Describe(ch)
}

func (m *MetricsCollector) trackRequest(request *http.Request, path string) *RequestObserver {
	if m == nil {
		return nil
	}

	requestLabelValues := make([]string, 0, len(m.requestLabels))
	for _, label := range m.requestLabels {
		requestLabelValues = append(requestLabelValues, label.value(request))
	}

	return &RequestObserver{
		collector:          m,
		method:             request.Method,
		path:               path,
		requestLabelValues: requestLabelValues,
		startAt:            time.Now(),
	}
}

type RequestObserver struct {
	collector          *MetricsCollector
	method             string
	path               string
	requestLabelValues []string
	startAt            time.Time
}

func (r *RequestObserver) Observe(requestBodySize, responseBodySize, code int) {
	if r == nil {
		return
	}

	r.collector.callsTotal.WithLabelValues(r.labelValues(strconv.Itoa(code))...).Inc()
	r.collector.durationHistogramVec.WithLabelValues(r.labelValues(statusClass(code))...).Observe(time.Since(r.startAt).Seconds())
	r.collector.requestBodySizeBytes.WithLabelValues(r.labelValues()...).Add(float64(requestBodySize))
	r.collector.responseBodySizeBytes.WithLabelValues(r.labelValues()...).Add(float64(responseBodySize))
}

// labelValues returns method and path, then the metric-specific values, then the request label values,
// matching the label name order declared in NewMetricsCollector.
func (r *RequestObserver) labelValues(metricSpecific ...string) []string {
	values := make([]string, 0, 2+len(metricSpecific)+len(r.requestLabelValues))
	values = append(values, r.method, r.path)
	values = append(values, metricSpecific...)

	return append(values, r.requestLabelValues...)
}

func statusClass(code int) string {
	if code < 100 || code > 599 {
		return "unknown"
	}

	return strconv.Itoa(code/100) + "xx"
}
