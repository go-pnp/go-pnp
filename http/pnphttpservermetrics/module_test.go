package pnphttpservermetrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

const itemsRoute = "/items/{id}"

func newTestRouter(handler http.HandlerFunc, opts ...optionutil.Option[options]) (*mux.Router, *MetricsCollector) {
	options := newOptions(opts...)
	collector := NewMetricsCollector(options)
	middleware := NewMiddleware(NewMuxHandlerRegistrarParams{MetricsCollector: collector, Options: options})

	router := mux.NewRouter()
	router.Use(middleware.Value)
	router.Path(itemsRoute).Handler(handler)

	return router, collector
}

func serve(router http.Handler, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func counterValue(t *testing.T, vec *prometheus.CounterVec, labelValues ...string) float64 {
	t.Helper()

	counter, err := vec.GetMetricWithLabelValues(labelValues...)
	require.NoError(t, err)

	return testutil.ToFloat64(counter)
}

func histogramSampleCount(t *testing.T, vec *prometheus.HistogramVec, labelValues ...string) uint64 {
	t.Helper()

	observer, err := vec.GetMetricWithLabelValues(labelValues...)
	require.NoError(t, err)

	metric, ok := observer.(prometheus.Metric)
	require.True(t, ok)

	var written dto.Metric
	require.NoError(t, metric.Write(&written))

	return written.GetHistogram().GetSampleCount()
}

func TestWriteWithoutWriteHeaderIsRecordedAsOK(t *testing.T) {
	router, collector := newTestRouter(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	recorder := serve(router, httptest.NewRequest(http.MethodGet, "/items/1", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1.0, counterValue(t, collector.callsTotal, http.MethodGet, itemsRoute, "200"))
	require.Equal(t, uint64(1), histogramSampleCount(t, collector.durationHistogramVec, http.MethodGet, itemsRoute, "2xx"))
	require.Equal(t, 5.0, counterValue(t, collector.responseBodySizeBytes, http.MethodGet, itemsRoute))
}

func TestEmptyResponseIsRecordedAsOK(t *testing.T) {
	router, collector := newTestRouter(func(http.ResponseWriter, *http.Request) {})

	serve(router, httptest.NewRequest(http.MethodGet, "/items/1", nil))

	require.Equal(t, 1.0, counterValue(t, collector.callsTotal, http.MethodGet, itemsRoute, "200"))
}

func TestExplicitStatusIsRecorded(t *testing.T) {
	router, collector := newTestRouter(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("missing"))
	})

	serve(router, httptest.NewRequest(http.MethodGet, "/items/1", nil))

	require.Equal(t, 1.0, counterValue(t, collector.callsTotal, http.MethodGet, itemsRoute, "404"))
	require.Equal(t, uint64(1), histogramSampleCount(t, collector.durationHistogramVec, http.MethodGet, itemsRoute, "4xx"))
}

func TestRequestBodySizeIsCounted(t *testing.T) {
	router, collector := newTestRouter(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	serve(router, httptest.NewRequest(http.MethodPost, "/items/1", strings.NewReader("payload")))

	require.Equal(t, 7.0, counterValue(t, collector.requestBodySizeBytes, http.MethodPost, itemsRoute))
}

func TestRequestLabelIsAddedToEveryMetric(t *testing.T) {
	clientLabel := WithRequestLabel("client", func(r *http.Request) string {
		return r.Header.Get("X-Client")
	})
	router, collector := newTestRouter(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}, clientLabel)

	request := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	request.Header.Set("X-Client", "ios")
	serve(router, request)

	require.Equal(t, 1.0, counterValue(t, collector.callsTotal, http.MethodGet, itemsRoute, "200", "ios"))
	require.Equal(t, uint64(1), histogramSampleCount(t, collector.durationHistogramVec, http.MethodGet, itemsRoute, "2xx", "ios"))
	require.Equal(t, 0.0, counterValue(t, collector.requestBodySizeBytes, http.MethodGet, itemsRoute, "ios"))
	require.Equal(t, 2.0, counterValue(t, collector.responseBodySizeBytes, http.MethodGet, itemsRoute, "ios"))
}

func TestStatusClass(t *testing.T) {
	require.Equal(t, "2xx", statusClass(http.StatusOK))
	require.Equal(t, "5xx", statusClass(http.StatusBadGateway))
	require.Equal(t, "unknown", statusClass(0))
}
