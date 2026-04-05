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
