package logger

import (
	"fmt"
	"go-stock/backend/apppath"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

func (r *Runtime) bootstrapSinks(cfg Config) error {
	if cfg.Paths.LogsDir == "" {
		return fmt.Errorf("logger config Paths.LogsDir is empty")
	}
	if err := os.MkdirAll(cfg.Paths.LogsDir, 0o755); err != nil {
		return fmt.Errorf("ensure logger logs dir %s: %w", cfg.Paths.LogsDir, err)
	}

	sinkPaths := buildSinkPaths(cfg.Paths)
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	for _, sink := range []Sink{SinkApp, SinkError, SinkHTTP, SinkAI, SinkTask, SinkDB, SinkFrontend, SinkPanic} {
		path, ok := sinkPaths[sink]
		if !ok {
			return fmt.Errorf("missing path for sink %s", sink)
		}
		r.sinks[sink] = newSinkLogger(path, encoder, cfg.EnableStdout)
	}
	return nil
}

func newSinkLogger(path string, encoder zapcore.Encoder, enableStdout bool) *zap.Logger {
	fileSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    10,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	})

	writer := fileSyncer
	if enableStdout {
		writer = zapcore.NewMultiWriteSyncer(fileSyncer, zapcore.AddSync(os.Stdout))
	}

	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	return zap.New(core, zap.AddCaller())
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
