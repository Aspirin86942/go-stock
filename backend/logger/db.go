package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	runtime                   *Runtime
	slowThreshold             time.Duration
	logLevel                  gormlogger.LogLevel
	ignoreRecordNotFoundError bool
}

func NewGormLogger(runtime *Runtime, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		runtime:                   runtime,
		slowThreshold:             slowThreshold,
		logLevel:                  gormlogger.Warn,
		ignoreRecordNotFoundError: true,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	if l == nil {
		return NewGormLogger(nil, 0)
	}
	clone := *l
	clone.logLevel = level
	return &clone
}

func (l *GormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l == nil || l.logLevel < gormlogger.Info {
		return
	}
	l.base().Info("db.gorm.info", "gorm info", String("message", fmt.Sprintf(msg, data...)))
}

func (l *GormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l == nil || l.logLevel < gormlogger.Warn {
		return
	}
	l.base().Warn("db.gorm.warn", "gorm warning", String("message", fmt.Sprintf(msg, data...)))
}

func (l *GormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l == nil || l.logLevel < gormlogger.Error {
		return
	}
	l.base().Error("db.gorm.error", "gorm error", String("message", fmt.Sprintf(msg, data...)))
}

func (l *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l == nil || l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	if err != nil && l.logLevel >= gormlogger.Error {
		if errors.Is(err, gormlogger.ErrRecordNotFound) && l.ignoreRecordNotFoundError {
			return
		}
		sql, rows := fc()
		l.base().Error(
			"db.query.failed",
			"gorm query failed",
			append(queryFields(sql, rows, elapsed), String("error_message", err.Error()))...,
		)
		return
	}

	if l.slowThreshold > 0 && elapsed >= l.slowThreshold && l.logLevel >= gormlogger.Warn {
		sql, rows := fc()
		l.base().Warn(
			"db.query.slow",
			"gorm query exceeded threshold",
			queryFields(sql, rows, elapsed)...,
		)
		return
	}

	if l.logLevel >= gormlogger.Info {
		sql, rows := fc()
		l.base().Info(
			"db.query.completed",
			"gorm query completed",
			queryFields(sql, rows, elapsed)...,
		)
	}
}

func (l *GormLogger) base() *Logger {
	var runtime *Runtime
	if l != nil {
		runtime = l.runtime
	}
	if runtime == nil {
		runtime = Default()
	}

	trace := TraceContext{Source: "gorm"}
	if runtime != nil {
		trace = runtime.NewTrace("gorm")
	}
	return runtime.ForSink(SinkDB, "gorm").WithTrace(trace)
}

func queryFields(sql string, rows int64, elapsed time.Duration) []zap.Field {
	return []zap.Field{
		String("sql", sql),
		Int64("rows_affected", rows),
		Int64("duration_ms", elapsed.Milliseconds()),
	}
}
