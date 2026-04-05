package logger

import (
	"encoding/json"
	"os"
	"strings"

	wailslogger "github.com/wailsapp/wails/v2/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FrontendErrorPayload 统一承接 Wails 前端错误事件，避免各平台重复解析 map 结构。
type FrontendErrorPayload struct {
	Page    string         `json:"page"`
	Route   string         `json:"route"`
	Message string         `json:"message"`
	Source  string         `json:"source"`
	Line    int            `json:"lineno"`
	Column  int            `json:"colno"`
	Stack   string         `json:"error"`
	Extra   map[string]any `json:"extra,omitempty"`
}

func NormalizeFrontendError(optionalData []interface{}) FrontendErrorPayload {
	if len(optionalData) == 0 || optionalData[0] == nil {
		return FrontendErrorPayload{}
	}

	raw, err := json.Marshal(optionalData[0])
	if err != nil {
		return FrontendErrorPayload{}
	}

	var payload FrontendErrorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return FrontendErrorPayload{}
	}
	return payload
}

// WailsLogger 将 Wails 内部日志接到统一 sink，避免继续写独立日志文件。
type WailsLogger struct {
	runtime *Runtime
	module  string
}

var _ wailslogger.Logger = (*WailsLogger)(nil)

func NewWailsLogger(runtime *Runtime, module string) *WailsLogger {
	if runtime == nil {
		runtime = Default()
	}
	if strings.TrimSpace(module) == "" {
		module = "wails"
	}
	return &WailsLogger{
		runtime: runtime,
		module:  module,
	}
}

func (l *WailsLogger) Print(message string) {
	l.write(SinkApp, "wails.print", zapcore.InfoLevel, message)
}

func (l *WailsLogger) Trace(message string) {
	l.write(SinkApp, "wails.trace", zapcore.DebugLevel, message)
}

func (l *WailsLogger) Debug(message string) {
	l.write(SinkApp, "wails.debug", zapcore.DebugLevel, message)
}

func (l *WailsLogger) Info(message string) {
	l.write(SinkApp, "wails.info", zapcore.InfoLevel, message)
}

func (l *WailsLogger) Warning(message string) {
	l.write(SinkApp, "wails.warning", zapcore.WarnLevel, message)
}

func (l *WailsLogger) Error(message string) {
	l.write(SinkError, "wails.error", zapcore.ErrorLevel, message)
}

func (l *WailsLogger) Fatal(message string) {
	l.write(SinkPanic, "wails.fatal", zapcore.ErrorLevel, message)
	os.Exit(1)
}

func (l *WailsLogger) write(sink Sink, event string, level zapcore.Level, message string) {
	trace := TraceContext{Source: "wails-runtime"}
	if l.runtime != nil {
		trace = l.runtime.NewTrace("wails-runtime")
	}

	base := l.runtime.getSinkLogger(sink).With(
		zap.String("module", l.module),
		zap.String("event", event),
		zap.String("trace_id", trace.TraceID),
		zap.String("span_id", trace.SpanID),
		zap.String("app_session_id", trace.AppSessionID),
		zap.String("source", trace.Source),
	)

	switch level {
	case zapcore.DebugLevel:
		base.Debug(message)
	case zapcore.WarnLevel:
		base.Warn(message)
	case zapcore.ErrorLevel:
		base.Error(message)
	default:
		base.Info(message)
	}
}
