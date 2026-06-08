package featureflags

import (
	"os"
	"strings"
	"sync"
)

// Flags holds the global feature flag state for evolution layers.
// All flags default to true (enabled) unless explicitly set to "false" via env vars.
type Flags struct {
	L1bCache        bool // MEMZENT_L1B_ENABLED — entity-keyed hot path cache
	OfflinePlane    bool // MEMZENT_OFFLINE_ENABLED — offline learning plane (O1/O2/O3 miners)
	OfflineStreams  bool // MEMZENT_OFFLINE_STREAMS — use Valkey Streams instead of in-memory channels
	WorkflowEngine  bool // MEMZENT_WORKFLOW_ENABLED — workflow registry + engine shortcut
	EntityMetrics   bool // MEMZENT_ENTITY_METRICS_ENABLED — entity quality + GPU avoidance counters
}

var (
	global *Flags
	once   sync.Once
)

// Load reads feature flags from environment variables.
// Missing or non-"false" values default to enabled (true).
func Load() *Flags {
	once.Do(func() {
		global = &Flags{
			L1bCache:       envBool("MEMZENT_L1B_ENABLED", true),
			OfflinePlane:   envBool("MEMZENT_OFFLINE_ENABLED", true),
			OfflineStreams: envBool("MEMZENT_OFFLINE_STREAMS", false),
			WorkflowEngine: envBool("MEMZENT_WORKFLOW_ENABLED", true),
			EntityMetrics:  envBool("MEMZENT_ENTITY_METRICS_ENABLED", true),
		}
	})
	return global
}

// Get returns the loaded flags (Load must be called first).
func Get() *Flags {
	if global == nil {
		return Load()
	}
	return global
}

// Reset clears the cached flags (for testing only).
func Reset() {
	once = sync.Once{}
	global = nil
}

func envBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return !strings.EqualFold(v, "false") && v != "0"
}
