package mcp

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// CompressToolOutput — rule-based pruning and truncation
// ---------------------------------------------------------------------------

func TestCompressToolOutput_ErrorIntentFiltersLines(t *testing.T) {
	raw := `INFO: system boot
ERROR: database connection failed
DEBUG: heartbeat ok
CRITICAL: memory threshold exceeded
INFO: nothing wrong here
FAIL: retry limit reached`

	result := CompressToolOutput(raw, "find all errors")

	// Must keep error/critical/fail lines
	if !strings.Contains(result, "ERROR: database connection failed") {
		t.Error("Should keep ERROR lines for error-intent queries")
	}
	if !strings.Contains(result, "CRITICAL: memory threshold exceeded") {
		t.Error("Should keep CRITICAL lines for error-intent queries")
	}
	if !strings.Contains(result, "FAIL: retry limit reached") {
		t.Error("Should keep FAIL lines for error-intent queries")
	}

	// Must strip unrelated lines
	if strings.Contains(result, "INFO: system boot") {
		t.Error("Should NOT keep plain INFO lines for error-intent queries")
	}
	if strings.Contains(result, "DEBUG: heartbeat ok") {
		t.Error("Should NOT keep DEBUG lines for error-intent queries")
	}
}

func TestCompressToolOutput_FailKeywordAlsoTriggersFilter(t *testing.T) {
	raw := `INFO: all good
ERROR: something failed`

	// "fail" in intent should trigger error-mode filtering
	result := CompressToolOutput(raw, "show me what failed")
	if !strings.Contains(result, "ERROR: something failed") {
		t.Error("'fail' keyword in intent should trigger error-mode line filtering")
	}
}

func TestCompressToolOutput_DefaultModeKeeps50Lines(t *testing.T) {
	// Build a 100-line output
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "line content here")
	}
	raw := strings.Join(lines, "\n")

	result := CompressToolOutput(raw, "summarize activity")
	resultLines := strings.Split(result, "\n")

	// Should be 51 lines: 50 content + 1 truncation footer
	if len(resultLines) > 52 {
		t.Errorf("Default mode should keep at most 50 lines + footer, got %d lines", len(resultLines))
	}
}

func TestCompressToolOutput_TruncationFooterAddedWhenCut(t *testing.T) {
	var lines []string
	for i := 0; i < 60; i++ {
		lines = append(lines, "data line")
	}
	raw := strings.Join(lines, "\n")

	result := CompressToolOutput(raw, "general query")
	if !strings.Contains(result, "Memzent Truncated") {
		t.Error("Should append truncation notice when output was cut")
	}
}

func TestCompressToolOutput_NoTruncationFooterWhenUnderLimit(t *testing.T) {
	raw := "line 1\nline 2\nline 3"

	result := CompressToolOutput(raw, "general query")
	if strings.Contains(result, "Memzent Truncated") {
		t.Error("Should NOT append truncation notice when output fits within limit")
	}
}

func TestCompressToolOutput_EmptyRawOutput(t *testing.T) {
	result := CompressToolOutput("", "anything")
	// Must not panic and must return a sensible (possibly empty) result
	if result == "" {
		return // Acceptable
	}
	// If not empty it also shouldn't have truncation notice
	if strings.Contains(result, "Memzent Truncated") {
		t.Error("Empty input should not produce a truncation footer")
	}
}

func TestCompressToolOutput_ErrorIntentCaseInsensitive(t *testing.T) {
	raw := "ERROR: something went wrong"

	// Both cased versions of "error" in intent must activate filtering
	r1 := CompressToolOutput(raw, "ERROR")
	r2 := CompressToolOutput(raw, "error")
	r3 := CompressToolOutput(raw, "Error")

	for i, r := range []string{r1, r2, r3} {
		if !strings.Contains(r, "ERROR: something went wrong") {
			t.Errorf("Case variant %d: error-mode filtering should be case-insensitive on intent", i+1)
		}
	}
}

func TestCompressToolOutput_SingleLine(t *testing.T) {
	raw := "single line of output"
	result := CompressToolOutput(raw, "general")
	if !strings.Contains(result, "single line of output") {
		t.Error("Single-line output should be preserved")
	}
	if strings.Contains(result, "Memzent Truncated") {
		t.Error("Single line should not trigger truncation footer")
	}
}

// ---------------------------------------------------------------------------
// MCPClient — nil guard on CallTool
// ---------------------------------------------------------------------------

func TestMCPClient_NilClientCallToolReturnsError(t *testing.T) {
	c := &MCPClient{client: nil}
	_, err := c.CallTool(nil, "some_tool", map[string]any{"key": "val"})
	if err == nil {
		t.Fatal("Expected error when MCP client is nil, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("Expected 'not configured' error, got: %v", err)
	}
}

func TestMCPClient_NilClientInitializeReturnsError(t *testing.T) {
	c := &MCPClient{client: nil}
	err := c.Initialize(nil)
	if err == nil {
		t.Fatal("Expected error when MCP client is nil, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("Expected 'not configured' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// NewMCPClient — MCP_SERVER_URL env resolution
// ---------------------------------------------------------------------------

func TestNewMCPClient_DefaultURL(t *testing.T) {
	// Just tests that construction succeeds when MCP_SERVER_URL is unset;
	// it does not open a real connection.
	t.Setenv("MCP_SERVER_URL", "")
	c, err := NewMCPClient()
	if err != nil {
		t.Fatalf("NewMCPClient() unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("NewMCPClient() returned nil client")
	}
}

func TestNewMCPClient_CustomURL(t *testing.T) {
	t.Setenv("MCP_SERVER_URL", "http://custom-mcp-host:9999/mcp")
	c, err := NewMCPClient()
	if err != nil {
		t.Fatalf("NewMCPClient() unexpected error with custom URL: %v", err)
	}
	if c == nil {
		t.Fatal("NewMCPClient() returned nil client with custom URL")
	}
}
