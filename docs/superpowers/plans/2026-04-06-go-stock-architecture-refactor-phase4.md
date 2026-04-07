# go-stock Architecture Refactor Phase 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `cron task + settings old prompt modal + shared prompt/AI config access` behind the new service boundaries while keeping assistant and multi-turn agent flows out of this slice.

**Architecture:** Keep the existing Wails method names stable, but stop letting [app.go](D:/codex_work/go-stock/app.go) and [app_common.go](D:/codex_work/go-stock/app_common.go) talk straight to `backend/data` and `backend/agent` for this domain. Introduce `backend/service/config` for settings/prompt/shared AI config access and `backend/service/task` for cron task orchestration, then route [frontend/src/components/settings.vue](D:/codex_work/go-stock/frontend/src/components/settings.vue) and [frontend/src/components/cron-task-manager.vue](D:/codex_work/go-stock/frontend/src/components/cron-task-manager.vue) through `frontend/src/services/configService.mjs` and `frontend/src/services/taskService.mjs`.

**Tech Stack:** Go 1.26, Wails v2, Vue 3, Naive UI, Node `--test`, Vite, `robfig/cron`

---

## Scope Note

This plan intentionally covers only the first executable slice of stage 4:

- shared config access:
  - app config read/write
  - AI config list read
  - prompt template read/CRUD
  - settings page legacy prompt modal
- task execution access:
  - cron task CRUD
  - enable / disable
  - immediate execute
  - next-run calculation and validation

This plan intentionally does **not** include:

- [frontend/src/components/FloatingAiAssistant.vue](D:/codex_work/go-stock/frontend/src/components/FloatingAiAssistant.vue)
- [frontend/src/components/FloatingAgentAssistant.vue](D:/codex_work/go-stock/frontend/src/components/FloatingAgentAssistant.vue)
- [frontend/src/components/agent-chat.vue](D:/codex_work/go-stock/frontend/src/components/agent-chat.vue)
- deep refactors inside [backend/agent/agent_api.go](D:/codex_work/go-stock/backend/agent/agent_api.go)
- analysis page cleanup beyond keeping existing Wails methods stable

## File Map

- Create: `D:\codex_work\go-stock\backend\service\config\store.go`
  - Thin adapter over existing `backend/data` settings and prompt APIs.
- Create: `D:\codex_work\go-stock\backend\service\config\service.go`
  - Shared config service for settings, AI config list, prompt template list/page/CRUD, and legacy prompt compatibility.
- Create: `D:\codex_work\go-stock\backend\service\config\service_test.go`
  - Unit tests for empty defaults, prompt mapping, prompt page delegation, and settings delegation.
- Create: `D:\codex_work\go-stock\backend\service\task\store.go`
  - Thin adapter over `backend/agent.CronTaskApi`.
- Create: `D:\codex_work\go-stock\backend\service\task\service.go`
  - Task service that owns stable scheduler keys, cron registration, enable/disable, and direct execute.
- Create: `D:\codex_work\go-stock\backend\service\task\service_test.go`
  - Unit tests for stable scheduler keys, register/unregister behavior, and task execution delegation.
- Create: `D:\codex_work\go-stock\app_task_scheduler.go`
  - App-level scheduler adapter that translates task-service registration into `App.cron` operations.
- Create: `D:\codex_work\go-stock\app_config_test.go`
  - Bridge tests that prove config and prompt Wails methods delegate to `configService`.
- Create: `D:\codex_work\go-stock\app_task_test.go`
  - Bridge tests that prove task Wails methods delegate to `taskService`.
- Modify: `D:\codex_work\go-stock\app.go`
  - Wire `configService` and `taskService`, keep `UpdateConfig` bridge-specific cron refresh, and delegate prompt/task methods.
- Modify: `D:\codex_work\go-stock\app_common.go`
  - Delegate prompt page CRUD through `configService`.
- Create: `D:\codex_work\go-stock\frontend\src\services\configService.mjs`
  - Frontend wrapper for settings, AI config list, prompt template list, legacy prompt modal actions, and config export/model discovery.
- Create: `D:\codex_work\go-stock\frontend\src\services\configService.test.mjs`
  - Node tests for config normalizers and prompt defaults.
- Create: `D:\codex_work\go-stock\frontend\src\services\taskService.mjs`
  - Frontend wrapper for cron task page data and task operations.
- Create: `D:\codex_work\go-stock\frontend\src\services\taskService.test.mjs`
  - Node tests for task normalizers.
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`
  - Move settings/prompt/AI-config calls into `configService.mjs`.
- Modify: `D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue`
  - Move shared config reads into `configService.mjs` and task calls into `taskService.mjs`.

### Task 1: Add The Backend `config` Service Boundary

**Files:**
- Create: `D:\codex_work\go-stock\backend\service\config\store.go`
- Create: `D:\codex_work\go-stock\backend\service\config\service.go`
- Create: `D:\codex_work\go-stock\backend\service\config\service_test.go`

- [ ] **Step 1: Write the failing config service tests**

```go
package config

import (
	"context"
	"errors"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type fakeStore struct {
	config        *data.SettingConfig
	prompts       *[]models.PromptTemplate
	promptPage    *models.PromptTemplatePageData
	promptPageErr error
	savedTemplate models.PromptTemplate
	deletedID     uint
	updatedConfig *data.SettingConfig
	updateResult  string
}

func (f *fakeStore) GetConfig(ctx context.Context) *data.SettingConfig { return f.config }
func (f *fakeStore) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	f.updatedConfig = cfg
	return f.updateResult
}
func (f *fakeStore) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.prompts
}
func (f *fakeStore) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	if f.promptPageErr != nil {
		return nil, f.promptPageErr
	}
	return f.promptPage, nil
}
func (f *fakeStore) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTemplate = template
	return "保存成功"
}
func (f *fakeStore) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedID = id
	return "删除成功"
}

func TestService_GetAiConfigsAndPromptsReturnStableDefaults(t *testing.T) {
	svc := NewService(&fakeStore{config: &data.SettingConfig{}, prompts: nil})
	configs := svc.GetAiConfigs(context.Background())
	if configs == nil || len(configs) != 0 {
		t.Fatalf("expected empty ai config slice, got %#v", configs)
	}
	prompts := svc.GetPromptTemplates(context.Background(), "", "")
	if prompts == nil || len(*prompts) != 0 {
		t.Fatalf("expected empty prompt list, got %#v", prompts)
	}
}

func TestService_SaveLegacyPromptMapsToPromptTemplate(t *testing.T) {
	store := &fakeStore{}
	svc := NewService(store)
	msg := svc.SaveLegacyPrompt(context.Background(), models.Prompt{
		ID: 8, Name: "系统模板", Type: "模型系统Prompt", Content: "先看风险",
	})
	if msg != "保存成功" {
		t.Fatalf("unexpected save message: %q", msg)
	}
	if store.savedTemplate.ID != 8 || store.savedTemplate.Content != "先看风险" {
		t.Fatalf("unexpected saved template: %#v", store.savedTemplate)
	}
}

func TestService_GetPromptTemplatePagePropagatesError(t *testing.T) {
	expected := errors.New("boom")
	svc := NewService(&fakeStore{promptPageErr: expected})
	if _, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{}); !errors.Is(err, expected) {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestService_UpdateConfigDelegates(t *testing.T) {
	store := &fakeStore{updateResult: "保存成功"}
	svc := NewService(store)
	cfg := &data.SettingConfig{}
	msg := svc.UpdateConfig(context.Background(), cfg)
	if msg != "保存成功" || store.updatedConfig != cfg {
		t.Fatalf("unexpected update delegation: msg=%q cfg=%#v", msg, store.updatedConfig)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/config -run "TestService_GetAiConfigsAndPromptsReturnStableDefaults|TestService_SaveLegacyPromptMapsToPromptTemplate|TestService_GetPromptTemplatePagePropagatesError|TestService_UpdateConfigDelegates" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing files under `backend/service/config`

- [ ] **Step 3: Write the config service and thin legacy store**

```go
// D:\codex_work\go-stock\backend\service\config\service.go
package config

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Store interface {
	GetConfig(ctx context.Context) *data.SettingConfig
	UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
}

type Service struct{ store Store }

func NewService(store Store) *Service {
	if store == nil {
		panic("config: store dependency is required")
	}
	return &Service{store: store}
}

func (s *Service) GetConfig(ctx context.Context) *data.SettingConfig {
	cfg := s.store.GetConfig(ctx)
	if cfg == nil {
		return &data.SettingConfig{AiConfigs: []*data.AIConfig{}}
	}
	if cfg.AiConfigs == nil {
		cfg.AiConfigs = []*data.AIConfig{}
	}
	return cfg
}

func (s *Service) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	return s.store.UpdateConfig(ctx, cfg)
}

func (s *Service) GetAiConfigs(ctx context.Context) []*data.AIConfig {
	return s.GetConfig(ctx).AiConfigs
}

func (s *Service) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	result := s.store.GetPromptTemplates(ctx, name, promptType)
	if result == nil {
		return &empty
	}
	return result
}

func (s *Service) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	page, err := s.store.GetPromptTemplatePage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.PromptTemplatePageData{}, nil
	}
	return page, nil
}

func (s *Service) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return s.store.SavePromptTemplate(ctx, template)
}

func (s *Service) DeletePromptTemplate(ctx context.Context, id uint) string {
	return s.store.DeletePromptTemplate(ctx, id)
}

func (s *Service) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	return s.SavePromptTemplate(ctx, models.PromptTemplate{
		ID: prompt.ID, Name: prompt.Name, Type: prompt.Type, Content: prompt.Content,
	})
}

func (s *Service) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	return s.DeletePromptTemplate(ctx, id)
}
```

```go
// D:\codex_work\go-stock\backend\service\config\store.go
package config

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type LegacyStore struct{}

func NewStore() *LegacyStore { return &LegacyStore{} }
func (s *LegacyStore) GetConfig(ctx context.Context) *data.SettingConfig { return data.GetSettingConfig() }
func (s *LegacyStore) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	return data.UpdateConfig(cfg)
}
func (s *LegacyStore) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
}
func (s *LegacyStore) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return data.NewPromptTemplateApi().GetPromptTemplateList(&query)
}
func (s *LegacyStore) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}
func (s *LegacyStore) DeletePromptTemplate(ctx context.Context, id uint) string {
	return data.NewPromptTemplateApi().DelPrompt(id)
}
```

- [ ] **Step 4: Run the config service tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/config -run "TestService_GetAiConfigsAndPromptsReturnStableDefaults|TestService_SaveLegacyPromptMapsToPromptTemplate|TestService_GetPromptTemplatePagePropagatesError|TestService_UpdateConfigDelegates" -count=1
```

Expected:

- command exits `0`
- all four tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/config/service.go backend/service/config/store.go backend/service/config/service_test.go
git commit -m "refactor: add shared config service boundary"
```

### Task 2: Bridge Settings And Prompt Methods Through `configService`

**Files:**
- Create: `D:\codex_work\go-stock\app_config_test.go`
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\app_common.go`

- [ ] **Step 1: Write the failing App bridge test**

```go
package main

import (
	"context"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type fakeConfigService struct {
	config       *data.SettingConfig
	aiConfigs    []*data.AIConfig
	prompts      *[]models.PromptTemplate
	promptPage   *models.PromptTemplatePageData
	updateResult string
	saveResult   string
	deleteResult string
	savedLegacy  models.Prompt
	savedTmpl    models.PromptTemplate
	deletedID    uint
}

func (f *fakeConfigService) GetConfig(ctx context.Context) *data.SettingConfig { return f.config }
func (f *fakeConfigService) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	f.config = cfg
	return f.updateResult
}
func (f *fakeConfigService) GetAiConfigs(ctx context.Context) []*data.AIConfig { return f.aiConfigs }
func (f *fakeConfigService) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.prompts
}
func (f *fakeConfigService) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return f.promptPage, nil
}
func (f *fakeConfigService) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTmpl = template
	return f.saveResult
}
func (f *fakeConfigService) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedID = id
	return f.deleteResult
}
func (f *fakeConfigService) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	f.savedLegacy = prompt
	return f.saveResult
}
func (f *fakeConfigService) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	f.deletedID = id
	return f.deleteResult
}

func TestApp_ConfigAndPromptMethodsDelegateToConfigService(t *testing.T) {
	app := NewApp()
	prompts := []models.PromptTemplate{{ID: 7, Name: "系统模板", Type: "模型系统Prompt", Content: "先看风险"}}
	cfg := &data.SettingConfig{AiConfigs: []*data.AIConfig{{ID: 3, Name: "deepseek", ModelName: "deepseek-chat"}}}
	fake := &fakeConfigService{
		config:       cfg,
		aiConfigs:    cfg.AiConfigs,
		prompts:      &prompts,
		promptPage:   &models.PromptTemplatePageData{List: prompts, Total: 1, Page: 1, PageSize: 10, TotalPages: 1},
		updateResult: "保存成功",
		saveResult:   "保存成功",
		deleteResult: "删除成功",
	}
	app.configService = fake

	if got := app.GetConfig(); got != cfg {
		t.Fatalf("expected config delegation, got %#v", got)
	}
	if got := app.GetAiConfigs(); len(got) != 1 || got[0].ID != 3 {
		t.Fatalf("unexpected ai configs: %#v", got)
	}
	if got := app.GetPromptTemplates("", ""); len(*got) != 1 || (*got)[0].ID != 7 {
		t.Fatalf("unexpected prompt list: %#v", got)
	}
	if msg := app.AddPrompt(models.Prompt{ID: 8, Name: "旧模板", Type: "模型用户Prompt", Content: "总结"}); msg != "保存成功" {
		t.Fatalf("unexpected legacy save message: %q", msg)
	}
	if fake.savedLegacy.ID != 8 || fake.savedLegacy.Name != "旧模板" {
		t.Fatalf("unexpected legacy prompt payload: %#v", fake.savedLegacy)
	}
	if page := app.GetPromptTemplateList(models.PromptTemplateQuery{Page: 1, PageSize: 10}); page.Total != 1 {
		t.Fatalf("unexpected prompt page: %#v", page)
	}
	if msg := app.AddPromptTemplate(models.PromptTemplate{ID: 9, Name: "新模板", Type: "模型系统Prompt", Content: "约束输出"}); msg != "保存成功" {
		t.Fatalf("unexpected add template message: %q", msg)
	}
	if fake.savedTmpl.ID != 9 || fake.savedTmpl.Name != "新模板" {
		t.Fatalf("unexpected prompt template payload: %#v", fake.savedTmpl)
	}
	if msg := app.DeletePromptTemplate(9); msg != "删除成功" || fake.deletedID != 9 {
		t.Fatalf("unexpected delete result: msg=%q id=%d", msg, fake.deletedID)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . -run "TestApp_ConfigAndPromptMethodsDelegateToConfigService" -count=1
```

Expected:

- command exits non-zero
- failure mentions `configService` does not exist on `App`

- [ ] **Step 3: Wire `configService` into `App` and delegate prompt methods**

```go
// Add to D:\codex_work\go-stock\app.go imports
import (
	configservice "go-stock/backend/service/config"
)

type configService interface {
	GetConfig(ctx context.Context) *data.SettingConfig
	UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string
	GetAiConfigs(ctx context.Context) []*data.AIConfig
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
	SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string
	DeleteLegacyPrompt(ctx context.Context, id uint) string
}

type App struct {
	// existing fields...
	configService configService
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

	app := &App{
		cache:              cache,
		cron:               c,
		cronEntrys:         make(map[string]cron.EntryID),
		AiTools:            tools,
		stockAlertLastSent: make(map[string]time.Time),
		priceAtAlertReset:  make(map[string]float64),
		marketReadService:  marketservice.NewService(marketsource.NewSource()),
		analysisService:    analysisservice.NewService(analysisProvider, analysisStore, analysisStore),
		configService:      configservice.NewService(configservice.NewStore()),
	}
	return app
}
```

```go
// Replace the settings and prompt methods in D:\codex_work\go-stock\app.go
func (a *App) UpdateConfig(settingConfig *data.SettingConfig) string {
	if settingConfig.RefreshInterval > 0 {
		if entryID, exists := a.getCronEntry("MonitorStockPrices"); exists {
			a.cron.Remove(entryID)
		}
		id, _ := a.cron.AddFunc(fmt.Sprintf("@every %ds", settingConfig.RefreshInterval), func() {
			MonitorStockPrices(a)
		})
		a.setCronEntry("MonitorStockPrices", id)
	}
	return a.configService.UpdateConfig(a.ctx, settingConfig)
}

func (a *App) GetConfig() *data.SettingConfig { return a.configService.GetConfig(a.ctx) }
func (a *App) GetAiConfigs() []*data.AIConfig { return a.configService.GetAiConfigs(a.ctx) }
func (a *App) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	return a.configService.GetPromptTemplates(a.ctx, name, promptType)
}
func (a *App) AddPrompt(prompt models.Prompt) string { return a.configService.SaveLegacyPrompt(a.ctx, prompt) }
func (a *App) DelPrompt(id uint) string              { return a.configService.DeleteLegacyPrompt(a.ctx, id) }
```

```go
// Replace the prompt page methods in D:\codex_work\go-stock\app_common.go
func (a *App) GetPromptTemplateList(query models.PromptTemplateQuery) *models.PromptTemplatePageData {
	page, err := a.configService.GetPromptTemplatePage(a.ctx, query)
	if err != nil {
		return &models.PromptTemplatePageData{}
	}
	return page
}

func (a *App) AddPromptTemplate(template models.PromptTemplate) string {
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) UpdatePromptTemplate(template models.PromptTemplate) string {
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) DeletePromptTemplate(id uint) string {
	return a.configService.DeletePromptTemplate(a.ctx, id)
}
```

- [ ] **Step 4: Run the bridge test**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/config -run "TestApp_ConfigAndPromptMethodsDelegateToConfigService|TestService_GetAiConfigsAndPromptsReturnStableDefaults|TestService_SaveLegacyPromptMapsToPromptTemplate" -count=1
```

Expected:

- command exits `0`
- App bridge test and config service tests pass

- [ ] **Step 5: Commit**

```powershell
git add app.go app_common.go app_config_test.go
git commit -m "refactor: bridge shared config access through app"
```

### Task 3: Add Frontend `configService` And Migrate Settings / Shared Config Reads

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\services\configService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\configService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\components\settings.vue`
- Modify: `D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue`

- [ ] **Step 1: Write the failing frontend config service tests**

```javascript
import test from 'node:test'
import assert from 'node:assert/strict'

import {
  normalizeAiConfigs,
  normalizePromptTemplates,
  normalizeSettingConfig,
} from './configService.mjs'

test('normalizeAiConfigs falls back to an empty array', () => {
  assert.deepEqual(normalizeAiConfigs(null), [])
})

test('normalizePromptTemplates keeps prompt rows stable', () => {
  assert.deepEqual(
    normalizePromptTemplates([{ ID: 7, name: '系统模板', type: '模型系统Prompt', content: '先看风险' }]),
    [{ ID: 7, name: '系统模板', type: '模型系统Prompt', content: '先看风险' }],
  )
})

test('normalizeSettingConfig always returns aiConfigs array', () => {
  assert.deepEqual(normalizeSettingConfig({ darkTheme: true }), {
    darkTheme: true,
    aiConfigs: [],
  })
})
```

- [ ] **Step 2: Run the frontend config service tests to verify they fail**

Run:

```powershell
node --test frontend/src/services/configService.test.mjs
```

Expected:

- command exits non-zero
- failure mentions `frontend/src/services/configService.mjs` does not exist

- [ ] **Step 3: Add the frontend config service module and migrate the two components**

```javascript
// D:\codex_work\go-stock\frontend\src\services\configService.mjs
import * as AppBindings from '../../wailsjs/go/main/App.js'

function toArray(value) {
  return Array.isArray(value) ? value : []
}

export function normalizeAiConfigs(value) {
  return toArray(value).map(item => ({
    ...item,
    ID: Number.isFinite(Number(item?.ID)) ? Number(item.ID) : 0,
    name: item?.name ?? '',
    modelName: item?.modelName ?? '',
  }))
}

export function normalizePromptTemplates(value) {
  return toArray(value).map(item => ({
    ID: Number.isFinite(Number(item?.ID)) ? Number(item.ID) : 0,
    name: item?.name ?? '',
    type: item?.type ?? '',
    content: item?.content ?? '',
  }))
}

export function normalizeSettingConfig(value) {
  const data = value ?? {}
  return {
    ...data,
    aiConfigs: normalizeAiConfigs(data.aiConfigs),
  }
}

export async function loadAppConfig() {
  return normalizeSettingConfig(await AppBindings.GetConfig())
}

export async function saveAppConfig(config) {
  return AppBindings.UpdateConfig(config)
}

export async function exportAppConfig() {
  return AppBindings.ExportConfig()
}

export async function fetchAiModels(baseUrl, apiKey) {
  return toArray(await AppBindings.FetchAiModels(baseUrl, apiKey)).map(String)
}

export async function loadAiConfigs() {
  return normalizeAiConfigs(await AppBindings.GetAiConfigs())
}

export async function loadPromptTemplates(name = '', type = '') {
  return normalizePromptTemplates(await AppBindings.GetPromptTemplates(name, type))
}

export async function saveLegacyPrompt(prompt) {
  return AppBindings.AddPrompt(prompt)
}

export async function deleteLegacyPrompt(id) {
  return AppBindings.DelPrompt(id)
}
```

```vue
<!-- Replace the App import block in D:\codex_work\go-stock\frontend\src\components\settings.vue -->
<script setup>
import { h, onBeforeUnmount, onMounted, ref } from "vue";
import { SendDingDingMessageByType } from "../../wailsjs/go/main/App";
import { data, models } from "../../wailsjs/go/models";
import { EventsEmit } from "../../wailsjs/runtime";
import { HelpCircleFilledIcon, HelpIcon } from "tdesign-icons-vue-next";
import { NTag, NTooltip, NIcon, useMessage } from "naive-ui";
import {
  deleteLegacyPrompt,
  exportAppConfig,
  fetchAiModels as fetchAvailableModels,
  loadAppConfig,
  loadPromptTemplates,
  saveAppConfig,
  saveLegacyPrompt,
} from "../services/configService.mjs";
</script>
```

```vue
<!-- Replace the settings-side async calls in D:\codex_work\go-stock\frontend\src\components\settings.vue -->
onMounted(() => {
  loadAppConfig().then(res => {
    formValue.value.ID = res.ID
    formValue.value.tushareToken = res.tushareToken
    formValue.value.dingPush = { enable: res.dingPushEnable, dingRobot: res.dingRobot }
    formValue.value.localPush = { enable: res.localPushEnable }
    formValue.value.updateBasicInfoOnStart = res.updateBasicInfoOnStart
    formValue.value.refreshInterval = res.refreshInterval
    formValue.value.openAI = {
      enable: res.openAiEnable,
      aiConfigs: res.aiConfigs || [],
      prompt: res.prompt,
      questionTemplate: res.questionTemplate ? res.questionTemplate : '{{stockName}}分析和总结',
      crawlTimeOut: res.crawlTimeOut,
      kDays: res.kDays,
      httpProxy:"",
      httpProxyEnabled:false,
    }
    formValue.value.enableDanmu = res.enableDanmu
    formValue.value.browserPath = res.browserPath
    formValue.value.enableNews = res.enableNews
    formValue.value.darkTheme = res.darkTheme
    formValue.value.enableFund = res.enableFund
    formValue.value.enablePushNews = res.enablePushNews
    formValue.value.enableOnlyPushRedNews = res.enableOnlyPushRedNews
    formValue.value.sponsorCode = res.sponsorCode
    formValue.value.httpProxy = res.httpProxy
    formValue.value.httpProxyEnabled = res.httpProxyEnabled
    formValue.value.enableAgent = res.enableAgent
    formValue.value.qgqpBId = res.qgqpBId
  })
})

async function fetchAiModels(aiConfig) {
  if (!aiConfig.baseUrl || !aiConfig.apiKey) {
    message.warning('请先填写接口地址和 apiKey')
    return
  }
  if (aiConfig._loadingModels) {
    return
  }
  aiConfig._loadingModels = true
  try {
    const list = await fetchAvailableModels(aiConfig.baseUrl, aiConfig.apiKey)
    const options = (list || []).map(id => ({ label: id, value: id }))
    aiConfig._modelOptions = options
    if (!aiConfig.modelName && options.length > 0) {
      aiConfig.modelName = options[0].value
    }
    if (!options.length) {
      message.warning('未从接口获取到可用模型，请检查地址和 apiKey')
    }
  } catch (e) {
    console.error('FetchAiModels error', e)
    message.error('获取模型列表失败，请检查接口地址和 apiKey')
  } finally {
    aiConfig._loadingModels = false
  }
}

function saveConfig() {
  let config = new data.SettingConfig({
    ID: formValue.value.ID,
    dingPushEnable: formValue.value.dingPush.enable,
    dingRobot: formValue.value.dingPush.dingRobot,
    localPushEnable: formValue.value.localPush.enable,
    updateBasicInfoOnStart: formValue.value.updateBasicInfoOnStart,
    refreshInterval: formValue.value.refreshInterval,
    openAiEnable: formValue.value.openAI.enable,
    aiConfigs: formValue.value.openAI.aiConfigs,
    tushareToken: formValue.value.tushareToken,
    prompt: formValue.value.openAI.prompt,
    questionTemplate: formValue.value.openAI.questionTemplate,
    crawlTimeOut: formValue.value.openAI.crawlTimeOut,
    kDays: formValue.value.openAI.kDays,
    enableDanmu: formValue.value.enableDanmu,
    browserPath: formValue.value.browserPath,
    enableNews: formValue.value.enableNews,
    darkTheme: formValue.value.darkTheme,
    enableFund: formValue.value.enableFund,
    enablePushNews: formValue.value.enablePushNews,
    enableOnlyPushRedNews: formValue.value.enableOnlyPushRedNews,
    sponsorCode: formValue.value.sponsorCode,
    httpProxy: formValue.value.httpProxy,
    httpProxyEnabled: formValue.value.httpProxyEnabled,
    enableAgent: formValue.value.enableAgent,
    qgqpBId: formValue.value.qgqpBId
  })
  saveAppConfig(config).then(res => {
    message.success(res)
    EventsEmit("updateSettings", config)
  })
}

function exportConfig() {
  exportAppConfig().then(res => {
    message.info(res)
  })
}

function savePrompt() {
  saveLegacyPrompt(formPrompt.value).then(res => {
    message.success(res)
    loadPromptTemplates("", "").then(res => {
      promptTemplates.value = res
    })
    showManagePromptsModal.value = false
  })
}

function deletePrompt(ID) {
  deleteLegacyPrompt(ID).then(res => {
    message.success(res)
    loadPromptTemplates("", "").then(res => {
      promptTemplates.value = res
    })
  })
}
```

```vue
<!-- Replace the shared config reads in D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue -->
<script setup>
import {
  loadAiConfigs,
  loadPromptTemplates,
} from '../services/configService.mjs'
</script>
```

```vue
<!-- Replace the shared config reads in D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue -->
const loadAiConfigsForTask = async () => {
  try {
    const configs = await loadAiConfigs()
    aiConfigOptions.value = configs.map(c => ({
      label: c.name + "[" + c.modelName + "]",
      value: c.ID
    }))
  } catch (error) {
    console.error('加载 AI 配置失败:', error)
  }
}

const loadPromptOptions = async () => {
  try {
    const userTemplates = await loadPromptTemplates('', '模型用户Prompt')
    promptTemplateOptions.value = userTemplates.map(t => ({ label: t.name, value: t.ID }))
    const sysTemplates = await loadPromptTemplates('', '模型系统Prompt')
    sysPromptOptions.value = sysTemplates.map(t => ({ label: t.name, value: t.ID }))
  } catch (error) {
    console.error('加载提示词模板失败:', error)
  }
}

onMounted(async () => {
  await Promise.all([
    loadTaskTypes(),
    loadAiConfigsForTask(),
    loadPromptOptions(),
    loadTaskList(),
  ])
})
```

- [ ] **Step 4: Run the frontend config tests**

Run:

```powershell
node --test frontend/src/services/configService.test.mjs
```

Expected:

- command exits `0`
- config service tests pass

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/services/configService.mjs frontend/src/services/configService.test.mjs frontend/src/components/settings.vue frontend/src/components/cron-task-manager.vue
git commit -m "refactor: route shared config access through frontend service"
```

### Task 4: Add The Backend `task` Service Boundary

**Files:**
- Create: `D:\codex_work\go-stock\backend\service\task\store.go`
- Create: `D:\codex_work\go-stock\backend\service\task\service.go`
- Create: `D:\codex_work\go-stock\backend\service\task\service_test.go`

- [ ] **Step 1: Write the failing task service tests**

```go
package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type fakeStore struct {
	task        *models.CronTask
	page        *models.CronTaskPageResp
	createErr   error
	updateErr   error
	deleteErr   error
	enableErr   error
	getErr      error
	runErr      error
	validateErr error
	createTask  *models.CronTask
	updateTask  *models.CronTask
	deleteID    uint
	enableID    uint
	enableValue bool
	runID       uint
	search      []models.CronTask
}

func (f *fakeStore) Create(task *models.CronTask) error {
	if f.createErr != nil { return f.createErr }
	f.createTask = task
	if task.ID == 0 { task.ID = 7 }
	return nil
}
func (f *fakeStore) Update(task *models.CronTask) error {
	if f.updateErr != nil { return f.updateErr }
	f.updateTask = task
	return nil
}
func (f *fakeStore) Delete(id uint) error { f.deleteID = id; return f.deleteErr }
func (f *fakeStore) GetByID(id uint) (*models.CronTask, error) {
	if f.getErr != nil { return nil, f.getErr }
	if f.task != nil { return f.task, nil }
	return &models.CronTask{ID: id, Name: "市场分析", CronExpr: "0 */5 * * * *", Enable: true}, nil
}
func (f *fakeStore) List(query *models.CronTaskQuery) *models.CronTaskPageResp { return f.page }
func (f *fakeStore) EnableTask(id uint, enable bool) error {
	f.enableID = id
	f.enableValue = enable
	return f.enableErr
}
func (f *fakeStore) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	f.runID = task.ID
	return f.runErr
}
func (f *fakeStore) GetTaskTypes() []lo.Tuple2[string, string] {
	return []lo.Tuple2[string, string]{{A: "market_analysis", B: "市场分析"}}
}
func (f *fakeStore) ValidateCronExpr(expr string) error { return f.validateErr }
func (f *fakeStore) CalculateNextRunTime(expr string) time.Time {
	return time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local)
}
func (f *fakeStore) CalculateNextRunTimes(expr string, count int) []time.Time {
	return []time.Time{
		time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 6, 10, 5, 0, 0, time.Local),
	}
}
func (f *fakeStore) Search(keyword string) []models.CronTask { return f.search }

type fakeScheduler struct {
	registeredKey  string
	registeredSpec string
	unregistered   []string
	registerErr    error
}

func (f *fakeScheduler) Register(key, spec string, job func()) error {
	f.registeredKey = key
	f.registeredSpec = spec
	return f.registerErr
}
func (f *fakeScheduler) Unregister(key string) { f.unregistered = append(f.unregistered, key) }

func TestService_CreateUsesStableSchedulerKey(t *testing.T) {
	store := &fakeStore{}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)
	msg := svc.Create(context.Background(), &models.CronTask{Name: "市场分析", CronExpr: "0 */5 * * * *", Enable: true})
	if msg != "创建成功" || scheduler.registeredKey != "cron-task-7" || scheduler.registeredSpec != "0 */5 * * * *" {
		t.Fatalf("unexpected create scheduling: msg=%q key=%q spec=%q", msg, scheduler.registeredKey, scheduler.registeredSpec)
	}
}

func TestService_UpdateReplacesExistingRegistration(t *testing.T) {
	store := &fakeStore{}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)
	msg := svc.Update(context.Background(), &models.CronTask{ID: 9, Name: "市场分析", CronExpr: "0 */10 * * * *", Enable: true})
	if msg != "更新成功" || len(scheduler.unregistered) != 1 || scheduler.unregistered[0] != "cron-task-9" || scheduler.registeredKey != "cron-task-9" {
		t.Fatalf("unexpected update scheduling: msg=%q unregister=%#v key=%q", msg, scheduler.unregistered, scheduler.registeredKey)
	}
}

func TestService_EnableAndDisableUseStableKey(t *testing.T) {
	store := &fakeStore{task: &models.CronTask{ID: 11, Name: "任务", CronExpr: "0 */15 * * * *", Enable: true}}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)
	if msg := svc.Enable(context.Background(), 11, false); msg != "操作成功" {
		t.Fatalf("unexpected disable message: %q", msg)
	}
	if len(scheduler.unregistered) == 0 || scheduler.unregistered[0] != "cron-task-11" {
		t.Fatalf("unexpected unregister list after disable: %#v", scheduler.unregistered)
	}
	scheduler.unregistered = nil
	if msg := svc.Enable(context.Background(), 11, true); msg != "操作成功" || scheduler.registeredKey != "cron-task-11" {
		t.Fatalf("unexpected enable scheduling: msg=%q key=%q", msg, scheduler.registeredKey)
	}
}

func TestService_RunNowDelegatesToStore(t *testing.T) {
	store := &fakeStore{task: &models.CronTask{ID: 12, Name: "执行任务"}}
	svc := NewService(store, &fakeScheduler{}, context.Background)
	if err := svc.RunNow(context.Background(), 12); err != nil || store.runID != 12 {
		t.Fatalf("unexpected run delegation: err=%v id=%d", err, store.runID)
	}
}

func TestService_ValidateCronExprPropagatesFailure(t *testing.T) {
	store := &fakeStore{validateErr: errors.New("bad cron")}
	svc := NewService(store, &fakeScheduler{}, context.Background)
	if msg := svc.ValidateCronExpr(context.Background(), "*"); msg == "有效表达式" {
		t.Fatalf("expected validation failure, got %q", msg)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/task -run "TestService_CreateUsesStableSchedulerKey|TestService_UpdateReplacesExistingRegistration|TestService_EnableAndDisableUseStableKey|TestService_RunNowDelegatesToStore|TestService_ValidateCronExprPropagatesFailure" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing files under `backend/service/task`

- [ ] **Step 3: Write the task service and thin agent-backed store**

```go
// D:\codex_work\go-stock\backend\service\task\service.go
package task

import (
	"context"
	"fmt"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type Store interface {
	Create(task *models.CronTask) error
	Update(task *models.CronTask) error
	Delete(id uint) error
	GetByID(id uint) (*models.CronTask, error)
	List(query *models.CronTaskQuery) *models.CronTaskPageResp
	EnableTask(id uint, enable bool) error
	ExecuteTask(ctx context.Context, task *models.CronTask) error
	GetTaskTypes() []lo.Tuple2[string, string]
	ValidateCronExpr(expr string) error
	CalculateNextRunTime(expr string) time.Time
	CalculateNextRunTimes(expr string, count int) []time.Time
	Search(keyword string) []models.CronTask
}

type Scheduler interface {
	Register(key, spec string, job func()) error
	Unregister(key string)
}

type Service struct {
	store       Store
	scheduler   Scheduler
	contextFunc func() context.Context
}

func NewService(store Store, scheduler Scheduler, contextFunc func() context.Context) *Service {
	if store == nil {
		panic("task: store dependency is required")
	}
	if scheduler == nil {
		panic("task: scheduler dependency is required")
	}
	if contextFunc == nil {
		panic("task: context func is required")
	}
	return &Service{store: store, scheduler: scheduler, contextFunc: contextFunc}
}

func scheduleKey(id uint) string {
	return fmt.Sprintf("cron-task-%d", id)
}

func (s *Service) syncSchedule(task *models.CronTask) error {
	if task == nil {
		return nil
	}
	key := scheduleKey(task.ID)
	s.scheduler.Unregister(key)
	if !task.Enable {
		return nil
	}
	return s.scheduler.Register(key, task.CronExpr, func() {
		latest, err := s.store.GetByID(task.ID)
		if err != nil || latest == nil {
			return
		}
		_ = s.store.ExecuteTask(s.contextFunc(), latest)
	})
}

func (s *Service) Create(ctx context.Context, task *models.CronTask) string {
	if err := s.store.Create(task); err != nil {
		return fmt.Sprintf("创建失败：%v", err)
	}
	if err := s.syncSchedule(task); err != nil {
		return "任务创建成功,但定时失败"
	}
	return "创建成功"
}

func (s *Service) Update(ctx context.Context, task *models.CronTask) string {
	if err := s.store.Update(task); err != nil {
		return fmt.Sprintf("更新失败：%v", err)
	}
	if err := s.syncSchedule(task); err != nil {
		return "更新成功,但定时失败"
	}
	return "更新成功"
}

func (s *Service) Delete(ctx context.Context, id uint) string {
	s.scheduler.Unregister(scheduleKey(id))
	if err := s.store.Delete(id); err != nil {
		return fmt.Sprintf("删除失败：%v", err)
	}
	return "删除成功"
}

func (s *Service) GetByID(ctx context.Context, id uint) (*models.CronTask, error) {
	return s.store.GetByID(id)
}

func (s *Service) List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp {
	return s.store.List(query)
}

func (s *Service) Enable(ctx context.Context, id uint, enable bool) string {
	if err := s.store.EnableTask(id, enable); err != nil {
		return fmt.Sprintf("操作失败：%v", err)
	}
	key := scheduleKey(id)
	s.scheduler.Unregister(key)
	if !enable {
		return "操作成功"
	}
	task, err := s.store.GetByID(id)
	if err != nil {
		return fmt.Sprintf("操作失败：%v", err)
	}
	task.Enable = enable
	if err := s.syncSchedule(task); err != nil {
		return "操作成功,但定时失败"
	}
	return "操作成功"
}

func (s *Service) RunNow(ctx context.Context, id uint) error {
	task, err := s.store.GetByID(id)
	if err != nil {
		return err
	}
	return s.store.ExecuteTask(ctx, task)
}

func (s *Service) GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string] {
	return s.store.GetTaskTypes()
}

func (s *Service) ValidateCronExpr(ctx context.Context, expr string) string {
	if err := s.store.ValidateCronExpr(expr); err != nil {
		return fmt.Sprintf("无效表达式：%v", err)
	}
	return "有效表达式"
}

func (s *Service) Search(ctx context.Context, keyword string) []models.CronTask {
	return s.store.Search(keyword)
}

func (s *Service) CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time {
	return s.store.CalculateNextRunTime(cronExpr)
}

func (s *Service) CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time {
	return s.store.CalculateNextRunTimes(cronExpr, count)
}
```

```go
// D:\codex_work\go-stock\backend\service\task\store.go
package task

import (
	"context"
	"time"

	"go-stock/backend/agent"
	"go-stock/backend/models"

	"github.com/samber/lo"
)

type AgentStore struct{}

func NewStore() *AgentStore { return &AgentStore{} }
func (s *AgentStore) Create(task *models.CronTask) error { return agent.NewCronTaskApi().Create(task) }
func (s *AgentStore) Update(task *models.CronTask) error { return agent.NewCronTaskApi().Update(task) }
func (s *AgentStore) Delete(id uint) error               { return agent.NewCronTaskApi().Delete(id) }
func (s *AgentStore) GetByID(id uint) (*models.CronTask, error) {
	return agent.NewCronTaskApi().GetByID(id)
}
func (s *AgentStore) List(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return agent.NewCronTaskApi().List(query)
}
func (s *AgentStore) EnableTask(id uint, enable bool) error {
	return agent.NewCronTaskApi().EnableTask(id, enable)
}
func (s *AgentStore) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	return agent.NewCronTaskApi().ExecuteTask(ctx, task)
}
func (s *AgentStore) GetTaskTypes() []lo.Tuple2[string, string] {
	return agent.NewCronTaskApi().GetTaskTypes()
}
func (s *AgentStore) ValidateCronExpr(expr string) error { return agent.NewCronTaskApi().ValidateCronExpr(expr) }
func (s *AgentStore) CalculateNextRunTime(expr string) time.Time {
	return agent.NewCronTaskApi().CalculateNextRunTime(expr)
}
func (s *AgentStore) CalculateNextRunTimes(expr string, count int) []time.Time {
	return agent.NewCronTaskApi().CalculateNextRunTimes(expr, count)
}
func (s *AgentStore) Search(keyword string) []models.CronTask { return agent.NewCronTaskApi().SearchTasks(keyword) }
```

- [ ] **Step 4: Run the task service tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/task -run "TestService_CreateUsesStableSchedulerKey|TestService_UpdateReplacesExistingRegistration|TestService_EnableAndDisableUseStableKey|TestService_RunNowDelegatesToStore|TestService_ValidateCronExprPropagatesFailure" -count=1
```

Expected:

- command exits `0`
- task service tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/task/service.go backend/service/task/store.go backend/service/task/service_test.go
git commit -m "refactor: add cron task service boundary"
```

### Task 5: Bridge Task Methods Through `App`

**Files:**
- Create: `D:\codex_work\go-stock\app_task_scheduler.go`
- Create: `D:\codex_work\go-stock\app_task_test.go`
- Modify: `D:\codex_work\go-stock\app.go`

- [ ] **Step 1: Write the failing task bridge test**

```go
package main

import (
	"context"
	"testing"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type fakeTaskService struct {
	createResult string
	updateResult string
	deleteResult string
	enableResult string
	runErr       error
	task         *models.CronTask
	page         *models.CronTaskPageResp
	types        []lo.Tuple2[string, string]
	search       []models.CronTask
	createTask   *models.CronTask
	updateTask   *models.CronTask
	deleteID     uint
	enableID     uint
	enableValue  bool
}

func (f *fakeTaskService) Create(ctx context.Context, task *models.CronTask) string { f.createTask = task; return f.createResult }
func (f *fakeTaskService) Update(ctx context.Context, task *models.CronTask) string { f.updateTask = task; return f.updateResult }
func (f *fakeTaskService) Delete(ctx context.Context, id uint) string                { f.deleteID = id; return f.deleteResult }
func (f *fakeTaskService) GetByID(ctx context.Context, id uint) (*models.CronTask, error) {
	return f.task, nil
}
func (f *fakeTaskService) List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp {
	return f.page
}
func (f *fakeTaskService) Enable(ctx context.Context, id uint, enable bool) string {
	f.enableID = id
	f.enableValue = enable
	return f.enableResult
}
func (f *fakeTaskService) RunNow(ctx context.Context, id uint) error { return f.runErr }
func (f *fakeTaskService) GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string] { return f.types }
func (f *fakeTaskService) ValidateCronExpr(ctx context.Context, expr string) string      { return "有效表达式" }
func (f *fakeTaskService) Search(ctx context.Context, keyword string) []models.CronTask   { return f.search }
func (f *fakeTaskService) CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time {
	return time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local)
}
func (f *fakeTaskService) CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time {
	return []time.Time{
		time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 6, 10, 5, 0, 0, time.Local),
	}
}

func TestApp_TaskMethodsDelegateToTaskService(t *testing.T) {
	app := NewApp()
	fake := &fakeTaskService{
		createResult: "创建成功",
		updateResult: "更新成功",
		deleteResult: "删除成功",
		enableResult: "操作成功",
		task:         &models.CronTask{ID: 8, Name: "市场分析"},
		page:         &models.CronTaskPageResp{Total: 1, Data: []models.CronTask{{ID: 8, Name: "市场分析"}}},
		types:        []lo.Tuple2[string, string]{{A: "market_analysis", B: "市场分析"}},
		search:       []models.CronTask{{ID: 8, Name: "市场分析"}},
	}
	app.taskService = fake

	if msg := app.CreateCronTask(&models.CronTask{Name: "创建任务"}); msg != "创建成功" {
		t.Fatalf("unexpected create message: %q", msg)
	}
	if fake.createTask == nil || fake.createTask.Name != "创建任务" {
		t.Fatalf("unexpected create payload: %#v", fake.createTask)
	}
	if msg := app.UpdateCronTask(&models.CronTask{ID: 8, Name: "更新任务"}); msg != "更新成功" {
		t.Fatalf("unexpected update message: %q", msg)
	}
	if msg := app.DeleteCronTask(8); msg != "删除成功" || fake.deleteID != 8 {
		t.Fatalf("unexpected delete result: msg=%q id=%d", msg, fake.deleteID)
	}
	if task := app.GetCronTaskByID(8); task == nil || task.ID != 8 {
		t.Fatalf("unexpected get task result: %#v", task)
	}
	if page := app.GetCronTaskList(&models.CronTaskQuery{Page: 1, PageSize: 10}); page.Total != 1 {
		t.Fatalf("unexpected task page: %#v", page)
	}
	if msg := app.EnableCronTask(8, true); msg != "操作成功" || fake.enableID != 8 || !fake.enableValue {
		t.Fatalf("unexpected enable result: msg=%q id=%d enable=%v", msg, fake.enableID, fake.enableValue)
	}
	if got := app.GetCronTaskTypes(); len(got) != 1 || got[0].A != "market_analysis" {
		t.Fatalf("unexpected task types: %#v", got)
	}
	if got := app.SearchCronTasks("市场"); len(got) != 1 {
		t.Fatalf("unexpected search result: %#v", got)
	}
}
```

- [ ] **Step 2: Run the bridge test to verify it fails**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . -run "TestApp_TaskMethodsDelegateToTaskService" -count=1
```

Expected:

- command exits non-zero
- failure mentions `taskService` field or methods do not delegate

- [ ] **Step 3: Add the scheduler adapter and delegate the App task methods**

```go
// Add to D:\codex_work\go-stock\app.go imports
import (
	taskservice "go-stock/backend/service/task"
)

type taskService interface {
	Create(ctx context.Context, task *models.CronTask) string
	Update(ctx context.Context, task *models.CronTask) string
	Delete(ctx context.Context, id uint) string
	GetByID(ctx context.Context, id uint) (*models.CronTask, error)
	List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp
	Enable(ctx context.Context, id uint, enable bool) string
	RunNow(ctx context.Context, id uint) error
	GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string]
	ValidateCronExpr(ctx context.Context, expr string) string
	Search(ctx context.Context, keyword string) []models.CronTask
	CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time
	CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time
}

type App struct {
	// existing fields...
	configService configService
	taskService   taskService
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

	app := &App{
		cache:              cache,
		cron:               c,
		cronEntrys:         make(map[string]cron.EntryID),
		AiTools:            tools,
		stockAlertLastSent: make(map[string]time.Time),
		priceAtAlertReset:  make(map[string]float64),
		marketReadService:  marketservice.NewService(marketsource.NewSource()),
		analysisService:    analysisservice.NewService(analysisProvider, analysisStore, analysisStore),
		configService:      configservice.NewService(configservice.NewStore()),
	}
	app.taskService = taskservice.NewService(taskservice.NewStore(), &appTaskScheduler{app: app}, func() context.Context {
		return app.ctx
	})
	return app
}
```

```go
// D:\codex_work\go-stock\app_task_scheduler.go
package main

import "go-stock/backend/service/task"

var _ task.Scheduler = (*appTaskScheduler)(nil)

type appTaskScheduler struct {
	app *App
}

func (s *appTaskScheduler) Register(key, spec string, job func()) error {
	if entryID, exists := s.app.getCronEntry(key); exists {
		s.app.cron.Remove(entryID)
	}
	entryID, err := s.app.cron.AddFunc(spec, job)
	if err != nil {
		return err
	}
	s.app.setCronEntry(key, entryID)
	return nil
}

func (s *appTaskScheduler) Unregister(key string) {
	if entryID, exists := s.app.getCronEntry(key); exists {
		s.app.cron.Remove(entryID)
		s.app.removeCronEntry(key)
	}
}
```

```go
// Replace the cron-task methods in D:\codex_work\go-stock\app.go
func (a *App) CreateCronTask(task *models.CronTask) string {
	return a.taskService.Create(a.ctx, task)
}

func (a *App) UpdateCronTask(task *models.CronTask) string {
	return a.taskService.Update(a.ctx, task)
}

func (a *App) DeleteCronTask(id uint) string {
	return a.taskService.Delete(a.ctx, id)
}

func (a *App) GetCronTaskByID(id uint) *models.CronTask {
	task, err := a.taskService.GetByID(a.ctx, id)
	if err != nil {
		return nil
	}
	return task
}

func (a *App) GetCronTaskList(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return a.taskService.List(a.ctx, query)
}

func (a *App) EnableCronTask(id uint, enable bool) string {
	return a.taskService.Enable(a.ctx, id, enable)
}

func (a *App) ExecuteCronTaskNow(id uint) string {
	go func() {
		err := a.taskService.RunNow(a.ctx, id)
		if err != nil {
			appError("trigger-cron-task", "cron.task_execute_failed", "trigger cron task failed", logger.Uint("task_id", id), logger.Err(err))
		}
	}()
	return "任务执行中"
}

func (a *App) GetCronTaskTypes() []lo.Tuple2[string, string] {
	return a.taskService.GetTaskTypes(a.ctx)
}

func (a *App) ValidateCronExpr(expr string) string {
	return a.taskService.ValidateCronExpr(a.ctx, expr)
}

func (a *App) SearchCronTasks(keyword string) []models.CronTask {
	return a.taskService.Search(a.ctx, keyword)
}

func (a *App) CalculateNextRunTime(cron string) string {
	return a.taskService.CalculateNextRunTime(a.ctx, cron).Format("2006-01-02 15:04:05")
}

func (a *App) CalculateNextRunTimes(cron string, count int) []string {
	times := a.taskService.CalculateNextRunTimes(a.ctx, cron, count)
	result := make([]string, 0, len(times))
	for _, t := range times {
		result = append(result, t.Format("2006-01-02 15:04:05"))
	}
	return result
}
```

- [ ] **Step 4: Run the bridge and task service tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/task -run "TestApp_TaskMethodsDelegateToTaskService|TestService_CreateUsesStableSchedulerKey|TestService_UpdateReplacesExistingRegistration|TestService_EnableAndDisableUseStableKey" -count=1
```

Expected:

- command exits `0`
- App bridge test and task service tests pass

- [ ] **Step 5: Commit**

```powershell
git add app.go app_task_scheduler.go app_task_test.go
git commit -m "refactor: bridge cron task access through app service"
```

### Task 6: Add Frontend `taskService` And Migrate The Cron Task Screen

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\services\taskService.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\services\taskService.test.mjs`
- Modify: `D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue`

- [ ] **Step 1: Write the failing frontend task service tests**

```javascript
import test from 'node:test'
import assert from 'node:assert/strict'

import {
  normalizeCronTask,
  normalizeCronTaskPage,
} from './taskService.mjs'

test('normalizeCronTask keeps stable camel-case fields', () => {
  assert.deepEqual(
    normalizeCronTask({ id: 7, name: '市场分析', cronExpr: '0 */5 * * * *', enable: true }),
    { id: 7, name: '市场分析', cronExpr: '0 */5 * * * *', enable: true, taskType: '', params: '', status: '', description: '', target: '', lastRunAt: '', nextRunAt: '', runCount: 0, lastRunResult: '' },
  )
})

test('normalizeCronTaskPage falls back to empty page', () => {
  assert.deepEqual(normalizeCronTaskPage(null), { total: 0, data: [] })
})
```

- [ ] **Step 2: Run the frontend task service tests to verify they fail**

Run:

```powershell
node --test frontend/src/services/taskService.test.mjs
```

Expected:

- command exits non-zero
- failure mentions `frontend/src/services/taskService.mjs` does not exist

- [ ] **Step 3: Add `taskService.mjs` and migrate the cron task page to it**

```javascript
// D:\codex_work\go-stock\frontend\src\services\taskService.mjs
import * as AppBindings from '../../wailsjs/go/main/App.js'

function toArray(value) {
  return Array.isArray(value) ? value : []
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback
}

export function normalizeCronTask(value) {
  const data = value ?? {}
  return {
    id: toNumber(data.id ?? data.ID, 0),
    name: data.name ?? '',
    cronExpr: data.cronExpr ?? '',
    taskType: data.taskType ?? '',
    target: data.target ?? '',
    params: data.params ?? '',
    enable: Boolean(data.enable),
    status: data.status ?? '',
    description: data.description ?? '',
    lastRunAt: data.lastRunAt ?? '',
    nextRunAt: data.nextRunAt ?? '',
    runCount: toNumber(data.runCount, 0),
    lastRunResult: data.lastRunResult ?? '',
  }
}

export function normalizeCronTaskPage(value) {
  const data = value ?? {}
  return {
    total: toNumber(data.total, 0),
    data: toArray(data.data).map(normalizeCronTask),
  }
}

export async function loadCronTask(id) { return normalizeCronTask(await AppBindings.GetCronTaskByID(id)) }
export async function loadCronTaskList(query) { return normalizeCronTaskPage(await AppBindings.GetCronTaskList(query)) }
export async function createCronTask(task) { return AppBindings.CreateCronTask(task) }
export async function updateCronTask(task) { return AppBindings.UpdateCronTask(task) }
export async function deleteCronTask(id) { return AppBindings.DeleteCronTask(id) }
export async function toggleCronTask(id, enable) { return AppBindings.EnableCronTask(id, enable) }
export async function executeCronTaskNow(id) { return AppBindings.ExecuteCronTaskNow(id) }
export async function loadCronTaskTypes() { return toArray(await AppBindings.GetCronTaskTypes()) }
export async function validateCronExpression(expr) { return AppBindings.ValidateCronExpr(expr) }
export async function calculateNextRunTime(expr) { return AppBindings.CalculateNextRunTime(expr) }
export async function calculateNextRunTimes(expr, count) { return toArray(await AppBindings.CalculateNextRunTimes(expr, count)) }
export async function searchCronTasks(keyword) { return toArray(await AppBindings.SearchCronTasks(keyword)).map(normalizeCronTask) }
```

```vue
<!-- Add to the import section in D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue -->
<script setup>
import {
  calculateNextRunTimes as loadNextRunTimes,
  createCronTask,
  deleteCronTask,
  executeCronTaskNow,
  loadCronTask,
  loadCronTaskList,
  loadCronTaskTypes,
  toggleCronTask,
  updateCronTask,
  validateCronExpression,
} from '../services/taskService.mjs'
</script>
```

```vue
<!-- Replace the task operations in D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue -->
const loadTaskTypes = async () => {
  try {
    const types = await loadCronTaskTypes()
    taskTypeOptions.value = types.map(t => ({ label: t.B, value: t.A }))
  } catch (error) {
    console.error('加载任务类型失败:', error)
  }
}

const loadTaskList = async () => {
  loading.value = true
  try {
    const result = await loadCronTaskList({
      page: currentPage.value,
      pageSize: pageSize.value,
      name: searchKeyword.value,
      taskType: filterTaskType.value,
      status: filterStatus.value
    })
    taskList.value = result.data || []
    total.value = result.total || 0
  } catch (error) {
    console.error('加载任务列表失败:', error)
    message.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const handleExecute = async (row) => {
  try {
    const result = await executeCronTaskNow(row.id)
    message.success(result)
  } catch (error) {
    message.error('执行任务失败：' + error.message)
  }
}

const handleToggleEnable = async (row) => {
  try {
    const newEnable = !row.enable
    const result = await toggleCronTask(row.id, newEnable)
    if (result === '操作成功') {
      message.success(newEnable ? '任务已启用' : '任务已禁用')
      await loadTaskList()
    } else {
      message.error(result)
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  }
}

watch([
  () => cronSecond.type,
  () => cronSecond.start,
  () => cronSecond.end,
  () => cronSecond.loopStart,
  () => cronSecond.loopStep,
  () => cronSecond.appoint,
  () => cronMinute.type,
  () => cronMinute.start,
  () => cronMinute.end,
  () => cronMinute.loopStart,
  () => cronMinute.loopStep,
  () => cronMinute.appoint,
  () => cronHour.type,
  () => cronHour.start,
  () => cronHour.end,
  () => cronHour.loopStart,
  () => cronHour.loopStep,
  () => cronHour.appoint,
  () => cronDay.type,
  () => cronDay.start,
  () => cronDay.end,
  () => cronMonth.type,
  () => cronMonth.start,
  () => cronMonth.end,
  () => cronWeek.type,
  () => cronWeek.days
], () => {
  generatedCronExpr.value = generateCronExpression()
  if (!generatedCronExpr.value) {
    calculateNextRunTime.value = ''
    nextRunTimes.value = []
    return
  }
  loadNextRunTimes(generatedCronExpr.value, 5)
    .then(res => {
      nextRunTimes.value = Array.isArray(res) ? res : []
      calculateNextRunTime.value = nextRunTimes.value[0] || ''
    })
    .catch(() => {
      nextRunTimes.value = []
      calculateNextRunTime.value = ''
    })
}, { deep: true })

const handleEdit = async (row) => {
  editingTask.value = true
  try {
    const task = await loadCronTask(row.id)
    if (task) {
      resetForm()
      formData.id = task.id
      formData.name = task.name
      formData.cronExpr = (task.cronExpr ?? '').trim()
      formData.taskType = task.taskType
      formData.target = task.target
      formData.params = task.params
      formData.enable = task.enable
      formData.status = task.status
      formData.description = task.description
      parseCronExpression(formData.cronExpr)
      if (task.taskType === 'stock_analysis' && task.params) {
        try {
          const parsed = JSON.parse(task.params)
          stockAnalysisParamsData.promptId = parsed.promptId ?? null
          stockAnalysisParamsData.aiConfigId = parsed.aiConfigId ?? null
          stockAnalysisParamsData.sysPromptId = parsed.sysPromptId ?? null
          stockAnalysisParamsData.thinking = parsed.thinking || false
          stockAnalysisParamsData.stockCode = parsed.stockCode || ''
          stockAnalysisParamsData.stockName = parsed.stockName || ''
        } catch (e) {
          console.error('解析参数失败:', e)
        }
      }
      if (task.taskType === 'market_analysis' && task.params) {
        try {
          const parsed = JSON.parse(task.params)
          marketAnalysisParamsData.promptId = parsed.promptId ?? null
          marketAnalysisParamsData.aiConfigId = parsed.aiConfigId ?? null
          marketAnalysisParamsData.sysPromptId = parsed.sysPromptId ?? null
          marketAnalysisParamsData.thinking = parsed.thinking || false
        } catch (e) {
          console.error('解析参数失败:', e)
        }
      }
      showCreateModal.value = true
    }
  } catch (error) {
    message.error('获取任务详情失败：' + error.message)
  }
}

const validateCronExpressionUI = async () => {
  if (!formData.cronExpr) return false
  try {
    const result = await validateCronExpression(formData.cronExpr)
    if (result.includes('有效')) {
      return true
    }
    message.error(result)
    return false
  } catch (error) {
    message.error('Cron 表达式无效：' + error.message)
    return false
  }
}

const checkCronInterval = async (cronExpr) => {
  if (!cronExpr) return { ok: true }
  try {
    const times = await loadNextRunTimes(cronExpr, 10)
    if (!Array.isArray(times) || times.length < 2) return { ok: true }
    let minSeconds = Infinity
    for (let i = 1; i < times.length; i++) {
      const prev = new Date(times[i - 1]).getTime()
      const curr = new Date(times[i]).getTime()
      if (!Number.isNaN(prev) && !Number.isNaN(curr)) {
        const sec = (curr - prev) / 1000
        if (sec < minSeconds) minSeconds = sec
      }
    }
    if (minSeconds !== Infinity && minSeconds < 60) {
      return { ok: false, minIntervalSeconds: Math.round(minSeconds) }
    }
    return { ok: true }
  } catch (_) {
    return { ok: true }
  }
}

const handleSubmit = async () => {
  try {
    if (formRef.value) {
      try {
        await formRef.value.validate()
      } catch {
        return
      }
    }
    if (!await validateCronExpressionUI()) {
      return
    }
    const intervalCheck = await checkCronInterval(formData.cronExpr)
    if (!intervalCheck.ok) {
      message.warning(`两次执行间隔过短（约 ${intervalCheck.minIntervalSeconds} 秒），请将间隔设置为至少 60 秒后再保存。`)
      return
    }
    formData.params = generatedParamsJson.value
    submitting.value = true
    const submitData = { ...formData }
    const result = formData.id ? await updateCronTask(submitData) : await createCronTask(submitData)
    if (result.includes('成功')) {
      message.success(result)
      showCreateModal.value = false
      await loadTaskList()
    } else {
      message.error(result)
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id) => {
  try {
    const result = await deleteCronTask(id)
    if (result === '删除成功') {
      message.success('任务已删除')
      await loadTaskList()
    } else {
      message.error(result)
    }
  } catch (error) {
    message.error('删除失败：' + error.message)
  }
}
```

- [ ] **Step 4: Run the frontend task service tests**

Run:

```powershell
node --test frontend/src/services/taskService.test.mjs
```

Expected:

- command exits `0`
- task service tests pass

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/services/taskService.mjs frontend/src/services/taskService.test.mjs frontend/src/components/cron-task-manager.vue
git commit -m "refactor: migrate cron task screen to frontend task service"
```

### Task 7: Run Slice Verification And Review The Result Against Phase 4 Goals

**Files:**
- Verify only:
  - `D:\codex_work\go-stock\backend\service\config\store.go`
  - `D:\codex_work\go-stock\backend\service\config\service.go`
  - `D:\codex_work\go-stock\backend\service\config\service_test.go`
  - `D:\codex_work\go-stock\backend\service\task\store.go`
  - `D:\codex_work\go-stock\backend\service\task\service.go`
  - `D:\codex_work\go-stock\backend\service\task\service_test.go`
  - `D:\codex_work\go-stock\app.go`
  - `D:\codex_work\go-stock\app_common.go`
  - `D:\codex_work\go-stock\app_task_scheduler.go`
  - `D:\codex_work\go-stock\app_config_test.go`
  - `D:\codex_work\go-stock\app_task_test.go`
  - `D:\codex_work\go-stock\frontend\src\services\configService.mjs`
  - `D:\codex_work\go-stock\frontend\src\services\configService.test.mjs`
  - `D:\codex_work\go-stock\frontend\src\services\taskService.mjs`
  - `D:\codex_work\go-stock\frontend\src\services\taskService.test.mjs`
  - `D:\codex_work\go-stock\frontend\src\components\settings.vue`
  - `D:\codex_work\go-stock\frontend\src\components\cron-task-manager.vue`

- [ ] **Step 1: Run the backend slice tests**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . ./backend/service/config ./backend/service/task -run "TestApp_ConfigAndPromptMethodsDelegateToConfigService|TestApp_TaskMethodsDelegateToTaskService|TestService_GetAiConfigsAndPromptsReturnStableDefaults|TestService_CreateUsesStableSchedulerKey" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 2: Run the frontend slice tests**

Run:

```powershell
node --test frontend/src/services/configService.test.mjs frontend/src/services/taskService.test.mjs
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
- no new frontend compile errors in `settings.vue` or `cron-task-manager.vue`

- [ ] **Step 4: Run the Wails desktop smoke build**

Run:

```powershell
$env:PATH='C:\Program Files\Go\bin;' + $env:PATH
& 'C:\Users\Aspir\go\bin\wails.exe' build --platform windows/amd64
```

Expected:

- command exits `0`
- no new binding methods are required for this slice, so generated Wails binding churn should be empty or minimal

- [ ] **Step 5: Review the diff is limited to the planned slice**

Run:

```powershell
git diff --stat <phase4-start-sha>..HEAD
```

Expected:

- diff is limited to the new config/task service files, App bridge wiring, the frontend config/task service wrappers, and the two migrated components
- `<phase4-start-sha>` is the commit recorded immediately before Task 1 starts, so review loops or extra fix commits do not hide part of the slice

## Self-Review

### Spec Coverage

- The overall spec’s phase 4 requirement for “任务执行链路有清晰服务边界” is covered by Task 4 and Task 5 via `backend/service/task` plus App delegation.
- The phase 4 requirement for “设置读取与保存具备稳定边界” is covered by Task 1, Task 2, and Task 3 via `backend/service/config` and `frontend/src/services/configService.mjs`.
- The user-approved slice “settings 页旧 prompt 弹窗 + cron 页 prompt / ai config 共享访问” is covered by Task 1, Task 2, and Task 3.
- The deliberately excluded assistant and multi-turn agent work matches the agreed scope for this plan and prevents the slice from widening.

### Placeholder Scan

- No `TODO`, `TBD`, “implement later”, or “similar to Task N” placeholders remain.
- Every code-changing step includes concrete file paths, code, and commands.
- Existing Wails method names are intentionally preserved to avoid unnecessary binding churn.

### Type Consistency

- Backend config service names are consistent across tasks:
  - `GetConfig`
  - `UpdateConfig`
  - `GetAiConfigs`
  - `GetPromptTemplates`
  - `GetPromptTemplatePage`
  - `SavePromptTemplate`
  - `DeletePromptTemplate`
  - `SaveLegacyPrompt`
  - `DeleteLegacyPrompt`
- Backend task service names are consistent across tasks:
  - `Create`
  - `Update`
  - `Delete`
  - `GetByID`
  - `List`
  - `Enable`
  - `RunNow`
  - `GetTaskTypes`
  - `ValidateCronExpr`
  - `Search`
  - `CalculateNextRunTime`
  - `CalculateNextRunTimes`
- Frontend service names are consistent across tasks:
  - `loadAppConfig`
  - `saveAppConfig`
  - `loadAiConfigs`
  - `loadPromptTemplates`
  - `saveLegacyPrompt`
  - `deleteLegacyPrompt`
  - `loadCronTask`
  - `loadCronTaskList`
  - `createCronTask`
  - `updateCronTask`
  - `deleteCronTask`
  - `toggleCronTask`
  - `executeCronTaskNow`
  - `validateCronExpression`
  - `calculateNextRunTime`
  - `calculateNextRunTimes`
