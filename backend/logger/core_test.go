package logger

import (
	"bytes"
	"io"
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
