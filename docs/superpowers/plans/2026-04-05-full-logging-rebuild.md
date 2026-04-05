# Full Logging Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace go-stock's scattered zap/stdlog/Wails logging with a default-on structured local log system that writes `app.log`, `error.log`, `http.log`, `ai.log`, `task.log`, `db.log`, `frontend.log`, `panic.log`, and `logs/payloads` under the shared runtime root.

**Architecture:** Build a new `backend/logger` facade with typed fields, per-sink JSONL writers, payload spill files, trace helpers, and panic wrappers, then wire `main`, Wails lifecycle handlers, `ai-assistant-web`, `backend/db`, frontend `frontendError` events, and the current `agent/data` log call sites onto that facade. Keep the rollout safe by using targeted tests around logger internals first, then migrate the highest-volume packages, and finish with repository guard tests that ban legacy log exits from non-test Go code.

**Tech Stack:** Go, Zap, Lumberjack, GORM, Wails, Vue 3, Node `--test`, PowerShell

---

## File Map

### New logger infrastructure

- `D:\codex_work\go-stock\backend\logger\types.go`
  - Sink names, typed field helpers, trace structs, frontend payload structs.
- `D:\codex_work\go-stock\backend\logger\bootstrap.go`
  - Shared runtime initialization, sink registry, default config.
- `D:\codex_work\go-stock\backend\logger\factory.go`
  - Module logger factory, `With(...)`, `WithTrace(...)`, `Info/Error/Debug`.
- `D:\codex_work\go-stock\backend\logger\sinks.go`
  - JSONL encoders, lumberjack writers, payload spill storage, retention helpers.
- `D:\codex_work\go-stock\backend\logger\runtime.go`
  - Panic recovery, goroutine wrappers, trace creation, duration helpers.
- `D:\codex_work\go-stock\backend\logger\http.go`
  - HTTP middleware for request/response capture and payload spill.
- `D:\codex_work\go-stock\backend\logger\db.go`
  - Custom GORM logger adapter.
- `D:\codex_work\go-stock\backend\logger\wails.go`
  - Wails/frontend-event normalization helpers.

### Logger tests

- `D:\codex_work\go-stock\backend\logger\core_test.go`
  - Sink path, common-field, and module logger tests.
- `D:\codex_work\go-stock\backend\logger\runtime_test.go`
  - Payload spill, panic wrapper, trace propagation tests.
- `D:\codex_work\go-stock\backend\logger\http_test.go`
  - HTTP middleware tests.
- `D:\codex_work\go-stock\backend\logger\db_test.go`
  - GORM adapter tests.
- `D:\codex_work\go-stock\backend\logger\legacy_usage_test.go`
  - Repository guard test that bans `fmt.Printf`, stdlib `log`, and direct `SugaredLogger` usage in non-test Go files.

### Existing files that must move to the new logger facade

- `D:\codex_work\go-stock\backend\logger\lgo.go`
  - Temporary compatibility shim during migration, then final thin wrapper or removal of direct globals.
- `D:\codex_work\go-stock\main.go`
  - Bootstraps the logger runtime, app session ID, panic handling, Wails logger bridge.
- `D:\codex_work\go-stock\app.go`
  - Main Wails lifecycle logs, cron task lifecycle logs, update/sync logs.
- `D:\codex_work\go-stock\app_common.go`
  - Agent chat panic logging and cross-platform helpers.
- `D:\codex_work\go-stock\app_windows.go`
- `D:\codex_work\go-stock\app_linux.go`
- `D:\codex_work\go-stock\app_darwin.go`
  - Platform startup, frontend error event listener, before-close and notification logs.
- `D:\codex_work\go-stock\bootstrap_stock_search_data.go`
  - Startup seed logs.
- `D:\codex_work\go-stock\backend\db\db.go`
  - Swap `gorm logger.Silent` for the custom adapter and remove stdlib `log`.
- `D:\codex_work\go-stock\ai-assistant-web\server.go`
  - Add middleware and structured server events.
- `D:\codex_work\go-stock\ai-assistant-web\cmd\ai-assistant-web\main.go`
  - Replace `log.Fatalf` with the shared logger.
- `D:\codex_work\go-stock\backend\agent\agent.go`
- `D:\codex_work\go-stock\backend\agent\agent_api.go`
- `D:\codex_work\go-stock\backend\agent\cron_task_api.go`
- `D:\codex_work\go-stock\backend\agent\chat_memory.go`
- `D:\codex_work\go-stock\backend\agent\tools\choice_stock_by_indicators_tool.go`
- `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper.go`
- `D:\codex_work\go-stock\backend\agent\tools\market_news_tool.go`
- `D:\codex_work\go-stock\backend\agent\tools\stock_code_tool.go`
  - Agent/session/task/tool-call logging.
- `D:\codex_work\go-stock\backend\data\alert_windows_api.go`
- `D:\codex_work\go-stock\backend\data\alert_darwin_api.go`
- `D:\codex_work\go-stock\backend\data\crawler_api.go`
- `D:\codex_work\go-stock\backend\data\dingding_api.go`
- `D:\codex_work\go-stock\backend\data\eastmoney_kline_example.go`
- `D:\codex_work\go-stock\backend\data\fund_data_api.go`
- `D:\codex_work\go-stock\backend\data\eastmoney_kline_chromedp.go`
- `D:\codex_work\go-stock\backend\data\eastmoney_kline_api.go`
- `D:\codex_work\go-stock\backend\data\openai_stream.go`
- `D:\codex_work\go-stock\backend\data\market_news_api.go`
- `D:\codex_work\go-stock\backend\data\openai_tools.go`
- `D:\codex_work\go-stock\backend\data\pool.go`
- `D:\codex_work\go-stock\backend\data\prompt_template_api.go`
- `D:\codex_work\go-stock\backend\data\search_stock_api.go`
- `D:\codex_work\go-stock\backend\data\openai_crawler.go`
- `D:\codex_work\go-stock\backend\data\settings_api.go`
- `D:\codex_work\go-stock\backend\data\stock_changes_api.go`
- `D:\codex_work\go-stock\backend\data\stock_data_api_darwin.go`
- `D:\codex_work\go-stock\backend\data\stock_data_api.go`
- `D:\codex_work\go-stock\backend\data\stock_data_api_linux.go`
- `D:\codex_work\go-stock\backend\data\stock_data_api_windows.go`
- `D:\codex_work\go-stock\backend\data\stock_sentiment_analysis.go`
- `D:\codex_work\go-stock\backend\data\tool_cailianpress_opinion.go`
- `D:\codex_work\go-stock\backend\data\tool_hot_tables.go`
- `D:\codex_work\go-stock\backend\data\tool_market_data.go`
- `D:\codex_work\go-stock\backend\data\tool_mutual_top10.go`
- `D:\codex_work\go-stock\backend\data\tool_interactive_answer.go`
- `D:\codex_work\go-stock\backend\data\tool_search_bk.go`
- `D:\codex_work\go-stock\backend\data\tool_search_etf_stock.go`
- `D:\codex_work\go-stock\backend\data\tool_stock_notice.go`
- `D:\codex_work\go-stock\backend\data\tool_stock_info.go`
- `D:\codex_work\go-stock\backend\data\tushare_data_api.go`
- `D:\codex_work\go-stock\backend\data\web_search_api.go`
- `D:\codex_work\go-stock\backend\data\utils.go`
  - External request, AI, chromedp, notification, and business-operation logs.

### Frontend files for normalized error capture

- `D:\codex_work\go-stock\frontend\src\utils\frontendLogger.mjs`
  - Shared frontend error payload builder and emitter.
- `D:\codex_work\go-stock\frontend\src\utils\frontendLogger.test.mjs`
  - Node tests for payload normalization and `EventsEmit` usage.
- `D:\codex_work\go-stock\frontend\src\main.js`
  - Install one shared error handler instead of ad hoc copies.
- `D:\codex_work\go-stock\frontend\src\App.vue`
- `D:\codex_work\go-stock\frontend\src\components\settings.vue`
- `D:\codex_work\go-stock\frontend\src\components\stock.vue`
  - Replace direct `window.onerror` blocks with the shared frontend logger helper.

### Existing tests to extend

- `D:\codex_work\go-stock\backend\db\db_test.go`
- `D:\codex_work\go-stock\ai-assistant-web\server_test.go`

## Task 1: Build the core logger facade and sink registry

**Files:**
- Create: `D:\codex_work\go-stock\backend\logger\types.go`
- Create: `D:\codex_work\go-stock\backend\logger\bootstrap.go`
- Create: `D:\codex_work\go-stock\backend\logger\factory.go`
- Create: `D:\codex_work\go-stock\backend\logger\sinks.go`
- Create: `D:\codex_work\go-stock\backend\logger\core_test.go`
- Modify: `D:\codex_work\go-stock\backend\logger\lgo.go`

- [ ] **Step 1: Write failing tests for sink paths and module fields**

```go
func TestBuildSinkPaths_UsesRuntimeLogsDir(t *testing.T) {
	root := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)
	paths := apppath.Paths{RootDir: root, LogsDir: filepath.Join(root, "logs")}

	sinkPaths := buildSinkPaths(paths)

	if sinkPaths[SinkApp] != filepath.Join(root, "logs", "app.log") {
		t.Fatalf("expected app sink under runtime logs dir, got %q", sinkPaths[SinkApp])
	}
	if sinkPaths[SinkPanic] != filepath.Join(root, "logs", "panic.log") {
		t.Fatalf("expected panic sink under runtime logs dir, got %q", sinkPaths[SinkPanic])
	}
}

func TestModuleLogger_AddsModuleEventAndTraceFields(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	log := runtime.ForSink(SinkApp, "main").WithTrace(TraceContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		AppSessionID: "session-1",
		Source:       "wails",
	})

	log.Info("startup.begin", "starting app", String("version", "dev"))

	line := buf.String()
	for _, want := range []string{`"module":"main"`, `"event":"startup.begin"`, `"trace_id":"trace-1"`, `"app_session_id":"session-1"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected log line to contain %s, got %s", want, line)
		}
	}
}
```

- [ ] **Step 2: Run the logger package tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger -run "TestBuildSinkPaths_UsesRuntimeLogsDir|TestModuleLogger_AddsModuleEventAndTraceFields" -count=1
```

Expected:

- compile failure because `SinkApp`, `SinkPanic`, `buildSinkPaths`, `TraceContext`, and the new runtime factory do not exist yet

- [ ] **Step 3: Implement the sink registry, typed fields, and module logger**

```go
type Sink string

const (
	SinkApp      Sink = "app"
	SinkError    Sink = "error"
	SinkHTTP     Sink = "http"
	SinkAI       Sink = "ai"
	SinkTask     Sink = "task"
	SinkDB       Sink = "db"
	SinkFrontend Sink = "frontend"
	SinkPanic    Sink = "panic"
)

type TraceContext struct {
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	AppSessionID string `json:"app_session_id"`
	Source       string `json:"source"`
}

type Runtime struct {
	sessionID string
	sinks     map[Sink]*zap.Logger
}

type Config struct {
	Paths        apppath.Paths
	EnableStdout bool
}

type Logger struct {
	runtime *Runtime
	sink    Sink
	module  string
	trace   TraceContext
	fields  []zap.Field
}

var defaultRuntime atomic.Pointer[Runtime]

func DefaultConfig(paths apppath.Paths) Config {
	return Config{
		Paths:        paths,
		EnableStdout: true,
	}
}

func MustInit(cfg Config) *Runtime {
	runtime := &Runtime{
		sessionID: uuid.NewString(),
		sinks:     make(map[Sink]*zap.Logger),
	}
	defaultRuntime.Store(runtime)
	return runtime
}

func Default() *Runtime {
	return defaultRuntime.Load()
}

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

func (r *Runtime) ForSink(sink Sink, module string) *Logger {
	return &Logger{runtime: r, sink: sink, module: module}
}

func String(key, value string) zap.Field   { return zap.String(key, value) }
func Any(key string, value any) zap.Field  { return zap.Any(key, value) }
func Int(key string, value int) zap.Field  { return zap.Int(key, value) }
func Int64(key string, value int64) zap.Field { return zap.Int64(key, value) }
func Uint(key string, value uint) zap.Field { return zap.Uint(key, value) }
func Duration(key string, value time.Duration) zap.Field { return zap.Duration(key, value) }
func Err(err error) zap.Field { return zap.NamedError("error", err) }

func (l *Logger) WithTrace(trace TraceContext) *Logger {
	clone := *l
	clone.trace = trace
	return &clone
}

func (l *Logger) Info(event, message string, fields ...zap.Field) {
	base := l.runtime.sinks[l.sink]
	base.With(
		zap.String("module", l.module),
		zap.String("event", event),
		zap.String("trace_id", l.trace.TraceID),
		zap.String("span_id", l.trace.SpanID),
		zap.String("app_session_id", l.trace.AppSessionID),
		zap.String("source", l.trace.Source),
	).Info(message, fields...)
}

func (l *Logger) Warn(event, message string, fields ...zap.Field) {
	l.runtime.sinks[l.sink].With(
		zap.String("module", l.module),
		zap.String("event", event),
	).Warn(message, fields...)
}

func (l *Logger) Error(event, message string, fields ...zap.Field) {
	l.runtime.sinks[l.sink].With(
		zap.String("module", l.module),
		zap.String("event", event),
	).Error(message, fields...)
}

func newRuntimeForTest(writer io.Writer) *Runtime {
	cfg := DefaultConfig(apppath.Paths{LogsDir: filepath.Join(os.TempDir(), "go-stock-test-logs")})
	runtime := MustInit(cfg)
	runtime.sinks[SinkApp] = zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(writer), zapcore.DebugLevel))
	runtime.sinks[SinkPanic] = runtime.sinks[SinkApp]
	runtime.sinks[SinkHTTP] = runtime.sinks[SinkApp]
	runtime.sinks[SinkDB] = runtime.sinks[SinkApp]
	runtime.sinks[SinkFrontend] = runtime.sinks[SinkApp]
	return runtime
}
```

- [ ] **Step 4: Re-run the logger package tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger -run "TestBuildSinkPaths_UsesRuntimeLogsDir|TestModuleLogger_AddsModuleEventAndTraceFields" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 5: Commit the core facade slice**

```powershell
git add backend/logger/types.go backend/logger/bootstrap.go backend/logger/factory.go backend/logger/sinks.go backend/logger/core_test.go backend/logger/lgo.go
git commit -m "feat: add structured logger facade"
```

## Task 2: Add payload spilling, trace helpers, and panic-safe execution wrappers

**Files:**
- Create: `D:\codex_work\go-stock\backend\logger\runtime.go`
- Create: `D:\codex_work\go-stock\backend\logger\runtime_test.go`
- Modify: `D:\codex_work\go-stock\backend\logger\bootstrap.go`
- Modify: `D:\codex_work\go-stock\backend\logger\sinks.go`

- [ ] **Step 1: Write failing tests for large payload spill and panic capture**

```go
func TestPayloadStore_SpillsLargeBodiesToPayloadDir(t *testing.T) {
	store := NewPayloadStore(filepath.Join(t.TempDir(), "logs"), 32, 5<<20)
	ref, err := store.Save("http-request", bytes.Repeat([]byte("x"), 128))
	if err != nil {
		t.Fatalf("save payload: %v", err)
	}
	if ref.File == "" {
		t.Fatalf("expected payload spill file, got inline payload %#v", ref)
	}
	if _, err := os.Stat(ref.File); err != nil {
		t.Fatalf("expected spill file to exist: %v", err)
	}
}

func TestGoWithRecover_WritesPanicEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)

	runtime.GoWithRecover("panic", "logger-test", func() {
		panic("boom")
	})

	if !strings.Contains(buf.String(), `"event":"logger-test"`) || !strings.Contains(buf.String(), `"error_class":"panic"`) {
		t.Fatalf("expected panic event in panic sink, got %s", buf.String())
	}
}
```

- [ ] **Step 2: Run the new runtime-focused tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger -run "TestPayloadStore_SpillsLargeBodiesToPayloadDir|TestGoWithRecover_WritesPanicEvent" -count=1
```

Expected:

- compile failure because `PayloadStore`, `Save`, and `GoWithRecover` do not exist yet

- [ ] **Step 3: Implement payload spill files and panic-safe wrappers**

```go
type PayloadRef struct {
	File   string `json:"file,omitempty"`
	Inline string `json:"inline,omitempty"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

type PayloadStore struct {
	root        string
	inlineLimit int
	maxTotal    int64
}

type Runtime struct {
	sessionID string
	sinks     map[Sink]*zap.Logger
	payloads  *PayloadStore
}

func NewPayloadStore(root string, inlineLimit int, maxTotal int64) *PayloadStore {
	return &PayloadStore{root: root, inlineLimit: inlineLimit, maxTotal: maxTotal}
}

func (s *PayloadStore) Save(kind string, body []byte) (PayloadRef, error) {
	sum := sha256.Sum256(body)
	ref := PayloadRef{Size: len(body), SHA256: hex.EncodeToString(sum[:])}
	if len(body) <= s.inlineLimit {
		ref.Inline = string(body)
		return ref, nil
	}
	dir := filepath.Join(s.root, "payloads", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return PayloadRef{}, err
	}
	file := filepath.Join(dir, fmt.Sprintf("%s-%d.log", kind, time.Now().UnixNano()))
	if err := os.WriteFile(file, body, 0o644); err != nil {
		return PayloadRef{}, err
	}
	ref.File = file
	return ref, nil
}

func (r *Runtime) NewTrace(source string) TraceContext {
	return TraceContext{
		TraceID:      uuid.NewString(),
		SpanID:       uuid.NewString(),
		AppSessionID: r.sessionID,
		Source:       source,
	}
}

func (r *Runtime) GoWithRecover(module, event string, fn func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				r.ForSink(SinkPanic, module).Error(
					event,
					"goroutine panic recovered",
					String("error_class", "panic"),
					Any("panic_value", recovered),
					String("stack", string(debug.Stack())),
				)
			}
		}()
		fn()
	}()
}

func (r *Runtime) AttachPayloadStore(store *PayloadStore) {
	r.payloads = store
}
```

- [ ] **Step 4: Re-run the runtime-focused tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger -run "TestPayloadStore_SpillsLargeBodiesToPayloadDir|TestGoWithRecover_WritesPanicEvent" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 5: Commit the payload and runtime helpers**

```powershell
git add backend/logger/runtime.go backend/logger/runtime_test.go backend/logger/bootstrap.go backend/logger/sinks.go
git commit -m "feat: add logger payload spill and panic wrappers"
```

## Task 3: Wire `main` and Wails lifecycle handlers onto the new logger

**Files:**
- Create: `D:\codex_work\go-stock\backend\logger\wails.go`
- Modify: `D:\codex_work\go-stock\main.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Modify: `D:\codex_work\go-stock\app_windows.go`
- Modify: `D:\codex_work\go-stock\app_linux.go`
- Modify: `D:\codex_work\go-stock\app_darwin.go`
- Modify: `D:\codex_work\go-stock\bootstrap_stock_search_data.go`

- [ ] **Step 1: Write a failing test for frontend-event normalization**

```go
func TestNormalizeFrontendError_PreservesRouteAndStack(t *testing.T) {
	payload := NormalizeFrontendError([]interface{}{
		map[string]interface{}{
			"page":    "stock.vue",
			"route":   "/stock",
			"message": "ResizeObserver loop limit exceeded",
			"error":   "stack-line-1",
		},
	})

	if payload.Page != "stock.vue" || payload.Route != "/stock" {
		t.Fatalf("expected normalized frontend payload, got %#v", payload)
	}
	if payload.Stack != "stack-line-1" {
		t.Fatalf("expected stack to be preserved, got %#v", payload)
	}
}
```

- [ ] **Step 2: Run the logger package test and confirm it fails**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestNormalizeFrontendError_PreservesRouteAndStack -count=1
```

Expected:

- compile failure because `NormalizeFrontendError` does not exist yet

- [ ] **Step 3: Add Wails helpers and swap startup hooks to module loggers**

```go
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
	if len(optionalData) == 0 {
		return FrontendErrorPayload{}
	}
	raw, _ := json.Marshal(optionalData[0])
	var payload FrontendErrorPayload
	_ = json.Unmarshal(raw, &payload)
	return payload
}

// main.go
paths, err := apppath.Ensure()
if err != nil {
	fmt.Fprintf(os.Stderr, "初始化运行时目录失败: %v\n", err)
	os.Exit(1)
}
runtimeLog := log.MustInit(log.DefaultConfig(paths))
runtimeLog.AttachPayloadStore(log.NewPayloadStore(paths.LogsDir, 32*1024, 5<<30))
appLog := runtimeLog.ForSink(log.SinkApp, "main")
trace := runtimeLog.NewTrace("bootstrap")
appLog.WithTrace(trace).Info("startup.begin", "starting app", log.String("version", Version), log.String("commit", VersionCommit))

// app_windows.go / app_linux.go / app_darwin.go
runtime.EventsOn(ctx, "frontendError", func(optionalData ...interface{}) {
	payload := logger.NormalizeFrontendError(optionalData)
	logger.Default().ForSink(logger.SinkFrontend, "frontend").WithTrace(logger.Default().NewTrace("wails-frontend")).Error(
		"frontend.error",
		"frontend runtime error",
		logger.String("page", payload.Page),
		logger.String("route", payload.Route),
		logger.String("error_message", payload.Message),
		logger.String("stack", payload.Stack),
	)
})
```

- [ ] **Step 4: Re-run the logger tests and the existing app-focused Go tests**

Run:

```powershell
conda run -n test go test ./backend/logger ./... -run "TestNormalizeFrontendError_PreservesRouteAndStack|TestSeedBundledStockSearchDataPopulatesEmptyRuntimeTables" -count=1
```

Expected:

- the logger test passes
- the targeted repository tests still compile with the new logger bootstrap in place

- [ ] **Step 5: Commit the Wails and startup wiring**

```powershell
git add backend/logger/wails.go main.go app.go app_common.go app_windows.go app_linux.go app_darwin.go bootstrap_stock_search_data.go
git commit -m "feat: wire app lifecycle into structured logger"
```

## Task 4: Normalize frontend error emission and remove duplicated `window.onerror` blocks

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\utils\frontendLogger.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\utils\frontendLogger.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\main.js`
- Modify: `D:\codex_work\go-stock\frontend\src\App.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\stock.vue`

- [ ] **Step 1: Write a failing Node test for payload normalization**

```js
import test from 'node:test'
import assert from 'node:assert/strict'

import { buildFrontendErrorPayload } from './frontendLogger.mjs'

test('buildFrontendErrorPayload normalizes page route and stack', () => {
  const payload = buildFrontendErrorPayload({
    page: 'stock.vue',
    route: '/stock',
    message: 'boom',
    error: new Error('boom'),
    extra: { panel: 'watchlist' },
  })

  assert.equal(payload.page, 'stock.vue')
  assert.equal(payload.route, '/stock')
  assert.match(payload.error, /boom/)
  assert.deepEqual(payload.extra, { panel: 'watchlist' })
})
```

- [ ] **Step 2: Run the new frontend logger test and confirm it fails**

Run:

```powershell
node --test .\frontend\src\utils\frontendLogger.test.mjs
```

Expected:

- module resolution or import failure because `frontendLogger.mjs` does not exist yet

- [ ] **Step 3: Implement the shared emitter and replace ad hoc handlers**

```js
import { EventsEmit } from '../../wailsjs/runtime/runtime'

export function buildFrontendErrorPayload({ page, route, message, source, lineno, colno, error, extra = {} }) {
  return {
    page,
    route: route || window.location.hash || window.location.pathname,
    message: typeof message === 'string' ? message : String(message),
    source: source || '',
    lineno: Number.isFinite(lineno) ? lineno : 0,
    colno: Number.isFinite(colno) ? colno : 0,
    error: error?.stack || error?.message || null,
    extra,
  }
}

export function emitFrontendError(input) {
  EventsEmit('frontendError', buildFrontendErrorPayload(input))
}

export function installFrontendErrorHandlers(page) {
  window.onerror = function (message, source, lineno, colno, error) {
    emitFrontendError({ page, message, source, lineno, colno, error })
    return true
  }

  window.addEventListener('unhandledrejection', (event) => {
    emitFrontendError({
      page,
      message: event.reason?.message || 'unhandledrejection',
      error: event.reason instanceof Error ? event.reason : new Error(String(event.reason)),
    })
    event.preventDefault()
  })
}
```

- [ ] **Step 4: Re-run the frontend logger test and verify the frontend still builds**

Run:

```powershell
node --test .\frontend\src\utils\frontendLogger.test.mjs
Set-Location .\frontend
npm run build
```

Expected:

- node test exits `0`
- Vite production build exits `0`

- [ ] **Step 5: Commit the frontend error bridge**

```powershell
git add frontend/src/utils/frontendLogger.mjs frontend/src/utils/frontendLogger.test.mjs frontend/src/main.js frontend/src/App.vue frontend/src/components/settings.vue frontend/src/components/stock.vue
git commit -m "feat: normalize frontend error logging"
```

## Task 5: Add HTTP middleware and a custom GORM logger

**Files:**
- Create: `D:\codex_work\go-stock\backend\logger\http.go`
- Create: `D:\codex_work\go-stock\backend\logger\http_test.go`
- Create: `D:\codex_work\go-stock\backend\logger\db.go`
- Create: `D:\codex_work\go-stock\backend\logger\db_test.go`
- Modify: `D:\codex_work\go-stock\backend\db\db.go`
- Modify: `D:\codex_work\go-stock\ai-assistant-web\server.go`
- Modify: `D:\codex_work\go-stock\ai-assistant-web\server_test.go`
- Modify: `D:\codex_work\go-stock\backend\db\db_test.go`

- [ ] **Step 1: Write failing tests for HTTP logging and DB logging**

```go
func TestHTTPMiddleware_WritesRequestResponseAndPayloadRef(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	runtime.AttachPayloadStore(NewPayloadStore(t.TempDir(), 32, 5<<20))
	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 128)))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/chat/summary-stream", strings.NewReader(strings.Repeat("q", 128)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	line := buf.String()
	for _, want := range []string{`"sink":"http"`, `"path":"/api/chat/summary-stream"`, `"payload_file"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected middleware log to contain %s, got %s", want, line)
		}
	}
}

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
}
```

- [ ] **Step 2: Run the HTTP/DB logger tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/logger ./backend/db ./ai-assistant-web -run "TestHTTPMiddleware_WritesRequestResponseAndPayloadRef|TestGormLogger_LogsErrorAndSlowStatements" -count=1
```

Expected:

- compile failure because the middleware and custom GORM logger do not exist yet

- [ ] **Step 3: Implement the middleware and GORM adapter, then wire them into the server and DB bootstrap**

```go
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *Runtime) HTTPMiddleware(module string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		reqBody, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))

		next.ServeHTTP(rec, req)

		reqRef, _ := r.payloads.Save("http-request", reqBody)
		resRef, _ := r.payloads.Save("http-response", rec.body.Bytes())
		r.ForSink(SinkHTTP, module).WithTrace(r.NewTrace("http")).Info(
			"http.request.completed",
			"handled http request",
			String("method", req.Method),
			String("path", req.URL.Path),
			Int("status_code", rec.status),
			Duration("duration_ms", time.Since(start)),
			Any("request_payload", reqRef),
			Any("response_payload", resRef),
		)
	})
}

type GormLogger struct {
	runtime       *Runtime
	slowThreshold time.Duration
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)
	fields := []zap.Field{String("sql", sql), Int64("rows_affected", rows), Duration("duration_ms", elapsed)}
	if err != nil {
		l.runtime.ForSink(SinkDB, "gorm").Error("db.query.failed", "gorm query failed", append(fields, String("error_message", err.Error()))...)
		return
	}
	if elapsed >= l.slowThreshold {
		l.runtime.ForSink(SinkDB, "gorm").Warn("db.query.slow", "gorm query exceeded threshold", fields...)
	}
}

// backend/db/db.go
dbLogger := logger.NewGormLogger(logger.Default(), 3*time.Second)
openDb, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
	Logger:                                   dbLogger,
	DisableForeignKeyConstraintWhenMigrating: true,
	SkipDefaultTransaction:                   true,
	PrepareStmt:                              true,
})

// ai-assistant-web/server.go
runtimeLogger := logger.Default()
return http.ListenAndServe(addr, runtimeLogger.HTTPMiddleware("ai-assistant-web", withCORS(mux)))
```

- [ ] **Step 4: Re-run the HTTP/DB tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/logger ./backend/db ./ai-assistant-web -run "TestHTTPMiddleware_WritesRequestResponseAndPayloadRef|TestGormLogger_LogsErrorAndSlowStatements|TestVipStatus_ReturnsOpenAccess" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 5: Commit the server and DB logging adapters**

```powershell
git add backend/logger/http.go backend/logger/http_test.go backend/logger/db.go backend/logger/db_test.go backend/db/db.go backend/db/db_test.go ai-assistant-web/server.go ai-assistant-web/server_test.go
git commit -m "feat: add http and db structured logging"
```

## Task 6: Migrate `main`, `agent`, and tool-call paths away from `logger.SugaredLogger`

**Files:**
- Modify: `D:\codex_work\go-stock\main.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Modify: `D:\codex_work\go-stock\app_windows.go`
- Modify: `D:\codex_work\go-stock\app_linux.go`
- Modify: `D:\codex_work\go-stock\app_darwin.go`
- Modify: `D:\codex_work\go-stock\bootstrap_stock_search_data.go`
- Modify: `D:\codex_work\go-stock\backend\agent\agent.go`
- Modify: `D:\codex_work\go-stock\backend\agent\agent_api.go`
- Modify: `D:\codex_work\go-stock\backend\agent\cron_task_api.go`
- Modify: `D:\codex_work\go-stock\backend\agent\chat_memory.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\choice_stock_by_indicators_tool.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\market_news_tool.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\stock_code_tool.go`
- Modify: `D:\codex_work\go-stock\ai-assistant-web\cmd\ai-assistant-web\main.go`
- Create: `D:\codex_work\go-stock\backend\logger\legacy_usage_test.go`

- [ ] **Step 1: Write the failing repository guard test for legacy logger exits**

```go
func TestNoLegacyLoggerUsageInNonTestGoFiles(t *testing.T) {
	for _, file := range scanRepoGoFiles(t) {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(data)
		for _, forbidden := range []string{
			"logger.SugaredLogger",
			"fmt.Printf(",
			"log.Fatalf(",
			"log.Printf(",
			"log.Println(",
			"log.Print(",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("found forbidden logger usage %q in %s", forbidden, file)
			}
		}
	}
}

func scanRepoGoFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", ".."))
	files := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "build" || base == "frontend" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repo files: %v", err)
	}
	return files
}
```

- [ ] **Step 2: Run the guard test and confirm it fails on the current call sites**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestNoLegacyLoggerUsageInNonTestGoFiles -count=1
```

Expected:

- FAIL listing files such as `main.go`, `app.go`, `backend/agent/agent_api.go`, and `ai-assistant-web/cmd/ai-assistant-web/main.go`

- [ ] **Step 3: Replace the app and agent call sites with module loggers**

```go
var (
	mainLog  = logger.Default().ForSink(logger.SinkApp, "main")
	agentLog = logger.Default().ForSink(logger.SinkAI, "agent")
	taskLog  = logger.Default().ForSink(logger.SinkTask, "cron-task")
)

mainLog.WithTrace(trace).Info("startup.begin", "starting app", logger.String("version", Version))
mainLog.Error("startup.seed.failed", "failed to initialize bundled stock search data", logger.Err(err))

agentLog.Error("agent.chat.failed", "stream error", logger.String("ai_session_id", sessionID), logger.Err(err))
taskLog.Info("task.execute.begin", "starting cron task", logger.Uint("task_id", task.ID), logger.String("task_name", task.Name))
taskLog.Error("task.execute.failed", "cron task failed", logger.Uint("task_id", task.ID), logger.Err(err))
```

- [ ] **Step 4: Re-run the guard test and focused Go tests for app and agent packages**

Run:

```powershell
conda run -n test go test ./backend/logger ./backend/agent ./... -run "TestNoLegacyLoggerUsageInNonTestGoFiles|TestBuildLogPaths_UsesSharedRuntimeRoot|TestGetAllTools" -count=1
```

Expected:

- the guard test passes for the migrated files
- targeted app/agent tests remain green

- [ ] **Step 5: Commit the app and agent migration**

```powershell
git add main.go app.go app_common.go app_windows.go app_linux.go app_darwin.go bootstrap_stock_search_data.go backend/agent/agent.go backend/agent/agent_api.go backend/agent/cron_task_api.go backend/agent/chat_memory.go backend/agent/tools/choice_stock_by_indicators_tool.go backend/agent/tools/data_tools_wrapper.go backend/agent/tools/market_news_tool.go backend/agent/tools/stock_code_tool.go ai-assistant-web/cmd/ai-assistant-web/main.go backend/logger/legacy_usage_test.go
git commit -m "refactor: migrate app and agent logging"
```

## Task 7: Sweep the `backend/data` package and finish full verification

**Files:**
- Modify: `D:\codex_work\go-stock\backend\data\alert_windows_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\alert_darwin_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\crawler_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\dingding_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_example.go`
- Modify: `D:\codex_work\go-stock\backend\data\fund_data_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_chromedp.go`
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\openai_stream.go`
- Modify: `D:\codex_work\go-stock\backend\data\market_news_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\openai_tools.go`
- Modify: `D:\codex_work\go-stock\backend\data\pool.go`
- Modify: `D:\codex_work\go-stock\backend\data\prompt_template_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\search_stock_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\openai_crawler.go`
- Modify: `D:\codex_work\go-stock\backend\data\settings_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_changes_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api_darwin.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api_linux.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_data_api_windows.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_sentiment_analysis.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_cailianpress_opinion.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_hot_tables.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_market_data.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_mutual_top10.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_interactive_answer.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_search_bk.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_search_etf_stock.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_stock_notice.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_stock_info.go`
- Modify: `D:\codex_work\go-stock\backend\data\tushare_data_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\web_search_api.go`
- Modify: `D:\codex_work\go-stock\backend\data\utils.go`

- [ ] **Step 1: Replace the remaining `backend/data` logger calls with sink-specific module loggers**

```go
var (
	dataLog = logger.Default().ForSink(logger.SinkApp, "data")
	httpLog = logger.Default().ForSink(logger.SinkHTTP, "data-http")
	aiLog   = logger.Default().ForSink(logger.SinkAI, "openai")
)

httpLog.Info("eastmoney.request.begin", "fetching eastmoney data", logger.String("url", url), logger.String("stock_code", stockCode))
httpLog.Error("eastmoney.request.failed", "eastmoney request failed", logger.String("url", url), logger.Err(err))
aiLog.Info("openai.stream.begin", "starting ai stream", logger.String("model", o.Model))
aiLog.Warn("openai.tool.skipped", "skip tool call with empty function name")
dataLog.Error("settings.save.failed", "failed to save settings", logger.Err(err))
```

- [ ] **Step 2: Run the guard test again and confirm no forbidden usage remains**

Run:

```powershell
conda run -n test go test ./backend/logger -run TestNoLegacyLoggerUsageInNonTestGoFiles -count=1
```

Expected:

- command exits `0`

- [ ] **Step 3: Run the full automated verification set**

Run:

```powershell
conda run -n test go test ./... -count=1
node --test .\frontend\src\utils\frontendLogger.test.mjs .\frontend\src\utils\stockCode.test.mjs .\frontend\src\utils\aiConfig.test.mjs .\frontend\src\components\stock-lightweight-kline\hoverTooltip.test.mjs
Set-Location .\frontend
npm run build
Set-Location ..
go build .
```

Expected:

- Go test exits `0`
- all Node tests exit `0`
- frontend Vite build exits `0`
- `go build .` exits `0`

- [ ] **Step 4: Perform the runtime log-file smoke test**

Run:

```powershell
Start-Process .\go-stock.exe
```

Manual verification checklist:

- `%LOCALAPPDATA%\go-stock\logs\app.log` exists and contains `startup.begin`
- `%LOCALAPPDATA%\go-stock\logs\frontend.log` receives a `frontend.error` record after triggering a browser-side exception
- `%LOCALAPPDATA%\go-stock\logs\http.log` receives `http.request.completed` after using `ai-assistant-web`
- `%LOCALAPPDATA%\go-stock\logs\db.log` contains at least one DB statement record
- `%LOCALAPPDATA%\go-stock\logs\panic.log` is created when forcing a panic through a temporary local test hook
- `%LOCALAPPDATA%\go-stock\logs\payloads\YYYY-MM-DD\` exists after a large AI or HTTP payload

- [ ] **Step 5: Commit the data-package sweep and final verification**

```powershell
git add backend/data backend/logger/legacy_usage_test.go
git commit -m "refactor: complete full logging rebuild"
```

## Self-Review

### Spec coverage

- Logger facade, sink split, payload spill, trace fields, and panic wrappers are covered by Tasks 1 and 2.
- `main`, Wails lifecycle, frontend-event listeners, and startup/shutdown logging are covered by Task 3.
- Frontend shared error emission is covered by Task 4.
- HTTP middleware and GORM logging are covered by Task 5.
- Agent/task/tool-call logging migration is covered by Task 6.
- Remaining `backend/data` call sites, guard tests, and runtime verification are covered by Task 7.

No spec section is currently uncovered.

### Placeholder scan

- No `TODO`, `TBD`, or “implement later” markers remain.
- Every task includes exact files, commands, and expected outcomes.
- All migration sweeps list the exact file paths captured from the current repository scan.

### Type consistency

- `TraceContext`, `PayloadStore`, `PayloadRef`, `FrontendErrorPayload`, `Runtime`, and `Logger` names are used consistently across Tasks 1 through 7.
- Sink names remain `SinkApp`, `SinkError`, `SinkHTTP`, `SinkAI`, `SinkTask`, `SinkDB`, `SinkFrontend`, and `SinkPanic` throughout the plan.
- Frontend payload field names remain `page`, `route`, `message`, `source`, `lineno`, `colno`, `error`, and `extra` throughout the plan.
