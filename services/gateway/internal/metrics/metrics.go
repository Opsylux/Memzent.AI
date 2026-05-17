package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "memzent_gateway_http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"path", "method", "status_code"},
	)

	RequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "memzent_gateway_http_request_duration_seconds",
			Help:    "Histogram of request durations.",
			Buckets: []float64{0.1, 0.3, 0.5, 1, 2, 5},
		},
		[]string{"path", "method"},
	)

	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_gateway_cache_hits_total",
			Help: "Total number of semantic cache hits.",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_gateway_cache_misses_total",
			Help: "Total number of semantic cache misses.",
		},
	)
)
