# go-stock Analysis Chain Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the bounded AI analysis chain onto the new `source / service / pages / services` structure without changing the Go/Wails/Vue stack or touching the AI assistant and cron subsystems.

**Architecture:** Keep the existing Wails method names for AI analysis so the bridge contract stays stable, but move their implementation behind a new `backend/service/analysis` boundary backed by `backend/source/analysis`. On the frontend, add an `analysisService.mjs` layer and route the stock and research screens through `pages`, so stock analysis, market summary, saved reports, and prompt-template management stop importing Wails bindings directly inside the feature components.

**Tech Stack:** Go 1.26, Wails v2, Vue 3, Naive UI, Node `--test`, Vite, `go test`

---

## Re-Alignment Summary

This plan is intentionally named after the current code boundary, not mechanically as “phase two” or “phase three”.

- Done after phase one:
  - `backend/source/marketnews` exists and is carrying real market-read code.
  - `backend/service/market` exists and is already the bridge target for a real read-only slice.
  - `frontend/src/pages/market-page.vue` and `frontend/src/services/marketService.mjs` proved the page/service pattern end-to-end.
- Partially done:
  - Page unification only covers `/market`; `/` and `/research` still point straight at heavy components.
  - `app.go` still holds large AI-analysis orchestration methods and still mixes bridge logic with analysis workflow code.
  - `app_common.go` still delegates AI result history and prompt-template queries straight to `backend/data`.
- Not started:
  - The bounded AI analysis chain from request -> result persistence -> result history -> prompt-template selection is still on the old path.
  - `frontend/src/services` has no reusable wrapper for analysis/history/prompt-template calls.
  - `backend/source` and `backend/service` have no analysis-specific package.
- Newly exposed next boundary from the current code:
  - `frontend/src/components/market.vue` now already splits read-only market data from AI summary behavior, so the remaining AI summary logic is a clean sub-slice.
  - `frontend/src/components/stock.vue`, `frontend/src/components/researchReport.vue`, and `frontend/src/components/promptTemplateList.vue` all still import Wails bindings directly for the same analysis-domain concerns.
  - `app.go` and `app_common.go` together already reveal a coherent analysis cluster:
    - `NewChatStream`
    - `SummaryStockNews`
    - `SaveAIResponseResult`
    - `GetAIResponseResult`
    - `GetAIResponseResultList`
    - `GetPromptTemplates`
    - `GetPromptTemplateList`
    - `AddPromptTemplate` / `UpdatePromptTemplate` / `DeletePromptTemplate`

## Slice Choice

The next most rational slice is the bounded analysis chain:

- stock AI analysis request and latest-result readback
- market news AI summary request and latest-result readback
- saved analysis report history
- prompt-template list and CRUD
- page wrappers for `/` and `/research`

This slice deliberately excludes:

- `FloatingAiAssistant.vue`
- `FloatingAgentAssistant.vue`
- `backend/agent` multi-turn chat flow
- cron-task AI execution
- settings-page prompt modal cleanup
- share/export helper rewrites in `app.go`

Those paths either belong to later task/settings work or would make this slice too wide to verify cleanly.

## File Map

- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\provider.go`
  - Thin adapter over the existing `backend/data` AI stream APIs.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\store.go`
  - Thin adapter over the existing AI result store and prompt-template store APIs.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\contracts.go`
  - Typed stock-analysis and market-summary request contracts plus stream chunk contract.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service.go`
  - Analysis service that owns stream translation, history parsing, result access, and prompt-template access.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service_test.go`
  - Unit tests for stream translation, history parsing, and result/prompt delegation.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_analysis_test.go`
  - App/AppCommon bridge tests that prove the non-stream analysis methods delegate to the analysis service.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go`
  - Wire the analysis service into `App`, move stock-analysis and market-summary stream orchestration behind the service, and delegate save/load prompt methods.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_common.go`
  - Delegate AI result history and prompt-template paging CRUD through the analysis service.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.mjs`
  - Frontend wrapper for AI analysis, report history, prompt-template list/CRUD, and related normalization helpers.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.test.mjs`
  - Node tests for the frontend analysis normalization helpers.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\stock-page.vue`
  - Page wrapper for the existing stock screen.
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\research-page.vue`
  - Page wrapper for the existing research screen.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\router\router.js`
  - Route `/` and `/research` through the new page wrappers.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\stock.vue`
  - Move AI-analysis-specific Wails calls into `analysisService.mjs`.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\market.vue`
  - Move AI-summary-specific Wails calls into `analysisService.mjs`.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\researchReport.vue`
  - Load saved analysis history through `analysisService.mjs`.
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\promptTemplateList.vue`
  - Load and mutate prompt templates through `analysisService.mjs`.

## Scope Note

This plan intentionally keeps the existing Wails method names for the AI-analysis surface:

- `NewChatStream`
- `SummaryStockNews`
- `SaveAIResponseResult`
- `GetAIResponseResult`
- `GetAIResponseResultList`
- `GetPromptTemplates`
- `GetPromptTemplateList`
- `AddPromptTemplate`
- `UpdatePromptTemplate`
- `DeletePromptTemplate`

That choice avoids unnecessary binding churn, keeps the slice focused, and still moves the implementation behind the new backend and frontend boundaries.

### Task 1: Add The Analysis Service Boundary And Lock The Contracts With Tests

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\contracts.go`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service.go`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service_test.go`

Task 1 must preserve the existing bridge semantics while extracting the service boundary:

- keep the market-summary history field aligned with the existing binding name `historyJSON`
- preserve already-consumed stream fields such as `reasoning_content` and passthrough `tool_calls`
- keep invalid history on the same fallback path as today, but make the drop observable instead of silently swallowing it
- fail fast when the analysis service is constructed without required dependencies so later App wiring cannot silently no-op

- [ ] **Step 1: Write the failing analysis service tests**

```go
package analysis

import (
	"context"
	"errors"
	"testing"

	"go-stock/backend/models"
)

type fakeStreams struct {
	stockRequests  []StockRequest
	marketRequests []MarketSummaryRequest
	marketHistory  []map[string]interface{}
	stockEvents    []map[string]any
	marketEvents   []map[string]any
}

func (f *fakeStreams) StockStream(ctx context.Context, request StockRequest) <-chan map[string]any {
	f.stockRequests = append(f.stockRequests, request)
	return toEventChannel(f.stockEvents)
}

func (f *fakeStreams) MarketSummaryStream(ctx context.Context, request MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any {
	f.marketRequests = append(f.marketRequests, request)
	f.marketHistory = append([]map[string]interface{}(nil), history...)
	return toEventChannel(f.marketEvents)
}

type fakeResults struct {
	latest  *models.AIResponseResult
	page    *models.AIResponseResultPageData
	pageErr error

	savedStockCode string
	savedStockName string
	savedResult    string
	savedChatID    string
	savedQuestion  string
	savedAIConfig  int

	deletedID uint
	batched   []uint
}

func (f *fakeResults) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.savedStockCode = stockCode
	f.savedStockName = stockName
	f.savedResult = result
	f.savedChatID = chatID
	f.savedQuestion = question
	f.savedAIConfig = aiConfigID
}

func (f *fakeResults) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return f.latest
}

func (f *fakeResults) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	if f.pageErr != nil {
		return nil, f.pageErr
	}
	return f.page, nil
}

func (f *fakeResults) DeleteResult(ctx context.Context, id uint) error {
	f.deletedID = id
	return nil
}

func (f *fakeResults) BatchDeleteResults(ctx context.Context, ids []uint) error {
	f.batched = append([]uint(nil), ids...)
	return nil
}

type fakePrompts struct {
	templates     *[]models.PromptTemplate
	page          *models.PromptTemplatePageData
	savedTemplate models.PromptTemplate
	deletedID     uint
	saveResult    string
	deleteResult  string
	pageErr       error
}

func (f *fakePrompts) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.templates
}

func (f *fakePrompts) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	if f.pageErr != nil {
		return nil, f.pageErr
	}
	return f.page, nil
}

func (f *fakePrompts) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTemplate = template
	return f.saveResult
}

func (f *fakePrompts) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedID = id
	return f.deleteResult
}

func TestService_ReadAndPromptMethodsDelegateToStores(t *testing.T) {
	templates := []models.PromptTemplate{
		{ID: 7, Name: "系统模板", Type: "模型系统Prompt", Content: "请先分析风险"},
	}

	results := &fakeResults{
		latest: &models.AIResponseResult{
			StockCode: "000001.SZ",
			StockName: "平安银行",
			Question:  "怎么看银行股",
			Content:   "分析完成",
		},
		page: &models.AIResponseResultPageData{
			List:       []models.AIResponseResult{{StockCode: "000001.SZ", Content: "page-item"}},
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	prompts := &fakePrompts{
		templates:    &templates,
		page:         &models.PromptTemplatePageData{List: templates, Total: 1, Page: 1, PageSize: 10, TotalPages: 1},
		saveResult:   "模板已保存",
		deleteResult: "删除成功",
	}
	svc := NewService(&fakeStreams{}, results, prompts)

	latest := svc.GetLatestResult(context.Background(), "000001.SZ")
	if latest.StockCode != "000001.SZ" || latest.Content != "分析完成" {
		t.Fatalf("unexpected latest result: %#v", latest)
	}

	page, err := svc.GetResultPage(context.Background(), models.AIResponseResultQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 {
		t.Fatalf("unexpected result page: %#v", page)
	}

	gotTemplates := svc.GetPromptTemplates(context.Background(), "", "")
	if len(*gotTemplates) != 1 || (*gotTemplates)[0].Name != "系统模板" {
		t.Fatalf("unexpected prompt templates: %#v", gotTemplates)
	}

	promptPage, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected prompt page error: %v", err)
	}
	if promptPage.Total != 1 || len(promptPage.List) != 1 {
		t.Fatalf("unexpected prompt page: %#v", promptPage)
	}

	svc.SaveResult(context.Background(), "000001.SZ", "平安银行", "新的分析", "chat-1", "更新一下", 12)
	if results.savedStockCode != "000001.SZ" || results.savedAIConfig != 12 {
		t.Fatalf("unexpected save args: %#v", results)
	}

	if msg := svc.SavePromptTemplate(context.Background(), models.PromptTemplate{Name: "用户模板", Type: "模型用户Prompt", Content: "请总结"}); msg != "模板已保存" {
		t.Fatalf("unexpected save prompt message: %q", msg)
	}

	if err := svc.DeleteResult(context.Background(), 9); err != nil || results.deletedID != 9 {
		t.Fatalf("unexpected delete result state: err=%v id=%d", err, results.deletedID)
	}

	if err := svc.BatchDeleteResults(context.Background(), []uint{1, 2}); err != nil || len(results.batched) != 2 {
		t.Fatalf("unexpected batch delete state: err=%v ids=%v", err, results.batched)
	}

	if msg := svc.DeletePromptTemplate(context.Background(), 7); msg != "删除成功" || prompts.deletedID != 7 {
		t.Fatalf("unexpected prompt delete state: msg=%q id=%d", msg, prompts.deletedID)
	}
}

func TestService_StartStreamsDelegateAndParseHistory(t *testing.T) {
	streams := &fakeStreams{
		stockEvents: []map[string]any{
			{
				"chatId":            "stock-1",
				"question":          "怎么看",
				"content":           "第一段",
				"reasoning_content": "stock reasoning",
				"tool_calls": []map[string]any{
					{"id": "call-1", "type": "function"},
				},
			},
		},
		marketEvents: []map[string]any{
			{"chatId": "summary-1", "question": "总结市场", "content": "市场第一段"},
			{
				"extraContent": "市场第二段",
				"model":        "deepseek-chat",
				"time":         "2026-04-06 10:00:00",
				"tool_calls": []map[string]any{
					{"id": "call-2", "function": map[string]any{"name": "tool"}},
				},
			},
		},
	}
	svc := NewService(streams, &fakeResults{}, &fakePrompts{})

	stockChunks := collectChunks(svc.StartStockAnalysis(context.Background(), StockRequest{
		StockName:   "平安银行",
		StockCode:   "000001.SZ",
		Question:    "怎么看",
		AIConfigID:  3,
		EnableTools: true,
		Think:       true,
	}))
	if len(stockChunks) != 1 || stockChunks[0].ChatID != "stock-1" || stockChunks[0].Content != "第一段" {
		t.Fatalf("unexpected stock chunks: %#v", stockChunks)
	}
	if len(streams.stockRequests) != 1 || streams.stockRequests[0].StockCode != "000001.SZ" {
		t.Fatalf("unexpected stock request capture: %#v", streams.stockRequests)
	}
	if !streams.stockRequests[0].EnableTools || !streams.stockRequests[0].Think {
		t.Fatalf("enable/think flags not preserved: %#v", streams.stockRequests[0])
	}
	if streams.stockRequests[0].AIConfigID != 3 || streams.stockRequests[0].Question != "怎么看" {
		t.Fatalf("unexpected stock request details: %#v", streams.stockRequests[0])
	}
	if stockChunks[0].ReasoningContent != "stock reasoning" {
		t.Fatalf("reasoning_content missing: %#v", stockChunks[0])
	}
	if len(stockChunks[0].ToolCalls) != 1 || stockChunks[0].ToolCalls[0]["id"] != "call-1" {
		t.Fatalf("tool_calls missing: %#v", stockChunks[0])
	}

	marketChunks := collectChunks(svc.StartMarketSummary(context.Background(), MarketSummaryRequest{
		Question:    "总结市场",
		AIConfigID:  8,
		EnableTools: true,
		Think:       false,
		HistoryJSON: `[{"role":"assistant","content":"旧回答","reasoning":"旧推理"}]`,
	}))
	if len(marketChunks) != 2 || marketChunks[1].ExtraContent != "市场第二段" {
		t.Fatalf("unexpected market chunks: %#v", marketChunks)
	}
	if len(marketChunks[1].ToolCalls) != 1 || marketChunks[1].ToolCalls[0]["id"] != "call-2" {
		t.Fatalf("market tool calls missing: %#v", marketChunks[1])
	}
	if len(streams.marketHistory) != 1 {
		t.Fatalf("unexpected market history: %#v", streams.marketHistory)
	}
	if got := streams.marketHistory[0]["reasoning_content"]; got != "旧推理" {
		t.Fatalf("unexpected reasoning_content: %#v", got)
	}
}

func TestService_GetResultPagePropagatesError(t *testing.T) {
	expected := errors.New("boom")
	svc := NewService(&fakeStreams{}, &fakeResults{pageErr: expected}, &fakePrompts{})
	if _, err := svc.GetResultPage(context.Background(), models.AIResponseResultQuery{}); !errors.Is(err, expected) {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestService_GetPromptTemplatePagePropagatesError(t *testing.T) {
	expected := errors.New("boom")
	svc := NewService(&fakeStreams{}, &fakeResults{}, &fakePrompts{pageErr: expected})
	if _, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{}); !errors.Is(err, expected) {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestService_GetPromptTemplatesHandlesNilSource(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{}, &fakePrompts{templates: nil})
	got := svc.GetPromptTemplates(context.Background(), "", "")
	if got == nil {
		t.Fatalf("expected non-nil template slice")
	}
	if len(*got) != 0 {
		t.Fatalf("expected empty slice when source returns nil, got %#v", got)
	}
}

func TestService_StartMarketSummaryInvalidHistory(t *testing.T) {
	streams := &fakeStreams{
		marketEvents: []map[string]any{
			{"chatId": "summary-1", "content": "段落"},
		},
	}
	svc := NewService(streams, &fakeResults{}, &fakePrompts{})
	chunks := collectChunks(svc.StartMarketSummary(context.Background(), MarketSummaryRequest{
		Question:    "总结市场",
		AIConfigID:  1,
		EnableTools: false,
		Think:       false,
		HistoryJSON: "{bad",
	}))
	if len(chunks) != 1 {
		t.Fatalf("expected chunk despite invalid history, got %#v", chunks)
	}
	if len(streams.marketHistory) != 0 {
		t.Fatalf("history should not be replayed on parse failure: %#v", streams.marketHistory)
	}
}

func TestMapChunksHandlesNil(t *testing.T) {
	if got := collectChunks(mapChunks(nil)); len(got) != 0 {
		t.Fatalf("expected empty slice for nil stream, got %#v", got)
	}
}

func TestParseHistoryTreatsWhitespaceAsEmpty(t *testing.T) {
	history, err := parseHistory("   \n\t")
	if err != nil {
		t.Fatalf("expected no error for blank history, got %v", err)
	}
	if history != nil {
		t.Fatalf("expected nil history, got %#v", history)
	}
}

func TestNewService_requiresDependencies(t *testing.T) {
	validStreams := &fakeStreams{}
	validResults := &fakeResults{}
	validPrompts := &fakePrompts{}

	cases := []struct {
		name    string
		streams StreamSource
		results ResultSource
		prompts PromptSource
	}{
		{name: "missingStreams", streams: nil, results: validResults, prompts: validPrompts},
		{name: "missingResults", streams: validStreams, results: nil, prompts: validPrompts},
		{name: "missingPrompts", streams: validStreams, results: validResults, prompts: nil},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for %s", tc.name)
				}
			}()
			NewService(tc.streams, tc.results, tc.prompts)
		})
	}
}

func TestService_GetResultPageHandlesNilPage(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{page: nil}, &fakePrompts{})
	page, err := svc.GetResultPage(context.Background(), models.AIResponseResultQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatalf("expected non-nil page result")
	}
	if page.Total != 0 || len(page.List) != 0 {
		t.Fatalf("expected empty page when source returns nil, got %#v", page)
	}
}

func TestService_GetPromptTemplatePageHandlesNilPage(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{}, &fakePrompts{page: nil})
	page, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatalf("expected non-nil prompt page result")
	}
	if page.Total != 0 || len(page.List) != 0 {
		t.Fatalf("expected empty prompt page when source returns nil, got %#v", page)
	}
}

func toEventChannel(items []map[string]any) <-chan map[string]any {
	ch := make(chan map[string]any, len(items))
	for _, item := range items {
		ch <- item
	}
	close(ch)
	return ch
}

func collectChunks(ch <-chan StreamChunk) []StreamChunk {
	var chunks []StreamChunk
	for item := range ch {
		chunks = append(chunks, item)
	}
	return chunks
}
```

- [ ] **Step 2: Run the analysis service tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis -count=1
```

Expected:

- command exits non-zero
- failure mentions missing package/files under `backend/service/analysis`

- [ ] **Step 3: Write the analysis contracts and contract-preserving service implementation**

```go
// D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\contracts.go
package analysis

type StockRequest struct {
	StockName   string `json:"stockName"`
	StockCode   string `json:"stockCode"`
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
}

type MarketSummaryRequest struct {
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
	HistoryJSON string `json:"historyJSON"`
}

type StreamChunk struct {
	ChatID       string `json:"chatId"`
	Question     string `json:"question"`
	Content      string `json:"content"`
	ExtraContent string `json:"extraContent"`
	Model        string `json:"model"`
	Time         string `json:"time"`
	ReasoningContent string           `json:"reasoning_content"`
	ToolCalls        []map[string]any `json:"tool_calls"`
}
```

```go
// D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service.go
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-stock/backend/logger"
	"go-stock/backend/models"
)

type StreamSource interface {
	StockStream(ctx context.Context, request StockRequest) <-chan map[string]any
	MarketSummaryStream(ctx context.Context, request MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any
}

type ResultSource interface {
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
	GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult
	GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error)
	DeleteResult(ctx context.Context, id uint) error
	BatchDeleteResults(ctx context.Context, ids []uint) error
}

type PromptSource interface {
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
}

type Service struct {
	streams StreamSource
	results ResultSource
	prompts PromptSource
}

func NewService(streams StreamSource, results ResultSource, prompts PromptSource) *Service {
	if streams == nil {
		panic("analysis: streams dependency is required")
	}
	if results == nil {
		panic("analysis: results dependency is required")
	}
	if prompts == nil {
		panic("analysis: prompts dependency is required")
	}
	return &Service{
		streams: streams,
		results: results,
		prompts: prompts,
	}
}

func (s *Service) StartStockAnalysis(ctx context.Context, request StockRequest) <-chan StreamChunk {
	return mapChunks(s.streams.StockStream(ctx, request))
}

func (s *Service) StartMarketSummary(ctx context.Context, request MarketSummaryRequest) <-chan StreamChunk {
	history, err := parseHistory(request.HistoryJSON)
	if err != nil {
		logHistoryParseError(err, request.HistoryJSON)
	}
	return mapChunks(s.streams.MarketSummaryStream(ctx, request, history))
}

func (s *Service) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	s.results.SaveResult(ctx, stockCode, stockName, result, chatID, question, aiConfigID)
}

func (s *Service) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	result := s.results.GetLatestResult(ctx, stockCode)
	if result == nil {
		return &models.AIResponseResult{}
	}
	return result
}

func (s *Service) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	page, err := s.results.GetResultPage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.AIResponseResultPageData{}, nil
	}
	return page, nil
}

func (s *Service) DeleteResult(ctx context.Context, id uint) error {
	return s.results.DeleteResult(ctx, id)
}

func (s *Service) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return s.results.BatchDeleteResults(ctx, ids)
}

func (s *Service) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	result := s.prompts.GetPromptTemplates(ctx, name, promptType)
	if result == nil {
		return &empty
	}
	return result
}

func (s *Service) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	page, err := s.prompts.GetPromptTemplatePage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.PromptTemplatePageData{}, nil
	}
	return page, nil
}

func (s *Service) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return s.prompts.SavePromptTemplate(ctx, template)
}

func (s *Service) DeletePromptTemplate(ctx context.Context, id uint) string {
	return s.prompts.DeletePromptTemplate(ctx, id)
}

func parseHistory(raw string) ([]map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var items []models.AiAssistantMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		entry := map[string]interface{}{
			"role":    item.Role,
			"content": item.Content,
		}
		if item.Role == "assistant" && item.Reasoning != "" {
			entry["reasoning_content"] = item.Reasoning
		}
		result = append(result, entry)
	}
	return result, nil
}

func mapChunks(raw <-chan map[string]any) <-chan StreamChunk {
	if raw == nil {
		ch := make(chan StreamChunk)
		close(ch)
		return ch
	}
	out := make(chan StreamChunk, 128)
	go func() {
		defer close(out)
		for item := range raw {
			out <- StreamChunk{
				ChatID:           asString(item["chatId"]),
				Question:         asString(item["question"]),
				Content:          asString(item["content"]),
				ExtraContent:     asString(item["extraContent"]),
				Model:            asString(item["model"]),
				Time:             asString(item["time"]),
				ReasoningContent: asString(item["reasoning_content"]),
				ToolCalls:        asToolCalls(item["tool_calls"]),
			}
		}
	}()
	return out
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func asToolCalls(value any) []map[string]any {
	if value == nil {
		return nil
	}
	if entries, ok := value.([]map[string]any); ok {
		copied := make([]map[string]any, len(entries))
		copy(copied, entries)
		return copied
	}
	rawSlice, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(rawSlice))
	for _, item := range rawSlice {
		if cast, ok := item.(map[string]any); ok {
			result = append(result, cast)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func logHistoryParseError(err error, raw string) {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return
	}
	const maxRawLength = 2048
	if len(raw) > maxRawLength {
		raw = raw[:maxRawLength]
	}
	runtimeLogger.ForSink(logger.SinkError, "analysis.service").Error(
		"analysis.market_history.parse_failed",
		"无法解析市场历史 JSON",
		logger.Err(err),
		logger.String("history_json", raw),
	)
}
```

- [ ] **Step 4: Run the analysis service tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis -count=1
```

Expected:

- command exits `0`
- all analysis service tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/analysis/contracts.go backend/service/analysis/service.go backend/service/analysis/service_test.go
git commit -m "refactor: add analysis service boundary"
```

### Task 2: Wire The Analysis Service Through `App` And `app_common`

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\provider.go`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\store.go`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_analysis_test.go`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_common.go`

Task 2 must keep the Wails bridge surface stable while moving the implementation behind the new source/service layer:

- `NewChatStream` and `SummaryStockNews` must forward the normalized stream fields already preserved by Task 1, including `reasoning_content` and `tool_calls`
- the bridge methods in `app.go` / `app_common.go` should delegate to `analysisService` instead of reaching back into `backend/data`
- prompt-template and analysis-result methods should preserve the existing public method names and return shapes

- [ ] **Step 1: Write the failing App bridge test**

```go
package main

import (
	"context"
	"testing"

	analysisservice "go-stock/backend/service/analysis"
	"go-stock/backend/models"
)

type fakeAnalysisService struct {
	latest     *models.AIResponseResult
	page       *models.AIResponseResultPageData
	prompts    *[]models.PromptTemplate
	promptPage *models.PromptTemplatePageData

	savedStockCode string
	savedStockName string
	savedContent   string
	savedChatID    string
	savedQuestion  string
	savedAIConfig  int

	deletedResultID uint
	batchedIDs      []uint
	savedTemplate   models.PromptTemplate
	deletedPromptID uint
}

func (f *fakeAnalysisService) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (f *fakeAnalysisService) StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (f *fakeAnalysisService) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.savedStockCode = stockCode
	f.savedStockName = stockName
	f.savedContent = result
	f.savedChatID = chatID
	f.savedQuestion = question
	f.savedAIConfig = aiConfigID
}

func (f *fakeAnalysisService) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return f.latest
}

func (f *fakeAnalysisService) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return f.page, nil
}

func (f *fakeAnalysisService) DeleteResult(ctx context.Context, id uint) error {
	f.deletedResultID = id
	return nil
}

func (f *fakeAnalysisService) BatchDeleteResults(ctx context.Context, ids []uint) error {
	f.batchedIDs = append([]uint(nil), ids...)
	return nil
}

func (f *fakeAnalysisService) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.prompts
}

func (f *fakeAnalysisService) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return f.promptPage, nil
}

func (f *fakeAnalysisService) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTemplate = template
	return "模板已保存"
}

func (f *fakeAnalysisService) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedPromptID = id
	return "删除成功"
}

func TestApp_AnalysisReadMethodsDelegateToService(t *testing.T) {
	prompts := []models.PromptTemplate{{ID: 3, Name: "系统模板", Type: "模型系统Prompt"}}
	app := NewApp()
	fake := &fakeAnalysisService{
		latest: &models.AIResponseResult{
			StockCode: "000001.SZ",
			StockName: "平安银行",
			Content:   "分析完成",
		},
		page: &models.AIResponseResultPageData{
			List:       []models.AIResponseResult{{StockCode: "000001.SZ"}},
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
		prompts: &prompts,
		promptPage: &models.PromptTemplatePageData{
			List:       prompts,
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	app.analysisService = fake

	app.SaveAIResponseResult("000001.SZ", "平安银行", "新的分析", "chat-1", "怎么看", 12)
	if fake.savedStockCode != "000001.SZ" || fake.savedAIConfig != 12 {
		t.Fatalf("unexpected save args: %#v", fake)
	}

	if got := app.GetAIResponseResult("000001.SZ"); got.StockCode != "000001.SZ" {
		t.Fatalf("unexpected latest result: %#v", got)
	}

	if got := app.GetAIResponseResultList(models.AIResponseResultQuery{Page: 1, PageSize: 10}); got.Total != 1 {
		t.Fatalf("unexpected result page: %#v", got)
	}

	if msg := app.DeleteAIResponseResult(9); msg != "删除成功" || fake.deletedResultID != 9 {
		t.Fatalf("unexpected delete result state: msg=%q id=%d", msg, fake.deletedResultID)
	}

	if msg := app.BatchDeleteAIResponseResult([]uint{1, 2}); msg != "删除成功" || len(fake.batchedIDs) != 2 {
		t.Fatalf("unexpected batch delete state: msg=%q ids=%v", msg, fake.batchedIDs)
	}

	if got := app.GetPromptTemplates("", ""); len(*got) != 1 || (*got)[0].Name != "系统模板" {
		t.Fatalf("unexpected prompt templates: %#v", got)
	}

	if got := app.GetPromptTemplateList(models.PromptTemplateQuery{Page: 1, PageSize: 10}); got.Total != 1 {
		t.Fatalf("unexpected prompt page: %#v", got)
	}

	if msg := app.AddPromptTemplate(models.PromptTemplate{Name: "系统模板", Type: "模型系统Prompt", Content: "请先分析风险"}); msg != "模板已保存" {
		t.Fatalf("unexpected add prompt template message: %q", msg)
	}

	if msg := app.UpdatePromptTemplate(models.PromptTemplate{ID: 3, Name: "系统模板", Type: "模型系统Prompt", Content: "请更新"}); msg != "模板已保存" {
		t.Fatalf("unexpected update prompt template message: %q", msg)
	}

	if msg := app.DeletePromptTemplate(3); msg != "删除成功" || fake.deletedPromptID != 3 {
		t.Fatalf("unexpected delete prompt template state: msg=%q id=%d", msg, fake.deletedPromptID)
	}

	if msg := app.AddPrompt(models.Prompt{Name: "用户模板", Type: "模型用户Prompt", Content: "请总结"}); msg != "模板已保存" {
		t.Fatalf("unexpected add prompt message: %q", msg)
	}

	if msg := app.DelPrompt(11); msg != "删除成功" || fake.deletedPromptID != 11 {
		t.Fatalf("unexpected del prompt state: msg=%q id=%d", msg, fake.deletedPromptID)
	}
}
```

- [ ] **Step 2: Run the targeted App bridge test to verify it fails**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . -run "TestApp_AnalysisReadMethodsDelegateToService" -count=1
```

Expected:

- command exits non-zero
- failure mentions `analysisService` wiring or missing source/service files

- [ ] **Step 3: Add the analysis source adapters and App bridge wiring**

```go
// D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\provider.go
package analysis

import (
	"context"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
)

type Provider struct {
	tools []data.Tool
}

func NewProvider(tools []data.Tool) *Provider {
	return &Provider{tools: append([]data.Tool(nil), tools...)}
}

func (p *Provider) StockStream(ctx context.Context, request analysisservice.StockRequest) <-chan map[string]any {
	if request.EnableTools {
		return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewChatStream(
			request.StockName,
			request.StockCode,
			request.Question,
			request.SysPromptID,
			p.tools,
			request.Think,
		)
	}
	return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewChatStream(
		request.StockName,
		request.StockCode,
		request.Question,
		request.SysPromptID,
		[]data.Tool{},
		request.Think,
	)
}

func (p *Provider) MarketSummaryStream(ctx context.Context, request analysisservice.MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any {
	if request.EnableTools {
		return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewSummaryStockNewsStreamWithTools(
			request.Question,
			request.SysPromptID,
			p.tools,
			request.Think,
			history,
		)
	}
	return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewSummaryStockNewsStream(
		request.Question,
		request.SysPromptID,
		request.Think,
		history,
	)
}
```

```go
// D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\store.go
package analysis

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	data.NewDeepSeekOpenAi(ctx, aiConfigID).SaveAIResponseResult(stockCode, stockName, result, chatID, question)
}

func (s *Store) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return data.NewDeepSeekOpenAi(ctx, 0).GetAIResponseResult(stockCode)
}

func (s *Store) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return data.NewAIResponseResultService().GetAIResponseResultList(query)
}

func (s *Store) DeleteResult(ctx context.Context, id uint) error {
	return data.NewAIResponseResultService().DeleteAIResponseResult(id)
}

func (s *Store) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return data.NewAIResponseResultService().BatchDeleteAIResponseResult(ids)
}

func (s *Store) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
}

func (s *Store) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return data.NewPromptTemplateApi().GetPromptTemplateList(&query)
}

func (s *Store) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (s *Store) DeletePromptTemplate(ctx context.Context, id uint) string {
	return data.NewPromptTemplateApi().DelPrompt(id)
}
```

```go
// Add to D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go imports
import (
	analysisservice "go-stock/backend/service/analysis"
	analysissource "go-stock/backend/source/analysis"
)

type analysisService interface {
	StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk
	StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
	GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult
	GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error)
	DeleteResult(ctx context.Context, id uint) error
	BatchDeleteResults(ctx context.Context, ids []uint) error
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
}

type App struct {
	// existing fields...
	marketReadService marketReadService
	analysisService   analysisService
}

func NewApp() *App {
	cacheSize := 512 * 1024
	cache := freecache.NewCache(cacheSize)
	c := cron.New(cron.WithSeconds())
	c.Start()
	var tools []data.Tool
	tools = data.Tools(tools)

	analysisProvider := analysissource.NewProvider(tools)
	analysisStore := analysissource.NewStore()

	return &App{
		cache:              cache,
		cron:               c,
		cronEntrys:         make(map[string]cron.EntryID),
		AiTools:            tools,
		marketReadService:  marketservice.NewService(marketsource.NewSource()),
		analysisService:    analysisservice.NewService(analysisProvider, analysisStore, analysisStore),
		stockAlertLastSent: make(map[string]time.Time),
		priceAtAlertReset:  make(map[string]float64),
	}
}
```

```go
// Replace in D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go
func (a *App) NewChatStream(stock, stockCode, question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool) {
	msgs := a.analysisService.StartStockAnalysis(a.ctx, analysisservice.StockRequest{
		StockName:   stock,
		StockCode:   stockCode,
		Question:    question,
		AIConfigID:  aiConfigId,
		SysPromptID: sysPromptId,
		EnableTools: enableTools,
		Think:       think,
	})
	for msg := range msgs {
		runtime.EventsEmit(a.ctx, "newChatStream", map[string]any{
			"chatId":            msg.ChatID,
			"question":          msg.Question,
			"content":           msg.Content,
			"extraContent":      msg.ExtraContent,
			"model":             msg.Model,
			"time":              msg.Time,
			"reasoning_content": msg.ReasoningContent,
			"tool_calls":        msg.ToolCalls,
		})
	}
	runtime.EventsEmit(a.ctx, "newChatStream", "DONE")
}

func (a *App) SaveAIResponseResult(stockCode, stockName, result, chatId, question string, aiConfigId int) {
	a.analysisService.SaveResult(a.ctx, stockCode, stockName, result, chatId, question, aiConfigId)
}

func (a *App) GetAIResponseResult(stock string) *models.AIResponseResult {
	return a.analysisService.GetLatestResult(a.ctx, stock)
}

func (a *App) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	return a.analysisService.GetPromptTemplates(a.ctx, name, promptType)
}

func (a *App) AddPrompt(prompt models.Prompt) string {
	return a.analysisService.SavePromptTemplate(a.ctx, models.PromptTemplate{
		ID:      prompt.ID,
		Name:    prompt.Name,
		Type:    prompt.Type,
		Content: prompt.Content,
	})
}

func (a *App) DelPrompt(id uint) string {
	return a.analysisService.DeletePromptTemplate(a.ctx, id)
}
```

```go
// Replace the stream body in D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go
func (a *App) SummaryStockNews(question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool, eventName string, historyJSON string) {
	ctx, cancel := context.WithCancel(a.ctx)

	a.summaryMu.Lock()
	if a.summaryCancel != nil {
		a.summaryCancel()
	}
	a.summaryCancel = cancel
	a.summaryMu.Unlock()

	if strings.TrimSpace(eventName) == "" {
		eventName = "summaryStockNews"
	}

	msgs := a.analysisService.StartMarketSummary(ctx, analysisservice.MarketSummaryRequest{
		Question:    question,
		AIConfigID:  aiConfigId,
		SysPromptID: sysPromptId,
		EnableTools: enableTools,
		Think:       think,
		HistoryJSON: historyJSON,
	})

	for msg := range msgs {
		runtime.EventsEmit(a.ctx, eventName, map[string]any{
			"chatId":            msg.ChatID,
			"question":          msg.Question,
			"content":           msg.Content,
			"extraContent":      msg.ExtraContent,
			"model":             msg.Model,
			"time":              msg.Time,
			"reasoning_content": msg.ReasoningContent,
			"tool_calls":        msg.ToolCalls,
		})
	}

	a.summaryMu.Lock()
	a.summaryCancel = nil
	a.summaryMu.Unlock()

	runtime.EventsEmit(a.ctx, eventName, "DONE")
}
```

```go
// Replace in D:\codex_work\go-stock\.worktrees\factor-phase-one\app_common.go
func (a *App) GetAIResponseResultList(query models.AIResponseResultQuery) *models.AIResponseResultPageData {
	page, err := a.analysisService.GetResultPage(a.ctx, query)
	if err != nil {
		return &models.AIResponseResultPageData{}
	}
	return page
}

func (a *App) DeleteAIResponseResult(id uint) string {
	if err := a.analysisService.DeleteResult(a.ctx, id); err != nil {
		return "删除失败"
	}
	return "删除成功"
}

func (a *App) BatchDeleteAIResponseResult(ids []uint) string {
	if err := a.analysisService.BatchDeleteResults(a.ctx, ids); err != nil {
		return "删除失败"
	}
	return "删除成功"
}

func (a *App) GetPromptTemplateList(query models.PromptTemplateQuery) *models.PromptTemplatePageData {
	page, err := a.analysisService.GetPromptTemplatePage(a.ctx, query)
	if err != nil {
		return &models.PromptTemplatePageData{}
	}
	return page
}

func (a *App) AddPromptTemplate(template models.PromptTemplate) string {
	return a.analysisService.SavePromptTemplate(a.ctx, template)
}

func (a *App) UpdatePromptTemplate(template models.PromptTemplate) string {
	return a.analysisService.SavePromptTemplate(a.ctx, template)
}

func (a *App) DeletePromptTemplate(id uint) string {
	return a.analysisService.DeletePromptTemplate(a.ctx, id)
}
```

- [ ] **Step 4: Run the bridge and service tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/analysis -run "TestApp_AnalysisReadMethodsDelegateToService|TestService_ReadAndPromptMethodsDelegateToStores|TestService_StartStreamsDelegateAndParseHistory" -count=1
```

Expected:

- command exits `0`
- App bridge test and analysis service tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/source/analysis/provider.go backend/source/analysis/store.go app.go app_common.go app_analysis_test.go
git commit -m "refactor: bridge analysis service through app"
```

### Task 3: Add Frontend Analysis Service Wrappers And Tests

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.mjs`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.test.mjs`

- [ ] **Step 1: Write the failing frontend analysis service tests**

```javascript
import test from 'node:test'
import assert from 'node:assert/strict'

import {
  normalizeAnalysisResult,
  normalizeAnalysisResultPage,
  normalizePromptTemplates,
  normalizePromptTemplatePage,
} from './analysisService.mjs'

test('normalizeAnalysisResult fills missing fields with stable defaults', () => {
  assert.deepEqual(
    normalizeAnalysisResult({ stockCode: '000001.SZ', content: '分析完成' }),
    {
      ID: 0,
      chatId: '',
      modelName: '',
      stockCode: '000001.SZ',
      stockName: '',
      question: '',
      content: '分析完成',
      CreatedAt: '',
      UpdatedAt: '',
    },
  )
})

test('normalizeAnalysisResultPage maps list and pagination defaults', () => {
  assert.deepEqual(
    normalizeAnalysisResultPage({
      list: [{ stockCode: '000001.SZ', content: '分析完成' }],
      total: 1,
      totalPages: 1,
    }),
    {
      list: [
        {
          ID: 0,
          chatId: '',
          modelName: '',
          stockCode: '000001.SZ',
          stockName: '',
          question: '',
          content: '分析完成',
          CreatedAt: '',
          UpdatedAt: '',
        },
      ],
      total: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1,
    },
  )
})

test('normalizePromptTemplates keeps only normalized prompt rows', () => {
  assert.deepEqual(
    normalizePromptTemplates([{ ID: 3, name: '系统模板', type: '模型系统Prompt', content: '请先分析风险' }]),
    [{ ID: 3, name: '系统模板', type: '模型系统Prompt', content: '请先分析风险', CreatedAt: '', UpdatedAt: '' }],
  )
})

test('normalizePromptTemplatePage falls back to an empty page', () => {
  assert.deepEqual(normalizePromptTemplatePage(null), {
    list: [],
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  })
})
```

- [ ] **Step 2: Run the frontend analysis service tests to verify they fail**

Run:

```powershell
node --test frontend/src/services/analysisService.test.mjs
```

Expected:

- command exits non-zero
- failure mentions `frontend/src/services/analysisService.mjs` does not exist

- [ ] **Step 3: Add the frontend analysis service module**

```javascript
// D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.mjs
import * as AppBindings from '../../wailsjs/go/main/App.js'

function toArray(value) {
  return Array.isArray(value) ? value : []
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback
}

export function normalizeAnalysisResult(value) {
  const data = value ?? {}
  return {
    ID: toNumber(data.ID, 0),
    chatId: data.chatId ?? '',
    modelName: data.modelName ?? '',
    stockCode: data.stockCode ?? '',
    stockName: data.stockName ?? '',
    question: data.question ?? '',
    content: data.content ?? '',
    CreatedAt: data.CreatedAt ?? '',
    UpdatedAt: data.UpdatedAt ?? '',
  }
}

export function normalizeAnalysisResultPage(value) {
  const data = value ?? {}
  return {
    list: toArray(data.list).map(normalizeAnalysisResult),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, 1),
    pageSize: toNumber(data.pageSize, 10),
    totalPages: toNumber(data.totalPages, 0),
  }
}

function normalizePromptTemplate(value) {
  const data = value ?? {}
  return {
    ID: toNumber(data.ID, 0),
    name: data.name ?? '',
    type: data.type ?? '',
    content: data.content ?? '',
    CreatedAt: data.CreatedAt ?? '',
    UpdatedAt: data.UpdatedAt ?? '',
  }
}

export function normalizePromptTemplates(value) {
  return toArray(value).map(normalizePromptTemplate)
}

export function normalizePromptTemplatePage(value) {
  const data = value ?? {}
  return {
    list: toArray(data.list).map(normalizePromptTemplate),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, 1),
    pageSize: toNumber(data.pageSize, 10),
    totalPages: toNumber(data.totalPages, 0),
  }
}

export async function startStockAnalysis({
  stockName,
  stockCode,
  question,
  aiConfigId,
  sysPromptId,
  enableTools = true,
  think = false,
}) {
  return AppBindings.NewChatStream(stockName, stockCode, question, aiConfigId, sysPromptId, enableTools, think)
}

export async function startMarketSummary({
  question,
  aiConfigId,
  sysPromptId,
  enableTools = true,
  think = false,
  eventName = 'summaryStockNews',
  historyJSON = '',
}) {
  return AppBindings.SummaryStockNews(question, aiConfigId, sysPromptId, enableTools, think, eventName, historyJSON)
}

export async function saveAnalysisResult({
  stockCode,
  stockName,
  content,
  chatId,
  question,
  aiConfigId,
}) {
  return AppBindings.SaveAIResponseResult(stockCode, stockName, content, chatId, question, aiConfigId)
}

export async function loadLatestAnalysisResult(stockCode) {
  return normalizeAnalysisResult(await AppBindings.GetAIResponseResult(stockCode))
}

export async function loadAnalysisResultPage(query) {
  return normalizeAnalysisResultPage(await AppBindings.GetAIResponseResultList(query))
}

export async function deleteAnalysisResult(id) {
  return AppBindings.DeleteAIResponseResult(id)
}

export async function loadPromptTemplates(name = '', type = '') {
  return normalizePromptTemplates(await AppBindings.GetPromptTemplates(name, type))
}

export async function loadPromptTemplatePage(query) {
  return normalizePromptTemplatePage(await AppBindings.GetPromptTemplateList(query))
}

export async function savePromptTemplate(template, { edit = false } = {}) {
  if (edit) {
    return AppBindings.UpdatePromptTemplate(template)
  }
  return AppBindings.AddPromptTemplate(template)
}

export async function deletePromptTemplate(id) {
  return AppBindings.DeletePromptTemplate(id)
}

export async function shareAnalysis(stockCode, stockName) {
  return AppBindings.ShareAnalysis(stockCode, stockName)
}

export async function saveAnalysisMarkdown(stockCode, stockName) {
  return AppBindings.SaveAsMarkdown(stockCode, stockName)
}
```

- [ ] **Step 4: Run the frontend analysis service tests**

Run:

```powershell
node --test frontend/src/services/analysisService.test.mjs
```

Expected:

- command exits `0`
- all four tests pass

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/services/analysisService.mjs frontend/src/services/analysisService.test.mjs
git commit -m "refactor: add frontend analysis service wrappers"
```

### Task 4: Route Stock And Research Through `pages` And Switch Analysis Screens To The Service Layer

**Files:**
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\stock-page.vue`
- Create: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\research-page.vue`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\router\router.js`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\stock.vue`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\market.vue`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\researchReport.vue`
- Modify: `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\promptTemplateList.vue`

- [ ] **Step 1: Confirm the current stock/research routes and direct analysis imports still use the old path**

Run:

```powershell
rg -n "components/stock\\.vue|components/researchIndex\\.vue|GetAIResponseResult|SaveAIResponseResult|SummaryStockNews|NewChatStream|GetPromptTemplates|GetPromptTemplateList|AddPromptTemplate|DeletePromptTemplate|UpdatePromptTemplate" frontend/src/router/router.js frontend/src/components/stock.vue frontend/src/components/market.vue frontend/src/components/researchReport.vue frontend/src/components/promptTemplateList.vue
```

Expected:

- `frontend/src/router/router.js` still points `/` straight at `../components/stock.vue`
- `frontend/src/router/router.js` still points `/research` straight at `../components/researchIndex.vue`
- the listed components still import analysis-domain calls directly from Wails bindings

- [ ] **Step 2: Add stock/research page wrappers and route through them**

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\stock-page.vue -->
<script setup>
import StockView from "../components/stock.vue";
</script>

<template>
  <StockView />
</template>
```

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\research-page.vue -->
<script setup>
import ResearchView from "../components/researchIndex.vue";
</script>

<template>
  <ResearchView />
</template>
```

```javascript
// D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\router\router.js
import stockPageView from '../pages/stock-page.vue'
import researchPageView from '../pages/research-page.vue'
import marketPageView from '../pages/market-page.vue'

const routes = [
  { path: '/', component: stockPageView, name: 'stock' },
  { path: '/fund', component: fundView, name: 'fund' },
  { path: '/settings', component: settingsView, name: 'settings' },
  { path: '/about', component: aboutView, name: 'about' },
  { path: '/market', component: marketPageView, name: 'market' },
  { path: '/agent', component: agentChat, name: 'agent' },
  { path: '/research', component: researchPageView, name: 'research' },
  { path: '/cron-tasks', component: cronTaskManager, name: 'cronTasks' },
]
```

- [ ] **Step 3: Switch the analysis-heavy components to `analysisService.mjs`**

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\stock.vue -->
<script setup>
import {
  AddGroup,
  AddStockGroup,
  Follow,
  GetAiConfigs,
  GetConfig,
  GetFollowList,
  GetGroupList,
  GetStockKLine,
  GetStockList,
  GetStockMinutePriceLineData,
  GetVersionInfo,
  Greet,
  InitializeGroupSort,
  OpenURL,
  RemoveGroup,
  RemoveStockGroup,
  SaveImage,
  SaveWordFile,
  SendDingDingMessageByType,
  SetAlarmChangePercent,
  SetCostPriceAndVolume,
  SetStockAICron,
  SetStockSort,
  SetTradingPrice,
  UnFollow,
  UpdateGroupSort
} from '../../wailsjs/go/main/App'
import {
  loadLatestAnalysisResult,
  loadPromptTemplates,
  saveAnalysisMarkdown,
  saveAnalysisResult,
  shareAnalysis,
  startStockAnalysis,
} from '../services/analysisService.mjs'

onBeforeMount(() => {
  loadPromptTemplates('', '').then(res => {
    promptTemplates.value = res
    sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统Prompt')
    userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户Prompt')
  })

  EventsOn("newChatStream", async (msg) => {
    data.loading = false
    if (msg === "DONE") {
      await saveAnalysisResult({
        stockCode: data.code,
        stockName: data.name,
        content: data.airesult,
        chatId: data.chatId,
        question: data.question,
        aiConfigId: data.aiConfigId,
      })
      message.info("AI分析完成！")
      message.destroyAll()
      return
    }
    if (msg.chatId) {
      data.chatId = msg.chatId
    }
    if (msg.question) {
      data.question = msg.question
    }
    if (msg.content) {
      data.airesult = data.airesult + msg.content
    }
    if (msg.extraContent) {
      data.airesult = data.airesult + msg.extraContent
    }
  })
})

function aiReCheckStock(stock, stockCode) {
  data.modelName = ""
  data.airesult = ""
  data.time = ""
  data.name = stock
  data.code = stockCode
  data.loading = true
  modalShow4.value = true
  message.loading("ai检测中...", { duration: 0 })

  startStockAnalysis({
    stockName: stock,
    stockCode,
    question: data.question,
    aiConfigId: data.aiConfigId,
    sysPromptId: data.sysPromptId,
    enableTools: enableTools.value,
    think: thinkingMode.value,
  })
}

function aiCheckStock(stock, stockCode) {
  loadLatestAnalysisResult(stockCode).then(result => {
    if (result.content) {
      data.modelName = result.modelName
      data.chatId = result.chatId
      data.question = result.question
      data.name = stock
      data.code = stockCode
      data.loading = false
      modalShow4.value = true
      data.airesult = result.content
      data.time = result.CreatedAt ? result.CreatedAt.replace('T', ' ').slice(0, 19) : ''
      return
    }
    data.modelName = ""
    data.question = ""
    data.airesult = ""
    data.time = ""
    data.name = stock
    data.code = stockCode
    data.loading = false
    modalShow4.value = true
  })
}

function saveAsMarkdown() {
  saveAnalysisMarkdown(data.code, data.name).then(result => {
    message.success(result)
  })
}

function share(code, name) {
  shareAnalysis(code, name).then(msg => {
    notify.info({
      avatar: () => h(NAvatar, { size: 'small', round: false, src: icon.value }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => h('div', { style: { 'text-align': 'left', 'font-size': '14px' } }, { default: () => msg }),
    })
  })
}
</script>
```

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\market.vue -->
<script setup>
import {
  GetAiConfigs,
  GetConfig,
} from "../../wailsjs/go/main/App";
import {
  loadLatestAnalysisResult,
  loadPromptTemplates,
  saveAnalysisMarkdown,
  saveAnalysisResult,
  shareAnalysis,
  startMarketSummary,
} from "../services/analysisService.mjs";
import {
  loadMarketFeeds,
  loadMarketGlobalIndexes,
  loadMarketIndustryRanks,
  refreshMarketFeed,
} from "../services/marketService.mjs";
```

```vue
<!-- Continue in D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\market.vue -->
onBeforeMount(() => {
  nowTab.value = route.query.name
  stockCode.value = route.query.stockCode
  GetConfig().then(result => {
    summaryBTN.value = result.openAiEnable
    darkTheme.value = result.darkTheme
    httpProxyEnabled.value = result.httpProxyEnabled
  })
  loadPromptTemplates("", "").then(res => {
    promptTemplates.value = res
    sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统Prompt')
    userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户Prompt')
  })
  GetAiConfigs().then(res => {
    aiConfigs.value = res
    aiConfigId.value = resolveFirstAiConfigId(res)
  })
  loadFeeds()
  getIndex()
  industryRank()
  indexInterval.value = setInterval(getIndex, 3000)
  indexIndustryRank.value = setInterval(() => {
    industryRank()
    ReFlesh("财联社电报")
    ReFlesh("新浪财经")
    ReFlesh("外媒")
  }, 1000 * 10)
})

function reAiSummary() {
  aiSummary.value = ""
  summaryModal.value = true
  loading.value = true
  startMarketSummary({
    question: question.value,
    aiConfigId: aiConfigId.value,
    sysPromptId: sysPromptId.value,
    enableTools: enableTools.value,
    think: thinkingMode.value,
    eventName: "summaryStockNews",
    historyJSON: "",
  })
}

function getAiSummary() {
  summaryModal.value = true
  loading.value = true
  loadLatestAnalysisResult("市场资讯").then(result => {
    loading.value = false
    if (!result.content) {
      aiSummaryTime.value = ""
      aiSummary.value = ""
      modelName.value = ""
      return
    }
    aiSummary.value = result.content
    question.value = result.question
    aiSummaryTime.value = result.CreatedAt ? result.CreatedAt.replace('T', ' ').slice(0, 19) : ''
    modelName.value = result.modelName
  })
}

EventsOn("summaryStockNews", async (msg) => {
  loading.value = false
  if (msg === "DONE") {
    await saveAnalysisResult({
      stockCode: "市场资讯",
      stockName: "市场资讯",
      content: aiSummary.value,
      chatId: chatId.value,
      question: question.value,
      aiConfigId: aiConfigId.value,
    })
    message.info("AI分析完成！")
    message.destroyAll()
    return
  }
  if (msg.chatId) {
    chatId.value = msg.chatId
  }
  if (msg.question) {
    question.value = msg.question
  }
  if (msg.content) {
    aiSummary.value = aiSummary.value + msg.content
  }
  if (msg.extraContent) {
    aiSummary.value = aiSummary.value + msg.extraContent
  }
  if (msg.model) {
    modelName.value = msg.model
  }
  if (msg.time) {
    aiSummaryTime.value = msg.time
  }
})
```

```vue
<!-- Finish the market component migration -->
function saveAsMarkdown() {
  saveAnalysisMarkdown('市场资讯', '市场资讯').then(result => {
    message.success(result)
  })
}

function share() {
  shareAnalysis('市场资讯', '市场资讯').then(msg => {
    notify.info({
      avatar: () => h(NAvatar, { size: 'small', round: false, src: icon.value }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => h('div', { style: { 'text-align': 'left', 'font-size': '14px' } }, { default: () => msg }),
    })
  })
}
</script>
```

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\researchReport.vue -->
<script setup>
import { GetConfig } from "../../wailsjs/go/main/App";
import {
  deleteAnalysisResult,
  loadAnalysisResultPage,
  saveAnalysisMarkdown,
  shareAnalysis,
} from "../services/analysisService.mjs";

function query({
  page,
  pageSize = 10,
  order = 'desc',
  keyword = "",
  startDate = "",
  endDate = ""
}) {
  return loadAnalysisResultPage({
    page,
    pageSize,
    modelName: keyword,
    question: keyword,
    stockName: keyword,
    stockCode: keyword,
    startDate,
    endDate,
  }).then((res) => ({
    pageCount: res.totalPages,
    data: res.list,
    total: res.total,
  }))
}

function share(code, name) {
  shareAnalysis(code, name).then(msg => {
    notify.info({
      avatar: () => h(NAvatar, { size: 'small', round: false, src: icon.value }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => h('div', { style: { 'text-align': 'left', 'font-size': '14px' } }, { default: () => msg }),
    })
  })
}

function saveAsMarkdown(code, name) {
  saveAnalysisMarkdown(code, name).then(result => {
    if (result !== "") {
      message.success(result)
    }
  })
}

function deleteAIResponseResult(id) {
  deleteAnalysisResult(id).then(result => {
    if (result !== "") {
      message.success(result)
      handleSearch()
    }
  })
}
</script>
```

```vue
<!-- D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\promptTemplateList.vue -->
<script setup>
import { GetConfig } from "../../wailsjs/go/main/App";
import { EventsEmit } from "../../wailsjs/runtime";
import {
  deletePromptTemplate as removePromptTemplate,
  loadPromptTemplatePage,
  savePromptTemplate as persistPromptTemplate,
} from "../services/analysisService.mjs";

function query({ page, pageSize = 10, name = "", type = "", content = "" }) {
  return loadPromptTemplatePage({
    page,
    pageSize,
    name,
    type,
    content,
  }).then((res) => ({
    data: res.list,
    total: res.total,
    totalPages: res.totalPages,
  }))
}

function savePromptTemplate() {
  if (!modalDataRef.formData.name || !modalDataRef.formData.type || !modalDataRef.formData.content) {
    message.warning('请填写完整信息')
    return
  }

  persistPromptTemplate({
    ID: modalDataRef.formData.ID,
    name: modalDataRef.formData.name,
    type: modalDataRef.formData.type,
    content: modalDataRef.formData.content,
  }, { edit: modalDataRef.isEdit }).then((res) => {
    message.info(res)
    modalDataRef.visible = false
    handleSearch()
    EventsEmit('promptTemplatesChanged')
  })
}

function deletePromptTemplate(id) {
  dialog.warning({
    title: '提示',
    content: '确定要删除这个模板吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: () => {
      removePromptTemplate(id).then((res) => {
        message.info(res)
        handleSearch()
        EventsEmit('promptTemplatesChanged')
      })
    }
  })
}
</script>
```

- [ ] **Step 4: Run focused frontend verification**

Run:

```powershell
node --test frontend/src/services/analysisService.test.mjs
npm --prefix frontend run build
```

Expected:

- `node --test` exits `0`
- frontend build exits `0`
- `/` and `/research` now route through page wrappers
- the four migrated components no longer import analysis-domain Wails bindings directly

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/pages/stock-page.vue frontend/src/pages/research-page.vue frontend/src/router/router.js frontend/src/components/stock.vue frontend/src/components/market.vue frontend/src/components/researchReport.vue frontend/src/components/promptTemplateList.vue
git commit -m "refactor: migrate analysis screens to page and service layers"
```

### Task 5: Run Slice Verification And Review The Result Against The Refactor Goals

**Files:**
- Verify only:
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\provider.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\source\analysis\store.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\contracts.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\backend\service\analysis\service_test.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\app.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_common.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\app_analysis_test.go`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.mjs`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\services\analysisService.test.mjs`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\stock-page.vue`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\pages\research-page.vue`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\router\router.js`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\stock.vue`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\market.vue`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\researchReport.vue`
  - `D:\codex_work\go-stock\.worktrees\factor-phase-one\frontend\src\components\promptTemplateList.vue`

- [ ] **Step 1: Run the targeted backend tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/analysis -run "TestApp_AnalysisReadMethodsDelegateToService|TestService_ReadAndPromptMethodsDelegateToStores|TestService_StartStreamsDelegateAndParseHistory" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 2: Run the full Go test suite**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./...
```

Expected:

- command exits `0`

- [ ] **Step 3: Run the frontend test and production build**

Run:

```powershell
node --test frontend/src/services/analysisService.test.mjs
npm --prefix frontend run build
```

Expected:

- both commands exit `0`

- [ ] **Step 4: Run the Wails desktop smoke build**

Run:

```powershell
$env:PATH='C:\Program Files\Go\bin;' + $env:PATH
& 'C:\Users\Aspir\go\bin\wails.exe' build --platform windows/amd64
```

Expected:

- command exits `0`
- no new Wails binding methods are required for this slice, so generated binding churn should be minimal or empty

- [ ] **Step 5: Review the diff is limited to the planned slice**

Run:

```powershell
git diff --stat HEAD~4..HEAD
```

Expected:

- diff is limited to the new analysis source/service files, App/AppCommon bridge wiring, the frontend analysis service, the two new page wrappers, the route updates, and the four analysis-heavy components

## Self-Review

### Spec Coverage

- The spec’s “AI 分析请求” requirement is covered by Task 1 and Task 2 via `backend/service/analysis` plus `NewChatStream` / `SummaryStockNews` delegation.
- The spec’s “AI 结果存储 / 读取” requirement is covered by Task 1, Task 2, and Task 4 via `SaveAIResponseResult`, `GetAIResponseResult`, and `GetAIResponseResultList`.
- The spec’s “提示词模板选择” requirement is covered by Task 2, Task 3, and Task 4 via prompt-template service wrappers and migrated prompt-template consumers.
- The spec’s “历史分析结果查看” requirement is covered by Task 4 through `researchReport.vue` and `/research` page routing.
- Page/service-layer expansion after phase one is covered by Task 3 and Task 4.
- The deliberately excluded assistant/cron/settings cleanup items match the overall spec’s instruction to keep AI assistant and task-system rewrites out of this slice.

### Placeholder Scan

- No `TODO`, `TBD`, “implement later”, or “similar to Task N” placeholders remain.
- Every code-changing step includes concrete file paths, code, and commands.
- The scope exclusions are explicit so the engineer does not accidentally widen the slice.

### Type Consistency

- Backend request and stream contract names are consistent across tasks:
  - `StockRequest`
  - `MarketSummaryRequest`
  - `StreamChunk`
- Backend service method names are consistent across tasks:
  - `StartStockAnalysis`
  - `StartMarketSummary`
  - `SaveResult`
  - `GetLatestResult`
  - `GetResultPage`
  - `GetPromptTemplates`
  - `GetPromptTemplatePage`
  - `SavePromptTemplate`
  - `DeletePromptTemplate`
- Frontend service names are consistent across tasks:
  - `startStockAnalysis`
  - `startMarketSummary`
  - `saveAnalysisResult`
  - `loadLatestAnalysisResult`
  - `loadAnalysisResultPage`
  - `loadPromptTemplates`
  - `loadPromptTemplatePage`
  - `savePromptTemplate`
  - `deletePromptTemplate`
