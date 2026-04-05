package logger

import (
	"time"

	"go.uber.org/zap"
)

func (r *Runtime) ForSink(sink Sink, module string) *Logger {
	return &Logger{runtime: r, sink: sink, module: module}
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Any(key string, value any) zap.Field {
	return zap.Any(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Uint(key string, value uint) zap.Field {
	return zap.Uint(key, value)
}

func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func Err(err error) zap.Field {
	return zap.NamedError("error", err)
}

func (l *Logger) WithTrace(trace TraceContext) *Logger {
	clone := *l
	clone.trace = trace
	return &clone
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	clone := *l
	if len(fields) == 0 {
		return &clone
	}
	clone.fields = append(append([]zap.Field{}, l.fields...), fields...)
	return &clone
}

func (l *Logger) Info(event, message string, fields ...zap.Field) {
	l.base(event).Info(message, mergeFields(l.fields, fields)...)
}

func (l *Logger) Warn(event, message string, fields ...zap.Field) {
	l.base(event).Warn(message, mergeFields(l.fields, fields)...)
}

func (l *Logger) Error(event, message string, fields ...zap.Field) {
	l.base(event).Error(message, mergeFields(l.fields, fields)...)
}

func (l *Logger) base(event string) *zap.Logger {
	return l.runtime.getSinkLogger(l.sink).With(
		zap.String("module", l.module),
		zap.String("event", event),
		zap.String("trace_id", l.trace.TraceID),
		zap.String("span_id", l.trace.SpanID),
		zap.String("app_session_id", l.trace.AppSessionID),
		zap.String("source", l.trace.Source),
	)
}

func mergeFields(base []zap.Field, extra []zap.Field) []zap.Field {
	switch {
	case len(base) == 0:
		return extra
	case len(extra) == 0:
		return base
	default:
		merged := make([]zap.Field, 0, len(base)+len(extra))
		merged = append(merged, base...)
		merged = append(merged, extra...)
		return merged
	}
}
