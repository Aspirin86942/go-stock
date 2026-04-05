package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"
)

func TestLogFrontendRuntimeError_ReusesPayloadTraceID(t *testing.T) {
	root, err := os.MkdirTemp("", "go-stock-frontend-log-*")
	if err != nil {
		t.Fatalf("create temp root: %v", err)
	}
	logsDir := filepath.Join(root, "logs")
	_ = logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: root,
			LogsDir: logsDir,
		},
		EnableStdout: false,
	})

	logFrontendRuntimeError([]interface{}{
		map[string]any{
			"page":    "stock.vue",
			"route":   "/stock",
			"message": "frontend boom",
			"source":  "main.ts",
			"lineno":  12,
			"colno":   7,
			"error":   "Error: frontend boom",
			"traceId": "trace-from-ui-123",
		},
	})

	content, err := os.ReadFile(filepath.Join(logsDir, "frontend.log"))
	if err != nil {
		t.Fatalf("read frontend log: %v", err)
	}

	entry := parseLatestJSONLogEntry(t, string(content))
	if got, _ := entry["event"].(string); got != "frontend.error" {
		t.Fatalf("expected event frontend.error, got %#v", got)
	}
	if got, _ := entry["trace_id"].(string); got != "trace-from-ui-123" {
		t.Fatalf("expected trace_id to reuse payload traceId, got %#v", got)
	}
	if got, _ := entry["error_class"].(string); got != "frontend_error" {
		t.Fatalf("expected error_class frontend_error, got %#v", got)
	}
}

func parseLatestJSONLogEntry(t *testing.T, content string) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected at least one log line, got empty content")
	}

	last := strings.TrimSpace(lines[len(lines)-1])
	if last == "" {
		t.Fatalf("expected non-empty latest log line")
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(last), &entry); err != nil {
		t.Fatalf("parse latest json log entry %q: %v", last, err)
	}
	return entry
}
