package main

import (
	"context"
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type fakeConfigService struct {
	getConfigResult *data.SettingConfig
	updateResult    string
	aiConfigsResult []*data.AIConfig
	promptsResult   *[]models.PromptTemplate
	promptPage      *models.PromptTemplatePageData

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
	return f.promptPage, nil
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
