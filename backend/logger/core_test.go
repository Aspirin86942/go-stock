package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestBuildSinkPaths_UsesRuntimeLogsDir(t *testing.T) {
	root := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)
	paths := apppath.Paths{RootDir: root, LogsDir: filepath.Join(root, "logs")}

	sinkPaths := buildSinkPaths(paths)

	if sinkPaths[SinkApp] != filepath.Join(root, "logs", "app.log") {
		t.Fatalf("expected app sink under runtime logs dir, got %q", sinkPaths[SinkApp])
	}
	if sinkPaths[SinkPanic] != filepath.Join(root, "logs", "panic.log") {
		t.Fatalf("expected panic sink under runtime logs dir, got %q", sinkPaths[SinkPanic])
	}
}

func TestModuleLogger_AddsModuleEventAndTraceFields(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	log := runtime.ForSink(SinkApp, "main").WithTrace(TraceContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		AppSessionID: "session-1",
		Source:       "wails",
	})

	log.Info("startup.begin", "starting app", String("version", "dev"))

	line := buf.String()
	for _, want := range []string{`"module":"main"`, `"event":"startup.begin"`, `"trace_id":"trace-1"`, `"app_session_id":"session-1"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected log line to contain %s, got %s", want, line)
		}
	}
}

func TestLegacySugaredLogger_ErrorGoesToErrorSink(t *testing.T) {
	appBuf := &bytes.Buffer{}
	errorBuf := &bytes.Buffer{}
	runtime := newRuntimeForTestWithSinks(appBuf, errorBuf)

	initLegacyGlobals(runtime)
	SugaredLogger.Errorf("legacy error: %s", "boom")

	if strings.Contains(appBuf.String(), "legacy error") {
		t.Fatalf("expected legacy error log not to go to app sink, got %s", appBuf.String())
	}
	if !strings.Contains(errorBuf.String(), "legacy error") {
		t.Fatalf("expected legacy error log to go to error sink, got %s", errorBuf.String())
	}
}

func TestForSinkError_RoutesToErrorSink(t *testing.T) {
	appBuf := &bytes.Buffer{}
	errorBuf := &bytes.Buffer{}
	runtime := newRuntimeForTestWithSinks(appBuf, errorBuf)

	runtime.ForSink(SinkError, "legacy").Error("legacy.err", "sink-error-message")

	if strings.Contains(appBuf.String(), "sink-error-message") {
		t.Fatalf("expected sink error message not to go to app sink, got %s", appBuf.String())
	}
	if !strings.Contains(errorBuf.String(), "sink-error-message") {
		t.Fatalf("expected sink error message to go to error sink, got %s", errorBuf.String())
	}
}

func TestNormalizeFrontendError_PreservesRouteAndStack(t *testing.T) {
	payload := NormalizeFrontendError([]interface{}{
		map[string]interface{}{
			"page":    "stock.vue",
			"route":   "/stock",
			"message": "ResizeObserver loop limit exceeded",
			"error":   "stack-line-1",
		},
	})

	if payload.Page != "stock.vue" || payload.Route != "/stock" {
		t.Fatalf("expected normalized frontend payload, got %#v", payload)
	}
	if payload.Stack != "stack-line-1" {
		t.Fatalf("expected stack to be preserved, got %#v", payload)
	}
}

func TestNormalizeFrontendError_PreservesUnknownTopLevelFieldsInExtra(t *testing.T) {
	payload := NormalizeFrontendError([]interface{}{
		map[string]interface{}{
			"page":      "stock.vue",
			"route":     "/stock",
			"message":   "ResizeObserver loop limit exceeded",
			"error":     "stack-line-1",
			"component": "StockPanel",
			"severity":  "warning",
			"extra": map[string]interface{}{
				"existing": "value",
			},
		},
	})

	if payload.Extra == nil {
		t.Fatalf("expected extra fields to be preserved, got %#v", payload)
	}
	if payload.Extra["existing"] != "value" {
		t.Fatalf("expected existing extra fields to be preserved, got %#v", payload.Extra)
	}
	if payload.Extra["component"] != "StockPanel" {
		t.Fatalf("expected unknown top-level field to be merged into extra, got %#v", payload.Extra)
	}
	if payload.Extra["severity"] != "warning" {
		t.Fatalf("expected unknown top-level field to be merged into extra, got %#v", payload.Extra)
	}
}

func TestMustInit_RebindsLegacyGlobalsToLatestRuntime(t *testing.T) {
	previousRuntime := Default()
	previousCoreLogger := CoreLogger
	previousSugaredLogger := SugaredLogger
	t.Cleanup(func() {
		defaultRuntime.Store(previousRuntime)
		CoreLogger = previousCoreLogger
		SugaredLogger = previousSugaredLogger
	})

	legacyAppBuf := &bytes.Buffer{}
	legacyErrorBuf := &bytes.Buffer{}
	initLegacyGlobals(newRuntimeForTestWithSinks(legacyAppBuf, legacyErrorBuf))

	rootDir, err := os.MkdirTemp("", "go-stock-logger-rebind-*")
	if err != nil {
		t.Fatalf("create temp dir for latest runtime: %v", err)
	}
	logsDir := filepath.Join(rootDir, "logs")
	latestRuntime := MustInit(Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: logsDir,
		},
		EnableStdout: false,
	})

	if Default() != latestRuntime {
		t.Fatalf("expected MustInit to publish latest runtime")
	}

	CoreLogger.Info("core-to-latest-runtime")
	SugaredLogger.Infof("sugar-to-latest-runtime")

	appLog, err := os.ReadFile(filepath.Join(logsDir, "app.log"))
	if err != nil {
		t.Fatalf("read app log from latest runtime: %v", err)
	}
	appLogContent := string(appLog)
	if !strings.Contains(appLogContent, "core-to-latest-runtime") {
		t.Fatalf("expected CoreLogger to write to latest runtime sink, got %s", appLogContent)
	}
	if !strings.Contains(appLogContent, "sugar-to-latest-runtime") {
		t.Fatalf("expected SugaredLogger to write to latest runtime sink, got %s", appLogContent)
	}
	if strings.Contains(legacyAppBuf.String(), "core-to-latest-runtime") || strings.Contains(legacyAppBuf.String(), "sugar-to-latest-runtime") {
		t.Fatalf("expected legacy globals to stop writing to previous runtime, got %s", legacyAppBuf.String())
	}
}

func TestMustInit_AppliesConfigFieldsToEverySink_PreservesSourceTraceFields(t *testing.T) {
	previousRuntime := Default()
	previousCoreLogger := CoreLogger
	previousSugaredLogger := SugaredLogger
	t.Cleanup(func() {
		defaultRuntime.Store(previousRuntime)
		CoreLogger = previousCoreLogger
		SugaredLogger = previousSugaredLogger
	})

	rootDir := t.TempDir()
	paths := apppath.Paths{
		RootDir: rootDir,
		LogsDir: filepath.Join(rootDir, "logs"),
	}

	runtime := MustInit(Config{
		Paths:        paths,
		EnableStdout: false,
		Fields: []zap.Field{
			String("execution_mode", "test"),
			String("test_suite", "logger-core"),
			String("source", "config-source"),
		},
	})
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Errorf("close runtime: %v", err)
		}
	})

	trace := TraceContext{
		TraceID:      "trace-config-fields",
		SpanID:       "span-config-fields",
		AppSessionID: "session-config-fields",
		Source:       "http",
	}
	sinks := []Sink{SinkApp, SinkError, SinkHTTP, SinkAI, SinkTask, SinkDB, SinkFrontend, SinkPanic}
	for _, sink := range sinks {
		runtime.ForSink(sink, "logger.core").
			WithTrace(trace).
			Info("runtime.field.check", "runtime config fields", String("sink", string(sink)))
	}

	sinkPaths := buildSinkPaths(paths)
	for _, sink := range sinks {
		logFile := sinkPaths[sink]
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("read sink log %s: %v", logFile, err)
		}
		entry := parseLastJSONLogEntry(t, content)
		if got, _ := entry["execution_mode"].(string); got != "test" {
			t.Fatalf("expected execution_mode=test in %s, got %#v", logFile, got)
		}
		if got, _ := entry["test_suite"].(string); got != "logger-core" {
			t.Fatalf("expected test_suite=logger-core in %s, got %#v", logFile, got)
		}
		if got, _ := entry["module"].(string); got != "logger.core" {
			t.Fatalf("expected module=logger.core in %s, got %#v", logFile, got)
		}
		if got, _ := entry["event"].(string); got != "runtime.field.check" {
			t.Fatalf("expected event=runtime.field.check in %s, got %#v", logFile, got)
		}
		if got, _ := entry["trace_id"].(string); got != "trace-config-fields" {
			t.Fatalf("expected trace_id=trace-config-fields in %s, got %#v", logFile, got)
		}
		if got, _ := entry["span_id"].(string); got != "span-config-fields" {
			t.Fatalf("expected span_id=span-config-fields in %s, got %#v", logFile, got)
		}
		if got, _ := entry["app_session_id"].(string); got != "session-config-fields" {
			t.Fatalf("expected app_session_id=session-config-fields in %s, got %#v", logFile, got)
		}
		if got, _ := entry["source"].(string); got != "http" {
			t.Fatalf("expected source from trace context in %s, got %#v", logFile, got)
		}
	}
}

func newRuntimeForTest(writer io.Writer) *Runtime {
	testCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		zapcore.DebugLevel,
	)
	testLogger := zap.New(testCore)

	runtime := &Runtime{sinks: make(map[Sink]*zap.Logger)}
	for _, sink := range []Sink{SinkApp, SinkPanic, SinkHTTP, SinkDB, SinkFrontend, SinkError, SinkAI, SinkTask} {
		runtime.sinks[sink] = testLogger
	}

	return runtime
}

func newRuntimeForTestWithSinks(appWriter io.Writer, errorWriter io.Writer) *Runtime {
	appCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(appWriter),
		zapcore.DebugLevel,
	)
	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(errorWriter),
		zapcore.DebugLevel,
	)
	appLogger := zap.New(appCore)
	errorLogger := zap.New(errorCore)

	runtime := &Runtime{sinks: make(map[Sink]*zap.Logger)}
	runtime.sinks[SinkApp] = appLogger
	runtime.sinks[SinkError] = errorLogger
	for _, sink := range []Sink{SinkPanic, SinkHTTP, SinkDB, SinkFrontend, SinkAI, SinkTask} {
		runtime.sinks[sink] = appLogger
	}
	return runtime
}

func parseLastJSONLogEntry(t *testing.T, content []byte) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) == "" {
		t.Fatalf("expected at least one JSON log line, got %q", string(content))
	}

	raw := lines[len(lines)-1]
	var entry map[string]any
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		t.Fatalf("decode log entry %q: %v", raw, err)
	}
	return entry
}
