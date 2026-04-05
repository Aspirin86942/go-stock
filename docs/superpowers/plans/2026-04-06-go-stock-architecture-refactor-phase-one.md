# go-stock Architecture Refactor Phase One Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the first `source / service / pages / services` boundaries and migrate one read-only market chain end-to-end without changing the Go/Wails/Vue stack.

**Architecture:** Introduce a thin `backend/source/marketnews` adapter over the existing `backend/data` implementation, then build a typed `backend/service/market` layer that the Wails bridge can call instead of directly touching `backend/data`. On the frontend, route `/market` through a page wrapper and move the read-only market data loading path into `frontend/src/services/marketService.mjs`, while keeping the existing large market screen reusable during the transition.

**Tech Stack:** Go 1.26, Wails v2, Vue 3, Naive UI, Node `--test`, Vite, `go test`

---

## File Map

- Create: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
  - Thin adapter over `backend/data.MarketNewsApi` for read-only market data.
- Create: `D:\codex_work\go-stock\backend\service\market\contracts.go`
  - Typed contracts for market feeds, global index groups, and industry rank rows.
- Create: `D:\codex_work\go-stock\backend\service\market\service.go`
  - Market read service that loads typed contracts from the new source adapter.
- Create: `D:\codex_work\go-stock\backend\service\market\service_test.go`
  - Unit tests for feed loading, refresh behavior, index normalization, and industry-rank normalization.
- Create: `D:\codex_work\go-stock\app_market_test.go`
  - Root-package bridge tests that prove new App methods delegate to the market read service.
- Modify: `D:\codex_work\go-stock\app.go`
  - Add bridge-level market read service wiring and new typed methods.
- Create: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
  - Frontend wrapper around the new Wails methods with normalization helpers.
- Create: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
  - Node tests for the frontend normalization helpers.
- Create: `D:\codex_work\go-stock\frontend\src\pages\market-page.vue`
  - Page-level wrapper for the existing market screen.
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
  - Point `/market` at the new page wrapper.
- Modify: `D:\codex_work\go-stock\frontend\src\components\market.vue`
  - Switch the read-only market data path to the new frontend service wrappers.
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.d.ts`
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.js`

## Scope Note

This plan intentionally covers only the first executable slice of the larger architecture refactor:

- Phase 1: establish new boundaries
- Phase 2: migrate one read-only market chain

AI analysis migration, cron/task migration, settings consolidation, and old-path cleanup remain follow-on plans after this slice is working and verified.

### Task 1: Add Typed Market Read Contracts And Service Tests

**Files:**
- Create: `D:\codex_work\go-stock\backend\service\market\contracts.go`
- Create: `D:\codex_work\go-stock\backend\service\market\service.go`
- Create: `D:\codex_work\go-stock\backend\service\market\service_test.go`

- [ ] **Step 1: Write the failing service tests**

```go
package market

import (
	"testing"

	"go-stock/backend/models"
)

type fakeSource struct {
	telegraphs   map[string][]*models.Telegraph
	indexes      map[string]any
	industryData map[string]any
	refreshCount int
}

func (f *fakeSource) GetTelegraphList(source string) *[]*models.Telegraph {
	items := append([]*models.Telegraph(nil), f.telegraphs[source]...)
	return &items
}

func (f *fakeSource) RefreshFeeds() {
	f.refreshCount++
}

func (f *fakeSource) GlobalStockIndexes(crawlTimeout uint) map[string]any {
	return f.indexes
}

func (f *fakeSource) GetIndustryRank(sort string, cnt int) map[string]any {
	return f.industryData
}

func TestService_LoadReadModel_NormalizesTypedContracts(t *testing.T) {
	svc := NewService(&fakeSource{
		telegraphs: map[string][]*models.Telegraph{
			"财联社电报": {{Content: "telegraph-1"}},
			"新浪财经":   {{Content: "sina-1"}},
			"外媒":     {{Content: "foreign-1"}},
		},
		indexes: map[string]any{
			"common": []any{
				map[string]any{
					"code":     "000001.SH",
					"name":     "上证指数",
					"location": "上海",
					"qtcode":   "sh000001",
					"state":    "open",
					"zdf":      "1.23",
					"zxj":      "3210.88",
					"img":      "https://example.com/sh.png",
				},
			},
		},
		industryData: map[string]any{
			"data": []any{
				map[string]any{
					"bd_code":  "BK0420",
					"bd_name":  "机器人",
					"bd_zdf":   "2.66",
					"bd_zdf5":  "5.00",
					"bd_zdf20": "10.50",
					"nzg_code": "300024",
					"nzg_name": "机器人龙头",
					"nzg_zdf":  "7.21",
					"nzg_zxj":  "15.66",
				},
			},
		},
	})

	feeds := svc.LoadFeeds()
	if len(feeds.Telegraph) != 1 || feeds.Telegraph[0].Content != "telegraph-1" {
		t.Fatalf("unexpected telegraph feed: %#v", feeds.Telegraph)
	}
	if len(feeds.Sina) != 1 || feeds.Sina[0].Content != "sina-1" {
		t.Fatalf("unexpected sina feed: %#v", feeds.Sina)
	}
	if len(feeds.Foreign) != 1 || feeds.Foreign[0].Content != "foreign-1" {
		t.Fatalf("unexpected foreign feed: %#v", feeds.Foreign)
	}

	indexes := svc.LoadGlobalIndexes(30)
	if len(indexes.Common) != 1 || indexes.Common[0].Name != "上证指数" {
		t.Fatalf("unexpected common index group: %#v", indexes.Common)
	}
	if indexes.Common[0].Region != "common" {
		t.Fatalf("expected region common, got %#v", indexes.Common[0])
	}

	ranks := svc.LoadIndustryRanks("0", 20)
	if len(ranks) != 1 || ranks[0].BoardName != "机器人" {
		t.Fatalf("unexpected industry ranks: %#v", ranks)
	}
	if ranks[0].LeaderCode != "300024" {
		t.Fatalf("unexpected leader code: %#v", ranks[0])
	}
}

func TestService_RefreshFeed_RefreshesBeforeReload(t *testing.T) {
	src := &fakeSource{
		telegraphs: map[string][]*models.Telegraph{
			"新浪财经": {{Content: "refresh-me"}},
		},
	}
	svc := NewService(src)

	feed := svc.RefreshFeed("新浪财经")
	if src.refreshCount != 1 {
		t.Fatalf("expected refresh to be called once, got %d", src.refreshCount)
	}
	if feed.Source != "新浪财经" {
		t.Fatalf("unexpected feed source: %#v", feed)
	}
	if len(feed.Items) != 1 || feed.Items[0].Content != "refresh-me" {
		t.Fatalf("unexpected feed items: %#v", feed.Items)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/market -run "TestService_LoadReadModel_NormalizesTypedContracts|TestService_RefreshFeed_RefreshesBeforeReload" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing package/files under `backend/service/market`

- [ ] **Step 3: Write the typed contracts and minimal service implementation**

```go
// D:\codex_work\go-stock\backend\service\market\contracts.go
package market

import "go-stock/backend/models"

type FeedSet struct {
	Telegraph []*models.Telegraph `json:"telegraph"`
	Sina      []*models.Telegraph `json:"sina"`
	Foreign   []*models.Telegraph `json:"foreign"`
}

type Feed struct {
	Source string              `json:"source"`
	Items  []*models.Telegraph `json:"items"`
}

type IndexSet struct {
	Common  []GlobalIndexEntry `json:"common"`
	America []GlobalIndexEntry `json:"america"`
	Europe  []GlobalIndexEntry `json:"europe"`
	Asia    []GlobalIndexEntry `json:"asia"`
	Other   []GlobalIndexEntry `json:"other"`
}

type GlobalIndexEntry struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Qtcode   string `json:"qtcode"`
	State    string `json:"state"`
	Zdf      string `json:"zdf"`
	Zxj      string `json:"zxj"`
	Img      string `json:"img"`
	Region   string `json:"region"`
}

type IndustryRankEntry struct {
	BoardCode             string `json:"boardCode"`
	BoardName             string `json:"boardName"`
	BoardChangePercent    string `json:"boardChangePercent"`
	BoardChangePercent5D  string `json:"boardChangePercent5D"`
	BoardChangePercent20D string `json:"boardChangePercent20D"`
	LeaderCode            string `json:"leaderCode"`
	LeaderName            string `json:"leaderName"`
	LeaderChangePercent   string `json:"leaderChangePercent"`
	LeaderPrice           string `json:"leaderPrice"`
}
```

```go
// D:\codex_work\go-stock\backend\service\market\service.go
package market

import (
	"go-stock/backend/models"

	"github.com/duke-git/lancet/v2/convertor"
)

type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeout uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
}

type Service struct {
	source Source
}

func NewService(source Source) *Service {
	return &Service{source: source}
}

func (s *Service) LoadFeeds() FeedSet {
	return FeedSet{
		Telegraph: cloneTelegraphs(s.source.GetTelegraphList("财联社电报")),
		Sina:      cloneTelegraphs(s.source.GetTelegraphList("新浪财经")),
		Foreign:   cloneTelegraphs(s.source.GetTelegraphList("外媒")),
	}
}

func (s *Service) RefreshFeed(source string) Feed {
	s.source.RefreshFeeds()
	return Feed{
		Source: source,
		Items:  cloneTelegraphs(s.source.GetTelegraphList(source)),
	}
}

func (s *Service) LoadGlobalIndexes(crawlTimeout uint) IndexSet {
	raw := s.source.GlobalStockIndexes(crawlTimeout)
	return IndexSet{
		Common:  mapIndexEntries("common", raw["common"]),
		America: mapIndexEntries("america", raw["america"]),
		Europe:  mapIndexEntries("europe", raw["europe"]),
		Asia:    mapIndexEntries("asia", raw["asia"]),
		Other:   mapIndexEntries("other", raw["other"]),
	}
}

func (s *Service) LoadIndustryRanks(sort string, cnt int) []IndustryRankEntry {
	raw := s.source.GetIndustryRank(sort, cnt)
	list, _ := raw["data"].([]any)
	result := make([]IndustryRankEntry, 0, len(list))
	for _, item := range list {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, IndustryRankEntry{
			BoardCode:             convertor.ToString(row["bd_code"]),
			BoardName:             convertor.ToString(row["bd_name"]),
			BoardChangePercent:    convertor.ToString(row["bd_zdf"]),
			BoardChangePercent5D:  convertor.ToString(row["bd_zdf5"]),
			BoardChangePercent20D: convertor.ToString(row["bd_zdf20"]),
			LeaderCode:            convertor.ToString(row["nzg_code"]),
			LeaderName:            convertor.ToString(row["nzg_name"]),
			LeaderChangePercent:   convertor.ToString(row["nzg_zdf"]),
			LeaderPrice:           convertor.ToString(row["nzg_zxj"]),
		})
	}
	return result
}

func cloneTelegraphs(ptr *[]*models.Telegraph) []*models.Telegraph {
	if ptr == nil {
		return []*models.Telegraph{}
	}
	return append([]*models.Telegraph(nil), (*ptr)...)
}

func mapIndexEntries(region string, raw any) []GlobalIndexEntry {
	list, _ := raw.([]any)
	result := make([]GlobalIndexEntry, 0, len(list))
	for _, item := range list {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, GlobalIndexEntry{
			Code:     convertor.ToString(row["code"]),
			Name:     convertor.ToString(row["name"]),
			Location: convertor.ToString(row["location"]),
			Qtcode:   convertor.ToString(row["qtcode"]),
			State:    convertor.ToString(row["state"]),
			Zdf:      convertor.ToString(row["zdf"]),
			Zxj:      convertor.ToString(row["zxj"]),
			Img:      convertor.ToString(row["img"]),
			Region:   region,
		})
	}
	return result
}
```

- [ ] **Step 4: Run the service tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/market -run "TestService_LoadReadModel_NormalizesTypedContracts|TestService_RefreshFeed_RefreshesBeforeReload" -count=1
```

Expected:

- command exits `0`
- both tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/market/contracts.go backend/service/market/service.go backend/service/market/service_test.go
git commit -m "refactor: add market read service contracts"
```

### Task 2: Bridge The New Market Service Through `app.go`

**Files:**
- Create: `D:\codex_work\go-stock\backend\source\marketnews\source.go`
- Create: `D:\codex_work\go-stock\app_market_test.go`
- Modify: `D:\codex_work\go-stock\app.go`

- [ ] **Step 1: Write the failing App bridge test**

```go
package main

import (
	"testing"

	marketservice "go-stock/backend/service/market"
	"go-stock/backend/models"
)

type fakeMarketReadService struct {
	feeds         marketservice.FeedSet
	refreshedFeed marketservice.Feed
	indexes       marketservice.IndexSet
	ranks         []marketservice.IndustryRankEntry
	refreshSource string
}

func (f *fakeMarketReadService) LoadFeeds() marketservice.FeedSet {
	return f.feeds
}

func (f *fakeMarketReadService) RefreshFeed(source string) marketservice.Feed {
	f.refreshSource = source
	return f.refreshedFeed
}

func (f *fakeMarketReadService) LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet {
	return f.indexes
}

func (f *fakeMarketReadService) LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return f.ranks
}

func TestApp_MarketReadMethodsDelegateToService(t *testing.T) {
	app := NewApp()
	fake := &fakeMarketReadService{
		feeds: marketservice.FeedSet{
			Telegraph: []*models.Telegraph{{Content: "telegraph-1"}},
		},
		refreshedFeed: marketservice.Feed{
			Source: "新浪财经",
			Items:  []*models.Telegraph{{Content: "refresh-1"}},
		},
		indexes: marketservice.IndexSet{
			Common: []marketservice.GlobalIndexEntry{{Name: "上证指数", Region: "common"}},
		},
		ranks: []marketservice.IndustryRankEntry{{BoardName: "机器人"}},
	}
	app.marketReadService = fake

	if got := app.GetMarketFeeds(); len(got.Telegraph) != 1 {
		t.Fatalf("unexpected market feeds: %#v", got)
	}
	if got := app.GetMarketGlobalIndexes(); len(got.Common) != 1 {
		t.Fatalf("unexpected market indexes: %#v", got)
	}
	if got := app.GetMarketIndustryRanks("0", 150); len(got) != 1 || got[0].BoardName != "机器人" {
		t.Fatalf("unexpected market industry ranks: %#v", got)
	}
	if got := app.RefreshMarketFeed("新浪财经"); got.Source != "新浪财经" {
		t.Fatalf("unexpected refreshed feed: %#v", got)
	}
	if fake.refreshSource != "新浪财经" {
		t.Fatalf("expected refresh source 新浪财经, got %q", fake.refreshSource)
	}
}
```

- [ ] **Step 2: Run the targeted test to verify it fails**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . -run "TestApp_MarketReadMethodsDelegateToService" -count=1
```

Expected:

- command exits non-zero
- failure mentions `marketReadService` or the new market bridge methods do not exist

- [ ] **Step 3: Add the source adapter and App bridge wiring**

```go
// D:\codex_work\go-stock\backend\source\marketnews\source.go
package marketnews

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Source struct {
	api data.MarketNewsApi
}

func NewSource() *Source {
	return &Source{api: data.NewMarketNewsApi()}
}

func (s *Source) GetTelegraphList(source string) *[]*models.Telegraph {
	return s.api.GetTelegraphList(source)
}

func (s *Source) RefreshFeeds() {
	go s.api.TelegraphList(30)
	go s.api.GetSinaNews(30)
	go s.api.TradingViewNews()
}

func (s *Source) GlobalStockIndexes(crawlTimeout uint) map[string]any {
	return s.api.GlobalStockIndexes(crawlTimeout)
}

func (s *Source) GetIndustryRank(sort string, cnt int) map[string]any {
	return s.api.GetIndustryRank(sort, cnt)
}
```

```go
// Add to D:\codex_work\go-stock\app.go imports
import (
	marketservice "go-stock/backend/service/market"
	marketsource "go-stock/backend/source/marketnews"
)

type marketReadService interface {
	LoadFeeds() marketservice.FeedSet
	RefreshFeed(source string) marketservice.Feed
	LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet
	LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry
}

type App struct {
	// existing fields...
	marketReadService marketReadService
}

func NewApp() *App {
	cacheSize := 512 * 1024
	cache := freecache.NewCache(cacheSize)
	c := cron.New(cron.WithSeconds())
	c.Start()
	var tools []data.Tool
	tools = data.Tools(tools)
	return &App{
		cache:              cache,
		cron:               c,
		cronEntrys:         make(map[string]cron.EntryID),
		AiTools:            tools,
		marketReadService:  marketservice.NewService(marketsource.NewSource()),
		stockAlertLastSent: make(map[string]time.Time),
		priceAtAlertReset:  make(map[string]float64),
	}
}

func (a *App) GetMarketFeeds() marketservice.FeedSet {
	return a.marketReadService.LoadFeeds()
}

func (a *App) RefreshMarketFeed(source string) marketservice.Feed {
	return a.marketReadService.RefreshFeed(source)
}

func (a *App) GetMarketGlobalIndexes() marketservice.IndexSet {
	return a.marketReadService.LoadGlobalIndexes(30)
}

func (a *App) GetMarketIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return a.marketReadService.LoadIndustryRanks(sort, cnt)
}
```

- [ ] **Step 4: Run the bridge and service tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/market -run "TestApp_MarketReadMethodsDelegateToService|TestService_LoadReadModel_NormalizesTypedContracts|TestService_RefreshFeed_RefreshesBeforeReload" -count=1
```

Expected:

- command exits `0`
- App test and market service tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/source/marketnews/source.go app.go app_market_test.go
git commit -m "refactor: bridge market read service through app"
```

### Task 3: Add Frontend Market Service Wrappers And Tests

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`

- [ ] **Step 1: Write the failing frontend service tests**

```javascript
import test from 'node:test'
import assert from 'node:assert/strict'

import {
  normalizeMarketFeed,
  normalizeMarketFeeds,
  normalizeMarketIndexes,
  normalizeMarketIndustryRanks,
} from './marketService.mjs'

test('normalizeMarketFeeds fills missing arrays with empty lists', () => {
  assert.deepEqual(normalizeMarketFeeds({ telegraph: [{ content: 't1' }] }), {
    telegraph: [{ content: 't1' }],
    sina: [],
    foreign: [],
  })
})

test('normalizeMarketFeed keeps the selected source and list', () => {
  assert.deepEqual(
    normalizeMarketFeed({
      source: '新浪财经',
      items: [{ content: 'n1' }],
    }),
    {
      source: '新浪财经',
      items: [{ content: 'n1' }],
    },
  )
})

test('normalizeMarketIndexes keeps only the known regions', () => {
  assert.deepEqual(
    normalizeMarketIndexes({
      common: [{ name: '上证指数' }],
      america: [{ name: '道琼斯' }],
    }),
    {
      common: [{ name: '上证指数' }],
      america: [{ name: '道琼斯' }],
      europe: [],
      asia: [],
      other: [],
    },
  )
})

test('normalizeMarketIndustryRanks falls back to an empty array', () => {
  assert.deepEqual(normalizeMarketIndustryRanks(null), [])
})
```

- [ ] **Step 2: Run the frontend service tests to verify they fail**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs
```

Expected:

- command exits non-zero
- failure mentions `frontend/src/services/marketService.mjs` does not exist

- [ ] **Step 3: Add the frontend market service module**

```javascript
// D:\codex_work\go-stock\frontend\src\services\marketService.mjs
import * as AppBindings from '../../wailsjs/go/main/App'

export function normalizeMarketFeeds(raw) {
  return {
    telegraph: Array.isArray(raw?.telegraph) ? raw.telegraph : [],
    sina: Array.isArray(raw?.sina) ? raw.sina : [],
    foreign: Array.isArray(raw?.foreign) ? raw.foreign : [],
  }
}

export function normalizeMarketFeed(raw) {
  return {
    source: String(raw?.source || ''),
    items: Array.isArray(raw?.items) ? raw.items : [],
  }
}

export function normalizeMarketIndexes(raw) {
  return {
    common: Array.isArray(raw?.common) ? raw.common : [],
    america: Array.isArray(raw?.america) ? raw.america : [],
    europe: Array.isArray(raw?.europe) ? raw.europe : [],
    asia: Array.isArray(raw?.asia) ? raw.asia : [],
    other: Array.isArray(raw?.other) ? raw.other : [],
  }
}

export function normalizeMarketIndustryRanks(raw) {
  return Array.isArray(raw) ? raw : []
}

export async function loadMarketFeeds() {
  return normalizeMarketFeeds(await AppBindings.GetMarketFeeds())
}

export async function loadMarketGlobalIndexes() {
  return normalizeMarketIndexes(await AppBindings.GetMarketGlobalIndexes())
}

export async function loadMarketIndustryRanks(sort = '0', count = 150) {
  return normalizeMarketIndustryRanks(await AppBindings.GetMarketIndustryRanks(sort, count))
}

export async function refreshMarketFeed(source) {
  return normalizeMarketFeed(await AppBindings.RefreshMarketFeed(source))
}
```

- [ ] **Step 4: Run the frontend service tests**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs
```

Expected:

- command exits `0`
- all four tests pass

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/services/marketService.mjs frontend/src/services/marketService.test.mjs
git commit -m "refactor: add frontend market service wrappers"
```

### Task 4: Route `/market` Through `pages` And Switch The Read-Only Chain To The New Service

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\pages\market-page.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\router\router.js`
- Modify: `D:\codex_work\go-stock\frontend\src\components\market.vue`
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.d.ts`
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.js`

- [ ] **Step 1: Confirm the current market route still points directly at the heavy component**

Run:

```powershell
rg -n "components/market\\.vue|GetTelegraphList|GlobalStockIndexes|GetIndustryRank|ReFleshTelegraphList" frontend/src/router/router.js frontend/src/components/market.vue
```

Expected:

- `frontend/src/router/router.js` still imports `../components/market.vue`
- `frontend/src/components/market.vue` still imports the raw Wails market read methods

- [ ] **Step 2: Add the page wrapper and route `/market` through it**

```vue
<!-- D:\codex_work\go-stock\frontend\src\pages\market-page.vue -->
<script setup>
import MarketScreen from '../components/market.vue'
</script>

<template>
  <MarketScreen />
</template>
```

```javascript
// D:\codex_work\go-stock\frontend\src\router\router.js
import marketPage from '../pages/market-page.vue'

const routes = [
  { path: '/', component: stockView, name: 'stock' },
  { path: '/fund', component: fundView, name: 'fund' },
  { path: '/settings', component: settingsView, name: 'settings' },
  { path: '/about', component: aboutView, name: 'about' },
  { path: '/market', component: marketPage, name: 'market' },
  { path: '/agent', component: agentChat, name: 'agent' },
  { path: '/research', component: research, name: 'research' },
  { path: '/cron-tasks', component: cronTaskManager, name: 'cronTasks' },
]
```

- [ ] **Step 3: Switch the read-only market data path in `market.vue` to the new frontend service**

```vue
<script setup>
import {
  GetAIResponseResult,
  GetConfig,
  GetPromptTemplates,
  SaveAIResponseResult,
  SaveAsMarkdown,
  ShareAnalysis,
  SummaryStockNews,
  GetAiConfigs,
} from "../../wailsjs/go/main/App";
import {
  loadMarketFeeds,
  loadMarketGlobalIndexes,
  loadMarketIndustryRanks,
  refreshMarketFeed,
} from "../services/marketService.mjs";

function applyMarketIndexes(indexes) {
  globalStockIndexes.value = indexes
  common.value = indexes.common
  america.value = indexes.america
  europe.value = indexes.europe
  asia.value = indexes.asia
  other.value = indexes.other
}

async function getIndex() {
  applyMarketIndexes(await loadMarketGlobalIndexes())
}

async function loadFeeds() {
  const feeds = await loadMarketFeeds()
  telegraphList.value = feeds.telegraph
  sinaNewsList.value = feeds.sina
  foreignNewsList.value = feeds.foreign
}

async function industryRank() {
  const ranks = await loadMarketIndustryRanks(sort.value, 150)
  if (ranks.length > 0) {
    industryRanks.value = ranks
    return
  }
  message.info("暂无数据")
}

onBeforeMount(() => {
  nowTab.value = route.query.name
  stockCode.value = route.query.stockCode
  GetConfig().then(result => {
    summaryBTN.value = result.openAiEnable
    darkTheme.value = result.darkTheme
    httpProxyEnabled.value = result.httpProxyEnabled
  })
  GetPromptTemplates("", "").then(res => {
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

function ReFlesh(source) {
  refreshMarketFeed(source).then(feed => {
    if (feed.source === "财联社电报") {
      telegraphList.value = feed.items
    }
    if (feed.source === "新浪财经") {
      sinaNewsList.value = feed.items
    }
    if (feed.source === "外媒") {
      foreignNewsList.value = feed.items
    }
  })
}
</script>
```

- [ ] **Step 4: Regenerate bindings and run focused verification**

Run:

```powershell
$env:PATH='C:\Program Files\Go\bin;' + $env:PATH
& 'C:\Users\Aspir\go\bin\wails.exe' build --platform windows/amd64
node --test frontend/src/services/marketService.test.mjs
npm --prefix frontend run build
```

Expected:

- Wails build exits `0`
- `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/main/App.js` include `GetMarketFeeds`, `RefreshMarketFeed`, `GetMarketGlobalIndexes`, and `GetMarketIndustryRanks`
- `node --test` exits `0`
- frontend build exits `0`

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/pages/market-page.vue frontend/src/router/router.js frontend/src/components/market.vue frontend/wailsjs/go/main/App.d.ts frontend/wailsjs/go/main/App.js
git commit -m "refactor: migrate market read path to page and service layers"
```

### Task 5: Run End-To-End Slice Verification And Review The Diff

**Files:**
- Verify only:
  - `D:\codex_work\go-stock\backend\source\marketnews\source.go`
  - `D:\codex_work\go-stock\backend\service\market\contracts.go`
  - `D:\codex_work\go-stock\backend\service\market\service.go`
  - `D:\codex_work\go-stock\backend\service\market\service_test.go`
  - `D:\codex_work\go-stock\app.go`
  - `D:\codex_work\go-stock\app_market_test.go`
  - `D:\codex_work\go-stock\frontend\src\services\marketService.mjs`
  - `D:\codex_work\go-stock\frontend\src\services\marketService.test.mjs`
  - `D:\codex_work\go-stock\frontend\src\pages\market-page.vue`
  - `D:\codex_work\go-stock\frontend\src\router\router.js`
  - `D:\codex_work\go-stock\frontend\src\components\market.vue`
  - `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.d.ts`
  - `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.js`

- [ ] **Step 1: Run the backend slice tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/market -run "TestApp_MarketReadMethodsDelegateToService|TestService_LoadReadModel_NormalizesTypedContracts|TestService_RefreshFeed_RefreshesBeforeReload" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 2: Run the frontend slice test**

Run:

```powershell
node --test frontend/src/services/marketService.test.mjs
```

Expected:

- command exits `0`

- [ ] **Step 3: Run the frontend production build**

Run:

```powershell
npm --prefix frontend run build
```

Expected:

- command exits `0`

- [ ] **Step 4: Review the diff is limited to the planned slice**

Run:

```powershell
git diff --stat HEAD~3..HEAD
```

Expected:

- diff is limited to the new market source/service files, App bridge wiring, the frontend market service, the market page wrapper, route updates, the market component, and generated Wails bindings

- [ ] **Step 5: Optional integration note for the follow-on plan**

Record in the next planning session that:

- the first `source / service / pages / services` slice now exists
- phase 3 should target the AI analysis chain next
- old direct market read methods can stay temporarily, but no new read-only market code should be added to the old path

## Self-Review

### Spec Coverage

- Phase 1 boundary creation is covered by Tasks 1, 2, and 3.
- Phase 2 first read-only market-chain migration is covered by Task 4.
- Logs/tests as the migration safety net are reflected in backend unit tests, App bridge tests, frontend node tests, and build verification in Tasks 1 through 5.
- Page-layer introduction for `/market` is covered by Task 4.
- Frontend service-layer introduction is covered by Task 3 and Task 4.

### Placeholder Scan

- No `TODO`, `TBD`, “implement later”, or “similar to Task N” placeholders remain.
- Every code-changing step includes concrete file paths, code, and commands.

### Type Consistency

- Backend service names are consistent across tasks:
  - `LoadFeeds`
  - `RefreshFeed`
  - `LoadGlobalIndexes`
  - `LoadIndustryRanks`
- App bridge names are consistent across tasks:
  - `GetMarketFeeds`
  - `RefreshMarketFeed`
  - `GetMarketGlobalIndexes`
  - `GetMarketIndustryRanks`
- Frontend service names are consistent across tasks:
  - `loadMarketFeeds`
  - `loadMarketGlobalIndexes`
  - `loadMarketIndustryRanks`
  - `refreshMarketFeed`
