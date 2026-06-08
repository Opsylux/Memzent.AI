package featureflags

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	Reset()
	// Clear all env vars to test defaults
	envVars := []string{
		"MEMZENT_L1B_ENABLED",
		"MEMZENT_OFFLINE_ENABLED",
		"MEMZENT_OFFLINE_STREAMS",
		"MEMZENT_WORKFLOW_ENABLED",
		"MEMZENT_ENTITY_METRICS_ENABLED",
		"MEMZENT_PATTERN_MINING_ENABLED",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	flags := Load()

	if !flags.L1bCache {
		t.Error("L1bCache should default to true")
	}
	if !flags.OfflinePlane {
		t.Error("OfflinePlane should default to true")
	}
	if flags.OfflineStreams {
		t.Error("OfflineStreams should default to false")
	}
	if !flags.WorkflowEngine {
		t.Error("WorkflowEngine should default to true")
	}
	if !flags.EntityMetrics {
		t.Error("EntityMetrics should default to true")
	}
	if flags.PatternMining {
		t.Error("PatternMining should default to false")
	}
}

func TestLoad_ExplicitFalse(t *testing.T) {
	Reset()
	os.Setenv("MEMZENT_L1B_ENABLED", "false")
	os.Setenv("MEMZENT_OFFLINE_ENABLED", "0")
	os.Setenv("MEMZENT_WORKFLOW_ENABLED", "FALSE")
	defer func() {
		os.Unsetenv("MEMZENT_L1B_ENABLED")
		os.Unsetenv("MEMZENT_OFFLINE_ENABLED")
		os.Unsetenv("MEMZENT_WORKFLOW_ENABLED")
	}()

	flags := Load()

	if flags.L1bCache {
		t.Error("L1bCache should be false when env is 'false'")
	}
	if flags.OfflinePlane {
		t.Error("OfflinePlane should be false when env is '0'")
	}
	if flags.WorkflowEngine {
		t.Error("WorkflowEngine should be false when env is 'FALSE'")
	}
}

func TestLoad_ExplicitTrue(t *testing.T) {
	Reset()
	os.Setenv("MEMZENT_PATTERN_MINING_ENABLED", "true")
	os.Setenv("MEMZENT_OFFLINE_STREAMS", "1")
	defer func() {
		os.Unsetenv("MEMZENT_PATTERN_MINING_ENABLED")
		os.Unsetenv("MEMZENT_OFFLINE_STREAMS")
	}()

	flags := Load()

	if !flags.PatternMining {
		t.Error("PatternMining should be true when env is 'true'")
	}
	if !flags.OfflineStreams {
		t.Error("OfflineStreams should be true when env is '1'")
	}
}

func TestGet_LazyLoads(t *testing.T) {
	Reset()
	flags := Get()
	if flags == nil {
		t.Fatal("Get() should never return nil")
	}
}

func TestReset_ClearsState(t *testing.T) {
	Reset()
	os.Setenv("MEMZENT_L1B_ENABLED", "false")
	f1 := Load()
	if f1.L1bCache {
		t.Error("Should be false after setting env")
	}

	Reset()
	os.Unsetenv("MEMZENT_L1B_ENABLED")
	f2 := Load()
	if !f2.L1bCache {
		t.Error("Should be true after reset + unset env")
	}
}

func TestEnvBool_EdgeCases(t *testing.T) {
	cases := []struct {
		value    string
		defVal   bool
		expected bool
	}{
		{"", true, true},
		{"", false, false},
		{"false", true, false},
		{"FALSE", true, false},
		{"False", true, false},
		{"0", true, false},
		{"true", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"anything", false, true},
	}

	for _, tc := range cases {
		os.Setenv("TEST_FLAG", tc.value)
		if tc.value == "" {
			os.Unsetenv("TEST_FLAG")
		}
		got := envBool("TEST_FLAG", tc.defVal)
		if got != tc.expected {
			t.Errorf("envBool(%q, %v) = %v, want %v", tc.value, tc.defVal, got, tc.expected)
		}
	}
	os.Unsetenv("TEST_FLAG")
}
