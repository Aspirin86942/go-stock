package logger

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGormLogger_LogsErrorAndSlowStatements(t *testing.T) {
	buf := &bytes.Buffer{}
	gormLogger := NewGormLogger(newRuntimeForTest(buf), 5*time.Millisecond)

	gormLogger.Trace(context.Background(), time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT 1", 1
	}, errors.New("boom"))

	line := buf.String()
	if !strings.Contains(line, `"event":"db.query.failed"`) || !strings.Contains(line, `"error_message":"boom"`) {
		t.Fatalf("expected db failure log, got %s", line)
	}

	buf.Reset()
	gormLogger.Trace(context.Background(), time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	line = buf.String()
	if !strings.Contains(line, `"event":"db.query.slow"`) || !strings.Contains(line, `"sql":"SELECT 1"`) {
		t.Fatalf("expected slow query log, got %s", line)
	}
}

func TestGormLogger_ReusesTraceFromContext(t *testing.T) {
	buf := &bytes.Buffer{}
	gormLogger := NewGormLogger(newRuntimeForTest(buf), 5*time.Millisecond)
	existing := TraceContext{
		TraceID:      "trace-db-1",
		SpanID:       "span-db-1",
		AppSessionID: "session-db-1",
		Source:       "http",
	}
	ctx := WithTraceContext(context.Background(), existing)

	gormLogger.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT 1", 1
	}, errors.New("boom"))

	line := buf.String()
	for _, want := range []string{
		`"event":"db.query.failed"`,
		`"trace_id":"trace-db-1"`,
		`"span_id":"span-db-1"`,
		`"app_session_id":"session-db-1"`,
		`"source":"http"`,
		`"error_class":"db_error"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected gorm trace log to contain %s, got %s", want, line)
		}
	}
}
