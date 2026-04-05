package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"
	"go-stock/backend/models"

	"github.com/cloudwego/eino/schema"
)

func TestModuleTrace_ReusesContextTrace(t *testing.T) {
	existing := logger.TraceContext{
		TraceID:      "trace-agent-1",
		SpanID:       "span-agent-1",
		AppSessionID: "session-agent-1",
		Source:       "http",
	}
	ctx := logger.WithTraceContext(context.Background(), existing)

	trace := moduleTrace(ctx, "agent-test")
	if trace != existing {
		t.Fatalf("expected moduleTrace to reuse %+v, got %+v", existing, trace)
	}
}

func TestCronTaskErrors_IncludeTaskErrorClass(t *testing.T) {
	logRoot, err := os.MkdirTemp("", "go-stock-cron-task-log-*")
	if err != nil {
		t.Fatalf("create temp log root: %v", err)
	}
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: logRoot,
			LogsDir: filepath.Join(logRoot, "logs"),
		},
		EnableStdout: false,
	})

	task := &models.CronTask{
		ID:       1,
		Name:     "broken-task",
		TaskType: "stock_analysis",
		Params:   "{invalid",
	}

	err = NewCronTaskApi().executeStockAnalysis(context.Background(), task)
	if err == nil {
		t.Fatalf("expected invalid params to return an error")
	}

	content, err := os.ReadFile(filepath.Join(logRoot, "logs", "task.log"))
	if err != nil {
		t.Fatalf("read task log: %v", err)
	}
	line := string(content)
	for _, want := range []string{
		`"event":"task.stock_analysis_params_invalid"`,
		`"error_class":"task_error"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected task log to contain %s, got %s", want, line)
		}
	}
}

func TestSafeSend_ReusesContextTrace(t *testing.T) {
	logRoot, err := os.MkdirTemp("", "go-stock-safe-send-log-*")
	if err != nil {
		t.Fatalf("create temp log root: %v", err)
	}
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: logRoot,
			LogsDir: filepath.Join(logRoot, "logs"),
		},
		EnableStdout: false,
	})

	trace := logger.TraceContext{
		TraceID:      "trace-safe-send-1",
		SpanID:       "span-safe-send-1",
		AppSessionID: "session-safe-send-1",
		Source:       "cron-task-execute",
	}
	ctx := logger.WithTraceContext(context.Background(), trace)

	ch := make(chan *schema.Message)
	close(ch)
	safeSend(ctx, ch, &schema.Message{Role: schema.Assistant, Content: "boom"})

	content, err := os.ReadFile(filepath.Join(logRoot, "logs", "ai.log"))
	if err != nil {
		t.Fatalf("read ai log: %v", err)
	}
	line := string(content)
	for _, want := range []string{
		`"event":"agent.channel_send_panic"`,
		`"trace_id":"trace-safe-send-1"`,
		`"span_id":"span-safe-send-1"`,
		`"app_session_id":"session-safe-send-1"`,
		`"source":"cron-task-execute"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected safeSend log to contain %s, got %s", want, line)
		}
	}
}
