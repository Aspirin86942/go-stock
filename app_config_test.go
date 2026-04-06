package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"

	"github.com/robfig/cron/v3"
)

type fakeConfigService struct {
	getConfigResult *data.SettingConfig
	updateResult    string
	aiConfigsResult []*data.AIConfig
	promptsResult   *[]models.PromptTemplate
	promptPage      *models.PromptTemplatePageData
	promptPageErr   error

	getConfigCalled          int
	updateConfigCalled       int
	getAiConfigsCalled       int
	getPromptTemplatesCalled int
	getPromptPageCalled      int
	savePromptTemplateCalled int
	deletePromptCalled       int
	saveLegacyPromptCalled   int
	deleteLegacyPromptCalled int

	lastCtx                  context.Context
	lastPromptName           string
	lastPromptType           string
	lastTemplate             models.PromptTemplate
	lastPromptQuery          models.PromptTemplateQuery
	lastDeletedTemplateID    uint
	lastSavedLegacyPrompt    models.Prompt
	lastDeletedLegacyPrompt  uint
	lastUpdatedSettingConfig *data.SettingConfig
}

func (f *fakeConfigService) GetConfig(ctx context.Context) *data.SettingConfig {
	f.getConfigCalled++
	f.lastCtx = ctx
	return f.getConfigResult
}

func (f *fakeConfigService) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	f.updateConfigCalled++
	f.lastCtx = ctx
	f.lastUpdatedSettingConfig = cfg
	return f.updateResult
}

func (f *fakeConfigService) GetAiConfigs(ctx context.Context) []*data.AIConfig {
	f.getAiConfigsCalled++
	f.lastCtx = ctx
	return f.aiConfigsResult
}

func (f *fakeConfigService) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	f.getPromptTemplatesCalled++
	f.lastCtx = ctx
	f.lastPromptName = name
	f.lastPromptType = promptType
	return f.promptsResult
}

func (f *fakeConfigService) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	f.getPromptPageCalled++
	f.lastCtx = ctx
	f.lastPromptQuery = query
	return f.promptPage, f.promptPageErr
}

func (f *fakeConfigService) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savePromptTemplateCalled++
	f.lastCtx = ctx
	f.lastTemplate = template
	return "模板已保存"
}

func (f *fakeConfigService) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletePromptCalled++
	f.lastCtx = ctx
	f.lastDeletedTemplateID = id
	return "模板已删除"
}

func (f *fakeConfigService) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	f.saveLegacyPromptCalled++
	f.lastCtx = ctx
	f.lastSavedLegacyPrompt = prompt
	return "旧版Prompt已保存"
}

func (f *fakeConfigService) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	f.deleteLegacyPromptCalled++
	f.lastCtx = ctx
	f.lastDeletedLegacyPrompt = id
	return "旧版Prompt已删除"
}

func TestApp_ConfigAndPromptMethodsDelegateToConfigService(t *testing.T) {
	ctx := context.Background()
	expectedConfig := &data.SettingConfig{
		Settings: &data.Settings{
			RefreshInterval: 15,
		},
	}
	expectedAiConfigs := []*data.AIConfig{{Name: "deepseek"}}
	expectedTemplates := []models.PromptTemplate{{ID: 1, Name: "系统模板", Type: "模型系统Prompt", Content: "分析风险"}}
	expectedTemplatePage := &models.PromptTemplatePageData{
		List:       expectedTemplates,
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	fake := &fakeConfigService{
		getConfigResult: expectedConfig,
		updateResult:    "更新成功",
		aiConfigsResult: expectedAiConfigs,
		promptsResult:   &expectedTemplates,
		promptPage:      expectedTemplatePage,
	}

	app := &App{
		ctx:           ctx,
		configService: fake,
	}

	if got := app.GetConfig(); got != expectedConfig {
		t.Fatalf("GetConfig() should delegate to configService, got=%#v want=%#v", got, expectedConfig)
	}
	if fake.getConfigCalled != 1 {
		t.Fatalf("GetConfig() should call configService.GetConfig exactly once, got=%d", fake.getConfigCalled)
	}

	newConfig := &data.SettingConfig{Settings: &data.Settings{RefreshInterval: 0}}
	if msg := app.UpdateConfig(newConfig); msg != "更新成功" {
		t.Fatalf("UpdateConfig() should return configService result, got=%q", msg)
	}
	if fake.updateConfigCalled != 1 {
		t.Fatalf("UpdateConfig() should call configService.UpdateConfig exactly once, got=%d", fake.updateConfigCalled)
	}
	if fake.lastUpdatedSettingConfig != newConfig {
		t.Fatalf("UpdateConfig() should pass the same settingConfig pointer")
	}

	if got := app.GetAiConfigs(); !reflect.DeepEqual(got, fake.aiConfigsResult) {
		t.Fatalf("GetAiConfigs() should delegate to configService, got=%#v want=%#v", got, fake.aiConfigsResult)
	}
	if fake.getAiConfigsCalled != 1 {
		t.Fatalf("GetAiConfigs() should call configService.GetAiConfigs exactly once, got=%d", fake.getAiConfigsCalled)
	}

	if got := app.GetPromptTemplates("系统模板", "模型系统Prompt"); got != fake.promptsResult {
		t.Fatalf("GetPromptTemplates() should delegate to configService, got=%#v want=%#v", got, fake.promptsResult)
	}
	if fake.getPromptTemplatesCalled != 1 || fake.lastPromptName != "系统模板" || fake.lastPromptType != "模型系统Prompt" {
		t.Fatalf("GetPromptTemplates() should pass through args, called=%d name=%q type=%q", fake.getPromptTemplatesCalled, fake.lastPromptName, fake.lastPromptType)
	}

	legacyPrompt := models.Prompt{ID: 9, Name: "旧提示词", Type: "模型用户Prompt", Content: "请总结"}
	if msg := app.AddPrompt(legacyPrompt); msg != "旧版Prompt已保存" {
		t.Fatalf("AddPrompt() should use SaveLegacyPrompt result, got=%q", msg)
	}
	if fake.saveLegacyPromptCalled != 1 || fake.lastSavedLegacyPrompt != legacyPrompt {
		t.Fatalf("AddPrompt() should call SaveLegacyPrompt with exact prompt, called=%d got=%#v", fake.saveLegacyPromptCalled, fake.lastSavedLegacyPrompt)
	}

	if msg := app.DelPrompt(101); msg != "旧版Prompt已删除" {
		t.Fatalf("DelPrompt() should use DeleteLegacyPrompt result, got=%q", msg)
	}
	if fake.deleteLegacyPromptCalled != 1 || fake.lastDeletedLegacyPrompt != 101 {
		t.Fatalf("DelPrompt() should call DeleteLegacyPrompt with id, called=%d id=%d", fake.deleteLegacyPromptCalled, fake.lastDeletedLegacyPrompt)
	}

	query := models.PromptTemplateQuery{Page: 2, PageSize: 20, Name: "系统"}
	if got := app.GetPromptTemplateList(query); got != expectedTemplatePage {
		t.Fatalf("GetPromptTemplateList() should delegate to configService, got=%#v want=%#v", got, expectedTemplatePage)
	}
	if fake.getPromptPageCalled != 1 || fake.lastPromptQuery != query {
		t.Fatalf("GetPromptTemplateList() should pass through query, called=%d query=%#v", fake.getPromptPageCalled, fake.lastPromptQuery)
	}

	template := models.PromptTemplate{ID: 2, Name: "新模板", Type: "模型系统Prompt", Content: "内容A"}
	if msg := app.AddPromptTemplate(template); msg != "模板已保存" {
		t.Fatalf("AddPromptTemplate() should delegate to SavePromptTemplate, got=%q", msg)
	}
	if fake.savePromptTemplateCalled != 1 || fake.lastTemplate != template {
		t.Fatalf("AddPromptTemplate() should call SavePromptTemplate once with template, called=%d template=%#v", fake.savePromptTemplateCalled, fake.lastTemplate)
	}

	updatedTemplate := models.PromptTemplate{ID: 2, Name: "新模板", Type: "模型系统Prompt", Content: "内容B"}
	if msg := app.UpdatePromptTemplate(updatedTemplate); msg != "模板已保存" {
		t.Fatalf("UpdatePromptTemplate() should delegate to SavePromptTemplate, got=%q", msg)
	}
	if fake.savePromptTemplateCalled != 2 || fake.lastTemplate != updatedTemplate {
		t.Fatalf("UpdatePromptTemplate() should call SavePromptTemplate twice total and update latest template, called=%d template=%#v", fake.savePromptTemplateCalled, fake.lastTemplate)
	}

	if msg := app.DeletePromptTemplate(88); msg != "模板已删除" {
		t.Fatalf("DeletePromptTemplate() should delegate to DeletePromptTemplate, got=%q", msg)
	}
	if fake.deletePromptCalled != 1 || fake.lastDeletedTemplateID != 88 {
		t.Fatalf("DeletePromptTemplate() should call DeletePromptTemplate with id, called=%d id=%d", fake.deletePromptCalled, fake.lastDeletedTemplateID)
	}

	if fake.lastCtx != ctx {
		t.Fatalf("all delegated methods should pass app ctx")
	}
}

type legacyPromptAwareAnalysisServiceStub struct {
	legacyTemplates      *[]models.PromptTemplate
	legacyTemplatePage   *models.PromptTemplatePageData
	legacySavedPrompt    models.Prompt
	legacySavedTemplate  models.PromptTemplate
	legacyDeletedPrompt  uint
	legacyDeletedTmplID  uint
	getPromptCalled      int
	getPromptPageCalled  int
	saveLegacyCalled     int
	saveTemplateCalled   int
	deleteLegacyCalled   int
	deleteTemplateCalled int
}

func (s *legacyPromptAwareAnalysisServiceStub) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (s *legacyPromptAwareAnalysisServiceStub) StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (s *legacyPromptAwareAnalysisServiceStub) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
}

func (s *legacyPromptAwareAnalysisServiceStub) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return nil
}

func (s *legacyPromptAwareAnalysisServiceStub) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return &models.AIResponseResultPageData{}, nil
}

func (s *legacyPromptAwareAnalysisServiceStub) DeleteResult(ctx context.Context, id uint) error {
	return nil
}

func (s *legacyPromptAwareAnalysisServiceStub) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return nil
}

func (s *legacyPromptAwareAnalysisServiceStub) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	s.getPromptCalled++
	return s.legacyTemplates
}

func (s *legacyPromptAwareAnalysisServiceStub) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	s.getPromptPageCalled++
	return s.legacyTemplatePage, nil
}

func (s *legacyPromptAwareAnalysisServiceStub) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	s.saveTemplateCalled++
	s.legacySavedTemplate = template
	return "legacy-template-saved"
}

func (s *legacyPromptAwareAnalysisServiceStub) DeletePromptTemplate(ctx context.Context, id uint) string {
	s.deleteTemplateCalled++
	s.legacyDeletedTmplID = id
	return "legacy-template-deleted"
}

func (s *legacyPromptAwareAnalysisServiceStub) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	s.saveLegacyCalled++
	s.legacySavedPrompt = prompt
	return "legacy-prompt-saved"
}

func (s *legacyPromptAwareAnalysisServiceStub) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	s.deleteLegacyCalled++
	s.legacyDeletedPrompt = id
	return "legacy-prompt-deleted"
}

func TestApp_PromptMethodsReturnFallbackDefaultsWhenConfigServiceNil(t *testing.T) {
	templates := []models.PromptTemplate{{ID: 8, Name: "legacy", Type: "模型系统Prompt"}}
	legacyPage := &models.PromptTemplatePageData{List: templates, Total: 1, Page: 1, PageSize: 10, TotalPages: 1}
	legacy := &legacyPromptAwareAnalysisServiceStub{
		legacyTemplates:    &templates,
		legacyTemplatePage: legacyPage,
	}
	app := &App{
		ctx:             context.Background(),
		analysisService: legacy,
		configService:   nil,
	}

	gotTemplates := app.GetPromptTemplates("legacy", "模型系统Prompt")
	if gotTemplates == nil {
		t.Fatalf("GetPromptTemplates() should return non-nil defaults when configService is nil")
	}
	if len(*gotTemplates) != 0 {
		t.Fatalf("GetPromptTemplates() should return empty defaults, got=%#v", gotTemplates)
	}

	prompt := models.Prompt{ID: 9, Name: "legacy prompt", Type: "模型用户Prompt", Content: "A"}
	if msg := app.AddPrompt(prompt); msg != "保存失败" {
		t.Fatalf("AddPrompt() should return nil-safe fallback, got=%q", msg)
	}
	if msg := app.DelPrompt(77); msg != "删除失败" {
		t.Fatalf("DelPrompt() should return nil-safe fallback, got=%q", msg)
	}

	query := models.PromptTemplateQuery{Page: 1, PageSize: 10}
	page := app.GetPromptTemplateList(query)
	if page == nil {
		t.Fatalf("GetPromptTemplateList() should return empty page instead of nil")
	}
	if len(page.List) != 0 || page.Total != 0 || page.Page != 0 || page.PageSize != 0 || page.TotalPages != 0 {
		t.Fatalf("GetPromptTemplateList() should return zero-value empty page, got=%#v", page)
	}

	template := models.PromptTemplate{ID: 11, Name: "legacy-template", Type: "模型系统Prompt", Content: "B"}
	if msg := app.AddPromptTemplate(template); msg != "保存失败" {
		t.Fatalf("AddPromptTemplate() should return nil-safe fallback, got=%q", msg)
	}
	if msg := app.UpdatePromptTemplate(template); msg != "保存失败" {
		t.Fatalf("UpdatePromptTemplate() should return nil-safe fallback, got=%q", msg)
	}
	if msg := app.DeletePromptTemplate(11); msg != "删除失败" {
		t.Fatalf("DeletePromptTemplate() should return nil-safe fallback, got=%q", msg)
	}

	if legacy.getPromptCalled != 0 || legacy.getPromptPageCalled != 0 || legacy.saveTemplateCalled != 0 || legacy.deleteTemplateCalled != 0 || legacy.saveLegacyCalled != 0 || legacy.deleteLegacyCalled != 0 {
		t.Fatalf("legacy analysis prompt methods should not be called after closeout, got=%#v", legacy)
	}
}

func TestApp_UpdateConfigRefreshesMonitorStockPricesAndDelegates(t *testing.T) {
	fake := &fakeConfigService{updateResult: "更新成功"}
	cronScheduler := cron.New(cron.WithSeconds())
	cronScheduler.Start()
	t.Cleanup(func() { cronScheduler.Stop() })

	app := &App{
		ctx:           context.Background(),
		cron:          cronScheduler,
		cronEntrys:    make(map[string]cron.EntryID),
		configService: fake,
	}

	oldID, err := app.cron.AddFunc("@every 100s", func() {})
	if err != nil {
		t.Fatalf("setup old cron entry failed: %v", err)
	}
	app.setCronEntry("MonitorStockPrices", oldID)

	cfg := &data.SettingConfig{Settings: &data.Settings{RefreshInterval: 5}}
	if msg := app.UpdateConfig(cfg); msg != "更新成功" {
		t.Fatalf("UpdateConfig() should return delegated result, got=%q", msg)
	}
	if fake.updateConfigCalled != 1 || fake.lastUpdatedSettingConfig != cfg {
		t.Fatalf("UpdateConfig() should delegate to configService.UpdateConfig, called=%d cfg=%#v", fake.updateConfigCalled, fake.lastUpdatedSettingConfig)
	}

	newID, exists := app.getCronEntry("MonitorStockPrices")
	if !exists {
		t.Fatalf("UpdateConfig() with RefreshInterval>0 should register MonitorStockPrices cron entry")
	}
	if newID == oldID {
		t.Fatalf("UpdateConfig() should refresh MonitorStockPrices cron entry id, old=%d new=%d", oldID, newID)
	}
}

func TestApp_GetPromptTemplateListFallsBackToEmptyPageOnError(t *testing.T) {
	fake := &fakeConfigService{
		promptPageErr: errors.New("query failed"),
	}
	app := &App{
		ctx:           context.Background(),
		configService: fake,
	}

	page := app.GetPromptTemplateList(models.PromptTemplateQuery{Page: 1, PageSize: 10})
	if page == nil {
		t.Fatalf("GetPromptTemplateList() should return empty page instead of nil on error")
	}
	if len(page.List) != 0 || page.Total != 0 {
		t.Fatalf("GetPromptTemplateList() on error should return zero-value empty page, got=%#v", page)
	}
	if fake.getPromptPageCalled != 1 {
		t.Fatalf("GetPromptTemplateList() should still call configService once, got=%d", fake.getPromptPageCalled)
	}
}
