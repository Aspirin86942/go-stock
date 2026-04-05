package data

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"
)

func TestDataModuleLog_WithContextReusesTrace(t *testing.T) {
	logRoot, err := os.MkdirTemp("", "go-stock-utils-trace-*")
	if err != nil {
		t.Fatalf("create temp log root: %v", err)
	}
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: logRoot,
			LogsDir: filepath.Join(logRoot, "logs"),
		},
		EnableStdout: false,
	})

	existing := logger.TraceContext{
		TraceID:      "trace-data-1",
		SpanID:       "span-data-1",
		AppSessionID: "session-data-1",
		Source:       "http",
	}
	ctx := logger.WithTraceContext(context.Background(), existing)

	dataModuleLogger(logger.SinkAI, "data.trace_test").
		WithContext(ctx, "data-test").
		Info("data.trace_test.reuse", "reuse trace from context")

	content, err := os.ReadFile(filepath.Join(logRoot, "logs", "ai.log"))
	if err != nil {
		t.Fatalf("read ai log: %v", err)
	}
	line := string(content)
	for _, want := range []string{
		`"event":"data.trace_test.reuse"`,
		`"trace_id":"trace-data-1"`,
		`"span_id":"span-data-1"`,
		`"app_session_id":"session-data-1"`,
		`"source":"http"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected log to contain %s, got %s", want, line)
		}
	}
}
