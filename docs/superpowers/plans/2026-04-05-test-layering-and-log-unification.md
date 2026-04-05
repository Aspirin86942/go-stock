# Test Layering And Log Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a layered test system with shared gates, unified test logging, centralized test scripts, CI split by `default` / `external` / `release`, and migrate ordinary tests away from legacy `logger.SugaredLogger`.

**Architecture:** Centralize test execution gates and logger-runtime helpers in a small `internal/testenv` package, then let `backend/logger` attach test execution metadata at the sink/runtime level so HTTP, DB, task, frontend, and AI logs all inherit the same test identity fields. Keep package tests close to code, move orchestration into `scripts/testing/`, split CI into three workflows, and sweep legacy test logging family-by-family before adding a guard that forbids new `SugaredLogger` usage outside logger compatibility coverage.

**Tech Stack:** Go 1.26, Go `testing`, `httptest`, `go.uber.org/zap`, PowerShell (`pwsh`), GitHub Actions, Node `--test`, Vite.

---

## File Map

- Create: `D:\codex_work\go-stock\internal\testenv\gates.go`
  - Shared gate helpers for external and release-smoke tests.
- Create: `D:\codex_work\go-stock\internal\testenv\gates_test.go`
  - Unit tests for gate-decision logic.
- Create: `D:\codex_work\go-stock\internal\testenv\runtime.go`
  - Minimal test logger/artifact helper that can write into `artifacts/testlogs`.
- Create: `D:\codex_work\go-stock\internal\testenv\runtime_test.go`
  - Tests for artifact-root override, path sanitization, and logger metadata fields.
- Modify: `D:\codex_work\go-stock\integration_test.go`
  - Root-package wrapper that delegates legacy `requireIntegrationTest` to shared external gate logic.
- Modify: `D:\codex_work\go-stock\backend\data\integration_test.go`
  - Same delegation for `backend/data`.
- Modify: `D:\codex_work\go-stock\backend\agent\integration_test.go`
  - Same delegation for `backend/agent`.
- Modify: `D:\codex_work\go-stock\backend\models\integration_test.go`
  - Same delegation for `backend/models`.
- Modify: `D:\codex_work\go-stock\backend\logger\types.go`
  - Add runtime-level config fields for shared execution metadata.
- Modify: `D:\codex_work\go-stock\backend\logger\sinks.go`
  - Bind config fields into every sink logger at bootstrap time.
- Modify: `D:\codex_work\go-stock\backend\logger\core_test.go`
  - Verify runtime-level config fields are present without clobbering trace/source fields.
- Modify: `D:\codex_work\go-stock\app_common_test.go`
  - Use shared test runtime helper and assert test metadata in frontend logs.
- Modify: `D:\codex_work\go-stock\ai-assistant-web\server_test.go`
  - Use shared test runtime helper and assert test metadata in HTTP logs.
- Modify: `D:\codex_work\go-stock\backend\logger\closeout_integration_test.go`
  - Use shared test runtime helper and assert metadata across HTTP/DB sink logs.
- Modify: `D:\codex_work\go-stock\backend\agent\cron_task_api_test.go`
  - Use shared test runtime helper and assert metadata across task/AI/DB sink logs.
- Create: `D:\codex_work\go-stock\app_release_smoke_test.go`
  - First explicit release-smoke Go test with logger output and frontend-bridge smoke coverage.
- Create: `D:\codex_work\go-stock\scripts\testing\default.ps1`
  - Repository-level default test entrypoint.
- Create: `D:\codex_work\go-stock\scripts\testing\external.ps1`
  - Repository-level external test entrypoint.
- Create: `D:\codex_work\go-stock\scripts\testing\release-smoke.ps1`
  - Repository-level release-smoke entrypoint.
- Modify: `D:\codex_work\go-stock\.gitignore`
  - Ignore `artifacts/testlogs/`.
- Create: `D:\codex_work\go-stock\.github\workflows\test-default.yml`
  - PR/push stable test workflow.
- Create: `D:\codex_work\go-stock\.github\workflows\test-external.yml`
  - Manual/nightly external test workflow.
- Create: `D:\codex_work\go-stock\.github\workflows\release-smoke.yml`
  - Manual release-smoke workflow.
- Create: `D:\codex_work\go-stock\docs\superpowers\checklists\2026-04-05-test-migration-inventory.md`
  - Explicit classification of retained default, external, release, and manual test files.
- Modify: `D:\codex_work\go-stock\app_test.go`
  - Remove legacy test logger usage, strengthen assertions, and move screen-resolution probe into release smoke.
- Modify: `D:\codex_work\go-stock\backend\agent\agent_test.go`
  - Replace `SugaredLogger` with test assertions and `t.Logf`, remove side-effect file writes.
- Modify: `D:\codex_work\go-stock\backend\models\models_test.go`
  - Replace `SugaredLogger` with `t.Logf` and explicit failure handling.
- Modify: `D:\codex_work\go-stock\backend\data\utils_test.go`
  - Replace logger prints with real assertions.
- Create: `D:\codex_work\go-stock\backend\data\external_test.go`
  - Small helper functions for external data tests (`logValue`, `requirePositiveLen`, `requireNotBlank`).
- Modify: `D:\codex_work\go-stock\backend\data\openai_api_test.go`
  - Remove legacy logger usage and strengthen non-empty assertions.
- Modify: `D:\codex_work\go-stock\backend\data\market_news_api_test.go`
  - Remove legacy logger usage and add representative data assertions.
- Modify: `D:\codex_work\go-stock\backend\data\search_stock_api_test.go`
  - Remove legacy logger usage and assert markdown/non-empty results.
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api_test.go`
  - Remove legacy logger usage and assert non-empty K-line responses.
- Modify: `D:\codex_work\go-stock\backend\data\crawler_api_test.go`
  - Remove legacy logger usage and assert crawled content is non-empty.
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api_test.go`
  - Remove legacy logger usage and strengthen representative data assertions.
- Modify: `D:\codex_work\go-stock\backend\data\stock_sentiment_analysis_test.go`
  - Remove legacy logger usage and assert summary output is non-empty.
- Modify: `D:\codex_work\go-stock\backend\data\alert_windows_api_test.go`
  - Move notification probe to release smoke gate and remove legacy logger usage.
- Modify: `D:\codex_work\go-stock\backend\data\alert_darwin_api_test.go`
  - Same for macOS notification probe.
- Create: `D:\codex_work\go-stock\backend\logger\legacy_test_usage_guard_test.go`
  - Guard test that bans `logger.SugaredLogger` in `_test.go` files outside explicit compatibility allowlist.

### Task 1: Centralize External And Release Gates

**Files:**
- Create: `D:\codex_work\go-stock\internal\testenv\gates.go`
- Create: `D:\codex_work\go-stock\internal\testenv\gates_test.go`
- Modify: `D:\codex_work\go-stock\integration_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\integration_test.go`
- Modify: `D:\codex_work\go-stock\backend\agent\integration_test.go`
- Modify: `D:\codex_work\go-stock\backend\models\integration_test.go`

- [ ] **Step 1: Write failing gate-decision tests**

```go
// D:\codex_work\go-stock\internal\testenv\gates_test.go
package testenv

import "testing"

func TestExternalGateDecision_UsesNewOrLegacyEnv(t *testing.T) {
	testCases := []struct {
		name        string
		shortMode   bool
		env         map[string]string
		wantEnabled bool
		wantReason  string
	}{
		{
			name:        "short mode wins",
			shortMode:   true,
			env:         map[string]string{ExternalEnv: "1"},
			wantEnabled: false,
			wantReason:  "skipping external test in short mode",
		},
		{
			name:        "new env enables external tests",
			env:         map[string]string{ExternalEnv: "1"},
			wantEnabled: true,
		},
		{
			name:        "legacy integration env still enables external tests",
			env:         map[string]string{LegacyIntegrationEnv: "1"},
			wantEnabled: true,
		},
		{
			name:        "missing env skips external tests",
			env:         map[string]string{},
			wantEnabled: false,
			wantReason:  "skipping external test; set GO_STOCK_RUN_EXTERNAL_TESTS=1 to enable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enabled, reason := externalGateDecision(func(key string) string {
				return tc.env[key]
			}, tc.shortMode)
			if enabled != tc.wantEnabled {
				t.Fatalf("externalGateDecision() enabled=%v, want %v", enabled, tc.wantEnabled)
			}
			if reason != tc.wantReason {
				t.Fatalf("externalGateDecision() reason=%q, want %q", reason, tc.wantReason)
			}
		})
	}
}

func TestReleaseSmokeGateDecision_RequiresDedicatedEnv(t *testing.T) {
	enabled, reason := releaseSmokeGateDecision(func(string) string { return "" }, false)
	if enabled {
		t.Fatalf("expected release smoke tests to stay disabled without env gate")
	}
	if reason != "skipping release smoke test; set GO_STOCK_RUN_RELEASE_SMOKE=1 to enable" {
		t.Fatalf("unexpected release smoke skip reason: %q", reason)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/testenv -run "TestExternalGateDecision_UsesNewOrLegacyEnv|TestReleaseSmokeGateDecision_RequiresDedicatedEnv" -count=1`

Expected: FAIL because `internal/testenv` and the gate helpers do not exist yet.

- [ ] **Step 3: Implement shared gate helpers and route legacy wrappers through them**

```go
// D:\codex_work\go-stock\internal\testenv\gates.go
package testenv

import (
	"os"
	"testing"
)

const (
	ExternalEnv          = "GO_STOCK_RUN_EXTERNAL_TESTS"
	LegacyIntegrationEnv = "GO_STOCK_RUN_INTEGRATION_TESTS"
	ReleaseSmokeEnv      = "GO_STOCK_RUN_RELEASE_SMOKE"
)

func RequireExternalTest(t *testing.T) {
	t.Helper()
	enabled, reason := externalGateDecision(os.Getenv, testing.Short())
	if enabled {
		return
	}
	t.Skip(reason)
}

func RequireReleaseSmokeTest(t *testing.T) {
	t.Helper()
	enabled, reason := releaseSmokeGateDecision(os.Getenv, testing.Short())
	if enabled {
		return
	}
	t.Skip(reason)
}

func externalGateDecision(getenv func(string) string, shortMode bool) (bool, string) {
	if shortMode {
		return false, "skipping external test in short mode"
	}
	if getenv(ExternalEnv) == "1" || getenv(LegacyIntegrationEnv) == "1" {
		return true, ""
	}
	return false, "skipping external test; set GO_STOCK_RUN_EXTERNAL_TESTS=1 to enable"
}

func releaseSmokeGateDecision(getenv func(string) string, shortMode bool) (bool, string) {
	if shortMode {
		return false, "skipping release smoke test in short mode"
	}
	if getenv(ReleaseSmokeEnv) == "1" {
		return true, ""
	}
	return false, "skipping release smoke test; set GO_STOCK_RUN_RELEASE_SMOKE=1 to enable"
}
```

```go
// D:\codex_work\go-stock\integration_test.go
package main

import (
	"testing"

	"go-stock/internal/testenv"
)

func requireIntegrationTest(t *testing.T) {
	t.Helper()
	testenv.RequireExternalTest(t)
}
```

```go
// D:\codex_work\go-stock\backend\data\integration_test.go
// D:\codex_work\go-stock\backend\agent\integration_test.go
// D:\codex_work\go-stock\backend\models\integration_test.go
func requireIntegrationTest(t *testing.T) {
	t.Helper()
	testenv.RequireExternalTest(t)
}
```

- [ ] **Step 4: Run the new unit tests and a compile-only smoke pass**

Run: `go test ./internal/testenv ./backend\data ./backend\agent ./backend\models . -run "TestExternalGateDecision_UsesNewOrLegacyEnv|TestReleaseSmokeGateDecision_RequiresDedicatedEnv|^$" -count=1`

Expected: PASS. The gate tests should pass, and the packages that keep `requireIntegrationTest` wrappers should compile against `internal/testenv`.

- [ ] **Step 5: Commit**

```bash
git add internal/testenv/gates.go internal/testenv/gates_test.go integration_test.go backend/data/integration_test.go backend/agent/integration_test.go backend/models/integration_test.go
git commit -m "test: centralize external and release gate helpers"
```

### Task 2: Add Runtime-Level Test Metadata And Artifact Helper

**Files:**
- Modify: `D:\codex_work\go-stock\backend\logger\types.go`
- Modify: `D:\codex_work\go-stock\backend\logger\sinks.go`
- Modify: `D:\codex_work\go-stock\backend\logger\core_test.go`
- Create: `D:\codex_work\go-stock\internal\testenv\runtime.go`
- Create: `D:\codex_work\go-stock\internal\testenv\runtime_test.go`

- [ ] **Step 1: Write the failing logger/runtime tests**

```go
// Add to D:\codex_work\go-stock\backend\logger\core_test.go
func TestMustInit_AppliesConfigFieldsToEverySink(t *testing.T) {
	rootDir := t.TempDir()
	logsDir := filepath.Join(rootDir, "logs")

	runtime := MustInit(Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: logsDir,
		},
		EnableStdout: false,
		Fields: []zap.Field{
			String("execution_mode", "test"),
			String("test_suite", "logger-core"),
		},
	})

	runtime.ForSink(SinkHTTP, "core").WithTrace(TraceContext{
		TraceID: "trace-http",
		Source:  "http",
	}).Info("http.request.completed", "request complete")

	content, err := os.ReadFile(filepath.Join(logsDir, "http.log"))
	if err != nil {
		t.Fatalf("read http log: %v", err)
	}

	for _, want := range []string{`"execution_mode":"test"`, `"test_suite":"logger-core"`, `"source":"http"`} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected http sink to contain %s, got %s", want, string(content))
		}
	}
}
```

```go
// D:\codex_work\go-stock\internal\testenv\runtime_test.go
package testenv_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/logger"
	"go-stock/internal/testenv"
)

func TestNewLoggerRuntime_UsesArtifactsDirEnvAndWritesMetadataFields(t *testing.T) {
	artifactRoot := t.TempDir()
	t.Setenv(testenv.ArtifactsDirEnv, artifactRoot)

	runtime, artifacts := testenv.NewLoggerRuntime(t, "backend-logger")
	runtime.ForSink(logger.SinkApp, "testenv").WithTrace(logger.TraceContext{
		TraceID: "trace-app",
		Source:  "app",
	}).Info("test.started", "writing scoped test log")

	if !strings.HasPrefix(artifacts.RootDir, artifactRoot) {
		t.Fatalf("expected artifact root under %s, got %s", artifactRoot, artifacts.RootDir)
	}

	content, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "app.log"))
	if err != nil {
		t.Fatalf("read app log: %v", err)
	}

	for _, want := range []string{
		`"execution_mode":"test"`,
		`"test_suite":"backend-logger"`,
		`"test_case":"TestNewLoggerRuntime_UsesArtifactsDirEnvAndWritesMetadataFields"`,
		`"source":"app"`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected app log to contain %s, got %s", want, string(content))
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/logger ./internal/testenv -run "TestMustInit_AppliesConfigFieldsToEverySink|TestNewLoggerRuntime_UsesArtifactsDirEnvAndWritesMetadataFields" -count=1`

Expected: FAIL because `logger.Config` has no `Fields` support yet and `testenv.NewLoggerRuntime` does not exist.

- [ ] **Step 3: Implement runtime-level metadata fields and the minimal test runtime helper**

```go
// D:\codex_work\go-stock\backend\logger\types.go
type Config struct {
	Paths        apppath.Paths
	EnableStdout bool
	Fields       []zap.Field
}
```

```go
// D:\codex_work\go-stock\backend\logger\sinks.go
func (r *Runtime) bootstrapSinks(cfg Config) error {
	if cfg.Paths.LogsDir == "" {
		return fmt.Errorf("logger config Paths.LogsDir is empty")
	}
	if err := os.MkdirAll(cfg.Paths.LogsDir, 0o755); err != nil {
		return fmt.Errorf("ensure logger logs dir %s: %w", cfg.Paths.LogsDir, err)
	}

	sinkPaths := buildSinkPaths(cfg.Paths)
	for _, sink := range []Sink{SinkApp, SinkError, SinkHTTP, SinkAI, SinkTask, SinkDB, SinkFrontend, SinkPanic} {
		path := sinkPaths[sink]
		baseLogger := newSinkLogger(path, cfg.EnableStdout)
		if len(cfg.Fields) > 0 {
			baseLogger = baseLogger.With(cfg.Fields...)
		}
		r.sinks[sink] = baseLogger
	}
	return nil
}
```

```go
// D:\codex_work\go-stock\internal\testenv\runtime.go
package testenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"
	"go.uber.org/zap"
)

const ArtifactsDirEnv = "GO_STOCK_TEST_ARTIFACTS_DIR"

type RuntimeArtifacts struct {
	RootDir   string
	LogsDir   string
	TestRunID string
	TestSuite string
	TestCase  string
}

func NewLoggerRuntime(t *testing.T, suite string) (*logger.Runtime, RuntimeArtifacts) {
	t.Helper()

	artifacts := RuntimeArtifacts{
		TestRunID: strings.NewReplacer(":", "-", ".", "-").Replace(time.Now().UTC().Format(time.RFC3339Nano)),
		TestSuite: sanitizePathFragment(suite),
		TestCase:  sanitizePathFragment(t.Name()),
	}

	rootDir := filepath.Join(t.TempDir(), artifacts.TestSuite, artifacts.TestCase)
	if override := strings.TrimSpace(os.Getenv(ArtifactsDirEnv)); override != "" {
		rootDir = filepath.Join(override, artifacts.TestSuite, artifacts.TestCase)
	}
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		t.Fatalf("create test artifact root %s: %v", rootDir, err)
	}

	artifacts.RootDir = rootDir
	artifacts.LogsDir = filepath.Join(rootDir, "logs")
	runtime := logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: artifacts.RootDir,
			LogsDir: artifacts.LogsDir,
		},
		EnableStdout: false,
		Fields: []zap.Field{
			logger.String("execution_mode", "test"),
			logger.String("test_run_id", artifacts.TestRunID),
			logger.String("test_suite", artifacts.TestSuite),
			logger.String("test_case", artifacts.TestCase),
		},
	})
	return runtime, artifacts
}

func sanitizePathFragment(value string) string {
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", " ", "_")
	sanitized := replacer.Replace(strings.TrimSpace(value))
	if sanitized == "" {
		return "unnamed"
	}
	return sanitized
}
```

- [ ] **Step 4: Run focused logger/testenv verification**

Run: `go test ./backend/logger ./internal/testenv -run "TestMustInit_AppliesConfigFieldsToEverySink|TestNewLoggerRuntime_UsesArtifactsDirEnvAndWritesMetadataFields" -count=1`

Expected: PASS. The sink logs should keep `source`/`trace_id` and add `execution_mode`, `test_run_id`, `test_suite`, and `test_case`.

- [ ] **Step 5: Commit**

```bash
git add backend/logger/types.go backend/logger/sinks.go backend/logger/core_test.go internal/testenv/runtime.go internal/testenv/runtime_test.go
git commit -m "test: add runtime-level test log metadata"
```

### Task 3: Move Existing Chain-Observation Tests Onto The Shared Test Runtime

**Files:**
- Modify: `D:\codex_work\go-stock\app_common_test.go`
- Modify: `D:\codex_work\go-stock\ai-assistant-web\server_test.go`
- Modify: `D:\codex_work\go-stock\backend\logger\closeout_integration_test.go`
- Modify: `D:\codex_work\go-stock\backend\agent\cron_task_api_test.go`

- [ ] **Step 1: Tighten the existing integration tests to require test metadata fields**

```go
// D:\codex_work\go-stock\app_common_test.go
if got, _ := entry["execution_mode"].(string); got != "test" {
	t.Fatalf("expected execution_mode=test, got %#v", got)
}
if got, _ := entry["test_case"].(string); got != "TestLogFrontendRuntimeError_ReusesPayloadTraceID" {
	t.Fatalf("expected test_case metadata, got %#v", got)
}
```

```go
// D:\codex_work\go-stock\ai-assistant-web\server_test.go
if !strings.Contains(string(httpLog), `"execution_mode":"test"`) {
	t.Fatalf("expected health request log to carry test metadata, got %s", string(httpLog))
}
```

```go
// D:\codex_work\go-stock\backend\logger\closeout_integration_test.go
if requireStringField(t, httpEvent, "execution_mode") != "test" {
	t.Fatalf("expected http.request.completed to carry execution_mode=test, got %#v", httpEvent)
}
if requireStringField(t, dbEvent, "test_case") != "TestHTTPTraceFlowsAcrossHTTPAndDBSinks" {
	t.Fatalf("expected db.integration.hit to carry test_case metadata, got %#v", dbEvent)
}
```

```go
// D:\codex_work\go-stock\backend\agent\cron_task_api_test.go
if requireStringField(t, taskEvent, "execution_mode") != "test" {
	t.Fatalf("expected task.execute_started to carry execution_mode=test, got %#v", taskEvent)
}
```

- [ ] **Step 2: Run the targeted tests to verify they fail**

Run: `go test . ./ai-assistant-web ./backend/logger ./backend/agent -run "TestLogFrontendRuntimeError_ReusesPayloadTraceID|TestNewHandler_HealthRouteStillWorksThroughLoggingMiddleware|TestHTTPTraceFlowsAcrossHTTPAndDBSinks|TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks" -count=1`

Expected: FAIL because these tests still create ad-hoc logger runtimes without the shared test metadata fields.

- [ ] **Step 3: Replace ad-hoc temp runtimes with `testenv.NewLoggerRuntime`**

```go
// D:\codex_work\go-stock\app_common_test.go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/internal/testenv"
)

func TestLogFrontendRuntimeError_ReusesPayloadTraceID(t *testing.T) {
	_, artifacts := testenv.NewLoggerRuntime(t, "app-common")

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

	content, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "frontend.log"))
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
	if got, _ := entry["execution_mode"].(string); got != "test" {
		t.Fatalf("expected execution_mode=test, got %#v", got)
	}
}
```

```go
// D:\codex_work\go-stock\ai-assistant-web\server_test.go
runtime, artifacts := testenv.NewLoggerRuntime(t, "ai-assistant-web")
handler, err := newHandler(runtime)
if err != nil {
	t.Fatalf("build server handler: %v", err)
}
req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
recorder := httptest.NewRecorder()
handler.ServeHTTP(recorder, req)
if recorder.Code != http.StatusOK {
	t.Fatalf("expected status 200, got %d", recorder.Code)
}
httpLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "http.log"))
```

```go
// D:\codex_work\go-stock\backend\logger\closeout_integration_test.go
runtime, artifacts := testenv.NewLoggerRuntime(t, "backend-logger")
runtime.AttachPayloadStore(NewPayloadStore(artifacts.LogsDir, 16, 5<<20))
httpLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "http.log"))
dbLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "db.log"))
```

```go
// D:\codex_work\go-stock\backend\agent\cron_task_api_test.go
runtime, artifacts := testenv.NewLoggerRuntime(t, "backend-agent")
db.Init(filepath.Join(artifacts.RootDir, "test.db"))
taskLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "task.log"))
aiLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "ai.log"))
dbLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "db.log"))
```

- [ ] **Step 4: Re-run the chain-observation tests with artifact output enabled**

Run: `$env:GO_STOCK_TEST_ARTIFACTS_DIR='artifacts/testlogs/task-3'; go test . ./ai-assistant-web ./backend/logger ./backend/agent -run "TestLogFrontendRuntimeError_ReusesPayloadTraceID|TestNewHandler_HealthRouteStillWorksThroughLoggingMiddleware|TestHTTPTraceFlowsAcrossHTTPAndDBSinks|TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks" -count=1`

Expected: PASS. All targeted logs should now contain `execution_mode=test`, `test_suite`, and `test_case`, and the logs should remain under `artifacts/testlogs/task-3`.

- [ ] **Step 5: Commit**

```bash
git add app_common_test.go ai-assistant-web/server_test.go backend/logger/closeout_integration_test.go backend/agent/cron_task_api_test.go
git commit -m "test: unify integration log runtimes"
```

### Task 4: Add Central Test Scripts, Artifact Ignore Rules, And The First Release-Smoke Test

**Files:**
- Create: `D:\codex_work\go-stock\app_release_smoke_test.go`
- Create: `D:\codex_work\go-stock\scripts\testing\default.ps1`
- Create: `D:\codex_work\go-stock\scripts\testing\external.ps1`
- Create: `D:\codex_work\go-stock\scripts\testing\release-smoke.ps1`
- Modify: `D:\codex_work\go-stock\.gitignore`

- [ ] **Step 1: Draft the first explicit release-smoke Go test and script entrypoints**

```go
// D:\codex_work\go-stock\app_release_smoke_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/internal/testenv"
)

func TestReleaseSmoke_FrontendBridgeWritesStructuredLog(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	_, artifacts := testenv.NewLoggerRuntime(t, "release-smoke")

	app := NewApp()
	if next := app.CalculateNextRunTime("0 0 0 * * ?"); next.IsZero() {
		t.Fatalf("expected cron expression to produce a non-zero next run time")
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
	for _, want := range []string{`"event":"frontend.error"`, `"trace_id":"trace-release-smoke"`, `"execution_mode":"test"`} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected frontend log to contain %s, got %s", want, string(content))
		}
	}
}
```

```powershell
# D:\codex_work\go-stock\scripts\testing\default.ps1
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/default'
  go test ./... -count=1
  node --test frontend/src/utils/frontendLogger.test.mjs frontend/src/utils/stockCode.test.mjs frontend/src/utils/aiConfig.test.mjs frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
} finally {
  Pop-Location
}
```

```powershell
# D:\codex_work\go-stock\scripts\testing\external.ps1
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/external'
  $env:GO_STOCK_RUN_EXTERNAL_TESTS = '1'
  go test ./... -count=1
  node --test frontend/src/utils/frontendLogger.test.mjs frontend/src/utils/stockCode.test.mjs frontend/src/utils/aiConfig.test.mjs frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
} finally {
  Pop-Location
}
```

```powershell
# D:\codex_work\go-stock\scripts\testing\release-smoke.ps1
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/release'
  $env:GO_STOCK_RUN_RELEASE_SMOKE = '1'
  go test . -run 'TestReleaseSmoke_' -count=1
  go test ./ai-assistant-web -run 'TestNewHandler_HealthRouteStillWorksThroughLoggingMiddleware' -count=1
  node --test frontend/src/utils/frontendLogger.test.mjs
  npm --prefix frontend run build
} finally {
  Pop-Location
}
```

```gitignore
# D:\codex_work\go-stock\.gitignore
artifacts/testlogs/
```

- [ ] **Step 2: Run the drafted release script once to catch path or import mistakes early**

Run: `pwsh -File .\scripts\testing\release-smoke.ps1`

Expected: on the first attempt, either PASS or a concrete failure pointing to a bad path/import/env wiring issue. Do not move on until the command runs cleanly.

- [ ] **Step 3: Save the final release smoke test, scripts, and artifact ignore rule**

```powershell
# Keep the script bodies exactly as shown in Step 1.
```

```go
// Keep TestReleaseSmoke_FrontendBridgeWritesStructuredLog exactly as shown in Step 1.
```

- [ ] **Step 4: Run the repository entrypoints**

Run: `pwsh -File .\scripts\testing\default.ps1`

Expected: PASS. Stable Go tests and the four frontend `node --test` files should pass, with logs written under `artifacts/testlogs/default`.

Run: `pwsh -File .\scripts\testing\release-smoke.ps1`

Expected: PASS. The gated release smoke test, `ai-assistant-web` smoke check, frontend logger test, and Vite build should all succeed.

- [ ] **Step 5: Commit**

```bash
git add app_release_smoke_test.go scripts/testing/default.ps1 scripts/testing/external.ps1 scripts/testing/release-smoke.ps1 .gitignore
git commit -m "test: add repository test entry scripts"
```

### Task 5: Split CI Into Default, External, And Release-Smoke Workflows

**Files:**
- Create: `D:\codex_work\go-stock\.github\workflows\test-default.yml`
- Create: `D:\codex_work\go-stock\.github\workflows\test-external.yml`
- Create: `D:\codex_work\go-stock\.github\workflows\release-smoke.yml`

- [ ] **Step 1: Draft the workflow files**

```yaml
# D:\codex_work\go-stock\.github\workflows\test-default.yml
name: Test Default

on:
  pull_request:
  push:
    tags-ignore:
      - '*-release'
      - '*-dev'
      - '*-pre'

jobs:
  test-default:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install frontend dependencies
        run: npm --prefix frontend ci
      - name: Run default tests
        shell: pwsh
        run: .\scripts\testing\default.ps1
      - name: Upload test logs
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: test-default-logs
          path: artifacts/testlogs
          if-no-files-found: ignore
```

```yaml
# D:\codex_work\go-stock\.github\workflows\test-external.yml
name: Test External

on:
  workflow_dispatch:
  schedule:
    - cron: '0 2 * * *'

jobs:
  test-external:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install frontend dependencies
        run: npm --prefix frontend ci
      - name: Run external tests
        shell: pwsh
        run: .\scripts\testing\external.ps1
      - name: Upload external test logs
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: test-external-logs
          path: artifacts/testlogs
          if-no-files-found: ignore
```

```yaml
# D:\codex_work\go-stock\.github\workflows\release-smoke.yml
name: Release Smoke

on:
  workflow_dispatch:

jobs:
  release-smoke:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install frontend dependencies
        run: npm --prefix frontend ci
      - name: Run release smoke checks
        shell: pwsh
        run: .\scripts\testing\release-smoke.ps1
      - name: Upload release smoke logs
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: release-smoke-logs
          path: artifacts/testlogs
          if-no-files-found: ignore
```

- [ ] **Step 2: Verify the workflow files point at the intended script entrypoints**

Run: `rg -n "scripts/testing/default.ps1|scripts/testing/external.ps1|scripts/testing/release-smoke.ps1" .github/workflows/test-default.yml .github/workflows/test-external.yml .github/workflows/release-smoke.yml`

Expected: each workflow should reference exactly one of the three repository test scripts shown above.

- [ ] **Step 3: Save the final workflow files exactly as above**

```yaml
# Keep the workflow contents exactly as shown in Step 1.
```

- [ ] **Step 4: Verify the local scripts still drive the same behavior as CI**

Run: `pwsh -File .\scripts\testing\default.ps1`

Expected: PASS. This is the same command the default workflow will run.

Run: `pwsh -File .\scripts\testing\release-smoke.ps1`

Expected: PASS. This is the same command the release-smoke workflow will run.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/test-default.yml .github/workflows/test-external.yml .github/workflows/release-smoke.yml
git commit -m "ci: split default external and release smoke workflows"
```

### Task 6: Inventory Current Test Families And Clean Root/Agent/Model/Utility Tests

**Files:**
- Create: `D:\codex_work\go-stock\docs\superpowers\checklists\2026-04-05-test-migration-inventory.md`
- Modify: `D:\codex_work\go-stock\app_test.go`
- Modify: `D:\codex_work\go-stock\backend\agent\agent_test.go`
- Modify: `D:\codex_work\go-stock\backend\models\models_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\utils_test.go`

- [ ] **Step 1: Write the migration inventory document**

```markdown
# Test Migration Inventory

## Default (retain in main gate)
- `app_common_test.go`
- `backend/logger/core_test.go`
- `backend/logger/closeout_integration_test.go`
- `backend/logger/http_test.go`
- `backend/logger/db_test.go`
- `backend/agent/cron_task_api_test.go`
- `backend/agent/trace_reuse_test.go`
- `ai-assistant-web/server_test.go`
- `backend/db/db_test.go`
- `backend/apppath/apppath_test.go`

## External (real network / real provider / real crawl)
- `app_test.go`
- `backend/agent/agent_test.go`
- `backend/models/models_test.go`
- `backend/data/openai_api_test.go`
- `backend/data/market_news_api_test.go`
- `backend/data/search_stock_api_test.go`
- `backend/data/eastmoney_kline_api_test.go`
- `backend/data/crawler_api_test.go`
- `backend/data/stock_data_api_test.go`
- `backend/data/stock_sentiment_analysis_test.go`

## Release Smoke
- `app_release_smoke_test.go`
- `backend/data/alert_windows_api_test.go`
- `backend/data/alert_darwin_api_test.go`
- `app_test.go: TestReleaseSmoke_GetScreenResolution`

## Manual / follow-up candidates
- Any remaining test that still prints output without stable assertions after the above sweep.
```

- [ ] **Step 2: Prove the current legacy logger usage exists in these test families**

Run: `rg -l "logger\.SugaredLogger|log\.SugaredLogger" -g "*_test.go" app_test.go backend/agent/agent_test.go backend/models/models_test.go backend/data/utils_test.go`

Expected: the four files appear in the output before migration.

- [ ] **Step 3: Replace legacy logger prints with assertions and `t.Logf`**

```go
// D:\codex_work\go-stock\app_test.go
func TestIsUSTradingTime(t *testing.T) {
	date := time.Now()
	hour, minute, _ := date.Clock()
	t.Logf("当前时间: %02d:%02d", hour, minute)
	t.Logf("美股交易时段=%v", IsUSTradingTime(time.Now()))
}

func TestUpdateCheck(t *testing.T) {
	requireIntegrationTest(t)
	releaseVersion := &models.GitHubReleaseVersion{}
	_, err := resty.New().R().
		SetResult(releaseVersion).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		Get("https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest")
	if err != nil {
		t.Fatalf("get github release version: %v", err)
	}
	if strings.TrimSpace(releaseVersion.TagName) == "" {
		t.Fatalf("expected latest release tag name, got %#v", releaseVersion)
	}
	t.Logf("releaseVersion=%+v", releaseVersion)
}

func TestReleaseSmoke_GetScreenResolution(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	x, y, w, h, err := getScreenResolution()
	if err != nil {
		t.Fatalf("get screen resolution: %v", err)
	}
	t.Logf("screen resolution x=%d y=%d w=%d h=%d", x, y, w, h)
}
```

```go
// D:\codex_work\go-stock\backend\agent\agent_test.go
if err != nil {
	t.Fatalf("stream error: %v", err)
}
defer sr.Close()
md := strings.Builder{}
for {
	msg, err := sr.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			break
		}
		t.Fatalf("recv stream chunk: %v", err)
	}
	t.Logf("stream recv: %#v", msg)
	if msg.ReasoningContent != "" {
		md.WriteString(msg.ReasoningContent)
	}
	if msg.Content != "" {
		md.WriteString(msg.Content)
	}
}
if md.Len() == 0 {
	t.Fatalf("expected streamed agent output to be non-empty")
}
outputPath := filepath.Join(t.TempDir(), "result.md")
fileutil.WriteStringToFile(outputPath, md.String(), false)
```

```go
// D:\codex_work\go-stock\backend\models\models_test.go
if err := json.Unmarshal(bs, v); err != nil {
	t.Fatalf("unmarshal hk fixture: %v", err)
}
for i, data := range *v.StockInfos {
	t.Logf("第%d条数据: %+v", i, data)
}
if err := db.Dao.Create(&hks).Error; err != nil {
	t.Fatalf("create hk rows: %v", err)
}
```

```go
// D:\codex_work\go-stock\backend\data\utils_test.go
func TestRemoveNonPrintable(t *testing.T) {
	if got := RemoveAllBlankChar("新 希 望"); got != "新希望" {
		t.Fatalf("RemoveAllBlankChar returned %q, want %q", got, "新希望")
	}
}

func TestConvertStockCodeToTushareCode(t *testing.T) {
	if got := ConvertStockCodeToTushareCode("sz000802"); got != "000802.SZ" {
		t.Fatalf("ConvertStockCodeToTushareCode returned %q", got)
	}
	if got := ConvertTushareCodeToStockCode("000802.SZ"); got != "sz000802" {
		t.Fatalf("ConvertTushareCodeToStockCode returned %q", got)
	}
}
```

- [ ] **Step 4: Run focused verification for the cleaned test families**

Run: `$env:GO_STOCK_RUN_EXTERNAL_TESTS='1'; go test . ./backend/agent ./backend/models ./backend/data -run "Test(IsUSTradingTime|UpdateCheck|ReleaseSmoke_GetScreenResolution|GetStockAiAgent|Agent|StockInfoHK|RemoveNonPrintable|ConvertStockCodeToTushareCode|ReplaceSensitiveWords)" -count=1`

Expected: PASS on the current platform, except `TestReleaseSmoke_GetScreenResolution`, which should remain skipped unless `GO_STOCK_RUN_RELEASE_SMOKE=1` is also set.

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/checklists/2026-04-05-test-migration-inventory.md app_test.go backend/agent/agent_test.go backend/models/models_test.go backend/data/utils_test.go
git commit -m "test: inventory and clean non-data legacy test logging"
```

### Task 7: Migrate External Data Tests Part 1 And Add Small Shared Helpers

**Files:**
- Create: `D:\codex_work\go-stock\backend\data\external_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\openai_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\market_news_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\search_stock_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api_test.go`

- [ ] **Step 1: Confirm the remaining legacy logger usage in these files**

Run: `rg -n "logger\.SugaredLogger|log\.SugaredLogger" backend/data/openai_api_test.go backend/data/market_news_api_test.go backend/data/search_stock_api_test.go backend/data/eastmoney_kline_api_test.go`

Expected: multiple matches across the four files.

- [ ] **Step 2: Add a tiny shared helper and rewrite the first batch of data external tests**

```go
// D:\codex_work\go-stock\backend\data\external_test.go
package data

import (
	"encoding/json"
	"strings"
	"testing"
)

func logValue(t *testing.T, label string, value any) {
	t.Helper()
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", label, err)
	}
	t.Logf("%s:\n%s", label, string(payload))
}

func requirePositiveLen(t *testing.T, label string, size int) {
	t.Helper()
	if size <= 0 {
		t.Fatalf("expected %s to contain data", label)
	}
}

func requireNotBlank(t *testing.T, label, value string) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Fatalf("expected %s to be non-blank", label)
	}
}
```

```go
// D:\codex_work\go-stock\backend\data\openai_api_test.go
func TestSearchGuShiTongStockInfo(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	msgs := SearchGuShiTongStockInfo("sh600745", 60)
	requirePositiveLen(t, "SearchGuShiTongStockInfo messages", len(*msgs))
	for _, msg := range *msgs {
		t.Logf("message=%s", msg)
	}
}
```

```go
// D:\codex_work\go-stock\backend\data\market_news_api_test.go
func TestGetSinaNews(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	news := NewMarketNewsApi().GetSinaNews(30)
	requirePositiveLen(t, "Sina news", len(*news))
	logValue(t, "first_sina_news", (*news)[0])
}

func TestStockNotice(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().StockNotice("600584,600900")
	requirePositiveLen(t, "stock notice rows", len(resp))
	type row struct {
		Title      string `md:"公告标题"`
		NoticeDate string `md:"公告日期"`
		ColumnName string `md:"公告类型"`
	}
	var rows []row
	for _, a := range resp {
		m, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if columns, ok := m["columns"].([]any); ok && len(columns) > 0 {
			column := columns[0].(map[string]any)
			rows = append(rows, row{
				Title:      convertor.ToString(m["title"]),
				NoticeDate: convertor.ToString(m["notice_date"]),
				ColumnName: convertor.ToString(column["column_name"]),
			})
			continue
		}
		rows = append(rows, row{
			Title:      convertor.ToString(m["title"]),
			NoticeDate: convertor.ToString(m["notice_date"]),
		})
	}
	md := util.MarkdownTableWithTitle("上市公司公告", rows)
	requireNotBlank(t, "stock notice markdown", md)
	t.Logf("stock notice markdown:\n%s", md)
}
```

```go
// D:\codex_work\go-stock\backend\data\search_stock_api_test.go
func TestGetStockFinancialInfo(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	res := NewStockDataApi().GetStockFinancialInfo("600519.SH")
	md := util.MarkdownTableWithTitle("600519.SH股票财报信息", res.Result.Data)
	requireNotBlank(t, "stock financial markdown", md)
	t.Logf("financial markdown:\n%s", md)
}
```

```go
// D:\codex_work\go-stock\backend\data\eastmoney_kline_api_test.go
func TestEastMoneyKLineApi_GetDayKLine(t *testing.T) {
	requireIntegrationTest(t)
	kLines := NewEastMoneyKLineApi().GetDayKLine("600519.SH", 30)
	requirePositiveLen(t, "day k lines", len(*kLines))
	t.Logf("获取到 %d 条日 K 线数据", len(*kLines))
}
```

- [ ] **Step 3: Run the first external data test batch**

Run: `$env:GO_STOCK_RUN_EXTERNAL_TESTS='1'; go test ./backend/data -run "Test(SearchGuShiTongStockInfo|GetSinaNews|StockNotice|GetStockFinancialInfo|EastMoneyKLineApi_GetDayKLine)" -count=1`

Expected: PASS. Each targeted test should now fail on empty responses instead of only printing them.

- [ ] **Step 4: Confirm the first batch is clean of `SugaredLogger`**

Run: `rg -n "logger\.SugaredLogger|log\.SugaredLogger" backend/data/openai_api_test.go backend/data/market_news_api_test.go backend/data/search_stock_api_test.go backend/data/eastmoney_kline_api_test.go`

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add backend/data/external_test.go backend/data/openai_api_test.go backend/data/market_news_api_test.go backend/data/search_stock_api_test.go backend/data/eastmoney_kline_api_test.go
git commit -m "test: migrate external data tests batch one"
```

### Task 8: Migrate External Data Tests Part 2 And Move OS Notification Probes To Release Smoke

**Files:**
- Modify: `D:\codex_work\go-stock\backend\data\crawler_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_sentiment_analysis_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\alert_windows_api_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\alert_darwin_api_test.go`

- [ ] **Step 1: Confirm the remaining legacy logger usage in the second data batch**

Run: `rg -n "logger\.SugaredLogger|log\.SugaredLogger" backend/data/crawler_api_test.go backend/data/stock_data_api_test.go backend/data/stock_sentiment_analysis_test.go backend/data/alert_windows_api_test.go backend/data/alert_darwin_api_test.go`

Expected: multiple matches across crawler, stock data, sentiment, and alert tests.

- [ ] **Step 2: Rewrite the remaining tests to use helper assertions and release gates where appropriate**

```go
// D:\codex_work\go-stock\backend\data\crawler_api_test.go
if err != nil {
	t.Fatalf("crawler request failed: %v", err)
}
requirePositiveLen(t, "crawler messages", len(messages))
t.Logf("messages=%d", len(messages))
```

```go
// D:\codex_work\go-stock\backend\data\stock_data_api_test.go
func TestGetTelegraph(t *testing.T) {
	requireIntegrationTest(t)
	telegraph := NewStockDataApi().GetTelegraph()
	requirePositiveLen(t, "telegraph", len(*telegraph))
	t.Logf("telegraph count=%d", len(*telegraph))
}

func TestGetFinancialReports(t *testing.T) {
	requireIntegrationTest(t)
	message := NewStockDataApi().GetFinancialReports("002594.SZ")
	requireNotBlank(t, "financial reports markdown", message)
	t.Logf("financial reports markdown:\n%s", message)
}
```

```go
// D:\codex_work\go-stock\backend\data\stock_sentiment_analysis_test.go
func TestAnalyzeSentiment(t *testing.T) {
	requireIntegrationTest(t)
	result := AnalyzeSentiment("新能源车销量继续增长，行业情绪显著改善")
	if result.Score == 0 && len(result.WordFreq) == 0 {
		t.Fatalf("expected sentiment analysis to return score or token stats, got %#v", result)
	}
	t.Logf("情感分析结果: score=%.2f freq=%v", result.Score, result.WordFreq)
}
```

```go
// D:\codex_work\go-stock\backend\data\alert_windows_api_test.go
func TestReleaseSmoke_Alert(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	if err := Alert("go-stock smoke", "release smoke alert"); err != nil {
		t.Fatalf("send windows alert: %v", err)
	}
}
```

```go
// D:\codex_work\go-stock\backend\data\alert_darwin_api_test.go
func TestReleaseSmoke_Alert(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	if err := Alert("go-stock smoke", "release smoke alert"); err != nil {
		t.Fatalf("send darwin alert: %v", err)
	}
}
```

- [ ] **Step 3: Run the second external/release batch**

Run: `$env:GO_STOCK_RUN_EXTERNAL_TESTS='1'; go test ./backend/data -run "Test(GetTelegraph|GetFinancialReports|AnalyzeSentiment|NewGuShiTongCrawler|GetHtml)" -count=1`

Expected: PASS. The targeted crawler/stock/sentiment tests should now assert non-empty results instead of only printing them.

Run: `$env:GO_STOCK_RUN_RELEASE_SMOKE='1'; go test ./backend/data -run "TestReleaseSmoke_Alert" -count=1`

Expected: PASS on the current platform-specific alert file, or SKIP on unsupported platforms.

- [ ] **Step 4: Confirm the second batch is clean of `SugaredLogger`**

Run: `rg -n "logger\.SugaredLogger|log\.SugaredLogger" backend/data/crawler_api_test.go backend/data/stock_data_api_test.go backend/data/stock_sentiment_analysis_test.go backend/data/alert_windows_api_test.go backend/data/alert_darwin_api_test.go`

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add backend/data/crawler_api_test.go backend/data/stock_data_api_test.go backend/data/stock_sentiment_analysis_test.go backend/data/alert_windows_api_test.go backend/data/alert_darwin_api_test.go
git commit -m "test: migrate external data tests batch two"
```

### Task 9: Add A Test-Only Legacy Logger Guard And Run Full Verification

**Files:**
- Create: `D:\codex_work\go-stock\backend\logger\legacy_test_usage_guard_test.go`

- [ ] **Step 1: Write the new guard test**

```go
// D:\codex_work\go-stock\backend\logger\legacy_test_usage_guard_test.go
package logger

import (
	"io/fs"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var allowedLegacyTestFiles = map[string]struct{}{
	"backend/logger/core_test.go":         {},
	"backend/logger/legacy_usage_test.go": {},
}

func TestNoLegacySugaredLoggerUsageInGoTestsOutsideCompatibilityCoverage(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	fset := token.NewFileSet()
	testFiles, err := collectLegacyTestTargetFiles(repoRoot)
	if err != nil {
		t.Fatalf("collect legacy test target files: %v", err)
	}

	var violations []string
	for _, relativePath := range testFiles {
		fileNode, err := parser.ParseFile(fset, filepath.Join(repoRoot, filepath.FromSlash(relativePath)), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", relativePath, err)
		}
		violations = append(violations, collectLegacyUsageViolations(fset, fileNode, relativePath)...)
	}

	if len(violations) > 0 {
		t.Fatalf("legacy SugaredLogger usage is forbidden in Go tests outside compatibility coverage:\n%s", strings.Join(violations, "\n"))
	}
}

func collectLegacyTestTargetFiles(repoRoot string) ([]string, error) {
	var testFiles []string
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relativePath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		normalized := filepath.ToSlash(relativePath)
		if shouldSkipGuardDir(normalized) {
			return nil
		}
		if _, allowed := allowedLegacyTestFiles[normalized]; allowed {
			return nil
		}
		testFiles = append(testFiles, normalized)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(testFiles)
	return testFiles, nil
}
```

- [ ] **Step 2: Verify the guard fails before the migration is complete**

Run: `go test ./backend/logger -run "TestNoLegacySugaredLoggerUsageInGoTestsOutsideCompatibilityCoverage" -count=1`

Expected: if any migrated file still contains `logger.SugaredLogger`, this test should FAIL and list exact file/line violations.

- [ ] **Step 3: Fix any remaining matches, then rerun the guard and the three repository entry scripts**

```bash
rg -l "logger\.SugaredLogger|log\.SugaredLogger" -g "*_test.go" .
```

Expected after cleanup: only `backend/logger/core_test.go` and `backend/logger/legacy_usage_test.go` remain.

Run: `go test ./backend/logger -run "TestNoLegacySugaredLoggerUsageInGoTestsOutsideCompatibilityCoverage" -count=1`

Expected: PASS.

Run: `pwsh -File .\scripts\testing\default.ps1`

Expected: PASS.

Run: `pwsh -File .\scripts\testing\external.ps1`

Expected: PASS when network and external dependencies are reachable.

Run: `pwsh -File .\scripts\testing\release-smoke.ps1`

Expected: PASS on the current platform.

- [ ] **Step 4: Commit the guard and final verification changes**

```bash
git add backend/logger/legacy_test_usage_guard_test.go
git commit -m "test: guard against legacy logger usage in tests"
```

- [ ] **Step 5: Final repo check**

Run: `git diff --stat`

Expected: diff is limited to `internal/testenv`, `backend/logger`, test files, `scripts/testing`, workflow YAMLs, `.gitignore`, and the migration inventory doc.

## Self-Review

### Spec Coverage

- Test layering and three entry modes are covered by Tasks 1, 4, and 5.
- Shared test logging and runtime-level metadata fields are covered by Tasks 2 and 3.
- Script centralization under `scripts/testing/` is covered by Task 4.
- CI split into `default` / `external` / `release-smoke` is covered by Task 5.
- Inventory-driven migration of historical test families is covered by Task 6.
- External data test cleanup and stronger assertions are covered by Tasks 7 and 8.
- Ban on ordinary test usage of legacy `SugaredLogger` is covered by Task 9.

### Placeholder Scan

- No `TODO`, `TBD`, “implement later”, or “similar to Task N” placeholders remain.
- Every task lists exact file paths, explicit commands, and concrete code snippets.

### Type And Naming Consistency

- Environment variables are consistent across tasks:
  - `GO_STOCK_RUN_EXTERNAL_TESTS`
  - `GO_STOCK_RUN_INTEGRATION_TESTS` as legacy alias
  - `GO_STOCK_RUN_RELEASE_SMOKE`
  - `GO_STOCK_TEST_ARTIFACTS_DIR`
- Shared helper names are consistent across tasks:
  - `RequireExternalTest`
  - `RequireReleaseSmokeTest`
  - `NewLoggerRuntime`
