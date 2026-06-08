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

	// Entity Extraction Quality Metrics (E5)
	EntityRegexSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_entity_regex_success_total",
			Help: "Total entity extractions that succeeded via regex (no LLM needed).",
		},
	)

	EntityRegexFailure = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_entity_regex_failure_total",
			Help: "Total entity extractions where regex produced no results.",
		},
	)

	EntityMismatchTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_entity_mismatch_total",
			Help: "Total cache guard rejections due to entity mismatch.",
		},
	)

	EntityLLMUsage = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_entity_llm_usage_total",
			Help: "Total requests where LLM was needed for entity extraction.",
		},
	)

	// Cache Layer Distribution (E5)
	CacheLayerHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "memzent_cache_layer_hits_total",
			Help: "Cache hits by layer (L1, L1b, L2, L5).",
		},
		[]string{"layer"},
	)

	// GPU Avoidance Rate — the key business metric
	GPUAvoidanceTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_gpu_avoidance_total",
			Help: "Total requests that avoided GPU/LLM invocation (served from cache/workflow).",
		},
	)

	GPUInvocationTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memzent_gpu_invocation_total",
			Help: "Total requests that required GPU/LLM invocation.",
		},
	)
)
