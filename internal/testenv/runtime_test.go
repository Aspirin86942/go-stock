package testenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/logger"
)

func TestNewLoggerRuntime_UsesArtifactsDirEnvAndWritesMetadataFields(t *testing.T) {
	artifactsRoot := filepath.Join(t.TempDir(), "artifacts-root")
	t.Setenv(ArtifactsDirEnv, artifactsRoot)

	suite := "backend/logger:runtime suite"
	runtime, artifacts := NewLoggerRuntime(t, suite)

	trace := logger.TraceContext{
		TraceID:      "trace-runtime-helper",
		SpanID:       "span-runtime-helper",
		AppSessionID: "session-runtime-helper",
		Source:       "http",
	}
	runtime.ForSink(logger.SinkApp, "runtime.helper").WithTrace(trace).Info("app.start", "app sink event")
	runtime.ForSink(logger.SinkHTTP, "runtime.helper").WithTrace(trace).Info("http.request.completed", "http sink event")

	if !strings.HasPrefix(filepath.Clean(artifacts.RootDir), filepath.Clean(artifactsRoot)) {
		t.Fatalf("expected artifacts root under %q, got %q", artifactsRoot, artifacts.RootDir)
	}

	rel, err := filepath.Rel(artifactsRoot, artifacts.RootDir)
	if err != nil {
		t.Fatalf("compute relative artifacts path: %v", err)
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		t.Fatalf("expected artifacts path to include test run/suite/case segments, got %q", rel)
	}
	segments := strings.Split(rel, string(os.PathSeparator))
	wantSegments := []string{
		sanitizePathFragment(artifacts.TestRunID),
		sanitizePathFragment(artifacts.TestSuite),
		sanitizePathFragment(artifacts.TestCase),
	}
	if len(segments) != len(wantSegments) {
		t.Fatalf("expected artifacts path segments %v, got %v", wantSegments, segments)
	}
	for index, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}
		if strings.ContainsAny(segment, ":/\\ \t\r\n") {
			t.Fatalf("expected sanitized artifact path segment, got %q in %q", segment, rel)
		}
		if segment != wantSegments[index] {
			t.Fatalf("expected artifacts path segment %d to be %q, got %q", index, wantSegments[index], segment)
		}
	}

	appPath := filepath.Join(artifacts.LogsDir, "app.log")
	appLog, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("read app log: %v", err)
	}
	httpPath := filepath.Join(artifacts.LogsDir, "http.log")
	httpLog, err := os.ReadFile(httpPath)
	if err != nil {
		t.Fatalf("read http log: %v", err)
	}

	appEntry := parseLastJSONEntry(t, appLog)
	httpEntry := parseLastJSONEntry(t, httpLog)
	for _, entry := range []map[string]any{appEntry, httpEntry} {
		if got, _ := entry["execution_mode"].(string); got != "test" {
			t.Fatalf("expected execution_mode=test, got %#v", got)
		}
		if got, _ := entry["test_run_id"].(string); got != artifacts.TestRunID {
			t.Fatalf("expected test_run_id=%q, got %#v", artifacts.TestRunID, got)
		}
		if got, _ := entry["test_suite"].(string); got != artifacts.TestSuite {
			t.Fatalf("expected test_suite=%q, got %#v", artifacts.TestSuite, got)
		}
		if got, _ := entry["test_case"].(string); got != artifacts.TestCase {
			t.Fatalf("expected test_case=%q, got %#v", artifacts.TestCase, got)
		}
		if got, _ := entry["source"].(string); got != "http" {
			t.Fatalf("expected source=http from trace context, got %#v", got)
		}
	}
}

func TestResolveArtifactsRoot_AnchorsRelativeOverrideAtRepoRoot(t *testing.T) {
	got := resolveArtifactsRoot(func(string) string {
		return filepath.Join("artifacts", "testlogs", "task-3")
	})
	want := filepath.Join(repoRoot(), "artifacts", "testlogs", "task-3")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("expected relative override to resolve under repo root %q, got %q", want, got)
	}
}

func TestResolveArtifactsRoot_StripsParentTraversalFromRelativeOverride(t *testing.T) {
	got := resolveArtifactsRoot(func(string) string {
		return filepath.Join("..", "tmp", "logs")
	})
	want := filepath.Join(repoRoot(), "tmp", "logs")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("expected traversal segments to be stripped under repo root %q, got %q", want, got)
	}
}

func parseLastJSONEntry(t *testing.T, content []byte) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Fatalf("expected JSON log content, got %q", string(content))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("decode log line %q: %v", lines[len(lines)-1], err)
	}
	return entry
}
