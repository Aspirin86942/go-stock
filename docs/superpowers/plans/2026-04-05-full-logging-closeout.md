# Full Logging Closeout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining full-logging gaps so `trace_id` is actually reused across key chains, frontend `console.error` is bridged into file logs, automated tests prove cross-sink trace reuse, and a developer can manually verify the real files under `logs` and `payloads`.

**Architecture:** Add one lightweight trace-context layer inside `backend/logger`, then make HTTP, Wails frontend error intake, task/AI/tool/DB adapters, and frontend error bridging consume that layer instead of calling `NewTrace()` ad hoc. Prove the closure with two real file-backed integration tests and one explicit manual-verification document that points developers to the runtime `logs` directory instead of adding any new UI or export path.

**Tech Stack:** Go, Wails, Zap, Go `testing`, Node `node:test`, Vite

---

## File Map

- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\trace_context.go`: new trace-context helpers for `context.Context` round-trip, `trace_id` reuse, and "reuse or create" runtime helpers.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\trace_context_test.go`: focused unit tests for trace-context round-trip and runtime reuse behavior.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\http.go`: bind one request trace into the request context and reuse it for both warning and completion events.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\http_test.go`: prove the middleware reuses the request trace instead of generating a fresh one per event.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\wails.go`: extend frontend payload normalization to keep optional `traceId`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\app_common.go`: reuse `traceId` from frontend payload when present and add the required `error_class`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\app_common_test.go`: real file-backed test for frontend error logging through `frontend.log`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\utils.go`: add a context-aware logger factory so AI/data code can reuse incoming trace context instead of allocating a new one.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\utils_trace_test.go`: prove the data helper reuses an injected trace.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\openai_stream.go`: switch AI stream logs to the context-aware helper and add `error_class` on failures.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\openai_tools.go`: switch AI/tool logs to the context-aware helper and add `error_class` on failures.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\agent.go`: change `moduleTrace` to reuse trace context from `context.Context`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\agent_api.go`: pass the active context through agent creation and stream processing logs.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\cron_task_api.go`: bind one task trace at `ExecuteTask`, reuse it downstream, and expose a tiny test-only executor override seam for integration tests.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\trace_reuse_test.go`: focused test for agent trace helper reuse.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\cron_task_api_test.go`: real file-backed task integration test for `task.log`, `ai.log`, and `db.log`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\choice_stock_by_indicators_tool.go`: make tool traces come from the invocation context.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\data_tools_wrapper.go`: make wrapper logs reuse the tool invocation trace.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\stock_code_tool.go`: make tool logs reuse the tool invocation trace and add `tool_error` on failures.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\market_news_tool.go`: reuse the tool invocation trace for iterative item logging.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\db.go`: reuse trace context from GORM `Trace(ctx, ...)` and attach `db_error` to failure events.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\db_test.go`: extend DB tests to prove GORM reuses the incoming trace.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\closeout_integration_test.go`: real file-backed HTTP integration test covering `http.log`, `db.log`, and `logs\payloads`.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\frontend\src\utils\frontendLogger.mjs`: add optional `traceId`, `console.error` proxying, and short-window dedupe.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\frontend\src\utils\frontendLogger.test.mjs`: prove `console.error` is forwarded, emitted once, and deduped.
- `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\docs\superpowers\checklists\2026-04-05-full-logging-manual-validation.md`: manual closeout checklist telling developers exactly which files under `logs` and `payloads` to inspect.

### Task 1: Add logger trace-context primitives

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\trace_context.go`
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\trace_context_test.go`

- [ ] **Step 1: Write the failing trace-context unit tests**

```go
func TestTraceContextRoundTripThroughContext(t *testing.T) {
	trace := TraceContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		AppSessionID: "session-1",
		Source:       "http",
	}

	ctx := WithTraceContext(context.Background(), trace)
	got, ok := TraceContextFromContext(ctx)
	if !ok {
		t.Fatalf("expected trace context to round-trip")
	}
	if got != trace {
		t.Fatalf("expected %#v, got %#v", trace, got)
	}
}

func TestRuntimeEnsureTraceContext_ReusesExistingTrace(t *testing.T) {
	runtime := newRuntimeForTest(io.Discard)
	existing := TraceContext{
		TraceID:      "trace-existing",
		SpanID:       "span-existing",
		AppSessionID: "session-existing",
		Source:       "http",
	}

	ctx, trace := runtime.EnsureTraceContext(WithTraceContext(context.Background(), existing), "db")
	if trace != existing {
		t.Fatalf("expected existing trace to be reused, got %#v", trace)
	}

	roundTrip, ok := TraceContextFromContext(ctx)
	if !ok || roundTrip != existing {
		t.Fatalf("expected context to keep existing trace, got %#v ok=%v", roundTrip, ok)
	}
}
```

- [ ] **Step 2: Run the logger package tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger -run 'TestTraceContextRoundTripThroughContext|TestRuntimeEnsureTraceContext_ReusesExistingTrace' -count=1
```

Expected: build failure because `WithTraceContext`, `TraceContextFromContext`, and `EnsureTraceContext` do not exist yet.

- [ ] **Step 3: Implement the minimal trace-context helpers**

```go
type traceContextKey struct{}

var loggerTraceContextKey traceContextKey

func WithTraceContext(ctx context.Context, trace TraceContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerTraceContextKey, trace)
}

func TraceContextFromContext(ctx context.Context) (TraceContext, bool) {
	if ctx == nil {
		return TraceContext{}, false
	}
	trace, ok := ctx.Value(loggerTraceContextKey).(TraceContext)
	if !ok || strings.TrimSpace(trace.TraceID) == "" {
		return TraceContext{}, false
	}
	return trace, true
}

func (r *Runtime) TraceOrNew(ctx context.Context, source string) TraceContext {
	if trace, ok := TraceContextFromContext(ctx); ok {
		return trace
	}
	return r.NewTrace(source)
}

func (r *Runtime) EnsureTraceContext(ctx context.Context, source string) (context.Context, TraceContext) {
	trace := r.TraceOrNew(ctx, source)
	return WithTraceContext(ctx, trace), trace
}
```

- [ ] **Step 4: Re-run the logger package tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger -run 'TestTraceContextRoundTripThroughContext|TestRuntimeEnsureTraceContext_ReusesExistingTrace' -count=1
```

Expected: command exits `0`.

- [ ] **Step 5: Commit the trace-context foundation**

```powershell
git add backend/logger/trace_context.go backend/logger/trace_context_test.go
git commit -m "feat: add logger trace context helpers"
```

### Task 2: Reuse one trace in HTTP middleware and Wails frontend-error intake

**Files:**
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\http.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\http_test.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\wails.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\app_common.go`
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\app_common_test.go`

- [ ] **Step 1: Write the failing reuse tests for HTTP and frontend error intake**

```go
func TestHTTPMiddleware_ReusesRequestContextTrace(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	existing := TraceContext{
		TraceID:      "trace-http",
		SpanID:       "span-http",
		AppSessionID: "session-http",
		Source:       "http",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/trace", nil)
	req.Body = errReader{}
	req = req.WithContext(WithTraceContext(req.Context(), existing))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected warning and completion log, got %d lines: %s", len(lines), buf.String())
	}
	for _, line := range lines {
		if !strings.Contains(line, `"trace_id":"trace-http"`) {
			t.Fatalf("expected middleware to reuse request trace, got %s", line)
		}
	}
}
```

```go
func TestLogFrontendRuntimeError_ReusesPayloadTraceID(t *testing.T) {
	rootDir := t.TempDir()
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})

	logFrontendRuntimeError([]interface{}{
		map[string]any{
			"page":    "stock.vue",
			"route":   "/stock",
			"message": "boom",
			"error":   "stack-line-1",
			"traceId": "trace-frontend",
		},
	})

	content, err := os.ReadFile(filepath.Join(rootDir, "logs", "frontend.log"))
	if err != nil {
		t.Fatalf("read frontend log: %v", err)
	}
	line := string(content)
	for _, want := range []string{
		`"trace_id":"trace-frontend"`,
		`"event":"frontend.error"`,
		`"error_class":"frontend_error"`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected frontend log to contain %s, got %s", want, line)
		}
	}
}
```

- [ ] **Step 2: Run the focused tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestHTTPMiddleware_ReusesRequestContextTrace -count=1
conda run -n test go test . -run TestLogFrontendRuntimeError_ReusesPayloadTraceID -count=1
```

Expected: the HTTP test fails because middleware still calls `NewTrace("http")`, and the frontend test fails because `traceId` is ignored and `error_class` is missing.

- [ ] **Step 3: Implement request-trace binding and frontend trace reuse**

```go
func (rt *Runtime) HTTPMiddleware(module string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ctx, trace := rt.EnsureTraceContext(req.Context(), "http")
		req = req.WithContext(ctx)

		recorder := newResponseRecorder(w, rt.payloads, "http-response")
		writer := wrapResponseWriter(recorder, w)

		requestBody, readErr := readRequestBody(req)
		if readErr != nil {
			rt.ForSink(SinkHTTP, module).WithTrace(trace).Warn(
				"http.request.body.read_failed",
				"failed to read http request body",
				String("error_class", "http_error"),
				String("method", req.Method),
				String("path", req.URL.Path),
				String("error_message", readErr.Error()),
			)
		}

		next.ServeHTTP(writer, req)

		responseRef, responseErr := recorder.PayloadRef()
		fields := []zap.Field{
			String("method", req.Method),
			String("path", req.URL.Path),
			Int("status_code", recorder.status),
			Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		fields = append(fields, payloadFields("request_payload", rt.savePayload("http-request", requestBody))...)
		fields = append(fields, payloadFields("response_payload", responseRef)...)
		if responseErr != nil {
			fields = append(fields, String("response_payload_error", responseErr.Error()))
		}

		rt.ForSink(SinkHTTP, module).WithTrace(trace).Info(
			"http.request.completed",
			"handled http request",
			fields...,
		)
	})
}
```

```go
type FrontendErrorPayload struct {
	Page    string         `json:"page"`
	Route   string         `json:"route"`
	Message string         `json:"message"`
	Source  string         `json:"source"`
	Line    int            `json:"lineno"`
	Column  int            `json:"colno"`
	Stack   string         `json:"error"`
	TraceID string         `json:"traceId"`
	Extra   map[string]any `json:"extra,omitempty"`
}
```

```go
func logFrontendRuntimeError(optionalData []interface{}) {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return
	}

	payload := logger.NormalizeFrontendError(optionalData)
	trace := runtimeLogger.NewTrace("wails-frontend")
	if strings.TrimSpace(payload.TraceID) != "" {
		trace.TraceID = payload.TraceID
	}

	runtimeLogger.ForSink(logger.SinkFrontend, "frontend").WithTrace(trace).Error(
		"frontend.error",
		"frontend runtime error",
		logger.String("error_class", "frontend_error"),
		logger.String("page", payload.Page),
		logger.String("route", payload.Route),
		logger.String("error_message", payload.Message),
		logger.String("source", payload.Source),
		logger.Int("lineno", payload.Line),
		logger.Int("colno", payload.Column),
		logger.String("stack", payload.Stack),
		logger.Any("extra", payload.Extra),
	)
}
```

- [ ] **Step 4: Re-run the focused tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestHTTPMiddleware_ReusesRequestContextTrace -count=1
conda run -n test go test . -run TestLogFrontendRuntimeError_ReusesPayloadTraceID -count=1
```

Expected: both commands exit `0`.

- [ ] **Step 5: Commit the HTTP and Wails trace reuse**

```powershell
git add backend/logger/http.go backend/logger/http_test.go backend/logger/wails.go app_common.go app_common_test.go
git commit -m "feat: reuse trace in http and frontend error logging"
```

### Task 3: Propagate trace reuse into task, AI, tool, and DB helpers

**Files:**
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\utils.go`
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\utils_trace_test.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\openai_stream.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\data\openai_tools.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\agent.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\agent_api.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\cron_task_api.go`
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\trace_reuse_test.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\choice_stock_by_indicators_tool.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\data_tools_wrapper.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\stock_code_tool.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\tools\market_news_tool.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\db.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\db_test.go`

- [ ] **Step 1: Write the failing helper-level reuse tests**

```go
func TestModuleTrace_ReusesContextTrace(t *testing.T) {
	rootDir := t.TempDir()
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})

	trace := logger.TraceContext{
		TraceID:      "trace-task",
		SpanID:       "span-task",
		AppSessionID: "session-task",
		Source:       "cron",
	}
	got := moduleTrace(logger.WithTraceContext(context.Background(), trace), "cron-task-execute")
	if got != trace {
		t.Fatalf("expected moduleTrace to reuse %#v, got %#v", trace, got)
	}
}
```

```go
func TestDataModuleLog_WithContextReusesTrace(t *testing.T) {
	rootDir := t.TempDir()
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})

	trace := logger.TraceContext{
		TraceID:      "trace-ai",
		SpanID:       "span-ai",
		AppSessionID: "session-ai",
		Source:       "cron",
	}
	dataModuleLogger(logger.SinkAI, "data.trace_test").
		WithContext(logger.WithTraceContext(context.Background(), trace), "openai-stream").
		Info("data.trace_test", "reused trace")

	content, err := os.ReadFile(filepath.Join(rootDir, "logs", "ai.log"))
	if err != nil {
		t.Fatalf("read ai log: %v", err)
	}
	if !strings.Contains(string(content), `"trace_id":"trace-ai"`) {
		t.Fatalf("expected ai log to reuse trace, got %s", string(content))
	}
}
```

```go
func TestGormLogger_ReusesTraceFromContext(t *testing.T) {
	buf := &bytes.Buffer{}
	gormLogger := NewGormLogger(newRuntimeForTest(buf), 5*time.Millisecond)
	ctx := WithTraceContext(context.Background(), TraceContext{
		TraceID: "trace-db",
		SpanID:  "span-db",
		Source:  "cron",
	})

	gormLogger.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT 1", 1
	}, errors.New("boom"))

	line := buf.String()
	for _, want := range []string{`"trace_id":"trace-db"`, `"error_class":"db_error"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected db log to contain %s, got %s", want, line)
		}
	}
}
```

- [ ] **Step 2: Run the focused helper tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/agent ./backend/data ./backend/logger -run 'TestModuleTrace_ReusesContextTrace|TestDataModuleLog_WithContextReusesTrace|TestGormLogger_ReusesTraceFromContext' -count=1
```

Expected: build failure because `moduleTrace` still has no `context.Context` input, `dataModuleLog.WithContext` does not exist, and `GormLogger` still ignores the incoming context trace.

- [ ] **Step 3: Implement context-aware helpers and wire them through task, AI, tool, and DB code**

```go
func (l dataModuleLog) WithContext(ctx context.Context, source string) *logger.Logger {
	runtime := logger.Default()
	if runtime == nil {
		panic("logger runtime not initialized")
	}
	return runtime.ForSink(l.sink, l.module).WithTrace(runtime.TraceOrNew(ctx, source))
}

func moduleTrace(ctx context.Context, source string) logger.TraceContext {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return logger.TraceContext{Source: source}
	}
	return runtimeLogger.TraceOrNew(ctx, source)
}

func toolTrace(ctx context.Context, source string) logger.TraceContext {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return logger.TraceContext{Source: source}
	}
	return runtimeLogger.TraceOrNew(ctx, source)
}
```

```go
func (a *CronTaskApi) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return fmt.Errorf("logger runtime not initialized")
	}

	ctx, trace := runtimeLogger.EnsureTraceContext(ctx, "cron-task-execute")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.execute_started",
			"started cron task execution",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
			logger.String("task_type", task.TaskType),
		)
	}

	err := a.executeTaskByType(ctx, task)
	if err != nil {
		taskLogger("agent.cron_task").WithTrace(trace).Error(
			"task.execute_failed",
			"cron task execution failed",
			logger.String("error_class", "task_error"),
			logger.String("error_message", err.Error()),
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
			logger.Err(err),
		)
	}
	return err
}
```

```go
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l == nil || l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	if err != nil && l.logLevel >= gormlogger.Error {
		if errors.Is(err, gormlogger.ErrRecordNotFound) && l.ignoreRecordNotFoundError {
			return
		}
		sql, rows := fc()
		l.base(ctx).Error(
			"db.query.failed",
			"gorm query failed",
			append(queryFields(sql, rows, elapsed),
				String("error_class", "db_error"),
				String("error_message", err.Error()),
			)...,
		)
		return
	}

	if l.slowThreshold > 0 && elapsed >= l.slowThreshold && l.logLevel >= gormlogger.Warn {
		sql, rows := fc()
		l.base(ctx).Warn(
			"db.query.slow",
			"gorm query exceeded threshold",
			queryFields(sql, rows, elapsed)...,
		)
		return
	}

	if l.logLevel >= gormlogger.Info {
		sql, rows := fc()
		l.base(ctx).Info(
			"db.query.completed",
			"gorm query completed",
			queryFields(sql, rows, elapsed)...,
		)
	}
}

func (l *GormLogger) base(ctx context.Context) *Logger {
	var runtime *Runtime
	if l != nil {
		runtime = l.runtime
	}
	if runtime == nil {
		runtime = Default()
	}

	trace := TraceContext{Source: "gorm"}
	if runtime != nil {
		trace = runtime.TraceOrNew(ctx, "gorm")
	}
	return runtime.ForSink(SinkDB, "gorm").WithTrace(trace)
}
```

```go
func (t *DataToolWrapper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if log := toolModuleLogger("agent.tool.data_wrapper"); log != nil {
		log.WithTrace(toolTrace(ctx, "tool-data-wrapper")).Info(
			"tool.data_wrapper.called",
			"data tool wrapper called",
			logger.String("tool_name", t.name),
			logger.String("arguments", argumentsInJSON),
		)
	}
	return t.handler(argumentsInJSON)
}
```

```go
trace := moduleTrace(ctx, "chat-with-context")

go func() {
	defer func() {
		if r := recover(); r != nil {
			if log := aiLogger("agent.api"); log != nil {
				log.WithTrace(trace).Error(
					"agent.chat_context_panic",
					"panic in chat with context",
					logger.String("error_class", "ai_error"),
					logger.Int("ai_config_id", aiConfigId),
					logger.Any("panic_value", r),
				)
			}
			ch <- &schema.Message{Role: schema.Assistant, Content: fmt.Sprintf("❌ 内部错误: %v", r)}
			close(ch)
		}
	}()

	stockAiAgent := receiver.newStockAiAgent(&ctx, aiConfigId, thinkingMode)
	if stockAiAgent == nil {
		ch <- &schema.Message{Role: schema.Assistant, Content: "❌ AI 配置不存在或无效，请检查 AI 配置"}
		close(ch)
		return
	}
}()
```

```go
if err != nil {
	openAIStreamLog.WithContext(o.Ctx(), "openai-stream").Error(
		"data.openai_stream.chat_failed",
		"stream request failed",
		logger.String("error_class", "ai_error"),
		logger.String("error_message", err.Error()),
		logger.Err(err),
	)
	ch <- map[string]any{
		"code":     0,
		"question": question,
		"content":  err.Error(),
	}
	return
}
```

```go
if hErr := handler(o, funcArguments, toolCtx); hErr != nil {
	openAIToolsLog.WithContext(o.Ctx(), "openai-tools").Error(
		"data.openai_tools.tool_call_failed",
		"tool call failed",
		logger.String("error_class", "tool_error"),
		logger.String("tool_name", funcName),
		logger.String("error_message", hErr.Error()),
		logger.Err(hErr),
	)
	ch <- map[string]any{
		"code":     0,
		"question": question,
		"content":  hErr.Error(),
	}
}
```

```go
trace := toolTrace(ctx, "tool-query-stock-code")
if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
	log.WithTrace(trace).Info(
		"tool.stock_code.called",
		"query stock code tool called",
		logger.String("arguments", argumentsInJSON),
	)
}

parms := map[string]any{}
err := json.Unmarshal([]byte(argumentsInJSON), &parms)
if err != nil {
	if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
		log.WithTrace(trace).Error(
			"tool.stock_code.arguments_invalid",
			"unmarshal stock code tool args failed",
			logger.String("error_class", "tool_error"),
			logger.String("error_message", err.Error()),
			logger.Err(err),
		)
	}
	return "", err
}
```

- [ ] **Step 4: Re-run the focused helper tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/agent ./backend/data ./backend/logger -run 'TestModuleTrace_ReusesContextTrace|TestDataModuleLog_WithContextReusesTrace|TestGormLogger_ReusesTraceFromContext' -count=1
```

Expected: command exits `0`.

- [ ] **Step 5: Commit the context-aware task and AI wiring**

```powershell
git add backend/data/utils.go backend/data/utils_trace_test.go backend/data/openai_stream.go backend/data/openai_tools.go backend/agent/agent.go backend/agent/agent_api.go backend/agent/cron_task_api.go backend/agent/trace_reuse_test.go backend/agent/tools/choice_stock_by_indicators_tool.go backend/agent/tools/data_tools_wrapper.go backend/agent/tools/stock_code_tool.go backend/agent/tools/market_news_tool.go backend/logger/db.go backend/logger/db_test.go
git commit -m "feat: propagate trace reuse through task ai tool db logs"
```

### Task 4: Bridge `console.error` without breaking local console output

**Files:**
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\frontend\src\utils\frontendLogger.mjs`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\frontend\src\utils\frontendLogger.test.mjs`

- [ ] **Step 1: Write the failing frontend tests for `traceId` and `console.error` proxying**

```javascript
test('buildFrontendErrorPayload keeps traceId when provided', () => {
  const payload = buildFrontendErrorPayload({
    page: 'stock.vue',
    route: '/stock',
    message: 'boom',
    error: new Error('boom'),
    traceId: 'trace-frontend',
  })

  assert.equal(payload.traceId, 'trace-frontend')
})
```

```javascript
test('console.error proxy keeps original output and emits once for the same error', () => {
  __resetFrontendErrorHandlersForTest()
  const previousWindow = globalThis.window
  const previousConsoleError = console.error
  const listeners = new Map()
  const emitted = []
  const forwarded = []

  globalThis.window = {
    location: { hash: '#/stock', pathname: '/stock' },
    runtime: {
      EventsEmit: (...args) => emitted.push(args),
    },
    addEventListener: (name, handler) => {
      listeners.set(name, handler)
    },
  }
  console.error = (...args) => {
    forwarded.push(args)
  }

  try {
    installFrontendErrorHandlers('main.js')
    const boom = new Error('console boom')

    console.error(boom)
    console.error(boom)

    assert.equal(forwarded.length, 2)
    assert.equal(emitted.length, 1)
    assert.equal(emitted[0][0], 'frontendError')
    assert.equal(emitted[0][1].message, 'console boom')
  } finally {
    console.error = previousConsoleError
    globalThis.window = previousWindow
    __resetFrontendErrorHandlersForTest()
  }
})
```

- [ ] **Step 2: Run the frontend tests and confirm they fail**

Run:

```powershell
node --test .\frontend\src\utils\frontendLogger.test.mjs
```

Expected: assertions fail because `traceId` is currently dropped and `installFrontendErrorHandlers` does not proxy `console.error`.

- [ ] **Step 3: Implement the proxy, dedupe window, and payload extension**

```javascript
let handlersInstalled = false
let originalConsoleError = null
let recentErrorObjects = new WeakSet()
const recentFingerprints = new Map()

function fingerprintConsoleArgs(args) {
  return args.map((item) => {
    if (item instanceof Error) {
      return item.stack || item.message
    }
    return String(item)
  }).join(' | ')
}

function markRecentlyReported(error, fingerprint) {
  if (error instanceof Error) {
    recentErrorObjects.add(error)
  }
  recentFingerprints.set(fingerprint, Date.now())
  setTimeout(() => {
    recentFingerprints.delete(fingerprint)
  }, 250)
}

function wasRecentlyReported(error, fingerprint) {
  if (error instanceof Error && recentErrorObjects.has(error)) {
    return true
  }
  return recentFingerprints.has(fingerprint)
}

export function buildFrontendErrorPayload({ page, route, message, source, lineno, colno, error, traceId, extra = {} }) {
  return {
    page,
    route: resolveRoute(route),
    message: resolveMessage(message, error),
    source: source || '',
    lineno: Number.isFinite(lineno) ? lineno : 0,
    colno: Number.isFinite(colno) ? colno : 0,
    error: resolveError(error),
    traceId: traceId || '',
    extra,
  }
}
```

```javascript
function installConsoleErrorProxy(page) {
  if (originalConsoleError) {
    return
  }

  originalConsoleError = console.error
  console.error = (...args) => {
    originalConsoleError(...args)

    const error = args.find((item) => item instanceof Error) || null
    const fingerprint = fingerprintConsoleArgs(args)
    if (wasRecentlyReported(error, fingerprint)) {
      return
    }

    const message = error?.message || fingerprint || 'console.error'
    const emitted = emitFrontendError({
      page,
      message,
      error: error || new Error(message),
      extra: { consoleArgs: args.map((item) => String(item)) },
    })
    if (emitted) {
      markRecentlyReported(error, fingerprint)
    }
  }
}

export function installFrontendErrorHandlers(page) {
  if (typeof window === 'undefined' || handlersInstalled) {
    return
  }
  handlersInstalled = true

  installConsoleErrorProxy(page)

  window.addEventListener('error', (event) => {
    const message = event?.message
    const error = event?.error
    if (isResizeObserverNoise(message) || isResizeObserverNoise(error?.message) || isResizeObserverNoise(error?.stack)) {
      event.preventDefault?.()
      return
    }
    const fingerprint = fingerprintConsoleArgs([error || message || 'window.error'])
    if (!wasRecentlyReported(error, fingerprint)) {
      const emitted = emitFrontendError({
        page,
        message,
        source: event?.filename,
        lineno: event?.lineno,
        colno: event?.colno,
        error,
      })
      if (emitted) {
        markRecentlyReported(error, fingerprint)
      }
    }
  })

  window.addEventListener('unhandledrejection', (event) => {
    const reason = event?.reason
    const reasonMessage = reason?.message || reason
    if (isResizeObserverNoise(reasonMessage)) {
      event.preventDefault?.()
      return
    }
    const rejectionError = reason instanceof Error ? reason : new Error(String(reason))
    const fingerprint = fingerprintConsoleArgs([rejectionError])
    if (!wasRecentlyReported(rejectionError, fingerprint)) {
      const emitted = emitFrontendError({
        page,
        message: rejectionError.message || 'unhandledrejection',
        error: rejectionError,
      })
      if (emitted) {
        markRecentlyReported(rejectionError, fingerprint)
      }
    }
    originalConsoleError?.('Unhandled promise rejection:', reason)
  })
}

export function __resetFrontendErrorHandlersForTest() {
  handlersInstalled = false
  if (originalConsoleError) {
    console.error = originalConsoleError
    originalConsoleError = null
  }
  recentErrorObjects = new WeakSet()
  recentFingerprints.clear()
}
```

- [ ] **Step 4: Re-run the frontend tests and confirm they pass**

Run:

```powershell
node --test .\frontend\src\utils\frontendLogger.test.mjs
```

Expected: command exits `0`.

- [ ] **Step 5: Build the frontend to catch any syntax or bundling regressions**

Run:

```powershell
Push-Location frontend
npm run build
Pop-Location
```

Expected: Vite build exits `0`.

- [ ] **Step 6: Commit the frontend error bridge**

```powershell
git add frontend/src/utils/frontendLogger.mjs frontend/src/utils/frontendLogger.test.mjs
git commit -m "feat: bridge console error into frontend logs"
```

### Task 5: Add real file-backed closeout integration tests for the HTTP and task chains

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\logger\closeout_integration_test.go`
- Modify: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\cron_task_api.go`
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\backend\agent\cron_task_api_test.go`

- [ ] **Step 1: Write the failing real file-backed integration tests**

```go
func TestHTTPTraceFlowsAcrossHTTPAndDBSinks(t *testing.T) {
	rootDir := t.TempDir()
	runtime := MustInit(Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})
	runtime.AttachPayloadStore(NewPayloadStore(filepath.Join(rootDir, "logs"), 16, 5<<20))

	handler := runtime.HTTPMiddleware("integration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.ForSink(SinkDB, "integration.db").WithTrace(runtime.TraceOrNew(r.Context(), "db")).Info(
			"db.integration.hit",
			"handled db step",
			String("table", "trace_probe"),
		)
		_, _ = w.Write([]byte(strings.Repeat("x", 128)))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/trace-probe", strings.NewReader(strings.Repeat("q", 128)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	httpLog, _ := os.ReadFile(filepath.Join(rootDir, "logs", "http.log"))
	dbLog, _ := os.ReadFile(filepath.Join(rootDir, "logs", "db.log"))
	traceID := extractTraceID(string(httpLog), "http.request.completed")
	if traceID == "" {
		t.Fatalf("expected http log trace id, got %s", string(httpLog))
	}
	if !strings.Contains(string(dbLog), traceID) {
		t.Fatalf("expected db log to reuse %s, got %s", traceID, string(dbLog))
	}
	payloadFiles, _ := filepath.Glob(filepath.Join(rootDir, "logs", "payloads", "*", "*.log"))
	if len(payloadFiles) == 0 {
		t.Fatalf("expected spilled payload file under logs/payloads")
	}
}
```

```go
func TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks(t *testing.T) {
	rootDir := t.TempDir()
	runtime := logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: rootDir,
			LogsDir: filepath.Join(rootDir, "logs"),
		},
		EnableStdout: false,
	})

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
		ID:       1,
		Name:     "trace probe",
		TaskType: "trace_probe",
		CronExpr: "0 * * * * *",
	}

	if err := api.ExecuteTask(context.Background(), task); err != nil {
		t.Fatalf("execute task: %v", err)
	}

	taskLog, _ := os.ReadFile(filepath.Join(rootDir, "logs", "task.log"))
	aiLog, _ := os.ReadFile(filepath.Join(rootDir, "logs", "ai.log"))
	dbLog, _ := os.ReadFile(filepath.Join(rootDir, "logs", "db.log"))
	traceID := extractTraceID(string(taskLog), "task.execute_started")
	if traceID == "" {
		t.Fatalf("expected task trace id, got %s", string(taskLog))
	}
	if !strings.Contains(string(aiLog), traceID) || !strings.Contains(string(dbLog), traceID) {
		t.Fatalf("expected ai/db logs to reuse %s, got ai=%s db=%s", traceID, string(aiLog), string(dbLog))
	}
}
```

- [ ] **Step 2: Run the focused integration tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger ./backend/agent -run 'TestHTTPTraceFlowsAcrossHTTPAndDBSinks|TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks' -count=1
```

Expected: the HTTP test fails on trace mismatch or missing spill evidence, and the task test fails because there is no task executor override seam yet or the downstream sinks still get new traces.

- [ ] **Step 3: Implement the minimal test seam and make the integration tests pass through real files**

```go
var taskExecutorOverrides = map[string]func(context.Context, *models.CronTask) error{}

func (a *CronTaskApi) executeTaskByType(ctx context.Context, task *models.CronTask) error {
	if override, ok := taskExecutorOverrides[task.TaskType]; ok {
		return override(ctx, task)
	}

	switch task.TaskType {
	case "stock_analysis":
		return a.executeStockAnalysis(ctx, task)
	case "market_analysis":
		return a.executeMarketAnalysis(ctx, task)
	case "global_stock_index_cache":
		return a.executeGlobalStockIndexCache(ctx, task)
	case "fund_analysis":
		return a.executeFundAnalysis(ctx, task)
	case "news_fetch":
		return a.executeNewsFetch(ctx, task)
	case "stock_monitor":
		return a.executeStockMonitor(ctx, task)
	case "stock_change_save":
		return a.executeStockChangeSave(ctx, task)
	case "custom":
		return a.executeCustomTask(ctx, task)
	default:
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(moduleTrace(ctx, "cron-task-dispatch")).Warn(
				"task.unknown_type",
				"unknown cron task type",
				logger.String("task_type", task.TaskType),
			)
		}
		return fmt.Errorf("未知任务类型：%s", task.TaskType)
	}
}
```

```go
func extractTraceID(content string, event string) string {
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		if !strings.Contains(line, `"`+"event"+`":"`+event+`"`) {
			continue
		}
		matches := regexp.MustCompile(`"trace_id":"([^"]+)"`).FindStringSubmatch(line)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}
```

- [ ] **Step 4: Re-run the focused integration tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger ./backend/agent -run 'TestHTTPTraceFlowsAcrossHTTPAndDBSinks|TestExecuteTask_TraceFlowsAcrossTaskAISinkAndDBSinks' -count=1
```

Expected: command exits `0`, `http.log` and `db.log` share one trace in the HTTP test, and `task.log`, `ai.log`, and `db.log` share one trace in the task test.

- [ ] **Step 5: Commit the closeout integration evidence**

```powershell
git add backend/logger/closeout_integration_test.go backend/agent/cron_task_api.go backend/agent/cron_task_api_test.go
git commit -m "test: prove closeout trace reuse across file sinks"
```

### Task 6: Add the manual validation checklist and run full verification

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\full-logging-rebuild\docs\superpowers\checklists\2026-04-05-full-logging-manual-validation.md`

- [ ] **Step 1: Write the manual validation document with concrete file checks**

```markdown
# 全量日志闭环手工验收

## 1. 日志目录

- Windows 桌面应用默认日志目录：`%LOCALAPPDATA%\go-stock\logs`
- 先确认存在：`app.log`、`error.log`、`http.log`、`ai.log`、`task.log`、`db.log`、`frontend.log`、`panic.log`
- 再确认存在：`payloads\YYYY-MM-DD\`

## 2. 前端错误

1. 打开任一页面，在浏览器控制台执行 `console.error(new Error('manual-frontend-probe'))`
2. 打开 `frontend.log`
3. 确认最新记录包含 `event=frontend.error`、`error_class=frontend_error`、`trace_id`

## 3. HTTP 请求

1. 发送一次本地请求：`Invoke-WebRequest http://127.0.0.1:18888/api/health`
2. 打开 `http.log`
3. 如请求或响应体超过阈值，继续检查 `payloads` 目录下是否出现新文件
4. 确认 `http.log` 中的 `trace_id` 可与同次链路的下游日志对应

## 4. AI 调用

1. 在应用里发起一次真实 AI 分析请求，例如输入 `测试日志链路`
2. 打开 `ai.log`
3. 确认存在本次请求的 `trace_id`、模型信息和关键事件

## 5. 任务失败

1. 在任务配置中执行一个 `stock_analysis` 任务，并故意将 `params` 写成非法 JSON
2. 打开 `task.log`
3. 确认最新失败记录包含 `error_class=task_error`、`error_message`、`trace_id`

## 6. 最终核对

- 同一条关键链路至少能在两个不同 sink 里看到同一个 `trace_id`
- 有大体积原文时，主日志只保留摘要字段，真实内容在 `payload_file`
- 开发者只需要查看 `logs` 和 `payloads`，不依赖额外 UI 或导出工具
```

- [ ] **Step 2: Run the project verification suite**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestNoLegacyLoggerUsageInNonTestGoFiles -count=1
conda run -n test go test ./... -count=1
node --test .\frontend\src\utils\frontendLogger.test.mjs
Push-Location frontend
npm run build
Pop-Location
```

Expected: every command exits `0`.

- [ ] **Step 3: Review only the intended closeout slice before the final commit**

Run:

```powershell
git diff -- backend/logger backend/agent backend/data app_common.go frontend/src/utils docs/superpowers/checklists/2026-04-05-full-logging-manual-validation.md docs/superpowers/plans/2026-04-05-full-logging-closeout.md
```

Expected: diff is limited to trace reuse, frontend error bridging, integration tests, and the manual verification document.

- [ ] **Step 4: Commit the final closeout docs and verification sweep**

```powershell
git add docs/superpowers/checklists/2026-04-05-full-logging-manual-validation.md docs/superpowers/plans/2026-04-05-full-logging-closeout.md
git commit -m "docs: add full logging closeout validation checklist"
```
