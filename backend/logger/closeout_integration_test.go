package logger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
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

	traceID := extractTraceID(string(httpLog), "http.request.completed")
	if traceID == "" {
		t.Fatalf("expected http log trace id, got %s", string(httpLog))
	}
	if !strings.Contains(string(dbLog), traceID) {
		t.Fatalf("expected db log to reuse %s, got %s", traceID, string(dbLog))
	}

	payloadFiles, err := filepath.Glob(filepath.Join(rootDir, "logs", "payloads", "*", "*.log"))
	if err != nil {
		t.Fatalf("glob payload files: %v", err)
	}
	if len(payloadFiles) == 0 {
		t.Fatalf("expected spilled payload file under logs/payloads")
	}
}

func extractTraceID(content string, event string) string {
	re := regexp.MustCompile(`"trace_id":"([^"]+)"`)
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		if !strings.Contains(line, `"`+"event"+`":"`+event+`"`) {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}
