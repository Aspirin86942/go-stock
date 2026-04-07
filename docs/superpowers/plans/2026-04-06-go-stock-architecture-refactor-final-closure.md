# go-stock Architecture Refactor Final Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining architecture gaps after phase 4B so `app.go` / `app_common.go` are bridge adapters with an explicit compatibility allowlist, remaining core pages route through `pages` + `services`, and residual direct `backend/data` / `wailsjs` usage is either migrated or intentionally fenced.

**Architecture:** Finish closure in four bounded passes. First, expand the existing `market` boundary to absorb the last read-only market, quote, search, and calendar surfaces that still bypass `service`. Second, move the remaining watchlist, price-alert, and research-center CRUD/monitor flows behind `watchlist` and a new `research` boundary. Third, finish route/page/service migration for `fund`, `about`, `agent`, and the floating assistants while explicitly freezing the multi-turn agent bridge as a compatibility island rather than silently expanding it. Fourth, add architecture guard tests plus full verification so phase 5 stays closed.

**Tech Stack:** Go, Wails, Vue 3, Naive UI, Node `--test`, `go test`, Vite

---

## Re-Alignment Summary

This plan starts **after** the already-landed slices:

- phase 1 / phase 2:
  - `backend/source/marketnews`
  - `backend/service/market`
  - `frontend/src/pages/market-page.vue`
  - `frontend/src/services/marketService.mjs`
- analysis-chain slice:
  - `backend/source/analysis`
  - `backend/service/analysis`
  - `frontend/src/pages/stock-page.vue`
  - `frontend/src/pages/research-page.vue`
  - `frontend/src/services/analysisService.mjs`
- phase 4 slice:
  - `backend/service/config`
  - `backend/service/task`
  - `frontend/src/services/configService.mjs`
  - `frontend/src/services/taskService.mjs`
- phase 4B / phase 5 closeout slice:
  - shared user-visible error contract
  - residual market money-rank helpers
  - analysis artifact helpers
  - config export closeout
  - notification service
  - stock AI cron closeout
  - `/settings` and `/cron-tasks` page wrappers

What still prevents the overall spec from being truly closed:

- `app.go` still owns direct business orchestration for:
  - followed-stock / group CRUD
  - price-alert evaluation
  - fund follow / refresh flows
  - multiple quote / K-line reads
- `app_common.go` still owns direct business orchestration for:
  - residual market/news/read-only helpers
  - stock changes and history
  - AI recommend stock record management
  - stock-info query CRUD
  - trading-record CRUD and statistics
- `frontend/src` still has direct `wailsjs/go/main/App` imports in many feature components.
- `/fund`, `/about`, and `/agent` still route directly to heavy components instead of `pages`.
- the assistant / agent multi-turn bridge is still a hidden compatibility path, not an explicitly fenced one.

## Closure Definition

Phase 5 is only considered closed when all of the following are true:

- business data reads and writes in `app.go` / `app_common.go` delegate to `service` boundaries instead of directly instantiating `backend/data` APIs
- direct `wailsjs/go/main/App` imports are limited to `frontend/src/services/**`
- all route components in `frontend/src/router/router.js` resolve to `frontend/src/pages/**`
- the remaining assistant compatibility island is explicitly documented, tested, and prevented from expanding accidentally
- the repo has guard tests that fail if these conditions regress later

## Compatibility Allowlist

The final closure deliberately keeps these bridge-only surfaces because the overall spec explicitly excluded a deep rewrite of the multi-turn agent subsystem and OS/runtime helpers:

- lifecycle and runtime helpers:
  - `startup`
  - `domReady`
  - `shutdown`
  - `OpenURL`
  - `SaveImage`
  - `SaveWordFile`
  - `CheckUpdate`
  - `GetVersionInfo`
  - `GetSponsorInfo`
  - `GetEffectiveSponsorVip`
  - `CheckSponsorCode`
  - `GetTimezone`
  - `FetchAiModels`
- assistant compatibility bridge:
  - `ChatWithAgent`
  - `AbortChatWithAgent`
  - `GetAiAssistantSession`
  - `SaveAiAssistantSession`

Everything else should move behind `service` or be removed.

## File Map

- Create: `D:\codex_work\go-stock\backend\service\research\contracts.go`
  - Stable research-domain contracts for stock changes, AI recommend records, stock-info pages, trading records, and alert monitor outputs.
- Create: `D:\codex_work\go-stock\backend\service\research\service.go`
  - Research service that owns stock-change history, AI recommend record management, stock-info query CRUD, trading-record CRUD, and AI recommend alert evaluation.
- Create: `D:\codex_work\go-stock\backend\service\research\service_test.go`
  - Unit tests for empty-page fallbacks, typed record mapping, alert evaluation, and mutation delegation.
- Create: `D:\codex_work\go-stock\backend\source\research\store.go`
  - Thin adapters over `backend/data.StockChangesApi`, `StockChangeHistoryService`, `AiRecommendStocksService`, and trading / stock-info APIs.
- Create: `D:\codex_work\go-stock\backend\service\fund\service.go`
  - Fund service for follow list, search, follow/unfollow, and background refresh.
- Create: `D:\codex_work\go-stock\backend\service\fund\service_test.go`
  - Unit tests for fund-list fallback and follow delegation.
- Create: `D:\codex_work\go-stock\backend\source\fund\adapter.go`
  - Thin adapter over `backend/data.FundApi`.
- Create: `D:\codex_work\go-stock\app_market_closeout_test.go`
  - Bridge tests proving residual market/read-only methods delegate to `marketService`.
- Create: `D:\codex_work\go-stock\app_watchlist_research_test.go`
  - Bridge tests for watchlist, price-alert, research, and fund delegation.
- Create: `D:\codex_work\go-stock\architecture_closeout_test.go`
  - Guard tests that scan `app.go`, `app_common.go`, `frontend/src`, and `frontend/src/router/router.js` for forbidden legacy paths.
- Modify: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
  - Extend the source adapter with the remaining market/news/search/quote reads still called directly from `app.go` / `app_common.go`.
- Modify: `D:\codex_work\go-stock\backend\service\market\contracts.go`
  - Add typed contracts for residual market/news/search/quote payloads while preserving current JSON keys expected by the frontend.
- Modify: `D:\codex_work\go-stock\backend\service\market\service.go`
  - Map residual read-only helpers into typed outputs and empty-safe defaults.
- Modify: `D:\codex_work\go-stock\backend\service\market\service_test.go`
  - Cover residual market helper mapping, realtime price fallback, and K-line normalization.
- Modify: `D:\codex_work\go-stock\backend\source\watchlist\store.go`
  - Add followed-stock, group, trading-target, alarm-threshold, and realtime quote reads/writes needed by the stock shell and price-alert monitor.
- Modify: `D:\codex_work\go-stock\backend\service\watchlist\service.go`
  - Extend the watchlist boundary with follow/group CRUD, alert-setting writes, and cost-alert evaluation.
- Modify: `D:\codex_work\go-stock\backend\service\watchlist\service_test.go`
  - Add tests for follow/group delegation and alert reset behavior.
- Modify: `D:\codex_work\go-stock\app.go`
  - Wire the final services, delegate residual old methods, move fund/background monitors to services, and add compatibility comments around the assistant allowlist.
- Modify: `D:\codex_work\go-stock\app_common.go`
  - Delegate residual market/news/search helpers and research CRUD helpers through the new service boundaries.
- Create: `D:\codex_work\go-stock\frontend\src\services\researchService.mjs`
  - Frontend wrapper for stock changes, AI recommend records, stock-info page/filter metadata, and trading-record CRUD.
- Create: `D:\codex_work\go-stock\frontend\src\services\researchService.test.mjs`
  - Node tests for research normalizers and page defaults.
- Create: `D:\codex_work\go-stock\frontend\src\services\fundService.mjs`
  - Frontend wrapper for fund list, follow list, follow/unfollow, and refresh actions.
- Create: `D:\codex_work\go-stock\frontend\src\services\fundService.test.mjs`
  - Node tests for fund-list and followed-fund normalizers.
- Create: `D:\codex_work\go-stock\frontend\src\services\watchlistService.mjs`
  - Frontend wrapper for followed-stock/group operations and alert-setting writes.
- Create: `D:\codex_work\go-stock\frontend\src\services\watchlistService.test.mjs`
  - Node tests for group / follow / alert payload normalization.
- Create: `D:\codex_work\go-stock\frontend\src\services\appShellService.mjs`
  - Frontend wrapper for shared shell reads such as config, version info, group list, update checks, and URL open.
- Create: `D:\codex_work\go-stock\frontend\src\services\appShellService.test.mjs`
  - Node tests for shell-data normalization defaults.
- Create: `D:\codex_work\go-stock\frontend\src\services\assistantService.mjs`
  - Frontend wrapper for the remaining assistant compatibility bridge: stream starts/stops, session load/save, and share actions.
- Create: `D:\codex_work\go-stock\frontend\src\services\assistantService.test.mjs`
  - Node tests for assistant-session defaults and share payload helpers.
- Create: `D:\codex_work\go-stock\frontend\src\pages\fund-page.vue`
  - Page wrapper for the existing fund screen.
- Create: `D:\codex_work\go-stock\frontend\src\pages\about-page.vue`
  - Page wrapper for the existing about screen.
- Create: `D:\codex_work\go-stock\frontend\src\pages\agent-page.vue`
  - Page wrapper for the existing agent screen.
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
  - Route the remaining direct routes through `pages`.
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
  - Add wrappers for the remaining market/news/search/quote readers.
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
  - Extend tests for the new market/news/search/quote normalizers.
- Modify: `D:\codex_work\go-stock\frontend\src\App.vue`
  - Stop importing main Wails bindings directly and consume shared shell/watchlist services.
- Modify: `D:\codex_work\go-stock\frontend\src\components\stock.vue`
  - Move residual read/write calls into `marketService`, `watchlistService`, `analysisService`, and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\SelectStock.vue`
  - Replace direct Wails imports with `marketService`, `watchlistService`, and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\KLineChart.vue`
  - Load K-line reads through `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
  - Load EastMoney K-line reads through `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\stockSparkLine.vue`
  - Load minute-line reads through `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\LongTigerRankList.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\IndustryResearchReportList.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockResearchReportList.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockNoticeList.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotStockList.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotEvents.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotTopics.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\InvestCalendarTimeLine.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\ClsCalendarTimeLine.vue`
  - Replace direct Wails import with `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\researchIndex.vue`
  - Consume `researchService` rather than mixing direct Wails imports.
- Modify: `D:\codex_work\go-stock\frontend\src\components\stockChangesMonitor.vue`
  - Replace direct Wails imports with `researchService` and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\aiRecommendStocksList.vue`
  - Replace direct Wails imports with `researchService`, `analysisService`, and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\allStockList.vue`
  - Replace direct Wails imports with `researchService` and `marketService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\allStockInfoList.vue`
  - Replace direct Wails imports with `researchService` and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\TradingRecordManager.vue`
  - Replace direct Wails imports with `researchService` and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\promptTemplateList.vue`
  - Replace residual direct config read with `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\researchReport.vue`
  - Replace residual direct config read with `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\fund.vue`
  - Replace direct Wails imports with `fundService` and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\about.vue`
  - Replace direct Wails imports with `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\agent-chat.vue`
  - Replace direct Wails imports with `assistantService` and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\FloatingAiAssistant.vue`
  - Replace direct Wails imports with `assistantService`, `configService`, and `appShellService`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\FloatingAgentAssistant.vue`
  - Replace direct Wails imports with `assistantService`, `configService`, and `appShellService`.
- Delete: `D:\codex_work\go-stock\frontend\src\components\agent-chat_bk.vue`
  - Unused backup file that still drags a direct Wails import into the tree.

## Scope Note

This plan intentionally focuses on **remaining closure work only**.

Included:

- residual market/news/search/quote read helpers still bypassing `service`
- watchlist/group CRUD and price-alert logic still living directly in `app.go`
- research-center CRUD/monitor flows still living directly in `app_common.go`
- remaining direct-route pages (`/fund`, `/about`, `/agent`)
- remaining direct frontend `wailsjs/go/main/App` imports
- explicit compatibility fencing for the assistant bridge
- automated guards for phase-5 regression

Explicitly excluded:

- deep rewrite of `backend/agent` reasoning or tool-planning internals
- visual redesign of existing pages
- Linux / macOS historical divergence cleanup beyond keeping the build green
- splitting `backend/models/models.go` or doing unrelated package churn

The core convergence files are intentionally revisited across tasks:

- `app.go`
- `app_common.go`
- `frontend/src/App.vue`
- `frontend/src/components/stock.vue`

That overlap is intentional because closure must converge on the same bridge and shell entry points rather than creating parallel wrappers.

### Task 1: Expand `market` To Absorb All Remaining Read-Only Market / Quote / Search Paths

**Files:**
- Modify: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
- Modify: `D:\codex_work\go-stock\backend\service\market\contracts.go`
- Modify: `D:\codex_work\go-stock\backend\service\market\service.go`
- Modify: `D:\codex_work\go-stock\backend\service\market\service_test.go`
- Create: `D:\codex_work\go-stock\app_market_closeout_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\appShellService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\appShellService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\App.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\SelectStock.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\KLineChart.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\stockSparkLine.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\LongTigerRankList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\IndustryResearchReportList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockResearchReportList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockNoticeList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotStockList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotEvents.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\HotTopics.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\InvestCalendarTimeLine.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\ClsCalendarTimeLine.vue`

- [ ] **Step 1: Write the failing backend tests for residual market reads**

```go
func TestService_LoadRealtimePriceFallsBackThroughBidLevels(t *testing.T) {
	source := &fakeSource{
		realtimePrices: []map[string]any{
			{
				"code":     "sz000001",
				"name":     "平安银行",
				"price":    "",
				"a1p":      "0",
				"b1p":      "12.34",
				"preClose": "12.10",
			},
		},
	}

	svc := NewService(source)
	price := svc.LoadRealtimePrice("sz000001")

	if price.Code != 0 {
		t.Fatalf("expected success code, got %d", price.Code)
	}
	if price.Price != 12.34 {
		t.Fatalf("expected fallback price 12.34, got %v", price.Price)
	}
}

func TestService_LoadHotTopicsReturnsEmptySliceWhenSourceReturnsNil(t *testing.T) {
	svc := NewService(&fakeSource{})
	items := svc.LoadHotTopics(10)
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(items))
	}
}

func TestApp_MarketLegacyReadsDelegateToService(t *testing.T) {
	app := &App{marketReadService: &stubMarketService{
		hotTopics: []map[string]any{{"title": "AI"}},
		longTiger: []models.LongTigerRankData{{Name: "示例"}},
	}}

	if got := app.HotTopic(10); len(got) != 1 {
		t.Fatalf("expected delegated hot topic result")
	}
	if got := app.LongTigerRank("2026-04-06"); len(*got) != 1 {
		t.Fatalf("expected delegated long-tiger result")
	}
}
```

- [ ] **Step 2: Run the backend tests to confirm the legacy reads are still uncovered**

Run:

```powershell
go test ./backend/service/market . -run "Test(Service_LoadRealtimePriceFallsBackThroughBidLevels|Service_LoadHotTopicsReturnsEmptySliceWhenSourceReturnsNil|App_MarketLegacyReadsDelegateToService)" -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Extend the `market` source and service with the remaining read-only helpers**

Add the residual source methods in `backend/source/marketnews/source.go` and map them in `backend/service/market/service.go`.

```go
type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeout uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
	GlobalStockIndexesReadable(crawlTimeout uint) string
	GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any
	GetMoneyRankSina(sort string) []map[string]any
	GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any
	LongTiger(date string) *[]models.LongTigerRankData
	StockResearchReport(stockCode string, days int) []any
	StockNotice(stockList string) []any
	IndustryResearchReport(industryCode string, days int) []any
	EMDictCode(code string) []any
	XueQiuHotStock(size int, marketType string) *[]models.HotItem
	HotEvent(size int) *[]models.HotEvent
	HotTopic(size int) []any
	InvestCalendar(yearMonth string) []any
	ClsCalendar() []any
	SearchStock(words string, size int) map[string]any
	HotStrategy() map[string]any
	GetStockKLine(stockCode string, days int64) *[]data.KLineData
	GetStockCommonKLine(stockCode string, days int64) *[]data.KLineData
	GetStockMinutePriceLineData(stockCode string) map[string]any
	GetStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) *[]data.KLineData
	GetStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) map[string]any
	GetStockRealtimePrice(stockCode string) map[string]any
}

func (s *Service) LoadRealtimePrice(stockCode string) RealtimePrice {
	raw := s.source.GetStockRealtimePrice(stockCode)
	price := firstPositiveFloat(raw["price"], raw["a1p"], raw["b1p"], raw["preClose"])
	if price == 0 {
		return RealtimePrice{Code: -1, Message: "获取股票价格失败"}
	}
	return RealtimePrice{
		Code:    0,
		Message: "success",
		Name:    asString(raw["name"]),
		Price:   price,
	}
}
```

- [ ] **Step 4: Expand the bridge interfaces and delegate the residual legacy methods**

Update the `marketReadService` interface in `app.go`, remove the remaining direct `data.NewMarketNewsApi` / `data.NewSearchStockApi` / quote reads in `app.go` and `app_common.go`, and delegate to `marketService` instead.

```go
type marketReadService interface {
	LoadFeeds() marketservice.FeedSet
	RefreshFeed(source string) marketservice.Feed
	LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet
	LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry
	LoadLongTiger(date string) []models.LongTigerRankData
	LoadStockResearchReport(stockCode string, days int) []any
	LoadStockNotice(stockCode string) []any
	LoadIndustryResearchReport(industryCode string, days int) []any
	LoadDictionaryCodes(code string) []any
	LoadHotStocks(marketType string, size int) []models.HotItem
	LoadHotEvents(size int) []models.HotEvent
	LoadHotTopics(size int) []any
	LoadInvestCalendar(yearMonth string) []any
	LoadClsCalendar() []any
	SearchStocks(words string, size int) map[string]any
	LoadHotStrategy() map[string]any
	LoadRealtimePrice(stockCode string) marketservice.RealtimePrice
	LoadStockMinuteLine(stockCode string) marketservice.MinuteLineResult
	LoadStockKLine(stockCode string, days int64) []marketservice.KLinePoint
	LoadStockCommonKLine(stockCode string, days int64) []marketservice.KLinePoint
	LoadEastMoneyKLinePage(stockCode, klt string, limit int, end string) []marketservice.KLinePoint
	LoadEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) marketservice.KLinePageResult
}

func (a *App) LongTigerRank(date string) *[]models.LongTigerRankData {
	items := a.marketReadService.LoadLongTiger(date)
	return &items
}

func (a *App) GetStockRealTimePrice(stockCode string) map[string]any {
	price := a.marketReadService.LoadRealtimePrice(stockCode)
	return map[string]any{
		"code":    price.Code,
		"message": price.Message,
		"price":   price.Price,
		"name":    price.Name,
	}
}
```

- [ ] **Step 5: Add frontend service wrappers for the residual market readers and shared shell reads**

Extend `frontend/src/services/marketService.mjs` and add `frontend/src/services/appShellService.mjs` so components stop importing Wails bindings directly.

```js
import * as AppBindings from '../../wailsjs/go/main/App.js'

export async function loadLongTigerRank(date) {
  return toArray(await AppBindings.LongTigerRank(date))
}

export async function loadHotTopics(size = 10) {
  return toArray(await AppBindings.HotTopic(size))
}

export async function loadRealtimePrice(stockCode) {
  const result = await AppBindings.GetStockRealTimePrice(stockCode)
  return {
    code: Number(result?.code ?? -1),
    message: result?.message ?? '获取股票价格失败',
    price: Number(result?.price ?? 0),
    name: result?.name ?? '',
  }
}
```

```js
import * as AppBindings from '../../wailsjs/go/main/App.js'
import { normalizeSettingConfig } from './configService.mjs'

export async function loadAppShellConfig() {
  return normalizeSettingConfig(await AppBindings.GetConfig())
}

export async function loadVersionInfo() {
  return await AppBindings.GetVersionInfo()
}

export async function loadGroupList() {
  const result = await AppBindings.GetGroupList()
  return Array.isArray(result) ? result : []
}
```

- [ ] **Step 6: Migrate the read-only frontend consumers**

Apply the wrappers to the residual readers, keeping the component behavior unchanged.

```vue
<script setup>
import { onMounted, ref } from 'vue'
import { loadHotTopics } from '../services/marketService.mjs'

const topics = ref([])

onMounted(async () => {
  topics.value = await loadHotTopics(10)
})
</script>
```

```vue
<script setup>
import { onMounted, ref } from 'vue'
import { loadAppShellConfig, loadVersionInfo } from '../services/appShellService.mjs'

const versionInfo = ref(null)
const shellConfig = ref(null)

onMounted(async () => {
  ;[versionInfo.value, shellConfig.value] = await Promise.all([
    loadVersionInfo(),
    loadAppShellConfig(),
  ])
})
</script>
```

- [ ] **Step 7: Run backend and frontend verification for the market closeout**

Run:

```powershell
go test ./backend/service/market . -run "Test(Service_LoadRealtimePriceFallsBackThroughBidLevels|Service_LoadHotTopicsReturnsEmptySliceWhenSourceReturnsNil|App_MarketLegacyReadsDelegateToService)" -count=1
node --test frontend/src/services/marketService.test.mjs frontend/src/services/appShellService.test.mjs
```

Expected:

```text
PASS
```

- [ ] **Step 8: Commit the market closeout slice**

```bash
git add backend/source/marketnews/source.go backend/service/market app.go app_common.go app_market_closeout_test.go frontend/src/services/marketService.mjs frontend/src/services/marketService.test.mjs frontend/src/services/appShellService.mjs frontend/src/services/appShellService.test.mjs frontend/src/App.vue frontend/src/components/SelectStock.vue frontend/src/components/KLineChart.vue frontend/src/components/StockLightweightKlineChart.vue frontend/src/components/stockSparkLine.vue frontend/src/components/LongTigerRankList.vue frontend/src/components/IndustryResearchReportList.vue frontend/src/components/StockResearchReportList.vue frontend/src/components/StockNoticeList.vue frontend/src/components/HotStockList.vue frontend/src/components/HotEvents.vue frontend/src/components/HotTopics.vue frontend/src/components/InvestCalendarTimeLine.vue frontend/src/components/ClsCalendarTimeLine.vue
git commit -m "refactor: close residual market read boundaries"
```

### Task 2: Move Watchlist, Price Alerts, And Research-Center CRUD Behind Service Boundaries

**Files:**
- Modify: `D:\codex_work\go-stock\backend\source\watchlist\store.go`
- Modify: `D:\codex_work\go-stock\backend\service\watchlist\service.go`
- Modify: `D:\codex_work\go-stock\backend\service\watchlist\service_test.go`
- Create: `D:\codex_work\go-stock\backend\source\research\store.go`
- Create: `D:\codex_work\go-stock\backend\service\research\contracts.go`
- Create: `D:\codex_work\go-stock\backend\service\research\service.go`
- Create: `D:\codex_work\go-stock\backend\service\research\service_test.go`
- Create: `D:\codex_work\go-stock\app_watchlist_research_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`
- Create: `D:\codex_work\go-stock\frontend\src\services\watchlistService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\watchlistService.test.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\researchService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\researchService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\components\stock.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\researchIndex.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\stockChangesMonitor.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\aiRecommendStocksList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\allStockList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\allStockInfoList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\TradingRecordManager.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\promptTemplateList.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\researchReport.vue`

- [ ] **Step 1: Write the failing backend tests for watchlist and research delegation**

```go
func TestWatchlistService_SetTradingPriceDelegatesToStore(t *testing.T) {
	store := &fakeStore{}
	svc := NewService(store, nil)

	result := svc.SetTradingPrice(context.Background(), "sz000001", 10, 12, 9, 10.5)

	if result != "操作成功" {
		t.Fatalf("expected success message, got %q", result)
	}
	if store.lastTradingStockCode != "sz000001" {
		t.Fatalf("expected stock code to be forwarded")
	}
}

func TestWatchlistService_EvaluateCostAlertsBuildsDeliveries(t *testing.T) {
	store := &fakeStore{
		follows: []data.FollowedStock{{StockCode: "sz000001", Name: "平安银行", CostPrice: 12}},
		quotes:  []marketservice.RealtimePrice{{Code: 0, Name: "平安银行", Price: 11.5}},
	}
	svc := NewService(store, nil)

	deliveries := svc.EvaluateCostAlerts(context.Background(), time.Now())

	if len(deliveries) != 1 {
		t.Fatalf("expected one alert delivery, got %d", len(deliveries))
	}
}

func TestResearchService_GetAiRecommendPageReturnsEmptyPageOnStoreError(t *testing.T) {
	svc := NewService(&fakeResearchStore{aiRecommendErr: errors.New("boom")})
	page := svc.GetAiRecommendPage(context.Background(), models.AiRecommendStocksQuery{})
	if page == nil || page.Total != 0 {
		t.Fatalf("expected empty page fallback")
	}
}

func TestApp_ResearchAndWatchlistMethodsDelegateToServices(t *testing.T) {
	app := &App{
		watchlistService: &stubWatchlistService{groups: []data.Group{{Name: "全部"}}},
		researchService:  &stubResearchService{markets: []string{"沪市"}},
	}
	if got := app.GetGroupList(); len(got) != 1 {
		t.Fatalf("expected delegated group list")
	}
	if got := app.GetAllMarkets(); len(got) != 1 {
		t.Fatalf("expected delegated market list")
	}
}
```

- [ ] **Step 2: Run the failing tests**

Run:

```powershell
go test ./backend/service/watchlist ./backend/service/research . -run "Test(WatchlistService_SetTradingPriceDelegatesToStore|WatchlistService_EvaluateCostAlertsBuildsDeliveries|ResearchService_GetAiRecommendPageReturnsEmptyPageOnStoreError|App_ResearchAndWatchlistMethodsDelegateToServices)" -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Extend `watchlist` for follow/group/alert operations**

Expand `backend/source/watchlist/store.go` so the stock shell and monitors stop touching `backend/data` directly.

```go
type Store interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string)
	GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock
	ListFollowedStocks(ctx context.Context) []data.FollowedStock
	Follow(ctx context.Context, stockCode string) string
	Unfollow(ctx context.Context, stockCode string) string
	ListGroups(ctx context.Context) []data.Group
	AddGroup(ctx context.Context, group data.Group) bool
	UpdateGroupSort(ctx context.Context, id int, newSort int) bool
	InitializeGroupSort(ctx context.Context) bool
	AddGroupStock(ctx context.Context, groupID int, stockCode string) bool
	RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) bool
	RemoveGroup(ctx context.Context, groupID int) bool
	SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string
	SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string
	SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string
	SetStockSort(ctx context.Context, stockCode string, sort int64)
}
```

Add `watchlist.Service` methods that return stable user messages and typed alert deliveries instead of firing notifications directly inside `app.go`.

```go
func (s *Service) EvaluateCostAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	_ = ctx
	_ = now
	return deliveries
}
```

- [ ] **Step 4: Add the new `research` service for stock changes, AI recommend records, stock-info CRUD, and trading records**

Implement `backend/source/research/store.go` and `backend/service/research/service.go`.

```go
type Service struct {
	store researchStore
}

func (s *Service) GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) StockChangePage {
	raw := s.store.GetStockChanges(ctx, changeTypes, pageIndex, pageSize)
	return normalizeStockChangePage(raw)
}

func (s *Service) GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	page, err := s.store.GetAiRecommendPage(ctx, query)
	if err != nil || page == nil {
		return &models.AiRecommendStocksPageData{}
	}
	return page
}

func (s *Service) EvaluateAiRecommendAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	_ = ctx
	_ = now
	return deliveries
}
```

- [ ] **Step 5: Delegate the remaining `app.go` / `app_common.go` methods and monitors**

Wire `researchService` into `App`, then delegate the remaining business methods and background monitors.

```go
type researchService interface {
	GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) researchservice.StockChangePage
	GetAllStockChanges(ctx context.Context, pageSize int) researchservice.StockChangePage
	GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData
	SaveStockChangesToHistory(ctx context.Context, changeTypes []int) string
	DeleteStockChangeHistory(ctx context.Context, days int) string
	GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData
	DeleteAiRecommend(ctx context.Context, id uint) string
	SetAiRecommendAlert(ctx context.Context, id uint, enable bool) string
	GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) *data.AllStockInfoPageData
	GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) *data.TradingRecordPageData
	EvaluateAiRecommendAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery
}

func MonitorAiRecommendStockPrices(a *App) {
	if a.researchService == nil || a.notificationService == nil {
		return
	}
	for _, delivery := range a.researchService.EvaluateAiRecommendAlerts(a.ctx, time.Now()) {
		a.notificationService.SendTyped(delivery.Message, delivery.StockCode, delivery.MsgType)
	}
}
```

- [ ] **Step 6: Add frontend watchlist and research service wrappers**

Create `frontend/src/services/watchlistService.mjs` and `frontend/src/services/researchService.mjs`.

```js
import * as AppBindings from '../../wailsjs/go/main/App.js'

export async function followStock(stockCode) {
  return AppBindings.Follow(stockCode)
}

export async function loadGroupList() {
  const result = await AppBindings.GetGroupList()
  return Array.isArray(result) ? result : []
}

export async function saveTradingPrice(stockCode, entryPrice, takeProfitPrice, stopLossPrice, costPrice) {
  return AppBindings.SetTradingPrice(stockCode, entryPrice, takeProfitPrice, stopLossPrice, costPrice)
}
```

```js
import * as AppBindings from '../../wailsjs/go/main/App.js'

export async function loadStockChanges(changeTypes, pageIndex, pageSize) {
  return normalizeStockChangePage(await AppBindings.GetStockChanges(changeTypes, pageIndex, pageSize))
}

export async function loadTradingRecordList(query) {
  return normalizeTradingRecordPage(await AppBindings.GetTradingRecordList(query))
}

export async function updateAiRecommendAlert(id, enableAlert) {
  return AppBindings.UpdateAiRecommendStocksAlert(id, enableAlert)
}
```

- [ ] **Step 7: Migrate the stock shell and research-center components**

Replace the remaining direct Wails imports with service calls.

```vue
<script setup>
import { followStock, unfollowStock, saveTradingPrice, loadGroupList } from '../services/watchlistService.mjs'
import { loadAppShellConfig, loadVersionInfo } from '../services/appShellService.mjs'
import { loadStockChanges, loadTradingRecordList } from '../services/researchService.mjs'
</script>
```

Keep the existing UI and event flow unchanged. This task is only about route/service boundary closure, not visual redesign.

- [ ] **Step 8: Verify the watchlist and research closeout**

Run:

```powershell
go test ./backend/service/watchlist ./backend/service/research . -run "Test(WatchlistService_SetTradingPriceDelegatesToStore|WatchlistService_EvaluateCostAlertsBuildsDeliveries|ResearchService_GetAiRecommendPageReturnsEmptyPageOnStoreError|App_ResearchAndWatchlistMethodsDelegateToServices)" -count=1
node --test frontend/src/services/watchlistService.test.mjs frontend/src/services/researchService.test.mjs
```

Expected:

```text
PASS
```

- [ ] **Step 9: Commit the watchlist and research closure**

```bash
git add backend/source/watchlist/store.go backend/service/watchlist/service.go backend/service/watchlist/service_test.go backend/source/research/store.go backend/service/research/contracts.go backend/service/research/service.go backend/service/research/service_test.go app.go app_common.go app_watchlist_research_test.go frontend/src/services/watchlistService.mjs frontend/src/services/watchlistService.test.mjs frontend/src/services/researchService.mjs frontend/src/services/researchService.test.mjs frontend/src/components/stock.vue frontend/src/components/researchIndex.vue frontend/src/components/stockChangesMonitor.vue frontend/src/components/aiRecommendStocksList.vue frontend/src/components/allStockList.vue frontend/src/components/allStockInfoList.vue frontend/src/components/TradingRecordManager.vue frontend/src/components/promptTemplateList.vue frontend/src/components/researchReport.vue
git commit -m "refactor: close watchlist and research boundaries"
```

### Task 3: Finish Remaining Route / Page Migration And Fence The Assistant Compatibility Island

**Files:**
- Create: `D:\codex_work\go-stock\backend\source\fund\adapter.go`
- Create: `D:\codex_work\go-stock\backend\service\fund\service.go`
- Create: `D:\codex_work\go-stock\backend\service\fund\service_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Create: `D:\codex_work\go-stock\frontend\src\services\fundService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\fundService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\services\appShellService.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\services\appShellService.test.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\assistantService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\assistantService.test.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\pages\fund-page.vue`
- Create: `D:\codex_work\go-stock\frontend\src\pages\about-page.vue`
- Create: `D:\codex_work\go-stock\frontend\src\pages\agent-page.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
- Modify: `D:\codex_work\go-stock\frontend\src\components\fund.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\about.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\agent-chat.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\FloatingAiAssistant.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\FloatingAgentAssistant.vue`
- Delete: `D:\codex_work\go-stock\frontend\src\components\agent-chat_bk.vue`

- [ ] **Step 1: Write failing tests for the fund boundary and assistant wrappers**

```go
func TestFundService_LoadFollowedFundsReturnsEmptySliceOnNil(t *testing.T) {
	svc := NewService(&fakeFundAdapter{})
	items := svc.LoadFollowedFunds(context.Background())
	if len(items) != 0 {
		t.Fatalf("expected empty slice")
	}
}

func TestApp_FundMethodsDelegateToService(t *testing.T) {
	app := &App{fundService: &stubFundService{funds: []data.FollowedFund{{Code: "510300"}}}}
	if got := app.GetFollowedFund(); len(got) != 1 {
		t.Fatalf("expected delegated followed fund list")
	}
}
```

```js
test('loadAssistantSession returns empty message list when backend returns null', async () => {
  AppBindings.GetAiAssistantSession = async () => null
  const result = await loadAssistantSession('session-1')
  assert.deepEqual(result.messages, [])
})
```

- [ ] **Step 2: Run the failing tests**

Run:

```powershell
go test ./backend/service/fund . -run "Test(FundService_LoadFollowedFundsReturnsEmptySliceOnNil|App_FundMethodsDelegateToService)" -count=1
node --test frontend/src/services/fundService.test.mjs frontend/src/services/assistantService.test.mjs
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Add the backend `fund` service and delegate fund operations**

```go
type fundService interface {
	LoadFundList(ctx context.Context, key string) []data.FundBasic
	LoadFollowedFunds(ctx context.Context) []data.FollowedFund
	FollowFund(ctx context.Context, fundCode string) string
	UnfollowFund(ctx context.Context, fundCode string) string
	RefreshFollowedFunds(ctx context.Context) error
}

func (a *App) GetfundList(key string) []data.FundBasic {
	return a.fundService.LoadFundList(a.ctx, key)
}

func MonitorFundPrices(a *App) {
	if a.fundService == nil {
		return
	}
	_ = a.fundService.RefreshFollowedFunds(a.ctx)
}
```

- [ ] **Step 4: Create the remaining page wrappers and route everything through `pages`**

Add page wrappers:

```vue
<script setup>
import FundView from '../components/fund.vue'
</script>

<template>
  <FundView />
</template>
```

```vue
<script setup>
import AboutView from '../components/about.vue'
</script>

<template>
  <AboutView />
</template>
```

```vue
<script setup>
import AgentView from '../components/agent-chat.vue'
</script>

<template>
  <AgentView />
</template>
```

Then update `frontend/src/router/router.js` so every route points at `../pages/**`.

- [ ] **Step 5: Wrap the shell and assistant compatibility bridge on the frontend**

Create `frontend/src/services/assistantService.mjs`:

```js
import * as AppBindings from '../../wailsjs/go/main/App.js'

export async function loadAssistantSession(sessionId) {
  const result = await AppBindings.GetAiAssistantSession(sessionId)
  return {
    sessionId,
    messages: Array.isArray(result?.messages) ? result.messages : [],
  }
}

export async function saveAssistantSession(sessionId, messages) {
  return AppBindings.SaveAiAssistantSession(sessionId, messages)
}

export async function startAgentChat(question, aiConfigId, sysPromptId, memoryMode, memoryCount, thinkingMode) {
  return AppBindings.ChatWithAgent(question, aiConfigId, sysPromptId, memoryMode, memoryCount, thinkingMode)
}

export async function abortAgentChat() {
  return AppBindings.AbortChatWithAgent()
}
```

Keep the backend compatibility allowlist explicit in `app.go` / `app_common.go` with comments near the retained methods:

```go
// Phase-5 compatibility allowlist:
// multi-turn assistant orchestration remains bridge-owned until a dedicated assistant refactor.
```

- [ ] **Step 6: Migrate the remaining route components and remove the stale backup file**

Update:

- `frontend/src/components/fund.vue`
- `frontend/src/components/about.vue`
- `frontend/src/components/agent-chat.vue`
- `frontend/src/components/FloatingAiAssistant.vue`
- `frontend/src/components/FloatingAgentAssistant.vue`

so they use `fundService`, `assistantService`, `configService`, and `appShellService` instead of direct Wails imports.

Delete:

```text
frontend/src/components/agent-chat_bk.vue
```

- [ ] **Step 7: Verify the route / compatibility closure**

Run:

```powershell
go test ./backend/service/fund . -run "Test(FundService_LoadFollowedFundsReturnsEmptySliceOnNil|App_FundMethodsDelegateToService)" -count=1
node --test frontend/src/services/fundService.test.mjs frontend/src/services/assistantService.test.mjs frontend/src/services/appShellService.test.mjs
```

Expected:

```text
PASS
```

- [ ] **Step 8: Commit the route / assistant closeout**

```bash
git add backend/source/fund/adapter.go backend/service/fund/service.go backend/service/fund/service_test.go app.go frontend/src/services/fundService.mjs frontend/src/services/fundService.test.mjs frontend/src/services/assistantService.mjs frontend/src/services/assistantService.test.mjs frontend/src/services/appShellService.mjs frontend/src/services/appShellService.test.mjs frontend/src/pages/fund-page.vue frontend/src/pages/about-page.vue frontend/src/pages/agent-page.vue frontend/src/router/router.js frontend/src/components/fund.vue frontend/src/components/about.vue frontend/src/components/agent-chat.vue frontend/src/components/FloatingAiAssistant.vue frontend/src/components/FloatingAgentAssistant.vue
git rm frontend/src/components/agent-chat_bk.vue
git commit -m "refactor: finish route pages and assistant compatibility fence"
```

### Task 4: Add Architecture Guards And Run Final Phase-5 Verification

**Files:**
- Create: `D:\codex_work\go-stock\architecture_closeout_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`

- [ ] **Step 1: Add a guard test for backend direct-data regression**

Create `architecture_closeout_test.go`:

```go
func TestAppBridgeNoLongerOwnsBusinessDataApis(t *testing.T) {
	appGo := mustReadFile(t, "app.go")
	appCommon := mustReadFile(t, "app_common.go")
	raw := appGo + "\n" + appCommon

	disallowed := []string{
		"data.NewMarketNewsApi()",
		"data.NewSearchStockApi(",
		"data.NewStockDataApi()",
		"data.NewAiRecommendStocksService()",
		"data.NewStockChangesApi()",
		"data.NewStockChangeHistoryService()",
		"data.NewFundApi()",
		"data.NewStockGroupApi(",
		"db.Dao",
	}

	for _, token := range disallowed {
		if strings.Contains(raw, token) {
			t.Fatalf("forbidden legacy bridge token still present: %s", token)
		}
	}
}
```

- [ ] **Step 2: Add a guard test for frontend direct Wails imports and route-page mapping**

Extend the same file:

```go
func TestFrontendDirectWailsImportsAreLimitedToServices(t *testing.T) {
	root := filepath.Join("frontend", "src")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".vue") && !strings.HasSuffix(path, ".mjs") && !strings.HasSuffix(path, ".js") {
			return nil
		}
		if strings.Contains(path, filepath.Join("frontend", "src", "services")) {
			return nil
		}
		raw := mustReadFile(t, path)
		if strings.Contains(raw, "wailsjs/go/main/App") {
			t.Fatalf("direct App binding import outside services: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRouterUsesPagesForAllRouteComponents(t *testing.T) {
	raw := mustReadFile(t, filepath.Join("frontend", "src", "router", "router.js"))
	if strings.Contains(raw, "../components/") {
		t.Fatalf("router still imports route components directly from components")
	}
}
```

- [ ] **Step 3: Run the architecture guards first**

Run:

```powershell
go test . -run "Test(AppBridgeNoLongerOwnsBusinessDataApis|FrontendDirectWailsImportsAreLimitedToServices|RouterUsesPagesForAllRouteComponents)" -count=1
```

Expected:

```text
PASS
```

- [ ] **Step 4: Run the final backend verification**

Run:

```powershell
go test ./backend/service/market ./backend/service/watchlist ./backend/service/research ./backend/service/fund ./backend/service/config ./backend/service/analysis ./backend/service/notification ./backend/service/task . -count=1
```

Expected:

```text
PASS
```

- [ ] **Step 5: Run the final frontend verification**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs frontend/src/services/appShellService.test.mjs frontend/src/services/watchlistService.test.mjs frontend/src/services/researchService.test.mjs frontend/src/services/fundService.test.mjs frontend/src/services/assistantService.test.mjs
```

Expected:

```text
PASS
```

- [ ] **Step 6: Run the production build verification**

Run:

```powershell
npm run build
```

Working directory:

```text
D:\codex_work\go-stock\frontend
```

Expected:

```text
vite build completed successfully
```

- [ ] **Step 7: Run the explicit closure greps**

Run:

```powershell
rg -n "wailsjs/go/main/App" frontend/src --glob '!frontend/src/services/**' --glob '!frontend/wailsjs/**'
rg -n "data\\.New|db\\.Dao" app.go app_common.go
rg -n "\\.\\./components/" frontend/src/router/router.js
```

Expected:

```text
no output
```

If any allowlisted helper still appears in `app.go` / `app_common.go`, fix the guard test and the grep expectation together. Do not leave undocumented exceptions.

- [ ] **Step 8: Commit the guard and verification closeout**

```bash
git add architecture_closeout_test.go app.go app_common.go
git commit -m "test: add architecture closeout guards"
```

## Self-Review

### Spec Coverage

- The overall spec’s phase 5 requirement to shrink old paths in `app.go` / `app_common.go` is covered by Tasks 1, 2, and 4.
- The requirement that core routes go through `pages` is covered by Task 3 and enforced again by Task 4.
- The requirement that frontend deep components stop calling Wails directly is covered by Tasks 1, 2, and 3, and guarded by Task 4.
- The phase 4 requirement that alert / notification logic stop scattering across bridge code is completed by Task 2.
- The overall spec’s non-goal for deep multi-turn assistant rewrite is respected by Task 3 via an explicit compatibility fence instead of a broad rewrite.

### Placeholder Scan

- No `TODO`, `TBD`, “implement later”, or “similar to Task N” placeholders remain.
- Every code-changing step names concrete files, method names, and commands.
- The final grep expectations are explicit rather than implied.

### Type Consistency

- Existing service names are intentionally preserved:
  - `marketService`
  - `analysisService`
  - `configService`
  - `taskService`
  - `notificationService`
  - `watchlistService`
- New service names are used consistently across tasks:
  - `researchService`
  - `fundService`
  - `assistantService` (frontend wrapper only)
  - `appShellService`
