package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	// Capture log output
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a simple test handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}

	// Wrap with logging middleware
	handler := LoggingMiddleware(logger)(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	// Execute request
	handler(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify request ID header was set
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}

	// Verify logs were written
	logContent := logOutput.String()
	if !strings.Contains(logContent, "request started") {
		t.Error("Expected 'request started' log entry")
	}
	if !strings.Contains(logContent, "request completed") {
		t.Error("Expected 'request completed' log entry")
	}

	// Parse log entries to verify structure
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 log lines, got %d", len(lines))
	}

	// Verify first log entry (request started)
	var startLog map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &startLog); err != nil {
		t.Errorf("Failed to parse start log: %v", err)
	}

	expectedFields := []string{"request_id", "method", "path", "query", "user_agent"}
	for _, field := range expectedFields {
		if _, exists := startLog[field]; !exists {
			t.Errorf("Expected field '%s' in start log", field)
		}
	}

	// Verify second log entry (request completed)
	var endLog map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &endLog); err != nil {
		t.Errorf("Failed to parse end log: %v", err)
	}

	expectedEndFields := []string{"request_id", "method", "path", "status_code", "duration"}
	for _, field := range expectedEndFields {
		if _, exists := endLog[field]; !exists {
			t.Errorf("Expected field '%s' in end log", field)
		}
	}

	// Verify request IDs match
	if startLog["request_id"] != endLog["request_id"] {
		t.Error("Request IDs should match between start and end logs")
	}
}

func TestGetRequestID(t *testing.T) {
	// Test with no request ID in context
	ctx := context.Background()
	if id := GetRequestID(ctx); id != "" {
		t.Errorf("Expected empty string, got %s", id)
	}

	// Test with request ID in context
	expectedID := "test-request-id"
	ctx = context.WithValue(ctx, requestIDKey, expectedID)
	if id := GetRequestID(ctx); id != expectedID {
		t.Errorf("Expected %s, got %s", expectedID, id)
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	// Should generate non-empty IDs
	if id1 == "" || id2 == "" {
		t.Error("Generated request IDs should not be empty")
	}

	// Should generate unique IDs
	if id1 == id2 {
		t.Error("Generated request IDs should be unique")
	}

	// Should be hex encoded (16 characters for 8 bytes)
	if len(id1) != 16 || len(id2) != 16 {
		t.Errorf("Expected 16 character hex string, got %d and %d", len(id1), len(id2))
	}
}
