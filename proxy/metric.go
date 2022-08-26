package proxy

type MetricPublisher interface {
	IncrementTotalRequestCount(method string, path string)
	IncrementSignedRequestCount(method string, path string)
	IncrementInternalErrorCount(method string, path string)
	MeasureSigningDuration(method string, path string, duration float64)
	MeasureTotalDuration(method string, path string, duration float64)
}
