package metric

import (
	"github.com/form3tech-oss/http-message-signing-proxy/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	promNamespace       = "signing_proxy"
	labelUpstreamTarget = "upstream_target"
	labelMethod         = "method"
	labelPath           = "path"
)

var (
	commonLabels = []string{
		labelUpstreamTarget,
		labelMethod,
		labelPath,
	}
	errorCounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: promNamespace,
			Name:      "total_internal_error_count",
			Help:      "Total number of internal errors",
		},
		commonLabels,
	)
	totalReqCounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: promNamespace,
			Name:      "total_request_count",
			Help:      "Total number of incoming requests",
		},
		commonLabels,
	)
	totalSignedReqCounterVec = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: promNamespace,
			Name:      "total_signed_request_count",
			Help:      "Total number of incoming requests that are signed",
		},
		commonLabels,
	)
)

var (
	signingDurationHistogramVec = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: promNamespace,
			Name:      "signing_duration_seconds",
			Help:      "Request signing duration time in seconds",
			// 20 buckets range from 2ms to 40ms, request signing is rather fast
			Buckets: prometheus.LinearBuckets(0.002, 0.002, 20),
		},
		commonLabels,
	)
	requestDurationHistogramVec = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: promNamespace,
			Name:      "request_duration_seconds",
			Help:      "Total request duration time in seconds, including signing and upstream processing",
			// 20 buckets range from 50ms to 1s, since upstream duration is unknown
			Buckets: prometheus.LinearBuckets(0.05, 0.05, 20),
		},
		commonLabels,
	)
)

type metricPublisher struct {
	upstreamTarget string
}

func NewMetricPublisher(upstreamTarget string) proxy.MetricPublisher {
	return &metricPublisher{
		upstreamTarget: upstreamTarget,
	}
}

func (m *metricPublisher) IncrementTotalRequestCount(method string, path string) {
	totalReqCounterVec.With(m.getCommonLabels(method, path)).Inc()
}

func (m *metricPublisher) IncrementSignedRequestCount(method string, path string) {
	totalSignedReqCounterVec.With(m.getCommonLabels(method, path)).Inc()
}

func (m *metricPublisher) IncrementInternalErrorCount(method string, path string) {
	errorCounterVec.With(m.getCommonLabels(method, path)).Inc()
}

func (m *metricPublisher) MeasureSigningDuration(method string, path string, duration float64) {
	signingDurationHistogramVec.With(m.getCommonLabels(method, path)).Observe(duration)
}

func (m *metricPublisher) MeasureTotalDuration(method string, path string, duration float64) {
	requestDurationHistogramVec.With(m.getCommonLabels(method, path)).Observe(duration)
}

func (m *metricPublisher) getCommonLabels(method string, path string) prometheus.Labels {
	return prometheus.Labels{
		labelUpstreamTarget: m.upstreamTarget,
		labelMethod:         method,
		labelPath:           path,
	}
}
