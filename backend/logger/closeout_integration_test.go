package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
)

func TestHTTPTraceFlowsAcrossHTTPAndDBSinks(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "go-stock-closeout-http-*")
	if err != nil {
		t.Fatalf("create temp root dir: %v", err)
	}
	runtime := MustInit(Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})
	runtime.AttachPayloadStore(NewPayloadStore(filepath.Join(rootDir, "logs"), 16, 5<<20))

	handler := runtime.HTTPMiddleware("integration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.ForSink(SinkDB, "integration.db").WithTrace(runtime.TraceOrNew(r.Context(), "db")).Info(
			"db.integration.hit",
			"handled db step",
			String("table", "trace_probe"),
		)
		_, _ = w.Write([]byte(strings.Repeat("x", 128)))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/trace-probe", strings.NewReader(strings.Repeat("q", 128)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	httpLog, err := os.ReadFile(filepath.Join(rootDir, "logs", "http.log"))
	if err != nil {
		t.Fatalf("read http log: %v", err)
	}
	dbLog, err := os.ReadFile(filepath.Join(rootDir, "logs", "db.log"))
	if err != nil {
		t.Fatalf("read db log: %v", err)
	}

	httpEvent := findLogEventByName(t, string(httpLog), "http.request.completed")
	dbEvent := findLogEventByName(t, string(dbLog), "db.integration.hit")
	traceID := requireStringField(t, httpEvent, "trace_id")
	if requireStringField(t, dbEvent, "trace_id") != traceID {
		t.Fatalf("expected db.integration.hit to reuse %s, got %#v", traceID, dbEvent)
	}

	payloadRoot := filepath.Join(rootDir, "logs", "payloads")
	for _, key := range []string{"request_payload_file", "response_payload_file"} {
		payloadPath := requireStringField(t, httpEvent, key)
		if !pathIsUnderRoot(payloadRoot, payloadPath) {
			t.Fatalf("expected %s under %s, got %s", key, payloadRoot, payloadPath)
		}
		if _, err := os.Stat(payloadPath); err != nil {
			t.Fatalf("stat %s %s: %v", key, payloadPath, err)
		}
	}
}

func findLogEventByName(t *testing.T, content string, event string) map[string]any {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry := map[string]any{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("parse log line for %s: %v\nline=%s", event, err, line)
		}
		if entry["event"] == event {
			return entry
		}
	}
	t.Fatalf("expected event %s in log: %s", event, content)
	return nil
}

func requireStringField(t *testing.T, entry map[string]any, key string) string {
	t.Helper()
	value, ok := entry[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		t.Fatalf("expected string field %s in %#v", key, entry)
	}
	return value
}

func pathIsUnderRoot(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
