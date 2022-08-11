package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestName(t *testing.T) {
	b := prometheus.ExponentialBucketsRange(0.005, 5, 20)
	t.Log(b)
	b = prometheus.LinearBuckets(0.05, 0.05, 20)
	t.Log(b)
}
