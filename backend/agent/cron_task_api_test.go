package agent

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
)

func TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "go-stock-closeout-task-*")
	if err != nil {
		t.Fatalf("create temp root dir: %v", err)
	}
	runtime := logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})

	// ExecuteTask 会在末尾调用 UpdateRunInfo，需要最小 sqlite 初始化避免 db.Dao 为空。
	db.Init(filepath.Join(rootDir, "test.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open sql db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.Dao.AutoMigrate(&models.CronTask{}); err != nil {
		t.Fatalf("migrate cron task table: %v", err)
	}

	taskExecutorOverrides["trace_probe"] = func(ctx context.Context, task *models.CronTask) error {
		runtime.ForSink(logger.SinkAI, "trace-probe.ai").WithTrace(runtime.TraceOrNew(ctx, "ai")).Info(
			"ai.trace_probe",
			"ai sink reached",
		)
		runtime.ForSink(logger.SinkDB, "trace-probe.db").WithTrace(runtime.TraceOrNew(ctx, "db")).Info(
			"db.trace_probe",
			"db sink reached",
		)
		return nil
	}
	defer delete(taskExecutorOverrides, "trace_probe")

	api := NewCronTaskApi()
	task := &models.CronTask{
		Name:     "trace probe",
		TaskType: "trace_probe",
		CronExpr: "0 * * * * *",
	}
	if err := db.Dao.Create(task).Error; err != nil {
		t.Fatalf("seed cron task row: %v", err)
	}

	if err := api.ExecuteTask(context.Background(), task); err != nil {
		t.Fatalf("execute task: %v", err)
	}

	taskLog, err := os.ReadFile(filepath.Join(rootDir, "logs", "task.log"))
	if err != nil {
		t.Fatalf("read task log: %v", err)
	}
	aiLog, err := os.ReadFile(filepath.Join(rootDir, "logs", "ai.log"))
	if err != nil {
		t.Fatalf("read ai log: %v", err)
	}
	dbLog, err := os.ReadFile(filepath.Join(rootDir, "logs", "db.log"))
	if err != nil {
		t.Fatalf("read db log: %v", err)
	}

	traceID := extractTraceID(string(taskLog), "task.execute_started")
	if traceID == "" {
		t.Fatalf("expected task trace id, got %s", string(taskLog))
	}
	if !strings.Contains(string(aiLog), traceID) || !strings.Contains(string(dbLog), traceID) {
		t.Fatalf("expected ai/db logs to reuse %s, got ai=%s db=%s", traceID, string(aiLog), string(dbLog))
	}
}

func extractTraceID(content string, event string) string {
	re := regexp.MustCompile(`"trace_id":"([^"]+)"`)
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		if !strings.Contains(line, `"`+"event"+`":"`+event+`"`) {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}
