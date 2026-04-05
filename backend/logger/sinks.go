package logger

import (
	"errors"
	"fmt"
	"go-stock/backend/apppath"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultPayloadInlineLimit = 4 << 10
	defaultPayloadMaxTotal    = int64(5 << 20)
)

func buildSinkPaths(paths apppath.Paths) map[Sink]string {
	return map[Sink]string{
		SinkApp:      filepath.Join(paths.LogsDir, "app.log"),
		SinkError:    filepath.Join(paths.LogsDir, "error.log"),
		SinkHTTP:     filepath.Join(paths.LogsDir, "http.log"),
		SinkAI:       filepath.Join(paths.LogsDir, "ai.log"),
		SinkTask:     filepath.Join(paths.LogsDir, "task.log"),
		SinkDB:       filepath.Join(paths.LogsDir, "db.log"),
		SinkFrontend: filepath.Join(paths.LogsDir, "frontend.log"),
		SinkPanic:    filepath.Join(paths.LogsDir, "panic.log"),
	}
}

func newDefaultPayloadStore(paths apppath.Paths) *PayloadStore {
	return NewPayloadStore(paths.LogsDir, defaultPayloadInlineLimit, defaultPayloadMaxTotal)
}

func (r *Runtime) bootstrapSinks(cfg Config) error {
	if cfg.Paths.LogsDir == "" {
		return fmt.Errorf("logger config Paths.LogsDir is empty")
	}
	if err := os.MkdirAll(cfg.Paths.LogsDir, 0o755); err != nil {
		return fmt.Errorf("ensure logger logs dir %s: %w", cfg.Paths.LogsDir, err)
	}

	commonFields := filterRuntimeConfigFields(cfg.Fields)
	sinkPaths := buildSinkPaths(cfg.Paths)
	for _, sink := range []Sink{SinkApp, SinkError, SinkHTTP, SinkAI, SinkTask, SinkDB, SinkFrontend, SinkPanic} {
		path, ok := sinkPaths[sink]
		if !ok {
			return fmt.Errorf("missing path for sink %s", sink)
		}
		baseLogger, closer := newSinkLogger(path, cfg.EnableStdout)
		if len(commonFields) > 0 {
			baseLogger = baseLogger.With(commonFields...)
		}
		r.sinks[sink] = baseLogger
		r.closers = append(r.closers, closer)
	}
	return nil
}

func filterRuntimeConfigFields(fields []zap.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	filtered := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		switch field.Key {
		case "module", "event", "trace_id", "span_id", "app_session_id", "source":
			continue
		default:
			filtered = append(filtered, field)
		}
	}
	return filtered
}

func newSinkLogger(path string, enableStdout bool) (*zap.Logger, io.Closer) {
	fileWriter := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    10,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	}
	fileSyncer := zapcore.AddSync(fileWriter)

	writer := fileSyncer
	if enableStdout {
		writer = zapcore.NewMultiWriteSyncer(fileSyncer, zapcore.AddSync(os.Stdout))
	}

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	return zap.New(core, zap.AddCaller()), fileWriter
}

func (r *Runtime) getSinkLogger(sink Sink) *zap.Logger {
	if r == nil {
		return zap.NewNop()
	}
	if log, ok := r.sinks[sink]; ok && log != nil {
		return log
	}
	if log, ok := r.sinks[SinkApp]; ok && log != nil {
		return log
	}
	return zap.NewNop()
}

func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}

	var err error
	for _, closer := range r.closers {
		if closer == nil {
			continue
		}
		err = errors.Join(err, closer.Close())
	}
	return err
}
