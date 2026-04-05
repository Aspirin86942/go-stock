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

func newRuntimeForTest(writer io.Writer) *Runtime {
	cfg := DefaultConfig(apppath.Paths{LogsDir: filepath.Join(os.TempDir(), "go-stock-test-logs")})
	runtime := MustInit(cfg)

	testCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		zapcore.DebugLevel,
	)
	testLogger := zap.New(testCore)

	for _, sink := range []Sink{SinkApp, SinkPanic, SinkHTTP, SinkDB, SinkFrontend, SinkError, SinkAI, SinkTask} {
		runtime.sinks[sink] = testLogger
	}

	return runtime
}
