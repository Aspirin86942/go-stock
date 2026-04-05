package logger

import (
	"bytes"
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
