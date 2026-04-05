package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"go-stock/internal/testenv"
)

func TestReleaseSmoke_FrontendBridgeWritesStructuredLog(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	_, artifacts := testenv.NewLoggerRuntime(t, "release-smoke")

	app := NewApp()
	if next := strings.TrimSpace(app.CalculateNextRunTime("0 0 0 * * ?")); next == "" {
		t.Fatalf("expected cron expression to produce a non-empty next run time")
	}

	logFrontendRuntimeError([]interface{}{
		map[string]any{
			"page":    "stock.vue",
			"route":   "/stock",
			"message": "release smoke frontend boom",
			"source":  "main.ts",
			"traceId": "trace-release-smoke",
			"error":   "Error: release smoke frontend boom",
		},
	})

	content, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "frontend.log"))
	if err != nil {
		t.Fatalf("read frontend log: %v", err)
	}
	for _, want := range []string{
		`"event":"frontend.error"`,
		`"trace_id":"trace-release-smoke"`,
		`"execution_mode":"test"`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected frontend log to contain %s, got %s", want, string(content))
		}
	}
}

func TestReleaseSmoke_AgentMessageBridgeUsesFrontendFieldNames(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)

	payload := agentMessageToFrontendMap(&schema.Message{
		Role:             schema.Assistant,
		Content:          "bridge ok",
		ReasoningContent: "thinking",
	})

	if got, _ := payload["role"].(string); got != string(schema.Assistant) {
		t.Fatalf("expected role=%q, got %#v", string(schema.Assistant), payload["role"])
	}
	if got, _ := payload["content"].(string); got != "bridge ok" {
		t.Fatalf("expected content to round-trip, got %#v", payload["content"])
	}
	if got, _ := payload["reasoning_content"].(string); got != "thinking" {
		t.Fatalf("expected reasoning_content to use frontend field name, got %#v", payload["reasoning_content"])
	}
	if _, exists := payload["ReasoningContent"]; exists {
		t.Fatalf("expected frontend bridge payload not to expose Go field names, got %#v", payload)
	}
}
