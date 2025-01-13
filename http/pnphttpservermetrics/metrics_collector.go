package pnphttpservermetrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector struct {
	durationHistogramVec  *prometheus.HistogramVec
	callsTotal            *prometheus.CounterVec
	requestBodySizeBytes  *prometheus.CounterVec
	responseBodySizeBytes *prometheus.CounterVec
}

// NewMetricsCollector returns an instance of the Client decorated with prometheus summary metric
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		callsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
			},
			[]string{"method", "path", "code"},
		),
		durationHistogramVec: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Buckets: []float64{0.05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		requestBodySizeBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_body_size_bytes",
			},
			[]string{"method", "path"},
		),
		responseBodySizeBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_response_body_size_bytes",
			},
			[]string{"method", "path"},
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

func (m *MetricsCollector) trackRequest(method, path string) *RequestObserver {
	if m == nil {
		return nil
	}
	return &RequestObserver{
		collector: m,
		method:    method,
		path:      path,
		startAt:   time.Now(),
	}
}

type RequestObserver struct {
	collector *MetricsCollector
	method    string
	path      string
	startAt   time.Time
}

func (r *RequestObserver) Observe(requestBodySize, responseBodySize, code int) {
	if r == nil {
		return
	}

	r.collector.callsTotal.WithLabelValues(r.method, r.path, fmt.Sprint(code)).Inc()
	r.collector.durationHistogramVec.WithLabelValues(r.method, r.path).Observe(time.Since(r.startAt).Seconds())
	r.collector.requestBodySizeBytes.WithLabelValues(r.method, r.path).Add(float64(requestBodySize))
	r.collector.responseBodySizeBytes.WithLabelValues(r.method, r.path).Add(float64(responseBodySize))
}
