# go-stock Architecture Refactor Phase 4B / Phase 5 Closeout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the Windows-first `phase4-b + phase5 closeout` slice by moving the remaining notification / timed-analysis / residual market-read helpers behind service boundaries, routing settings and cron screens through `pages`, and deleting the legacy prompt fallback path so the final repo-wide closure plan can target a much smaller surface.

**Architecture:** Reuse the existing `market / analysis / config / task` service anchors instead of inventing a new umbrella layer. Extend `backend/service/market` to absorb the remaining bridge-only read helpers, extend `backend/service/analysis` and `backend/service/config` to cover result artifacts and config export, add a narrow `backend/service/notification` plus `backend/service/watchlist` slice for DingTalk / desktop notification and stock AI cron execution, then make `app.go` a thin delegator that keeps current Wails method names stable while dropping direct `backend/data` calls in the already-migrated domains.

**Tech Stack:** Go 1.26, Wails v2, Vue 3, Naive UI, Node `--test`, Vite, `robfig/cron`, `freecache`, `resty`

---

## Scope Note

This plan is intentionally narrower than the final full-repo closure:

- Included in this slice:
  - residual market read helpers still exposed directly from `app.go`
  - analysis artifact helpers used by `ShareAnalysis` and `SaveAsMarkdown`
  - config export and prompt bridge closeout
  - DingTalk / desktop notification bridge helpers
  - stock AI cron scheduling and execution
  - `settings` / `cron-tasks` route migration to `pages`
  - frontend market child widgets still importing Wails bindings directly
  - first shared user-visible error contract for the new service boundaries
- Explicitly excluded from this slice:
  - `FloatingAiAssistant.vue`
  - `FloatingAgentAssistant.vue`
  - `agent-chat.vue`
  - the broad CRUD surfaces still living in `app_common.go`
  - a repo-wide rewrite of every remaining direct `wailsjs` import
  - Linux / macOS legacy divergence cleanup

This is a Windows-first closeout slice because the current production path and release flow are Windows-centric. The plan must not widen into repairing all non-Windows historical branches.

## File Map

- Create: `D:\codex_work\go-stock\backend\service\contract\error_model.go`
  - Shared `UserVisibleError` contract with explicit `stage` and `retryable`.
- Create: `D:\codex_work\go-stock\backend\service\contract\error_model_test.go`
  - Unit tests for the shared error contract defaults and `error` behavior.
- Modify: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
  - Extend the market source adapter with the residual read helpers still called from `app.go`.
- Modify: `D:\codex_work\go-stock\backend\service\market\contracts.go`
  - Add typed contracts for residual industry money rank, money rank, and stock money trend rows while keeping the current JSON keys stable.
- Modify: `D:\codex_work\go-stock\backend\service\market\service.go`
  - Map the residual market helper outputs into typed contracts and readable strings.
- Create: `D:\codex_work\go-stock\backend\service\market\legacy_reads_test.go`
  - Unit tests for the new residual market helper normalization.
- Modify: `D:\codex_work\go-stock\backend\service\analysis\contracts.go`
  - Add an analysis artifact contract used by share / markdown bridge methods.
- Modify: `D:\codex_work\go-stock\backend\service\analysis\service.go`
  - Add `GetResultArtifact` with shared error contract output.
- Modify: `D:\codex_work\go-stock\backend\service\analysis\service_test.go`
  - Tests for analysis artifact fallback and filename generation.
- Modify: `D:\codex_work\go-stock\backend\service\config\store.go`
  - Add config export access to the config store boundary.
- Modify: `D:\codex_work\go-stock\backend\service\config\service.go`
  - Add exported-config access while keeping legacy prompt compatibility on the config boundary.
- Modify: `D:\codex_work\go-stock\backend\service\config\service_test.go`
  - Tests for config export delegation.
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter.go`
  - Thin adapter over DingTalk send APIs, stock snapshot lookup, and platform-neutral local notification dispatch.
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter_windows.go`
  - Windows / macOS local notification implementation backed by `data.NewAlertWindowsApi`.
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter_other.go`
  - Non-Windows / non-macOS no-op local notification implementation so the package still builds outside the Windows release path.
- Create: `D:\codex_work\go-stock\backend\service\notification\service.go`
  - Notification service with cache-based dedupe and typed notification delivery payloads.
- Create: `D:\codex_work\go-stock\backend\service\notification\service_test.go`
  - Unit tests for TTL dedupe and typed notification delivery.
- Create: `D:\codex_work\go-stock\backend\source\watchlist\store.go`
  - Thin adapter over followed stock reads and stock AI cron persistence.
- Create: `D:\codex_work\go-stock\backend\service\watchlist\service.go`
  - Service that persists stock AI cron text, restores scheduled watchlist jobs, and runs scheduled stock analysis through the analysis boundary.
- Create: `D:\codex_work\go-stock\backend\service\watchlist\service_test.go`
  - Unit tests for cron persistence, normalized stock-code lookup, list filtering, and scheduled analysis execution.
- Modify: `D:\codex_work\go-stock\app.go`
  - Wire the new closeout helpers, remove `legacyPromptBridge`, delegate residual old methods, and move notification / stock-AI-cron helpers to service-backed code.
- Modify: `D:\codex_work\go-stock\app_common.go`
  - Delegate prompt page CRUD straight through `configService` without legacy fallback.
- Modify: `D:\codex_work\go-stock\app_config_test.go`
  - Replace legacy prompt fallback expectations with nil-safe closeout expectations.
- Create: `D:\codex_work\go-stock\app_closeout_test.go`
  - Bridge tests for residual market methods, analysis artifact helpers, config export, notification helpers, and stock AI cron scheduling helpers.
- Create: `D:\codex_work\go-stock\frontend\src\pages\settings-page.vue`
  - Page wrapper for the existing settings screen.
- Create: `D:\codex_work\go-stock\frontend\src\pages\cron-task-page.vue`
  - Page wrapper for the existing cron task screen.
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
  - Route `/settings` and `/cron-tasks` through page wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
  - Add wrappers for the residual market helper methods that still hit raw Wails bindings.
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
  - Tests for the residual market helper normalization.
- Modify: `D:\codex_work\go-stock\frontend\src\services\configService.mjs`
  - Add wrapper for the settings-page notification test-send action.
- Modify: `D:\codex_work\go-stock\frontend\src\components\market.vue`
  - Replace residual direct config/prompt Wails imports with service wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\AnalyzeMartket.vue`
  - Replace direct Wails imports with market service wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\industryMoneyRank.vue`
  - Replace direct Wails imports with market service wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\moneyTrend.vue`
  - Replace direct Wails imports with market service wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\rankTable.vue`
  - Replace direct Wails imports with market service wrappers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`
  - Stop importing the notification Wails binding directly and use `configService.mjs`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\newsList.vue`
  - Remove the stale direct market refresh import that is no longer used.

### Task 1: Add Shared Error Contract And Close Out Residual `market` Service Reads

**Files:**
- Create: `D:\codex_work\go-stock\backend\service\contract\error_model.go`
- Create: `D:\codex_work\go-stock\backend\service\contract\error_model_test.go`
- Modify: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
- Modify: `D:\codex_work\go-stock\backend\service\market\contracts.go`
- Modify: `D:\codex_work\go-stock\backend\service\market\service.go`
- Create: `D:\codex_work\go-stock\backend\service\market\legacy_reads_test.go`

- [ ] **Step 1: Write the failing shared-error and residual-market tests**

```go
// D:\codex_work\go-stock\backend\service\contract\error_model_test.go
package contract

import "testing"

func TestUserVisibleError_ErrorUsesMessage(t *testing.T) {
	err := UserVisibleError{
		Code:      "analysis.result_missing",
		Message:   "分析结果不存在",
		Retryable: false,
		Stage:     StageService,
	}

	if err.Error() != "分析结果不存在" {
		t.Fatalf("expected Error() to return message, got %q", err.Error())
	}
}

func TestNewUserVisibleErrorBuildsStablePayload(t *testing.T) {
	err := NewUserVisibleError("market.fetch.failed", "市场数据获取失败", true, StageSource)
	if err.Code != "market.fetch.failed" || err.Message != "市场数据获取失败" || !err.Retryable || err.Stage != StageSource {
		t.Fatalf("unexpected user visible error: %#v", err)
	}
}
```

```go
// D:\codex_work\go-stock\backend\service\market\legacy_reads_test.go
package market

import (
	"testing"

	"go-stock/backend/models"
)

type residualSourceStub struct{}

func (s *residualSourceStub) GetTelegraphList(source string) *[]*models.Telegraph { return nil }
func (s *residualSourceStub) RefreshFeeds()                                        {}
func (s *residualSourceStub) GlobalStockIndexes(crawlTimeout uint) map[string]any  { return map[string]any{} }
func (s *residualSourceStub) GetIndustryRank(sort string, cnt int) map[string]any  { return map[string]any{} }
func (s *residualSourceStub) GlobalStockIndexesReadable(crawlTimeout uint) string {
	return "  亚洲市场：上证指数 +1.23%\n"
}
func (s *residualSourceStub) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return []map[string]any{{
		"category":        "board-1",
		"name":            "机器人",
		"avg_changeratio": 0.0312,
		"inamount":        2500000.0,
		"outamount":       1250000.0,
		"netamount":       1250000.0,
		"ratioamount":     0.22,
		"ts_name":         "机器人龙头",
		"ts_symbol":       "300024",
		"ts_changeratio":  0.0415,
		"ts_trade":        15.66,
		"ts_ratioamount":  0.19,
	}}
}
func (s *residualSourceStub) GetMoneyRankSina(sort string) []map[string]any {
	return []map[string]any{{
		"symbol":      "600519",
		"name":        "贵州茅台",
		"trade":       1688.88,
		"changeratio": 0.015,
		"turnover":    0.031,
		"amount":      5530000.0,
		"outamount":   1650000.0,
		"inamount":    3880000.0,
		"netamount":   2230000.0,
		"ratioamount": 0.015,
		"r0_out":      730000.0,
		"r0_in":       910000.0,
		"r0_net":      180000.0,
		"r0_ratio":    0.07,
		"r3_out":      300000.0,
		"r3_in":       210000.0,
		"r3_net":      -90000.0,
		"r3_ratio":    -0.03,
	}}
}
func (s *residualSourceStub) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	return []map[string]any{
		{"opendate": "2026-04-02", "trade": 18.90, "netamount": 260000.0, "r0_net": 135000.0},
		{"opendate": "2026-04-01", "trade": 18.32, "netamount": 250000.0, "r0_net": 120000.0},
	}
}

func TestService_LoadResidualMarketReads_NormalizesTypedContracts(t *testing.T) {
	svc := NewService(&residualSourceStub{})

	if got := svc.LoadGlobalIndexesReadable(30); got != "亚洲市场：上证指数 +1.23%" {
		t.Fatalf("unexpected readable indexes: %q", got)
	}

	industry := svc.LoadIndustryMoneyRanks("0", "netamount")
	if len(industry) != 1 || industry[0].Name != "机器人" || industry[0].TSSymbol != "300024" || industry[0].NetAmount != 1250000 {
		t.Fatalf("unexpected industry money rank: %#v", industry)
	}

	moneyRanks := svc.LoadMoneyRanks("netamount")
	if len(moneyRanks) != 1 || moneyRanks[0].Symbol != "600519" || moneyRanks[0].R0Net != 180000 {
		t.Fatalf("unexpected money rank data: %#v", moneyRanks)
	}

	trend := svc.LoadStockMoneyTrend("600519", 20)
	if len(trend) != 2 || trend[0].OpenDate != "2026-04-01" || trend[1].OpenDate != "2026-04-02" {
		t.Fatalf("unexpected stock money trend order: %#v", trend)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/contract ./backend/service/market -run "TestUserVisibleError_ErrorUsesMessage|TestNewUserVisibleErrorBuildsStablePayload|TestService_LoadResidualMarketReads_NormalizesTypedContracts" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing `backend/service/contract` package members and missing residual market service methods

- [ ] **Step 3: Add the shared error contract and residual market helper mappings**

```go
// D:\codex_work\go-stock\backend\service\contract\error_model.go
package contract

type Stage string

const (
	StageSource  Stage = "source"
	StageService Stage = "service"
	StageBridge  Stage = "bridge"
)

type UserVisibleError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Stage     Stage  `json:"stage"`
}

func NewUserVisibleError(code, message string, retryable bool, stage Stage) UserVisibleError {
	return UserVisibleError{
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Stage:     stage,
	}
}

func (e UserVisibleError) Error() string {
	return e.Message
}
```

```go
// Append to D:\codex_work\go-stock\backend\service\market\contracts.go
type IndustryMoneyRankRow struct {
	Category       string  `json:"category"`
	Name           string  `json:"name"`
	AvgChangeRatio float64 `json:"avg_changeratio"`
	InAmount       float64 `json:"inamount"`
	OutAmount      float64 `json:"outamount"`
	NetAmount      float64 `json:"netamount"`
	RatioAmount    float64 `json:"ratioamount"`
	TSName         string  `json:"ts_name"`
	TSSymbol       string  `json:"ts_symbol"`
	TSChangeRatio  float64 `json:"ts_changeratio"`
	TSTrade        float64 `json:"ts_trade"`
	TSRatioAmount  float64 `json:"ts_ratioamount"`
}

type MoneyRankRow struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Trade       float64 `json:"trade"`
	ChangeRatio float64 `json:"changeratio"`
	Turnover    float64 `json:"turnover"`
	Amount      float64 `json:"amount"`
	OutAmount   float64 `json:"outamount"`
	InAmount    float64 `json:"inamount"`
	NetAmount   float64 `json:"netamount"`
	RatioAmount float64 `json:"ratioamount"`
	R0Out       float64 `json:"r0_out"`
	R0In        float64 `json:"r0_in"`
	R0Net       float64 `json:"r0_net"`
	R0Ratio     float64 `json:"r0_ratio"`
	R3Out       float64 `json:"r3_out"`
	R3In        float64 `json:"r3_in"`
	R3Net       float64 `json:"r3_net"`
	R3Ratio     float64 `json:"r3_ratio"`
}

type StockMoneyTrendRow struct {
	OpenDate  string  `json:"opendate"`
	Trade     float64 `json:"trade"`
	NetAmount float64 `json:"netamount"`
	R0Net     float64 `json:"r0_net"`
}
```

```go
// Update D:\codex_work\go-stock\backend\source\marketnews\source.go
func (s *Source) GlobalStockIndexesReadable(crawlTimeout uint) string {
	return s.api.GlobalStockIndexesReadable(crawlTimeout)
}

func (s *Source) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return s.api.GetIndustryMoneyRankSina(fenlei, sort)
}

func (s *Source) GetMoneyRankSina(sort string) []map[string]any {
	return s.api.GetMoneyRankSina(sort)
}

func (s *Source) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	return s.api.GetStockMoneyTrendByDay(stockCode, days)
}
```

```go
// Replace the Source interface and append the residual methods in D:\codex_work\go-stock\backend\service\market\service.go
type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeOut uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
	GlobalStockIndexesReadable(crawlTimeout uint) string
	GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any
	GetMoneyRankSina(sort string) []map[string]any
	GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any
}

func (s *Service) LoadGlobalIndexesReadable(crawlTimeout uint) string {
	return strings.TrimSpace(s.source.GlobalStockIndexesReadable(crawlTimeout))
}

func (s *Service) LoadIndustryMoneyRanks(fenlei, sort string) []IndustryMoneyRankRow {
	raw := s.source.GetIndustryMoneyRankSina(fenlei, sort)
	if len(raw) == 0 {
		return []IndustryMoneyRankRow{}
	}

	result := make([]IndustryMoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, IndustryMoneyRankRow{
			Category:       convertor.ToString(row["category"]),
			Name:           convertor.ToString(row["name"]),
			AvgChangeRatio: toFloat(row["avg_changeratio"]),
			InAmount:       toFloat(row["inamount"]),
			OutAmount:      toFloat(row["outamount"]),
			NetAmount:      toFloat(row["netamount"]),
			RatioAmount:    toFloat(row["ratioamount"]),
			TSName:         convertor.ToString(row["ts_name"]),
			TSSymbol:       convertor.ToString(row["ts_symbol"]),
			TSChangeRatio:  toFloat(row["ts_changeratio"]),
			TSTrade:        toFloat(row["ts_trade"]),
			TSRatioAmount:  toFloat(row["ts_ratioamount"]),
		})
	}
	return result
}

func (s *Service) LoadMoneyRanks(sort string) []MoneyRankRow {
	raw := s.source.GetMoneyRankSina(sort)
	if len(raw) == 0 {
		return []MoneyRankRow{}
	}

	result := make([]MoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, MoneyRankRow{
			Symbol:      convertor.ToString(row["symbol"]),
			Name:        convertor.ToString(row["name"]),
			Trade:       toFloat(row["trade"]),
			ChangeRatio: toFloat(row["changeratio"]),
			Turnover:    toFloat(row["turnover"]),
			Amount:      toFloat(row["amount"]),
			OutAmount:   toFloat(row["outamount"]),
			InAmount:    toFloat(row["inamount"]),
			NetAmount:   toFloat(row["netamount"]),
			RatioAmount: toFloat(row["ratioamount"]),
			R0Out:       toFloat(row["r0_out"]),
			R0In:        toFloat(row["r0_in"]),
			R0Net:       toFloat(row["r0_net"]),
			R0Ratio:     toFloat(row["r0_ratio"]),
			R3Out:       toFloat(row["r3_out"]),
			R3In:        toFloat(row["r3_in"]),
			R3Net:       toFloat(row["r3_net"]),
			R3Ratio:     toFloat(row["r3_ratio"]),
		})
	}
	return result
}

func (s *Service) LoadStockMoneyTrend(stockCode string, days int) []StockMoneyTrendRow {
	raw := s.source.GetStockMoneyTrendByDay(stockCode, days)
	if len(raw) == 0 {
		return []StockMoneyTrendRow{}
	}

	result := make([]StockMoneyTrendRow, 0, len(raw))
	for idx := len(raw) - 1; idx >= 0; idx-- {
		row := raw[idx]
		result = append(result, StockMoneyTrendRow{
			OpenDate:  convertor.ToString(row["opendate"]),
			Trade:     toFloat(row["trade"]),
			NetAmount: toFloat(row["netamount"]),
			R0Net:     toFloat(row["r0_net"]),
		})
	}
	return result
}

func toFloat(value any) float64 {
	v, err := convertor.ToFloat(value)
	if err != nil {
		return 0
	}
	return v
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/contract ./backend/service/market -run "TestUserVisibleError_ErrorUsesMessage|TestNewUserVisibleErrorBuildsStablePayload|TestService_LoadResidualMarketReads_NormalizesTypedContracts" -count=1
```

Expected:

- command exits `0`
- both packages pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/contract/error_model.go backend/service/contract/error_model_test.go backend/source/marketnews/source.go backend/service/market/contracts.go backend/service/market/service.go backend/service/market/legacy_reads_test.go
git commit -m "refactor: close out residual market service reads"
```

### Task 2: Close Out Analysis Artifacts, Config Export, And Legacy Prompt Fallback

**Files:**
- Modify: `D:\codex_work\go-stock\backend\service\analysis\contracts.go`
- Modify: `D:\codex_work\go-stock\backend\service\analysis\service.go`
- Modify: `D:\codex_work\go-stock\backend\service\analysis\service_test.go`
- Modify: `D:\codex_work\go-stock\backend\service\config\store.go`
- Modify: `D:\codex_work\go-stock\backend\service\config\service.go`
- Modify: `D:\codex_work\go-stock\backend\service\config\service_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Modify: `D:\codex_work\go-stock\app_config_test.go`
- Create: `D:\codex_work\go-stock\app_closeout_test.go`

- [ ] **Step 1: Write the failing analysis/config closeout tests**

```go
// Replace the import block in D:\codex_work\go-stock\backend\service\analysis\service_test.go with this block, then append the tests below
import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go-stock/backend/models"
	contractservice "go-stock/backend/service/contract"
)

func TestService_GetResultArtifactBuildsStableShareAndMarkdownFields(t *testing.T) {
	results := &fakeResults{
		latest: &models.AIResponseResult{
			StockCode: "000001.SZ",
			StockName: "平安银行",
			Content:   strings.Repeat("分析结论", 30),
		},
	}
	results.latest.CreatedAt = time.Date(2026, 4, 6, 9, 30, 0, 0, time.Local)
	svc := NewService(&fakeStreams{}, results, &fakePrompts{})

	artifact, userErr := svc.GetResultArtifact(context.Background(), "000001.SZ", "")
	if userErr != nil {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
	if artifact.StockCode != "000001.SZ" || artifact.StockName != "平安银行" {
		t.Fatalf("unexpected artifact identity: %#v", artifact)
	}
	if artifact.AnalysisDate != "2026/04/06" {
		t.Fatalf("unexpected share analysis date: %q", artifact.AnalysisDate)
	}
	if artifact.MarkdownFilename != "平安银行[000001.SZ]AI分析结果_2026-04-06_09_30_00.md" {
		t.Fatalf("unexpected markdown filename: %q", artifact.MarkdownFilename)
	}
}

func TestService_GetResultArtifactReturnsUserVisibleErrorWhenResultMissing(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{latest: nil}, &fakePrompts{})

	artifact, userErr := svc.GetResultArtifact(context.Background(), "000001.SZ", "平安银行")
	if userErr == nil {
		t.Fatalf("expected user visible error, got artifact %#v", artifact)
	}
	if userErr.Code != "analysis.result_missing" || userErr.Stage != contractservice.StageService || userErr.Retryable {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
}
```

```go
// Extend fakeStore and append ExportConfig support in D:\codex_work\go-stock\backend\service\config\service_test.go
type fakeStore struct {
	cfg             *data.SettingConfig
	getConfigCtx    context.Context
	updateArg       *data.SettingConfig
	updateCtx       context.Context
	updateResult    string
	getPromptsCtx   context.Context
	templates       *[]models.PromptTemplate
	getPageCtx      context.Context
	page            *models.PromptTemplatePageData
	pageErr         error
	saveCtx         context.Context
	savedTemplate   models.PromptTemplate
	saveResult      string
	deleteCtx       context.Context
	deletedID       uint
	deleteResult    string
	exportCtx       context.Context
	exportResult    string
}

func (f *fakeStore) ExportConfig(ctx context.Context) string {
	f.exportCtx = ctx
	return f.exportResult
}

func TestService_ExportConfigDelegates(t *testing.T) {
	store := &fakeStore{
		exportResult: "{\"darkTheme\":true}",
	}
	svc := NewService(store)
	bg := context.Background()

	if got := svc.ExportConfig(bg); got != "{\"darkTheme\":true}" {
		t.Fatalf("unexpected export result: %q", got)
	}
	if store.exportCtx != bg {
		t.Fatalf("expected background context passed to export")
	}
}
```

```go
// D:\codex_work\go-stock\app_closeout_test.go
package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type closeoutMarketServiceStub struct {
	loadFeedsResult             marketservice.FeedSet
	refreshFeedResult           marketservice.Feed
	globalIndexesResult         marketservice.IndexSet
	industryRanksResult         []marketservice.IndustryRankEntry
	globalIndexesReadableResult string
	industryMoneyRanksResult    []marketservice.IndustryMoneyRankRow
	moneyRanksResult            []marketservice.MoneyRankRow
	stockMoneyTrendResult       []marketservice.StockMoneyTrendRow
}

func (s *closeoutMarketServiceStub) LoadFeeds() marketservice.FeedSet { return s.loadFeedsResult }
func (s *closeoutMarketServiceStub) RefreshFeed(source string) marketservice.Feed { return s.refreshFeedResult }
func (s *closeoutMarketServiceStub) LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet {
	return s.globalIndexesResult
}
func (s *closeoutMarketServiceStub) LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return s.industryRanksResult
}
func (s *closeoutMarketServiceStub) LoadGlobalIndexesReadable(crawlTimeout uint) string {
	return s.globalIndexesReadableResult
}
func (s *closeoutMarketServiceStub) LoadIndustryMoneyRanks(fenlei, sort string) []marketservice.IndustryMoneyRankRow {
	return s.industryMoneyRanksResult
}
func (s *closeoutMarketServiceStub) LoadMoneyRanks(sort string) []marketservice.MoneyRankRow {
	return s.moneyRanksResult
}
func (s *closeoutMarketServiceStub) LoadStockMoneyTrend(stockCode string, days int) []marketservice.StockMoneyTrendRow {
	return s.stockMoneyTrendResult
}

type closeoutAnalysisServiceStub struct {
	artifact    analysisservice.ResultArtifact
	artifactErr *contractservice.UserVisibleError
}

func (s *closeoutAnalysisServiceStub) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}
func (s *closeoutAnalysisServiceStub) StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}
func (s *closeoutAnalysisServiceStub) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {}
func (s *closeoutAnalysisServiceStub) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return nil
}
func (s *closeoutAnalysisServiceStub) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return &models.AIResponseResultPageData{}, nil
}
func (s *closeoutAnalysisServiceStub) DeleteResult(ctx context.Context, id uint) error { return nil }
func (s *closeoutAnalysisServiceStub) BatchDeleteResults(ctx context.Context, ids []uint) error { return nil }
func (s *closeoutAnalysisServiceStub) GetResultArtifact(ctx context.Context, stockCode, stockName string) (analysisservice.ResultArtifact, *contractservice.UserVisibleError) {
	return s.artifact, s.artifactErr
}

type closeoutConfigServiceStub struct {
	exportResult string
}

func (s *closeoutConfigServiceStub) GetConfig(ctx context.Context) *data.SettingConfig {
	return &data.SettingConfig{Settings: &data.Settings{}, AiConfigs: []*data.AIConfig{}}
}
func (s *closeoutConfigServiceStub) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string { return "更新成功" }
func (s *closeoutConfigServiceStub) GetAiConfigs(ctx context.Context) []*data.AIConfig { return []*data.AIConfig{} }
func (s *closeoutConfigServiceStub) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	return &empty
}
func (s *closeoutConfigServiceStub) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return &models.PromptTemplatePageData{}, nil
}
func (s *closeoutConfigServiceStub) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return "模板已保存"
}
func (s *closeoutConfigServiceStub) DeletePromptTemplate(ctx context.Context, id uint) string { return "模板已删除" }
func (s *closeoutConfigServiceStub) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	return "旧版Prompt已保存"
}
func (s *closeoutConfigServiceStub) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	return "旧版Prompt已删除"
}
func (s *closeoutConfigServiceStub) ExportConfig(ctx context.Context) string { return s.exportResult }

func TestApp_CloseoutMarketHelpersDelegateToMarketService(t *testing.T) {
	telegraphItem := &models.Telegraph{Title: "telegraph"}
	market := &closeoutMarketServiceStub{
		loadFeedsResult: marketservice.FeedSet{
			Telegraph: []*models.Telegraph{telegraphItem},
		},
		refreshFeedResult: marketservice.Feed{
			Source: "财联社电报",
			Items:  []*models.Telegraph{telegraphItem},
		},
		globalIndexesResult: marketservice.IndexSet{
			Common: []marketservice.GlobalIndexEntry{{Code: "000001", Name: "上证指数"}},
		},
		industryRanksResult:         []marketservice.IndustryRankEntry{{BoardCode: "BK001", BoardName: "算力"}},
		globalIndexesReadableResult: "亚洲市场：上证指数 +1.23%",
		industryMoneyRanksResult:    []marketservice.IndustryMoneyRankRow{{Name: "机器人", TSSymbol: "300024"}},
		moneyRanksResult:            []marketservice.MoneyRankRow{{Symbol: "600519", Name: "贵州茅台"}},
		stockMoneyTrendResult:       []marketservice.StockMoneyTrendRow{{OpenDate: "2026-04-01", Trade: 18.32}},
	}
	app := &App{marketReadService: market}

	if got := app.GetTelegraphList("财联社电报"); len(*got) != 1 || (*got)[0].Title != "telegraph" {
		t.Fatalf("unexpected telegraph list: %#v", got)
	}
	if got := app.ReFleshTelegraphList("财联社电报"); len(*got) != 1 || (*got)[0].Title != "telegraph" {
		t.Fatalf("unexpected refreshed telegraph list: %#v", got)
	}
	if got := app.GlobalStockIndexes(); !reflect.DeepEqual(got, market.globalIndexesResult) {
		t.Fatalf("unexpected indexes: %#v", got)
	}
	if got := app.GlobalStockIndexesReadable(); got != "亚洲市场：上证指数 +1.23%" {
		t.Fatalf("unexpected readable indexes: %q", got)
	}
	if got := app.GetIndustryRank("0", 10); len(got) != 1 || got[0].BoardName != "算力" {
		t.Fatalf("unexpected industry rank list: %#v", got)
	}
	if got := app.GetIndustryMoneyRankSina("0", "netamount"); len(got) != 1 || got[0].TSSymbol != "300024" {
		t.Fatalf("unexpected industry money ranks: %#v", got)
	}
	if got := app.GetMoneyRankSina("netamount"); len(got) != 1 || got[0].Symbol != "600519" {
		t.Fatalf("unexpected money ranks: %#v", got)
	}
	if got := app.GetStockMoneyTrendByDay("600519", 20); len(got) != 1 || got[0].OpenDate != "2026-04-01" {
		t.Fatalf("unexpected stock money trend: %#v", got)
	}
}

func TestApp_ShareAnalysisUsesArtifactUploader(t *testing.T) {
	analysis := &closeoutAnalysisServiceStub{
		artifact: analysisservice.ResultArtifact{
			StockCode:        "000001.SZ",
			StockName:        "平安银行",
			Content:          "长文本分析",
			AnalysisDate:     "2026/04/06",
			MarkdownFilename: "平安银行[000001.SZ]AI分析结果_2026-04-06_09_30_00.md",
		},
	}
	gotArtifact := analysisservice.ResultArtifact{}
	app := &App{
		ctx:                  context.Background(),
		analysisService:      analysis,
		shareAnalysisUploader: func(artifact analysisservice.ResultArtifact) (string, error) {
			gotArtifact = artifact
			return "shared-ok", nil
		},
	}

	if msg := app.ShareAnalysis("000001.SZ", "平安银行"); msg != "shared-ok" {
		t.Fatalf("unexpected share result: %q", msg)
	}
	if gotArtifact.StockCode != "000001.SZ" || gotArtifact.AnalysisDate != "2026/04/06" {
		t.Fatalf("unexpected uploaded artifact: %#v", gotArtifact)
	}
}

func TestApp_SaveAsMarkdownUsesArtifactAndWriter(t *testing.T) {
	analysis := &closeoutAnalysisServiceStub{
		artifact: analysisservice.ResultArtifact{
			StockCode:        "000001.SZ",
			StockName:        "平安银行",
			Content:          "# 分析结果",
			AnalysisDate:     "2026/04/06",
			MarkdownFilename: "平安银行[000001.SZ]AI分析结果_2026-04-06_09_30_00.md",
		},
	}
	var savedPath string
	var savedContent string
	app := &App{
		ctx:             context.Background(),
		analysisService: analysis,
		saveFileDialog: func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
			if options.DefaultFilename != "平安银行[000001.SZ]AI分析结果_2026-04-06_09_30_00.md" {
				t.Fatalf("unexpected default filename: %q", options.DefaultFilename)
			}
			return "D:\\temp\\analysis.md", nil
		},
		writeFile: func(name string, data []byte, perm os.FileMode) error {
			savedPath = name
			savedContent = string(data)
			return nil
		},
	}

	if msg := app.SaveAsMarkdown("000001.SZ", "平安银行"); msg != "已保存至：D:\\temp\\analysis.md" {
		t.Fatalf("unexpected save result: %q", msg)
	}
	if savedPath != "D:\\temp\\analysis.md" || savedContent != "# 分析结果" {
		t.Fatalf("unexpected saved file state: path=%q content=%q", savedPath, savedContent)
	}
}

func TestApp_ExportConfigUsesConfigServiceExport(t *testing.T) {
	var savedPath string
	var savedContent string
	app := &App{
		ctx:           context.Background(),
		configService: &closeoutConfigServiceStub{exportResult: "{\"darkTheme\":true}"},
		saveFileDialog: func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
			if options.DefaultFilename != "config.json" {
				t.Fatalf("unexpected export filename: %q", options.DefaultFilename)
			}
			return "D:\\temp\\config.json", nil
		},
		writeFile: func(name string, data []byte, perm os.FileMode) error {
			savedPath = name
			savedContent = string(data)
			return nil
		},
	}

	if msg := app.ExportConfig(); msg != "导出成功:D:\\temp\\config.json" {
		t.Fatalf("unexpected export result: %q", msg)
	}
	if savedPath != "D:\\temp\\config.json" || savedContent != "{\"darkTheme\":true}" {
		t.Fatalf("unexpected exported content: path=%q content=%q", savedPath, savedContent)
	}
}
```

```go
// Replace the legacy fallback block in D:\codex_work\go-stock\app_config_test.go with this test
func TestApp_PromptMethodsReturnFallbackDefaultsWhenConfigServiceNil(t *testing.T) {
	app := &App{
		ctx:           context.Background(),
		configService: nil,
	}

	templates := app.GetPromptTemplates("legacy", "模型系统Prompt")
	if templates == nil || len(*templates) != 0 {
		t.Fatalf("expected empty templates when configService is nil, got %#v", templates)
	}

	if msg := app.AddPrompt(models.Prompt{ID: 9, Name: "legacy"}); msg != "保存失败" {
		t.Fatalf("expected save failure when configService is nil, got %q", msg)
	}

	if msg := app.DelPrompt(77); msg != "删除失败" {
		t.Fatalf("expected delete failure when configService is nil, got %q", msg)
	}

	page := app.GetPromptTemplateList(models.PromptTemplateQuery{Page: 1, PageSize: 10})
	if page == nil || len(page.List) != 0 || page.Total != 0 {
		t.Fatalf("expected empty prompt page when configService is nil, got %#v", page)
	}

	if msg := app.AddPromptTemplate(models.PromptTemplate{ID: 1}); msg != "保存失败" {
		t.Fatalf("expected add template failure when configService is nil, got %q", msg)
	}

	if msg := app.UpdatePromptTemplate(models.PromptTemplate{ID: 1}); msg != "保存失败" {
		t.Fatalf("expected update template failure when configService is nil, got %q", msg)
	}

	if msg := app.DeletePromptTemplate(1); msg != "删除失败" {
		t.Fatalf("expected delete template failure when configService is nil, got %q", msg)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis ./backend/service/config . -run "TestService_(GetResultArtifactBuildsStableShareAndMarkdownFields|GetResultArtifactReturnsUserVisibleErrorWhenResultMissing|ExportConfigDelegates)|TestApp_(CloseoutMarketHelpersDelegateToMarketService|ShareAnalysisUsesArtifactUploader|SaveAsMarkdownUsesArtifactAndWriter|ExportConfigUsesConfigServiceExport|PromptMethodsReturnFallbackDefaultsWhenConfigServiceNil)" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing `GetResultArtifact`, missing `ExportConfig`, or still hits the legacy prompt fallback path

- [ ] **Step 3: Implement artifact/config closeout and delete the legacy prompt fallback**

```go
// Append to D:\codex_work\go-stock\backend\service\analysis\contracts.go
type ResultArtifact struct {
	StockCode        string `json:"stockCode"`
	StockName        string `json:"stockName"`
	Content          string `json:"content"`
	AnalysisDate     string `json:"analysisDate"`
	MarkdownFilename string `json:"markdownFilename"`
}
```

```go
// Update the imports and append GetResultArtifact in D:\codex_work\go-stock\backend\service\analysis\service.go
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-stock/backend/logger"
	"go-stock/backend/models"
	contractservice "go-stock/backend/service/contract"
)

func (s *Service) GetResultArtifact(ctx context.Context, stockCode, stockName string) (ResultArtifact, *contractservice.UserVisibleError) {
	result := s.results.GetLatestResult(ctx, stockCode)
	if result == nil || len(strings.TrimSpace(result.Content)) < 100 {
		err := contractservice.NewUserVisibleError("analysis.result_missing", "分析结果异常", false, contractservice.StageService)
		return ResultArtifact{}, &err
	}

	resolvedName := strings.TrimSpace(stockName)
	if resolvedName == "" {
		resolvedName = strings.TrimSpace(result.StockName)
	}
	if resolvedName == "" {
		resolvedName = stockCode
	}

	return ResultArtifact{
		StockCode:        stockCode,
		StockName:        resolvedName,
		Content:          result.Content,
		AnalysisDate:     result.CreatedAt.Format("2006/01/02"),
		MarkdownFilename: fmt.Sprintf("%s[%s]AI分析结果_%s.md", resolvedName, stockCode, result.CreatedAt.Format("2006-01-02_15_04_05")),
	}, nil
}
```

```go
// Add export support to D:\codex_work\go-stock\backend\service\config\store.go
func (s *DataStore) ExportConfig(ctx context.Context) string {
	_ = ctx
	return data.NewSettingsApi().Export()
}
```

```go
// Extend the Store interface and append ExportConfig in D:\codex_work\go-stock\backend\service\config\service.go
type Store interface {
	GetConfig(ctx context.Context) *data.SettingConfig
	UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
	ExportConfig(ctx context.Context) string
}

func (s *Service) ExportConfig(ctx context.Context) string {
	return s.store.ExportConfig(ctx)
}
```

```go
// Replace the closeout helper declarations near the top of D:\codex_work\go-stock\app.go
type marketResidualReadService interface {
	LoadGlobalIndexesReadable(crawlTimeout uint) string
	LoadIndustryMoneyRanks(fenlei, sort string) []marketservice.IndustryMoneyRankRow
	LoadMoneyRanks(sort string) []marketservice.MoneyRankRow
	LoadStockMoneyTrend(stockCode string, days int) []marketservice.StockMoneyTrendRow
}

type analysisArtifactService interface {
	GetResultArtifact(ctx context.Context, stockCode, stockName string) (analysisservice.ResultArtifact, *contractservice.UserVisibleError)
}

type configExportService interface {
	ExportConfig(ctx context.Context) string
}

func (a *App) residualMarketReads() marketResidualReadService {
	if a.marketReadService == nil {
		return nil
	}
	service, _ := a.marketReadService.(marketResidualReadService)
	return service
}

func (a *App) artifactService() analysisArtifactService {
	if a.analysisService == nil {
		return nil
	}
	service, _ := a.analysisService.(analysisArtifactService)
	return service
}

func (a *App) exportConfigSource() configExportService {
	if a.configService == nil {
		return nil
	}
	service, _ := a.configService.(configExportService)
	return service
}
```

```go
// Replace the App field block and NewApp wiring in D:\codex_work\go-stock\app.go
type App struct {
	ctx                   context.Context
	cache                 *freecache.Cache
	cron                  *cron.Cron
	cronEntrys            map[string]cron.EntryID
	cronEntrysMu          sync.Mutex
	AiTools               []data.Tool
	SponsorInfo           map[string]any
	VipLevel              int64
	summaryMu             sync.Mutex
	summaryCancel         context.CancelFunc
	agentMu               sync.Mutex
	agentCancel           context.CancelFunc
	stockAlertMu          sync.Mutex
	stockAlertLastSent    map[string]time.Time
	priceAtAlertReset     map[string]float64
	marketReadService     marketReadService
	analysisService       analysisService
	configService         configService
	taskService           taskService
	saveFileDialog        func(ctx context.Context, options runtime.SaveDialogOptions) (string, error)
	writeFile             func(name string, data []byte, perm os.FileMode) error
	shareAnalysisUploader func(artifact analysisservice.ResultArtifact) (string, error)
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
	analysisSvc := analysisservice.NewService(analysisProvider, analysisStore, analysisStore)
	configSvc := configservice.NewService(configservice.NewStore())
	app := &App{
		cache:                 cache,
		cron:                  c,
		cronEntrys:            make(map[string]cron.EntryID),
		AiTools:               tools,
		stockAlertLastSent:    make(map[string]time.Time),
		priceAtAlertReset:     make(map[string]float64),
		marketReadService:     marketservice.NewService(marketsource.NewSource()),
		analysisService:       analysisSvc,
		configService:         configSvc,
		saveFileDialog:        runtime.SaveFileDialog,
		writeFile:             os.WriteFile,
		shareAnalysisUploader: uploadSharedAnalysis,
	}
	app.taskService = taskservice.NewService(taskservice.NewStore(), &appTaskScheduler{app: app}, func() context.Context {
		return app.ctx
	})
	return app
}
```

```go
// Add this helper near the closeout methods in D:\codex_work\go-stock\app.go
func uploadSharedAnalysis(artifact analysisservice.ResultArtifact) (string, error) {
	response, err := resty.New().SetHeader("ua-x", "go-stock").R().SetFormData(map[string]string{
		"text":         artifact.Content,
		"stockCode":    artifact.StockCode,
		"stockName":    artifact.StockName,
		"analysisTime": artifact.AnalysisDate,
	}).Post("http://go-stock.sparkmemory.top:16688/upload")
	if err != nil {
		return "", err
	}
	return response.String(), nil
}
```

```go
// Replace the residual market helpers and analysis/config artifact methods in D:\codex_work\go-stock\app.go
func feedItemsForSource(feeds marketservice.FeedSet, source string) []*models.Telegraph {
	switch source {
	case "财联社电报":
		return append([]*models.Telegraph(nil), feeds.Telegraph...)
	case "新浪财经":
		return append([]*models.Telegraph(nil), feeds.Sina...)
	case "外媒":
		return append([]*models.Telegraph(nil), feeds.Foreign...)
	default:
		return []*models.Telegraph{}
	}
}

func (a *App) GetTelegraphList(source string) *[]*models.Telegraph {
	items := feedItemsForSource(a.marketReadService.LoadFeeds(), source)
	return &items
}

func (a *App) ReFleshTelegraphList(source string) *[]*models.Telegraph {
	feed := a.marketReadService.RefreshFeed(source)
	items := append([]*models.Telegraph(nil), feed.Items...)
	return &items
}

func (a *App) GlobalStockIndexes() marketservice.IndexSet {
	return a.marketReadService.LoadGlobalIndexes(30)
}

func (a *App) GlobalStockIndexesReadable() string {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadGlobalIndexesReadable(30)
	}
	return ""
}

func (a *App) GetIndustryRank(sort string, cnt int) []marketservice.IndustryRankEntry {
	return a.marketReadService.LoadIndustryRanks(sort, cnt)
}

func (a *App) GetIndustryMoneyRankSina(fenlei, sort string) []marketservice.IndustryMoneyRankRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadIndustryMoneyRanks(fenlei, sort)
	}
	return []marketservice.IndustryMoneyRankRow{}
}

func (a *App) GetMoneyRankSina(sort string) []marketservice.MoneyRankRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadMoneyRanks(sort)
	}
	return []marketservice.MoneyRankRow{}
}

func (a *App) GetStockMoneyTrendByDay(stockCode string, days int) []marketservice.StockMoneyTrendRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadStockMoneyTrend(stockCode, days)
	}
	return []marketservice.StockMoneyTrendRow{}
}

func (a *App) ExportConfig() string {
	exporter := a.exportConfigSource()
	if exporter == nil {
		return "导出失败"
	}

	config := exporter.ExportConfig(a.ctx)
	file, err := a.saveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "导出配置文件",
		CanCreateDirectories: true,
		DefaultFilename:      "config.json",
	})
	if err != nil {
		appError("export-config", "config.export_dialog_failed", "export config dialog failed", logger.Err(err))
		return err.Error()
	}
	if err := a.writeFile(file, []byte(config), os.ModePerm); err != nil {
		appError("export-config", "config.export_write_failed", "write exported config failed", logger.String("file", file), logger.Err(err))
		return err.Error()
	}
	return "导出成功:" + file
}

func (a *App) ShareAnalysis(stockCode, stockName string) string {
	artifacts := a.artifactService()
	if artifacts == nil {
		return "分析结果异常"
	}
	artifact, userErr := artifacts.GetResultArtifact(a.ctx, stockCode, stockName)
	if userErr != nil {
		return userErr.Message
	}
	msg, err := a.shareAnalysisUploader(artifact)
	if err != nil {
		return err.Error()
	}
	return msg
}

func (a *App) SaveAsMarkdown(stockCode, stockName string) string {
	artifacts := a.artifactService()
	if artifacts == nil {
		return "分析结果异常,无法保存。"
	}
	artifact, userErr := artifacts.GetResultArtifact(a.ctx, stockCode, stockName)
	if userErr != nil {
		return userErr.Message + ",无法保存。"
	}
	file, err := a.saveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存为Markdown",
		DefaultFilename: artifact.MarkdownFilename,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown",
				Pattern:     "*.md;*.markdown",
			},
		},
	})
	if err != nil {
		return err.Error()
	}
	if err := a.writeFile(file, []byte(artifact.Content), 0644); err != nil {
		return err.Error()
	}
	return "已保存至：" + file
}

func (a *App) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	if a.configService == nil {
		empty := []models.PromptTemplate{}
		return &empty
	}
	return a.configService.GetPromptTemplates(a.ctx, name, promptType)
}

func (a *App) AddPrompt(prompt models.Prompt) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SaveLegacyPrompt(a.ctx, prompt)
}

func (a *App) DelPrompt(id uint) string {
	if a.configService == nil {
		return "删除失败"
	}
	return a.configService.DeleteLegacyPrompt(a.ctx, id)
}
```

```go
// Replace the prompt page methods in D:\codex_work\go-stock\app_common.go
func (a *App) GetPromptTemplateList(query models.PromptTemplateQuery) *models.PromptTemplatePageData {
	if a.configService == nil {
		return &models.PromptTemplatePageData{}
	}
	page, err := a.configService.GetPromptTemplatePage(a.ctx, query)
	if err != nil {
		return &models.PromptTemplatePageData{}
	}
	return page
}

func (a *App) AddPromptTemplate(template models.PromptTemplate) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) UpdatePromptTemplate(template models.PromptTemplate) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) DeletePromptTemplate(id uint) string {
	if a.configService == nil {
		return "删除失败"
	}
	return a.configService.DeletePromptTemplate(a.ctx, id)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis ./backend/service/config . -run "TestService_(GetResultArtifactBuildsStableShareAndMarkdownFields|GetResultArtifactReturnsUserVisibleErrorWhenResultMissing|ExportConfigDelegates)|TestApp_(CloseoutMarketHelpersDelegateToMarketService|ShareAnalysisUsesArtifactUploader|SaveAsMarkdownUsesArtifactAndWriter|ExportConfigUsesConfigServiceExport|PromptMethodsReturnFallbackDefaultsWhenConfigServiceNil)" -count=1
```

Expected:

- command exits `0`
- analysis service, config service, and root closeout tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/analysis/contracts.go backend/service/analysis/service.go backend/service/analysis/service_test.go backend/service/config/store.go backend/service/config/service.go backend/service/config/service_test.go app.go app_common.go app_config_test.go app_closeout_test.go
git commit -m "refactor: close out analysis artifacts and prompt fallback"
```

### Task 3: Move Notification And Stock AI Cron Helpers Behind Service Boundaries

**Files:**
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter.go`
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter_windows.go`
- Create: `D:\codex_work\go-stock\backend\source\notification\adapter_other.go`
- Create: `D:\codex_work\go-stock\backend\service\notification\service.go`
- Create: `D:\codex_work\go-stock\backend\service\notification\service_test.go`
- Create: `D:\codex_work\go-stock\backend\source\watchlist\store.go`
- Create: `D:\codex_work\go-stock\backend\service\watchlist\service.go`
- Create: `D:\codex_work\go-stock\backend\service\watchlist\service_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_closeout_test.go`

- [ ] **Step 1: Write the failing notification/watchlist tests**

```go
// D:\codex_work\go-stock\backend\service\notification\service_test.go
package notification

import (
	"testing"

	"go-stock/backend/data"
)

type fakeCache struct {
	ttl    map[string]int64
	setTTL map[string]int
}

func (f *fakeCache) TTL(key []byte) (int64, error) {
	if f.ttl == nil {
		return 0, nil
	}
	return f.ttl[string(key)], nil
}

func (f *fakeCache) Set(key []byte, value []byte, expireSeconds int) error {
	if f.setTTL == nil {
		f.setTTL = map[string]int{}
	}
	f.setTTL[string(key)] = expireSeconds
	return nil
}

type fakeAdapter struct {
	dingResult   string
	localTitle   string
	localContent string
	stockInfo    *data.StockInfo
}

func (f *fakeAdapter) SendDingTalk(message string) string { return f.dingResult }
func (f *fakeAdapter) SendLocal(title, content string) bool {
	f.localTitle = title
	f.localContent = content
	return true
}
func (f *fakeAdapter) LoadStockInfo(stockCode string) *data.StockInfo { return f.stockInfo }

func TestService_SendDingTalkSkipsWhileTTLActive(t *testing.T) {
	cache := &fakeCache{ttl: map[string]int64{"sh600519": 30}}
	adapter := &fakeAdapter{dingResult: "发送成功"}
	svc := NewService(cache, adapter)

	if got := svc.SendDingTalk("body", "sh600519"); got != "" {
		t.Fatalf("expected empty result when ttl is active, got %q", got)
	}
	if _, exists := cache.setTTL["sh600519"]; exists {
		t.Fatalf("did not expect cache reset when ttl is already active")
	}
}

func TestService_SendTypedDeliversLocalAndDingTalk(t *testing.T) {
	cache := &fakeCache{}
	adapter := &fakeAdapter{
		dingResult: "发送钉钉消息成功",
		stockInfo: &data.StockInfo{
			Name:     "平安银行",
			Price:    "12.34",
			PreClose: "12.00",
			Date:     "2026-04-06",
			Time:     "10:00:00",
		},
	}
	svc := NewService(cache, adapter)

	result := svc.SendTyped("body", "sz000001", 1)
	if result.DingResult != "发送钉钉消息成功" {
		t.Fatalf("unexpected ding result: %#v", result)
	}
	if adapter.localTitle != "涨跌报警" {
		t.Fatalf("unexpected local title: %q", adapter.localTitle)
	}
	if adapter.localContent == "" || result.EventContent == "" {
		t.Fatalf("expected non-empty local/event content: adapter=%q result=%#v", adapter.localContent, result)
	}
	if cache.setTTL["sz000001"] != 300 {
		t.Fatalf("unexpected ttl seconds: %d", cache.setTTL["sz000001"])
	}
}
```

```go
// D:\codex_work\go-stock\backend\service\watchlist\service_test.go
package watchlist

import (
	"context"
	"testing"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
)

type fakeStore struct {
	savedCronText string
	savedCode     string
	follow        data.FollowedStock
	list          []data.FollowedStock
}

func (f *fakeStore) SaveStockAICron(ctx context.Context, cronText, stockCode string) {
	f.savedCronText = cronText
	f.savedCode = stockCode
}

func (f *fakeStore) GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock {
	if f.follow.StockCode == stockCode {
		return f.follow
	}
	return data.FollowedStock{}
}

func (f *fakeStore) ListFollowedStocks(ctx context.Context) []data.FollowedStock {
	return append([]data.FollowedStock(nil), f.list...)
}

type fakeAnalyzer struct {
	request  analysisservice.StockRequest
	saveCode string
	saveName string
	saveBody string
	saveChat string
	saveQ    string
	saveAIID int
	stream   []analysisservice.StreamChunk
}

func (f *fakeAnalyzer) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	f.request = request
	ch := make(chan analysisservice.StreamChunk, len(f.stream))
	for _, item := range f.stream {
		ch <- item
	}
	close(ch)
	return ch
}

func (f *fakeAnalyzer) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.saveCode = stockCode
	f.saveName = stockName
	f.saveBody = result
	f.saveChat = chatID
	f.saveQ = question
	f.saveAIID = aiConfigID
}

func TestService_SaveStockAICronNormalizesUsCode(t *testing.T) {
	store := &fakeStore{
		follow: data.FollowedStock{StockCode: "usaapl", Name: "Apple", AiConfigId: 7},
	}
	svc := NewService(store, &fakeAnalyzer{})

	result, userErr := svc.SaveStockAICron(context.Background(), "0 */5 * * * *", "gb_aapl")
	if userErr != nil {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
	if store.savedCode != "usaapl" {
		t.Fatalf("expected normalized stock code, got %q", store.savedCode)
	}
	if result.StockCode != "usaapl" || result.Name != "Apple" || result.Cron != "0 */5 * * * *" {
		t.Fatalf("unexpected save result: %#v", result)
	}
}

func TestService_ListScheduledStocksFiltersBlankCron(t *testing.T) {
	cronText := "0 */10 * * * *"
	store := &fakeStore{
		list: []data.FollowedStock{
			{StockCode: "sz000001", Name: "平安银行", Cron: &cronText, AiConfigId: 1},
			{StockCode: "sh600519", Name: "贵州茅台", Cron: nil, AiConfigId: 2},
		},
	}
	svc := NewService(store, &fakeAnalyzer{})

	result := svc.ListScheduledStocks(context.Background())
	if len(result) != 1 || result[0].StockCode != "sz000001" {
		t.Fatalf("unexpected scheduled stock list: %#v", result)
	}
}

func TestService_RunScheduledAnalysisSavesCollectedResult(t *testing.T) {
	store := &fakeStore{
		follow: data.FollowedStock{StockCode: "sz000001", Name: "平安银行", AiConfigId: 3},
	}
	analyzer := &fakeAnalyzer{
		stream: []analysisservice.StreamChunk{
			{ExtraContent: "第一段", ChatID: "chat-1", Question: "默认问题"},
			{Content: "第二段"},
		},
	}
	svc := NewService(store, analyzer)

	result, userErr := svc.RunScheduledAnalysis(context.Background(), "sz000001")
	if userErr != nil {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
	if result.StockCode != "sz000001" || result.Name != "平安银行" {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if analyzer.saveCode != "sz000001" || analyzer.saveAIID != 3 || analyzer.saveBody != "第一段\n第二段" {
		t.Fatalf("unexpected save delegation: %#v", analyzer)
	}
}

func TestService_RunScheduledAnalysisReturnsUserVisibleErrorWhenStockMissing(t *testing.T) {
	svc := NewService(&fakeStore{}, &fakeAnalyzer{})

	_, userErr := svc.RunScheduledAnalysis(context.Background(), "gb_aapl")
	if userErr == nil {
		t.Fatal("expected user visible error for missing followed stock")
	}
	if userErr.Code != "watchlist.stock_not_followed" || userErr.Stage != contractservice.StageService {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
}
```

```go
// Replace the import block in D:\codex_work\go-stock\app_closeout_test.go with this block, then append the notification/watchlist stubs and tests below
import (
	"context"
	"os"
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"
	watchlistservice "go-stock/backend/service/watchlist"

	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type closeoutNotificationServiceStub struct {
	sendResult    string
	typedResult   notificationservice.Delivery
	lastMessage   string
	lastStockCode string
	lastType      int
}

func (s *closeoutNotificationServiceStub) SendDingTalk(message, stockCode string) string {
	s.lastMessage = message
	s.lastStockCode = stockCode
	return s.sendResult
}

func (s *closeoutNotificationServiceStub) SendTyped(message, stockCode string, msgType int) notificationservice.Delivery {
	s.lastMessage = message
	s.lastStockCode = stockCode
	s.lastType = msgType
	return s.typedResult
}

type closeoutWatchlistServiceStub struct {
	saveCronText  string
	saveStockCode string
	saveResult    watchlistservice.ScheduledStock
	listResult    []watchlistservice.ScheduledStock
	runStockCode  string
	runResult     watchlistservice.ScheduledStock
}

func (s *closeoutWatchlistServiceStub) SaveStockAICron(ctx context.Context, cronText, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	s.saveCronText = cronText
	s.saveStockCode = stockCode
	return s.saveResult, nil
}

func (s *closeoutWatchlistServiceStub) ListScheduledStocks(ctx context.Context) []watchlistservice.ScheduledStock {
	return append([]watchlistservice.ScheduledStock(nil), s.listResult...)
}

func (s *closeoutWatchlistServiceStub) RunScheduledAnalysis(ctx context.Context, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	s.runStockCode = stockCode
	return s.runResult, nil
}

func TestApp_NotificationHelpersDelegateToService(t *testing.T) {
	notification := &closeoutNotificationServiceStub{
		sendResult: "普通通知成功",
		typedResult: notificationservice.Delivery{
			DingResult:   "类型通知成功",
			EventTitle:   "涨跌报警",
			EventContent: "[平安银行] 12.34 2.83% 2026-04-06 10:00:00",
		},
	}
	var eventName string
	var eventPayload any
	app := &App{
		ctx:                 context.Background(),
		notificationService: notification,
		emitEvent: func(ctx context.Context, name string, data ...interface{}) {
			eventName = name
			if len(data) > 0 {
				eventPayload = data[0]
			}
		},
	}

	if got := app.SendDingDingMessage("body-1", "sz000001"); got != "普通通知成功" {
		t.Fatalf("unexpected normal notification result: %q", got)
	}
	if notification.lastMessage != "body-1" || notification.lastStockCode != "sz000001" {
		t.Fatalf("unexpected normal notification args: %#v", notification)
	}

	if got := app.SendDingDingMessageByType("body-2", "sz000001", 1); got != "类型通知成功" {
		t.Fatalf("unexpected typed notification result: %q", got)
	}
	if eventName != "newsPush" {
		t.Fatalf("expected newsPush event, got %q", eventName)
	}
	payload, ok := eventPayload.(map[string]any)
	if !ok || payload["content"] != "[平安银行] 12.34 2.83% 2026-04-06 10:00:00" {
		t.Fatalf("unexpected event payload: %#v", eventPayload)
	}
}

func TestApp_SetStockAICronRegistersAndRestoresJobsThroughWatchlistService(t *testing.T) {
	cronScheduler := cron.New(cron.WithSeconds())
	cronScheduler.Start()
	t.Cleanup(func() { cronScheduler.Stop() })

	watchlist := &closeoutWatchlistServiceStub{
		saveResult: watchlistservice.ScheduledStock{
			StockCode:  "usaapl",
			Name:       "Apple",
			Cron:       "0 */5 * * * *",
			AIConfigID: 7,
		},
		listResult: []watchlistservice.ScheduledStock{
			{StockCode: "sz000001", Name: "平安银行", Cron: "0 */10 * * * *", AIConfigID: 3},
		},
		runResult: watchlistservice.ScheduledStock{
			StockCode:  "usaapl",
			Name:       "Apple",
			Cron:       "0 */5 * * * *",
			AIConfigID: 7,
		},
	}
	var events []string
	app := &App{
		ctx:              context.Background(),
		cron:             cronScheduler,
		cronEntrys:       make(map[string]cron.EntryID),
		watchlistService: watchlist,
		emitEvent: func(ctx context.Context, name string, data ...interface{}) {
			if len(data) == 0 {
				return
			}
			if text, ok := data[0].(string); ok {
				events = append(events, text)
			}
		},
	}

	app.SetStockAICron("0 */5 * * * *", "gb_aapl")
	if watchlist.saveCronText != "0 */5 * * * *" || watchlist.saveStockCode != "gb_aapl" {
		t.Fatalf("unexpected watchlist save args: %#v", watchlist)
	}
	if _, exists := app.getCronEntry("usaapl"); !exists {
		t.Fatal("expected usaapl cron entry after SetStockAICron")
	}

	app.restoreStockAICronSchedules()
	if _, exists := app.getCronEntry("sz000001"); !exists {
		t.Fatal("expected restored sz000001 cron entry")
	}

	job := app.buildStockAICronJob("usaapl")
	job()
	if watchlist.runStockCode != "usaapl" {
		t.Fatalf("unexpected scheduled run stock code: %q", watchlist.runStockCode)
	}
	if len(events) < 2 || events[0] != "开始自动分析Apple_usaapl" || events[1] != "AI分析完成：Apple_usaapl" {
		t.Fatalf("unexpected event log: %#v", events)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/notification ./backend/service/watchlist . -run "Test(Service_(SendDingTalkSkipsWhileTTLActive|SendTypedDeliversLocalAndDingTalk|SaveStockAICronNormalizesUsCode|ListScheduledStocksFiltersBlankCron|RunScheduledAnalysisSavesCollectedResult|RunScheduledAnalysisReturnsUserVisibleErrorWhenStockMissing)|App_(NotificationHelpersDelegateToService|SetStockAICronRegistersAndRestoresJobsThroughWatchlistService))" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing `notification` / `watchlist` service packages or still uses old `App` helper implementations

- [ ] **Step 3: Implement notification/watchlist services and wire `app.go` to them**

```go
// D:\codex_work\go-stock\backend\source\notification\adapter.go
package notification

import (
	"go-stock/backend/data"
	"go-stock/backend/db"
)

type Adapter struct{}

func NewAdapter() *Adapter { return &Adapter{} }

func (a *Adapter) SendDingTalk(message string) string {
	return data.NewDingDingAPI().SendDingDingMessage(message)
}

func (a *Adapter) LoadStockInfo(stockCode string) *data.StockInfo {
	stockInfo := &data.StockInfo{}
	db.Dao.Model(stockInfo).Where("code = ?", stockCode).First(stockInfo)
	return stockInfo
}

func (a *Adapter) SendLocal(title, content string) bool {
	return sendLocalNotification("go-stock消息通知", title, content, "")
}
```

```go
// D:\codex_work\go-stock\backend\source\notification\adapter_windows.go
//go:build windows || darwin

package notification

import "go-stock/backend/data"

func sendLocalNotification(appID, title, content, icon string) bool {
	return data.NewAlertWindowsApi(appID, title, content, icon).SendNotification()
}
```

```go
// D:\codex_work\go-stock\backend\source\notification\adapter_other.go
//go:build !windows && !darwin

package notification

func sendLocalNotification(appID, title, content, icon string) bool {
	return false
}
```

```go
// D:\codex_work\go-stock\backend\service\notification\service.go
package notification

import (
	"strings"

	"go-stock/backend/data"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
)

type Cache interface {
	TTL(key []byte) (int64, error)
	Set(key []byte, value []byte, expireSeconds int) error
}

type Adapter interface {
	SendDingTalk(message string) string
	SendLocal(title, content string) bool
	LoadStockInfo(stockCode string) *data.StockInfo
}

type Delivery struct {
	DingResult   string `json:"dingResult"`
	EventTitle   string `json:"eventTitle"`
	EventContent string `json:"eventContent"`
}

type Service struct {
	cache   Cache
	adapter Adapter
}

func NewService(cache Cache, adapter Adapter) *Service {
	if cache == nil {
		panic("notification: cache dependency is required")
	}
	if adapter == nil {
		panic("notification: adapter dependency is required")
	}
	return &Service{cache: cache, adapter: adapter}
}

func (s *Service) SendDingTalk(message, stockCode string) string {
	ttl, _ := s.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return ""
	}
	if err := s.cache.Set([]byte(stockCode), []byte("1"), 60*5); err != nil {
		return ""
	}
	return s.adapter.SendDingTalk(message)
}

func (s *Service) SendTyped(message, stockCode string, msgType int) Delivery {
	ttl, _ := s.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return Delivery{}
	}
	if err := s.cache.Set([]byte(stockCode), []byte("1"), ttlForMessageType(msgType)); err != nil {
		return Delivery{}
	}

	content := formatNotificationContent(s.adapter.LoadStockInfo(stockCode))
	if strings.TrimSpace(content) != "" {
		s.adapter.SendLocal(messageTypeName(msgType), content)
	}

	return Delivery{
		DingResult:   s.adapter.SendDingTalk(message),
		EventTitle:   messageTypeName(msgType),
		EventContent: content,
	}
}

func ttlForMessageType(msgType int) int {
	switch msgType {
	case 1, 4, 5:
		return 60 * 5
	case 2, 3:
		return 60 * 30
	default:
		return 60 * 5
	}
}

func messageTypeName(msgType int) string {
	switch msgType {
	case 1:
		return "涨跌报警"
	case 2:
		return "股价报警"
	case 3:
		return "成本价报警"
	case 4:
		return "止盈报警"
	case 5:
		return "止损报警"
	default:
		return "未知类型"
	}
}

func formatNotificationContent(stockInfo *data.StockInfo) string {
	if stockInfo == nil {
		return ""
	}
	price, err := convertor.ToFloat(stockInfo.Price)
	if err != nil {
		price = 0
	}
	preClose, err := convertor.ToFloat(stockInfo.PreClose)
	if err != nil {
		preClose = 0
	}
	changePercent := float64(0)
	if preClose > 0 {
		changePercent = mathutil.RoundToFloat(((price-preClose)/preClose)*100, 2)
	}
	return "[" + stockInfo.Name + "] " + stockInfo.Price + " " + convertor.ToString(changePercent) + "% " + stockInfo.Date + " " + stockInfo.Time
}
```

```go
// D:\codex_work\go-stock\backend\source\watchlist\store.go
package watchlist

import (
	"context"

	"go-stock/backend/data"
)

type Store struct{}

func NewStore() *Store { return &Store{} }

func (s *Store) SaveStockAICron(ctx context.Context, cronText, stockCode string) {
	_ = ctx
	data.NewStockDataApi().SetStockAICron(cronText, stockCode)
}

func (s *Store) GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock {
	_ = ctx
	return data.NewStockDataApi().GetFollowedStockByStockCode(stockCode)
}

func (s *Store) ListFollowedStocks(ctx context.Context) []data.FollowedStock {
	_ = ctx
	result := data.NewStockDataApi().GetFollowList(0)
	if result == nil {
		return []data.FollowedStock{}
	}
	return append([]data.FollowedStock(nil), (*result)...)
}
```

```go
// D:\codex_work\go-stock\backend\service\watchlist\service.go
package watchlist

import (
	"context"
	"strings"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
)

type Store interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string)
	GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock
	ListFollowedStocks(ctx context.Context) []data.FollowedStock
}

type Analyzer interface {
	StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
}

type ScheduledStock struct {
	StockCode  string `json:"stockCode"`
	Name       string `json:"name"`
	Cron       string `json:"cron"`
	AIConfigID int    `json:"aiConfigId"`
}

type Service struct {
	store    Store
	analyzer Analyzer
}

func NewService(store Store, analyzer Analyzer) *Service {
	if store == nil {
		panic("watchlist: store dependency is required")
	}
	if analyzer == nil {
		panic("watchlist: analyzer dependency is required")
	}
	return &Service{store: store, analyzer: analyzer}
}

func NormalizeStockCode(stockCode string) string {
	code := strings.TrimSpace(stockCode)
	upper := strings.ToUpper(code)
	if strings.HasPrefix(upper, "GB_") {
		return strings.ToLower(strings.Replace(upper, "GB_", "us", 1))
	}
	return strings.ToLower(code)
}

func (s *Service) SaveStockAICron(ctx context.Context, cronText, stockCode string) (ScheduledStock, *contractservice.UserVisibleError) {
	normalized := NormalizeStockCode(stockCode)
	s.store.SaveStockAICron(ctx, cronText, normalized)
	follow := s.store.GetFollowedStock(ctx, normalized)
	if strings.TrimSpace(follow.StockCode) == "" {
		err := contractservice.NewUserVisibleError("watchlist.stock_not_followed", "股票未关注", false, contractservice.StageService)
		return ScheduledStock{}, &err
	}
	return ScheduledStock{
		StockCode:  follow.StockCode,
		Name:       follow.Name,
		Cron:       cronText,
		AIConfigID: follow.AiConfigId,
	}, nil
}

func (s *Service) ListScheduledStocks(ctx context.Context) []ScheduledStock {
	follows := s.store.ListFollowedStocks(ctx)
	result := make([]ScheduledStock, 0, len(follows))
	for _, follow := range follows {
		if follow.Cron == nil || strings.TrimSpace(*follow.Cron) == "" {
			continue
		}
		result = append(result, ScheduledStock{
			StockCode:  follow.StockCode,
			Name:       follow.Name,
			Cron:       *follow.Cron,
			AIConfigID: follow.AiConfigId,
		})
	}
	return result
}

func (s *Service) RunScheduledAnalysis(ctx context.Context, stockCode string) (ScheduledStock, *contractservice.UserVisibleError) {
	follow := s.store.GetFollowedStock(ctx, NormalizeStockCode(stockCode))
	if strings.TrimSpace(follow.StockCode) == "" {
		err := contractservice.NewUserVisibleError("watchlist.stock_not_followed", "股票未关注", false, contractservice.StageService)
		return ScheduledStock{}, &err
	}

	stream := s.analyzer.StartStockAnalysis(ctx, analysisservice.StockRequest{
		StockName:   follow.Name,
		StockCode:   follow.StockCode,
		Question:    "",
		AIConfigID:  follow.AiConfigId,
		EnableTools: true,
		Think:       true,
	})

	var content strings.Builder
	chatID := ""
	question := ""
	for chunk := range stream {
		if chunk.ExtraContent != "" {
			content.WriteString(chunk.ExtraContent)
			content.WriteString("\n")
		}
		if chunk.Content != "" {
			content.WriteString(chunk.Content)
		}
		if chunk.ChatID != "" {
			chatID = chunk.ChatID
		}
		if chunk.Question != "" {
			question = chunk.Question
		}
	}

	s.analyzer.SaveResult(ctx, follow.StockCode, follow.Name, content.String(), chatID, question, follow.AiConfigId)
	return ScheduledStock{
		StockCode:  follow.StockCode,
		Name:       follow.Name,
		Cron:       valueOrEmpty(follow.Cron),
		AIConfigID: follow.AiConfigId,
	}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
```

```go
// Update the App fields, imports, startup/domReady helpers, and notification methods in D:\codex_work\go-stock\app.go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-stock/backend/agent/tools"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
	configservice "go-stock/backend/service/config"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"
	taskservice "go-stock/backend/service/task"
	watchlistservice "go-stock/backend/service/watchlist"
	analysissource "go-stock/backend/source/analysis"
	marketsource "go-stock/backend/source/marketnews"
	notificationsource "go-stock/backend/source/notification"
	watchlistsource "go-stock/backend/source/watchlist"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type notificationService interface {
	SendDingTalk(message, stockCode string) string
	SendTyped(message, stockCode string, msgType int) notificationservice.Delivery
}

type watchlistService interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError)
	ListScheduledStocks(ctx context.Context) []watchlistservice.ScheduledStock
	RunScheduledAnalysis(ctx context.Context, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError)
}

type App struct {
	ctx                   context.Context
	cache                 *freecache.Cache
	cron                  *cron.Cron
	cronEntrys            map[string]cron.EntryID
	cronEntrysMu          sync.Mutex
	AiTools               []data.Tool
	SponsorInfo           map[string]any
	VipLevel              int64
	summaryMu             sync.Mutex
	summaryCancel         context.CancelFunc
	agentMu               sync.Mutex
	agentCancel           context.CancelFunc
	stockAlertMu          sync.Mutex
	stockAlertLastSent    map[string]time.Time
	priceAtAlertReset     map[string]float64
	marketReadService     marketReadService
	analysisService       analysisService
	configService         configService
	taskService           taskService
	notificationService   notificationService
	watchlistService      watchlistService
	saveFileDialog        func(ctx context.Context, options runtime.SaveDialogOptions) (string, error)
	writeFile             func(name string, data []byte, perm os.FileMode) error
	shareAnalysisUploader func(artifact analysisservice.ResultArtifact) (string, error)
	emitEvent             func(ctx context.Context, name string, data ...interface{})
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
	analysisSvc := analysisservice.NewService(analysisProvider, analysisStore, analysisStore)
	configSvc := configservice.NewService(configservice.NewStore())
	app := &App{
		cache:                 cache,
		cron:                  c,
		cronEntrys:            make(map[string]cron.EntryID),
		AiTools:               tools,
		stockAlertLastSent:    make(map[string]time.Time),
		priceAtAlertReset:     make(map[string]float64),
		marketReadService:     marketservice.NewService(marketsource.NewSource()),
		analysisService:       analysisSvc,
		configService:         configSvc,
		notificationService:   notificationservice.NewService(cache, notificationsource.NewAdapter()),
		watchlistService:      watchlistservice.NewService(watchlistsource.NewStore(), analysisSvc),
		saveFileDialog:        runtime.SaveFileDialog,
		writeFile:             os.WriteFile,
		shareAnalysisUploader: uploadSharedAnalysis,
		emitEvent:             runtime.EventsEmit,
	}
	app.taskService = taskservice.NewService(taskservice.NewStore(), &appTaskScheduler{app: app}, func() context.Context {
		return app.ctx
	})
	return app
}

func (a *App) restoreStockAICronSchedules() {
	if a.watchlistService == nil {
		return
	}
	for _, follow := range a.watchlistService.ListScheduledStocks(a.ctx) {
		a.registerStockAICron(follow.StockCode, follow.Cron)
	}
}

func (a *App) registerStockAICron(stockCode, cronText string) {
	if entryID, exists := a.getCronEntry(stockCode); exists {
		a.cron.Remove(entryID)
	}
	if strings.TrimSpace(cronText) == "" {
		a.removeCronEntry(stockCode)
		return
	}
	id, err := a.cron.AddFunc(cronText, a.buildStockAICronJob(stockCode))
	if err != nil {
		appError("watchlist-cron", "watchlist.cron_add.failed", "add stock ai cron failed", logger.String("stock_code", stockCode), logger.String("cron_expr", cronText), logger.Err(err))
		return
	}
	a.setCronEntry(stockCode, id)
}

func (a *App) buildStockAICronJob(stockCode string) func() {
	return func() {
		if a.watchlistService == nil {
			return
		}
		result, userErr := a.watchlistService.RunScheduledAnalysis(a.ctx, stockCode)
		if userErr != nil {
			if a.emitEvent != nil {
				a.emitEvent(a.ctx, "warnMsg", "AI分析失败："+userErr.Message)
			}
			return
		}
		if a.emitEvent != nil {
			a.emitEvent(a.ctx, "warnMsg", "开始自动分析"+result.Name+"_"+result.StockCode)
			a.emitEvent(a.ctx, "warnMsg", "AI分析完成："+result.Name+"_"+result.StockCode)
		}
	}
}

// Replace the followed-stock restore loop near the end of D:\codex_work\go-stock\app.go domReady with this call:
a.restoreStockAICronSchedules()

func (a *App) SendDingDingMessage(message string, stockCode string) string {
	if a.notificationService == nil {
		return ""
	}
	return a.notificationService.SendDingTalk(message, stockCode)
}

func (a *App) SendDingDingMessageByType(message string, stockCode string, msgType int) string {
	if strutil.HasPrefixAny(stockCode, []string{"SZ", "SH", "sh", "sz"}) && (!isTradingTime(time.Now())) {
		return "非A股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"hk", "HK"}) && (!IsHKTradingTime(time.Now())) {
		return "非港股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "US", "gb_"}) && (!IsUSTradingTime(time.Now())) {
		return "非美股交易时间"
	}
	if a.notificationService == nil {
		return ""
	}
	result := a.notificationService.SendTyped(message, stockCode, msgType)
	if a.emitEvent != nil && strings.TrimSpace(result.EventContent) != "" {
		go a.emitEvent(a.ctx, "newsPush", map[string]any{
			"time":    "📈 " + result.EventTitle,
			"isRed":   true,
			"source":  "go-stock",
			"content": result.EventContent,
		})
	}
	return result.DingResult
}

func (a *App) SetStockAICron(cronText, stockCode string) {
	if a.watchlistService == nil {
		return
	}
	result, userErr := a.watchlistService.SaveStockAICron(a.ctx, cronText, stockCode)
	if userErr != nil {
		if a.emitEvent != nil {
			a.emitEvent(a.ctx, "warnMsg", "AI分析任务保存失败："+userErr.Message)
		}
		return
	}
	a.registerStockAICron(result.StockCode, result.Cron)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/notification ./backend/service/watchlist . -run "Test(Service_(SendDingTalkSkipsWhileTTLActive|SendTypedDeliversLocalAndDingTalk|SaveStockAICronNormalizesUsCode|ListScheduledStocksFiltersBlankCron|RunScheduledAnalysisSavesCollectedResult|RunScheduledAnalysisReturnsUserVisibleErrorWhenStockMissing)|App_(NotificationHelpersDelegateToService|SetStockAICronRegistersAndRestoresJobsThroughWatchlistService))" -count=1
```

Expected:

- command exits `0`
- notification service, watchlist service, and app closeout bridge tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/source/notification/adapter.go backend/source/notification/adapter_windows.go backend/source/notification/adapter_other.go backend/service/notification/service.go backend/service/notification/service_test.go backend/source/watchlist/store.go backend/service/watchlist/service.go backend/service/watchlist/service_test.go app.go app_closeout_test.go
git commit -m "refactor: move notification and stock ai cron behind services"
```

### Task 4: Route Closeout Screens Through `pages` And Migrate Residual Frontend Consumers

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\pages\settings-page.vue`
- Create: `D:\codex_work\go-stock\frontend\src\pages\cron-task-page.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\services\configService.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\components\market.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\AnalyzeMartket.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\industryMoneyRank.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\moneyTrend.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\rankTable.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\newsList.vue`

- [ ] **Step 1: Write the failing frontend service normalization tests**

```js
// Replace D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs with:
import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeMarketFeeds,
  normalizeMarketFeed,
  normalizeMarketIndexes,
  normalizeMarketIndustryRanks,
  normalizeIndustryMoneyRanks,
  normalizeMoneyRanks,
  normalizeStockMoneyTrend,
} from './marketService.mjs';

test('normalizeMarketFeeds 缺失字段时补空数组', () => {
  assert.deepEqual(normalizeMarketFeeds({ telegraph: ['t1'] }), {
    telegraph: ['t1'],
    sina: [],
    foreign: [],
  });
});

test('normalizeMarketFeed 只保留 source 和 items', () => {
  assert.deepEqual(
    normalizeMarketFeed({
      source: 'telegraph',
      items: [{ id: 1 }],
      ignored: 'x',
    }),
    {
      source: 'telegraph',
      items: [{ id: 1 }],
    },
  );
});

test('normalizeMarketIndexes 只保留已知区域并补空数组', () => {
  assert.deepEqual(
    normalizeMarketIndexes({
      common: [{ code: '000001' }],
      europe: [{ code: 'DAX' }],
      unknown: [{ code: 'XXX' }],
    }),
    {
      common: [{ code: '000001' }],
      america: [],
      europe: [{ code: 'DAX' }],
      asia: [],
      other: [],
    },
  );
});

test('normalizeMarketIndustryRanks 对空值回退为数组', () => {
  assert.deepEqual(normalizeMarketIndustryRanks(undefined), []);
  assert.deepEqual(normalizeMarketIndustryRanks(null), []);
});

test('normalizeIndustryMoneyRanks 稳定输出当前页面需要的字段', () => {
  assert.deepEqual(
    normalizeIndustryMoneyRanks([{ name: '机器人', avg_changeratio: '0.0312', ts_trade: '15.66' }]),
    [{
      category: '',
      name: '机器人',
      avg_changeratio: 0.0312,
      inamount: 0,
      outamount: 0,
      netamount: 0,
      ratioamount: 0,
      ts_name: '',
      ts_symbol: '',
      ts_changeratio: 0,
      ts_trade: 15.66,
      ts_ratioamount: 0,
    }],
  );
});

test('normalizeMoneyRanks 稳定输出资金流排行数值字段', () => {
  assert.deepEqual(
    normalizeMoneyRanks([{ symbol: '600519', name: '贵州茅台', r0_net: '180000' }]),
    [{
      symbol: '600519',
      name: '贵州茅台',
      trade: 0,
      changeratio: 0,
      turnover: 0,
      amount: 0,
      outamount: 0,
      inamount: 0,
      netamount: 0,
      ratioamount: 0,
      r0_out: 0,
      r0_in: 0,
      r0_net: 180000,
      r0_ratio: 0,
      r3_out: 0,
      r3_in: 0,
      r3_net: 0,
      r3_ratio: 0,
    }],
  );
});

test('normalizeStockMoneyTrend 对空值回退并保留字段名', () => {
  assert.deepEqual(
    normalizeStockMoneyTrend([{ opendate: '2026-04-01', trade: '18.32', netamount: '250000', r0_net: '120000' }]),
    [{
      opendate: '2026-04-01',
      trade: 18.32,
      netamount: 250000,
      r0_net: 120000,
    }],
  );
});
```

- [ ] **Step 2: Run the frontend tests to verify they fail**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs
```

Expected:

- command exits non-zero
- failure mentions missing market wrapper normalizers or exports

- [ ] **Step 3: Add frontend wrappers, route through `pages`, and remove residual direct Wails imports from this slice**

```vue
<!-- D:\codex_work\go-stock\frontend\src\pages\settings-page.vue -->
<script setup>
import SettingsView from "../components/settings.vue";
</script>

<template>
  <SettingsView />
</template>
```

```vue
<!-- D:\codex_work\go-stock\frontend\src\pages\cron-task-page.vue -->
<script setup>
import CronTaskManager from "../components/cron-task-manager.vue";
</script>

<template>
  <CronTaskManager />
</template>
```

```js
// Replace D:\codex_work\go-stock\frontend\src\router\router.js with:
import { createRouter, createWebHashHistory } from 'vue-router'

import stockPageView from '../pages/stock-page.vue'
import settingsPageView from '../pages/settings-page.vue'
import aboutView from "../components/about.vue";
import fundView from "../components/fund.vue";
import marketPageView from "../pages/market-page.vue";
import agentChat from "../components/agent-chat.vue"
import researchPageView from "../pages/research-page.vue";
import cronTaskPageView from "../pages/cron-task-page.vue"

const routes = [
  { path: '/', component: stockPageView, name: 'stock' },
  { path: '/fund', component: fundView, name: 'fund' },
  { path: '/settings', component: settingsPageView, name: 'settings' },
  { path: '/about', component: aboutView, name: 'about' },
  { path: '/market', component: marketPageView, name: 'market' },
  { path: '/agent', component: agentChat, name: 'agent' },
  { path: '/research', component: researchPageView, name: 'research' },
  { path: '/cron-tasks', component: cronTaskPageView, name: 'cronTasks' },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
```

```js
// Replace D:\codex_work\go-stock\frontend\src\services\marketService.mjs with:
import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback;
}

export function normalizeMarketFeeds(value) {
  const data = value ?? {};
  return {
    telegraph: toArray(data.telegraph),
    sina: toArray(data.sina),
    foreign: toArray(data.foreign),
  };
}

export function normalizeMarketFeed(value) {
  const data = value ?? {};
  return {
    source: data.source ?? '',
    items: toArray(data.items),
  };
}

export function normalizeMarketIndexes(value) {
  const data = value ?? {};
  return {
    common: toArray(data.common),
    america: toArray(data.america),
    europe: toArray(data.europe),
    asia: toArray(data.asia),
    other: toArray(data.other),
  };
}

export function normalizeMarketIndustryRanks(value) {
  return toArray(value);
}

export function normalizeIndustryMoneyRanks(value) {
  return toArray(value).map((item) => ({
    category: item?.category ?? '',
    name: item?.name ?? '',
    avg_changeratio: toNumber(item?.avg_changeratio, 0),
    inamount: toNumber(item?.inamount, 0),
    outamount: toNumber(item?.outamount, 0),
    netamount: toNumber(item?.netamount, 0),
    ratioamount: toNumber(item?.ratioamount, 0),
    ts_name: item?.ts_name ?? '',
    ts_symbol: item?.ts_symbol ?? '',
    ts_changeratio: toNumber(item?.ts_changeratio, 0),
    ts_trade: toNumber(item?.ts_trade, 0),
    ts_ratioamount: toNumber(item?.ts_ratioamount, 0),
  }));
}

export function normalizeMoneyRanks(value) {
  return toArray(value).map((item) => ({
    symbol: item?.symbol ?? '',
    name: item?.name ?? '',
    trade: toNumber(item?.trade, 0),
    changeratio: toNumber(item?.changeratio, 0),
    turnover: toNumber(item?.turnover, 0),
    amount: toNumber(item?.amount, 0),
    outamount: toNumber(item?.outamount, 0),
    inamount: toNumber(item?.inamount, 0),
    netamount: toNumber(item?.netamount, 0),
    ratioamount: toNumber(item?.ratioamount, 0),
    r0_out: toNumber(item?.r0_out, 0),
    r0_in: toNumber(item?.r0_in, 0),
    r0_net: toNumber(item?.r0_net, 0),
    r0_ratio: toNumber(item?.r0_ratio, 0),
    r3_out: toNumber(item?.r3_out, 0),
    r3_in: toNumber(item?.r3_in, 0),
    r3_net: toNumber(item?.r3_net, 0),
    r3_ratio: toNumber(item?.r3_ratio, 0),
  }));
}

export function normalizeStockMoneyTrend(value) {
  return toArray(value).map((item) => ({
    opendate: item?.opendate ?? '',
    trade: toNumber(item?.trade, 0),
    netamount: toNumber(item?.netamount, 0),
    r0_net: toNumber(item?.r0_net, 0),
  }));
}

export async function analyzeMarketSentiment(keyword = '') {
  return AppBindings.AnalyzeSentimentWithFreqWeight(keyword);
}

export async function loadMarketFeeds() {
  const result = await AppBindings.GetMarketFeeds();
  return normalizeMarketFeeds(result);
}

export async function loadMarketGlobalIndexes() {
  const result = await AppBindings.GetMarketGlobalIndexes();
  return normalizeMarketIndexes(result);
}

export async function loadMarketIndustryRanks(sort = '0', count = 150) {
  const result = await AppBindings.GetMarketIndustryRanks(sort, count);
  return normalizeMarketIndustryRanks(result);
}

export async function loadIndustryMoneyRanks(fenlei = '0', sort = 'netamount') {
  const result = await AppBindings.GetIndustryMoneyRankSina(fenlei, sort);
  return normalizeIndustryMoneyRanks(result);
}

export async function loadMoneyRanks(sort = 'netamount') {
  const result = await AppBindings.GetMoneyRankSina(sort);
  return normalizeMoneyRanks(result);
}

export async function loadStockMoneyTrend(stockCode, days = 20) {
  const result = await AppBindings.GetStockMoneyTrendByDay(stockCode, days);
  return normalizeStockMoneyTrend(result);
}

export async function refreshMarketFeed(source) {
  const result = await AppBindings.RefreshMarketFeed(source);
  return normalizeMarketFeed(result);
}
```

```js
// Append to D:\codex_work\go-stock\frontend\src\services\configService.mjs
export async function sendTypedNotification(message, stockCode, msgType) {
  return AppBindings.SendDingDingMessageByType(message, stockCode, msgType);
}
```

```vue
<!-- Replace the entire <script setup> block in D:\codex_work\go-stock\frontend\src\components\AnalyzeMartket.vue -->
<script setup>
import * as echarts from "echarts";
import { onMounted, onUnmounted, ref } from "vue";
import _ from "lodash";
import {
  analyzeMarketSentiment,
  loadMarketGlobalIndexes,
} from "../services/marketService.mjs";
const { name, darkTheme, kDays, chartHeight } = defineProps({
  name: {
    type: String,
    default: '',
  },
  kDays: {
    type: Number,
    default: 14,
  },
  chartHeight: {
    type: Number,
    default: 500,
  },
  darkTheme: {
    type: Boolean,
    default: false,
  },
})
const common = ref([])
const america = ref([])
const europe = ref([])
const asia = ref([])
const mainIndex = ref([])
const chinaIndex = ref([])
const other = ref([])
const globalStockIndexes = ref(null)
const chartRef = ref(null);
const gaugeChartRef = ref(null);
const triggerAreas = ref(["main", "extra", "arrow"])
let handleChartInterval = null
let handleIndexInterval = null
onMounted(() => {
  handleChart()
  getIndex()
  handleChartInterval = setInterval(function () {
    handleChart()
  }, 1000 * 60)

  handleIndexInterval = setInterval(function () {
    getIndex()
  }, 1000 * 2)
})

onUnmounted(() => {
  clearInterval(handleChartInterval)
  clearInterval(handleIndexInterval)
})

function getIndex() {
  loadMarketGlobalIndexes().then((res) => {
    globalStockIndexes.value = res
    common.value = res.common
    america.value = res.america
    europe.value = res.europe
    asia.value = res.asia
    other.value = res.other
    mainIndex.value = asia.value.filter(function (item) {
      return ['上海', "深圳", "香港", "台湾", "北京", "东京", "首尔", "纽约", "纳斯达克"].includes(item.location)
    }).concat(america.value.filter(function (item) {
      return ['上海', "深圳", "香港", "台湾", "北京", "东京", "首尔", "纽约", "纳斯达克"].includes(item.location)
    }))

    chinaIndex.value = asia.value.filter(function (item) {
      return ['上海', "深圳", "香港", "台湾", "北京"].includes(item.location)
    })
  })
}

function handleChart() {
  const formatUtil = echarts.format;
  analyzeMarketSentiment("").then((res) => {
    const treemapchart = echarts.init(chartRef.value);
    const gaugeChart = echarts.init(gaugeChartRef.value);
    let data = res['frequencies'].map(item => ({
      name: item.Word,
      frequency: item.Frequency,
      weight: item.Weight,
      value: item.Score,
    }));

    let data2 = res['frequencies'].map(item => ({
      name: item.Word,
      value: item.Frequency,
      frequency: item.Frequency,
      weight: item.Weight,
    }));

    let data3 = res['frequencies'].map(item => ({
      name: item.Word,
      value: item.Weight,
      frequency: item.Frequency,
      weight: item.Weight,
    }));

    let option = {
      darkMode: darkTheme,
      title: {
        text: name,
        left: 'center',
        textStyle: {
          color: darkTheme ? '#ccc' : '#456'
        }
      },
      legend: {
        show: false
      },
      toolbox: {
        left: '20px',
        tooltip: {
          textStyle: {
            color: darkTheme ? '#ccc' : '#456'
          }
        },
        feature: {
          saveAsImage: { title: '保存图片' },
          restore: {
            title: '默认',
          },
          myTool2: {
            show: true,
            title: '按权重',
            icon: "path://M393.8816 148.1216a29.3376 29.3376 0 0 1-15.2576 38.0928c-43.776 17.152-81.92 43.8272-114.2784 76.2368A345.7536 345.7536 0 0 0 159.5392 512 352.8704 352.8704 0 0 0 512 864.4608a351.744 351.744 0 0 0 249.5488-102.912 353.536 353.536 0 0 0 76.2368-114.2784c5.6832-15.2576 22.8352-20.992 38.0928-15.2576 15.2576 5.7344 20.992 22.8864 15.2576 38.0928a421.2224 421.2224 0 0 1-89.6 133.376A412.6208 412.6208 0 0 1 512 921.6c-226.7136 0-409.6-182.8864-409.6-409.6 0-108.544 41.9328-211.456 120.0128-289.5872A421.2224 421.2224 0 0 1 355.84 132.864a29.3376 29.3376 0 0 1 38.0928 15.2576zM512 102.4c226.7136 0 409.6 182.8864 409.6 409.6 0 15.2576-13.312 28.5696-28.5696 28.5696H512A29.2864 29.2864 0 0 1 483.4304 512V130.9696c0-15.2576 13.312-28.5696 28.5696-28.5696z m28.5696 59.0336v321.9968h321.9968a350.976 350.976 0 0 0-321.9968-321.9968z",
            onclick: function () {
              treemapchart.setOption({ series: { data: data3 } })
            }
          },
          myTool1: {
            show: true,
            title: '按频次',
            icon: "path://M895.466667 476.8l-87.424-87.424v-123.626667a49.770667 49.770667 0 0 0-49.770667-49.770666h-123.626667L547.2 128.533333a49.792 49.792 0 0 0-70.4 0l-87.424 87.424h-123.626667a49.770667 49.770667 0 0 0-49.770666 49.770667v123.626667L128.533333 476.8a49.792 49.792 0 0 0 0 70.4l87.424 87.424v123.626667a49.770667 49.770667 0 0 0 49.770667 49.770666h123.626667l87.424 87.424a49.792 49.792 0 0 0 70.4 0l87.424-87.424h123.626666a49.770667 49.770667 0 0 0 49.770667-49.770666v-123.626667l87.424-87.424a49.749333 49.749333 0 0 0 0.042667-70.4z m-137.216 137.194667v144.256h-144.256L512 860.266667l-101.994667-101.994667h-144.256v-144.256L163.733333 512l101.994667-101.994667v-144.256h144.256L512 163.733333l101.994667 101.994667h144.256v144.256L860.266667 512l-102.016 101.994667z M414.378667 514.730667l28.672 10.922666c-18.090667 47.445333-38.229333 92.16-60.757334 133.802667l-30.037333-13.653333a1042.133333 1042.133333 0 0 0 62.122667-131.072zM381.952 367.616L355.669333 384c25.258667 26.282667 45.056 50.176 60.074667 72.021333l25.6-17.749333c-13.994667-20.48-33.792-44.032-59.392-70.656zM537.258667 455.338667c-0.682667 43.690667-6.144 79.189333-16.725334 106.837333-14.336 32.768-44.373333 60.416-89.429333 82.944l21.162667 25.941333c52.224-26.624 85.333333-60.074667 99.328-100.693333 1.706667-5.12 3.413333-10.24 4.778666-15.36 21.504 45.738667 52.906667 83.968 93.866667 115.370667l21.504-24.917334c-51.2-34.474667-86.357333-81.237333-105.813333-140.288 1.706667-15.701333 2.730667-32.085333 2.730666-49.834666h-31.402666z M508.586667 434.858667h115.712c-6.826667 25.258667-15.018667 47.786667-24.917334 66.901333l31.744 8.874667a627.008 627.008 0 0 0 27.989334-85.674667v-21.162667H517.12c3.413333-14.336 6.144-29.354667 8.874667-45.738666l-32.426667-5.12c-7.850667 59.392-25.6 105.813333-52.906667 139.264l26.965334 19.114666c16.725333-19.114667 30.378667-44.373333 40.96-76.458666z",
            onclick: function () {
              treemapchart.setOption({ series: { data: data2 } })
            }
          }
        }
      },
      tooltip: {
        formatter: function (info) {
          var value = info.value.toFixed(2);
          var frequency = info.data.frequency;
          var weight = info.data.weight;
          return [
            '<div class="tooltip-title">' + info.name + '</div>',
            '热度: ' + formatUtil.addCommas(value) + '',
            '<div class="tooltip-title">频次: ' + formatUtil.addCommas(frequency) + '</div>',
            '<div class="tooltip-title">权重: ' + formatUtil.addCommas(weight) + '</div>',
          ].join('');
        }
      },
      series: [
        {
          type: 'treemap',
          breadcrumb: { show: false },
          left: '0',
          top: '40',
          right: '0',
          bottom: '0',
          tooltip: {
            show: true
          },
          data: data
        }
      ]
    };
    treemapchart.setOption(option);

    let option2 = {
      darkMode: darkTheme,
      series: [
        {
          type: 'gauge',
          startAngle: 180,
          endAngle: 0,
          center: ['50%', '75%'],
          radius: '90%',
          min: -100,
          max: 100,
          splitNumber: 8,
          axisLine: {
            lineStyle: {
              width: 6,
              color: [
                [0.25, '#03fb6a'],
                [0.5, '#58e1f9'],
                [0.75, '#ef5922'],
                [1, '#f11d29'],
              ]
            }
          },
          pointer: {
            icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
            length: '12%',
            width: 20,
            offsetCenter: [0, '-60%'],
            itemStyle: {
              color: 'auto'
            }
          },
          axisTick: {
            length: 12,
            lineStyle: {
              color: 'auto',
              width: 2
            }
          },
          splitLine: {
            length: 20,
            lineStyle: {
              color: 'auto',
              width: 5
            }
          },
          axisLabel: {
            color: darkTheme ? '#ccc' : '#456',
            fontSize: 20,
            distance: -45,
            rotate: 'tangential',
            formatter: function (value) {
              if (value === 100) {
                return '极热';
              } else if (value === 50) {
                return '乐观';
              } else if (value === 0) {
                return '中性';
              } else if (value === -50) {
                return '谨慎';
              } else if (value === -100) {
                return '冰点';
              }
              return '';
            }
          },
          title: {
            offsetCenter: [0, '-10%'],
            fontSize: 20
          },
          detail: {
            fontSize: 30,
            offsetCenter: [0, '-35%'],
            valueAnimation: true,
            formatter: function (value) {
              return value.toFixed(2) + '';
            },
            color: 'inherit'
          },
          data: [
            {
              value: res.result.Score * 0.2,
              name: '市场情绪强弱'
            }
          ]
        }
      ]
    };
    gaugeChart.setOption(option2);
  })
}
</script>
```

```vue
<!-- Replace the App import and GetRankData in D:\codex_work\go-stock\frontend\src\components\industryMoneyRank.vue -->
<script setup>
import { CaretDown, CaretUp, RefreshCircleOutline } from "@vicons/ionicons5";
import { NText, useMessage } from "naive-ui";
import { onBeforeUnmount, onMounted, ref } from "vue";
import { loadIndustryMoneyRanks } from "../services/marketService.mjs";
import KLineChart from "./KLineChart.vue";

function GetRankData() {
  message.loading("正在刷新数据...")
  loadIndustryMoneyRanks(fenlei.value, sort.value).then(result => {
    if (result.length > 0) {
      dataList.value = result
    }
  })
}
</script>
```

```vue
<!-- Replace the entire <script setup lang="ts"> block in D:\codex_work\go-stock\frontend\src\components\moneyTrend.vue -->
<script setup lang="ts">
import { onMounted, ref } from "vue";
import { loadStockMoneyTrend } from "../services/marketService.mjs";
import * as echarts from "echarts";

const { code, name, darkTheme, days, chartHeight } = defineProps({
  code: {
    type: String,
    default: ''
  },
  name: {
    type: String,
    default: ''
  },
  days: {
    type: Number,
    default: 14
  },
  chartHeight: {
    type: Number,
    default: 500
  },
  darkTheme: {
    type: Boolean,
    default: false
  }
})
const LineChartRef = ref(null);

onMounted(
  () => {
    handleLine(code, days)
  }
)

const handleLine = (code, days) => {
  loadStockMoneyTrend(code, days).then(result => {
    const chart = echarts.init(LineChartRef.value);
    const categoryData = [];
    const netamount_values = [];
    const r0_net_values = [];
    const trades_values = [];
    let volume = []

    let min = 0
    let max = 0
    for (let i = 0; i < result.length; i++) {
      let resultElement = result[i]
      categoryData.push(resultElement.opendate)
      let netamount = (resultElement.netamount / 10000).toFixed(2);
      netamount_values.push(netamount)
      let price = Number(resultElement.trade);
      trades_values.push(price)
      r0_net_values.push((resultElement.r0_net / 10000).toFixed(2))

      if (min === 0 || min > price) {
        min = price
      }
      if (max < price) {
        max = price
      }

      if (i > 0) {
        let b = Number(Number(result[i].netamount) + Number(result[i - 1].netamount)) / 10000
        volume.push(b.toFixed(2))
      } else {
        volume.push((Number(result[i].netamount) / 10000).toFixed(2))
      }
    }
    const option = {
      title: {
        text: name,
        left: '20px',
        textStyle: {
          color: darkTheme ? '#ccc' : '#456'
        }
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          lineStyle: {
            color: '#376df4',
            width: 1,
            opacity: 1
          }
        },
        borderWidth: 2,
        borderColor: darkTheme ? '#456' : '#ccc',
        backgroundColor: darkTheme ? '#456' : '#fff',
        padding: 10,
        textStyle: {
          color: darkTheme ? '#ccc' : '#456'
        },
      },
      axisPointer: {
        link: [
          {
            xAxisIndex: 'all'
          }
        ],
        label: {
          backgroundColor: '#888'
        }
      },
      legend: {
        show: true,
        data: ['当日净流入', '主力当日净流入', '累计净流入', '股价'],
        selected: {
          '当日净流入': true,
          '主力当日净流入': true,
          '累计净流入': true,
          '股价': true,
        },
        textStyle: {
          color: darkTheme ? 'rgb(253,252,252)' : '#456'
        },
        right: 150,
      },
      dataZoom: [
        {
          type: 'inside',
          xAxisIndex: [0, 1],
          start: 86,
          end: 100
        },
        {
          show: true,
          xAxisIndex: [0, 1],
          type: 'slider',
          top: '90%',
          start: 86,
          end: 100
        }
      ],
      grid: [
        {
          left: '8%',
          right: '8%',
          height: '50%',
        },
        {
          left: '8%',
          right: '8%',
          top: '74%',
          height: '15%'
        },
      ],
      xAxis: [
        {
          type: 'category',
          data: categoryData,
          axisPointer: {
            z: 100
          },
          boundaryGap: false,
          axisLine: { onZero: false },
          splitLine: { show: false },
          min: 'dataMin',
          max: 'dataMax',
        },
        {
          gridIndex: 1,
          type: 'category',
          data: categoryData,
          axisLabel: {
            show: false
          },
        }
      ],
      yAxis: [
        {
          name: '当日净流入/万',
          type: 'value',
          axisLine: {
            show: true
          },
          splitLine: {
            show: false
          },
        },
        {
          name: '股价',
          type: 'value',
          min: min - 1,
          max: max + 1,
          minInterval: 0.01,
          axisLine: {
            show: true
          },
          splitLine: {
            show: false
          },
        },
        {
          gridIndex: 1,
          name: '累计净流入/万',
          type: 'value',
          axisLine: {
            show: true
          },
          splitLine: {
            show: false
          },
        },
      ],
      series: [
        {
          yAxisIndex: 0,
          name: '当日净流入',
          data: netamount_values,
          smooth: false,
          showSymbol: false,
          lineStyle: {
            width: 2
          },
          markPoint: {
            symbol: 'arrow',
            symbolRotate: 90,
            symbolSize: [10, 20],
            symbolOffset: [10, 0],
            itemStyle: {
              color: '#0d7dfc'
            },
            label: {
              position: 'right',
            },
            data: [
              { type: 'max', name: 'Max' },
              { type: 'min', name: 'Min' }
            ]
          },
          markLine: {
            data: [
              {
                type: 'average',
                name: 'Average',
                lineStyle: {
                  color: '#0077ff',
                  width: 0.5
                },
              },
            ]
          },
          type: 'line'
        },
        {
          yAxisIndex: 0,
          name: '主力当日净流入',
          data: r0_net_values,
          smooth: false,
          showSymbol: false,
          lineStyle: {
            width: 2
          },
          type: 'bar'
        },
        {
          yAxisIndex: 1,
          name: '股价',
          type: 'line',
          data: trades_values,
          smooth: true,
          showSymbol: false,
          lineStyle: {
            width: 3
          },
          markPoint: {
            symbol: 'arrow',
            symbolRotate: 90,
            symbolSize: [10, 20],
            symbolOffset: [10, 0],
            itemStyle: {
              color: '#f39509'
            },
            label: {
              position: 'right',
            },
            data: [
              { type: 'max', name: 'Max' },
              { type: 'min', name: 'Min' }
            ]
          },
          markLine: {
            data: [
              {
                type: 'average',
                name: 'Average',
                lineStyle: {
                  color: '#f39509',
                  width: 0.5
                },
              },
            ]
          },
        },
        {
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 2,
          name: '累计净流入',
          data: volume,
          smooth: true,
          showSymbol: false,
          lineStyle: {
            width: 2
          },
          markPoint: {
            symbol: 'arrow',
            symbolRotate: 90,
            symbolSize: [10, 20],
            symbolOffset: [10, 0],
            label: {
              position: 'right',
            },
            data: [
              { type: 'max', name: 'Max' },
              { type: 'min', name: 'Min' }
            ]
          },
        },
      ]
    };
    chart.setOption(option);
  })
}
</script>
```

```vue
<!-- Replace the App import and GetMoneyRankSinaData in D:\codex_work\go-stock\frontend\src\components\rankTable.vue -->
<script setup>
import { CaretDown, CaretUp, RefreshCircleOutline } from "@vicons/ionicons5";
import { NText, useMessage } from "naive-ui";
import { onBeforeUnmount, onMounted, ref } from "vue";
import { loadMoneyRanks } from "../services/marketService.mjs";
import KLineChart from "./KLineChart.vue";

function GetMoneyRankSinaData() {
  message.loading("正在刷新数据...")
  loadMoneyRanks(sort.value).then(result => {
    if (result.length > 0) {
      dataList.value = result
    }
  })
}
</script>
```

```vue
<!-- Replace the config/prompt imports and the onBeforeMount config reads in D:\codex_work\go-stock\frontend\src\components\market.vue -->
<script setup>
import * as echarts from "echarts";
import { computed, h, onBeforeMount, onBeforeUnmount, onMounted, onUnmounted, ref } from 'vue'
import { EventsOff, EventsOn } from "../../wailsjs/runtime";
import NewsList from "./newsList.vue";
import KLineChart from "./KLineChart.vue";
import StockLightweightKlineChart from "./StockLightweightKlineChart.vue";
import { CaretDown, CaretUp, PulseOutline } from "@vicons/ionicons5";
import { NAvatar, NButton, NFlex, NText, useMessage, useNotification } from "naive-ui";
import { MdPreview } from "md-editor-v3";
import { useRoute } from 'vue-router'
import RankTable from "./rankTable.vue";
import IndustryMoneyRank from "./industryMoneyRank.vue";
import StockResearchReportList from "./StockResearchReportList.vue";
import StockNoticeList from "./StockNoticeList.vue";
import LongTigerRankList from "./LongTigerRankList.vue";
import IndustryResearchReportList from "./IndustryResearchReportList.vue";
import HotStockList from "./HotStockList.vue";
import HotEvents from "./HotEvents.vue";
import HotTopics from "./HotTopics.vue";
import InvestCalendarTimeLine from "./InvestCalendarTimeLine.vue";
import ClsCalendarTimeLine from "./ClsCalendarTimeLine.vue";
import SelectStock from "./SelectStock.vue";
import Stockhotmap from "./stockhotmap.vue";
import { resolveFirstAiConfigId } from "../utils/aiConfig.mjs";
import {
  loadLatestAnalysisResult,
  saveAnalysisMarkdown,
  saveAnalysisResult,
  shareAnalysis,
  startMarketSummary,
} from "../services/analysisService.mjs";
import {
  loadAiConfigs,
  loadAppConfig,
  loadPromptTemplates,
} from "../services/configService.mjs";
import {
  loadMarketFeeds,
  loadMarketGlobalIndexes,
  loadMarketIndustryRanks,
  refreshMarketFeed,
} from "../services/marketService.mjs";

onBeforeMount(() => {
  nowTab.value = route.query.name
  stockCode.value = route.query.stockCode
  loadAppConfig().then(result => {
    summaryBTN.value = result.openAiEnable
    darkTheme.value = result.darkTheme
    httpProxyEnabled.value = result.httpProxyEnabled
  })
  loadPromptTemplates("", "").then(res => {
    promptTemplates.value = res
    sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统Prompt')
    userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户Prompt')
  })
  loadAiConfigs().then(res => {
    aiConfigs.value = res
    aiConfigId.value = resolveFirstAiConfigId(res)
  })
  loadFeeds()
  getIndex()
  industryRank()
  indexInterval.value = setInterval(() => {
    getIndex()
  }, 3000)
  indexIndustryRank.value = setInterval(() => {
    industryRank()
    ReFlesh("财联社电报")
    ReFlesh("新浪财经")
    ReFlesh("外媒")
  }, 1000 * 10)
})
</script>
```

```vue
<!-- Replace the App import and sendTestNotice call in D:\codex_work\go-stock\frontend\src\components\settings.vue -->
<script setup>
import { h, onBeforeUnmount, onMounted, ref } from "vue";
import { NTag, NTooltip, NIcon, useMessage } from "naive-ui";
import { data } from "../../wailsjs/go/models";
import { EventsEmit } from "../../wailsjs/runtime";
import { HelpCircleFilledIcon } from "tdesign-icons-vue-next";
import {
  deleteLegacyPrompt,
  exportAppConfig,
  fetchAiModels as fetchAvailableModels,
  loadAppConfig,
  loadPromptTemplates,
  saveAppConfig,
  saveLegacyPrompt,
  sendTypedNotification,
} from "../services/configService.mjs";

function sendTestNotice() {
  let markdown = "### go-stock test\n" + new Date()
  let msg = '{' +
      '     "msgtype": "markdown",' +
      '     "markdown": {' +
      '         "title":"go-stock' + new Date() + '",' +
      '         "text": "' + markdown + '"' +
      '     },' +
      '      "at": {' +
      '          "isAtAll": true' +
      '      }' +
      ' }'

  sendTypedNotification(msg, "test-" + new Date().getTime(), 1).then(res => {
    message.info(res)
  })
}
</script>
```

```vue
<!-- Replace the import block in D:\codex_work\go-stock\frontend\src\components\newsList.vue -->
<script setup>
import { RefreshCircleSharp } from "@vicons/ionicons5";
import { onMounted, onUnmounted, ref } from 'vue'

const { headerTitle, newsList } = defineProps({
  headerTitle: {
    type: String,
    default: '市场资讯'
  },
  newsList: {
    type: Array,
    default: () => []
  },
})

const emits = defineEmits(['update:message'])

const updateMessage = () => {
  emits('update:message', headerTitle)
}

const time = ref(new Date())

const updateTime = () => {
  time.value = new Date()
}

let timer = null

onMounted(() => {
  if (headerTitle === '财联社电报') {
    timer = setInterval(updateTime, 1000)
  }
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>
```

- [ ] **Step 4: Run the frontend tests and build to verify they pass**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs
```

Run:

```powershell
Set-Location 'D:\codex_work\go-stock\frontend'
npm run build
```

Expected:

- the Node test command exits `0`
- `npm run build` exits `0`
- no new frontend compile errors in `market.vue`, `settings.vue`, or the new page wrappers

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/pages/settings-page.vue frontend/src/pages/cron-task-page.vue frontend/src/router/router.js frontend/src/services/marketService.mjs frontend/src/services/marketService.test.mjs frontend/src/services/configService.mjs frontend/src/components/market.vue frontend/src/components/AnalyzeMartket.vue frontend/src/components/industryMoneyRank.vue frontend/src/components/moneyTrend.vue frontend/src/components/rankTable.vue frontend/src/components/settings.vue frontend/src/components/newsList.vue
git commit -m "refactor: route closeout screens through pages and services"
```

### Task 5: Verify The Closeout Slice And Review The Final Diff

**Files:**
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\components\market.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`

- [ ] **Step 1: Run the combined Go verification**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/contract ./backend/service/market ./backend/service/analysis ./backend/service/config ./backend/service/notification ./backend/service/watchlist . -count=1
```

Expected:

- command exits `0`
- root closeout bridge tests and all touched service packages pass together

- [ ] **Step 2: Run the frontend verification**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs frontend/src/services/configService.test.mjs frontend/src/services/taskService.test.mjs
```

Run:

```powershell
Set-Location 'D:\codex_work\go-stock\frontend'
npm run build
```

Expected:

- Node service tests exit `0`
- Vite build exits `0`

- [ ] **Step 3: Search for the old dual-path markers that this slice is supposed to remove**

Run:

```powershell
rg -n "legacyPromptBridge|shouldFallbackToLegacyPromptBridge" app.go app_common.go
```

Run:

```powershell
rg -n "../../wailsjs/go/main/App" frontend/src/components/market.vue frontend/src/components/AnalyzeMartket.vue frontend/src/components/industryMoneyRank.vue frontend/src/components/moneyTrend.vue frontend/src/components/rankTable.vue frontend/src/components/settings.vue frontend/src/components/newsList.vue frontend/src/pages/settings-page.vue frontend/src/pages/cron-task-page.vue
```

Run:

```powershell
rg -n "data.New(SettingsApi|DeepSeekOpenAi|MarketNewsApi|StockDataApi|DingDingAPI)" app.go app_common.go
```

Expected:

- the first command prints nothing
- the second command prints nothing for the listed closeout files
- the third command no longer shows direct `backend/data` calls for config export, market residual reads, prompt fallback, notification send helpers, or stock AI cron helpers

- [ ] **Step 4: Review the aggregate diff for the four implementation commits**

Run:

```powershell
git diff --stat HEAD~4..HEAD
```

Run:

```powershell
git diff --name-only HEAD~4..HEAD
```

Expected:

- the diff is limited to the files listed in this plan
- no unrelated assistant / floating UI files appear
- the change surface is visibly smaller than a repo-wide closure and matches the intended `phase4-b / phase5 closeout` slice
