package config

import (
	"errors"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type fakeStore struct {
	cfg             *data.SettingConfig
	updateArg       *data.SettingConfig
	updateResult    string
	templates       *[]models.PromptTemplate
	page            *models.PromptTemplatePageData
	pageErr         error
	savedTemplate   models.PromptTemplate
	saveResult      string
	deletedID       uint
	deleteResult    string
}

func (f *fakeStore) GetConfig() *data.SettingConfig {
	return f.cfg
}

func (f *fakeStore) UpdateConfig(cfg *data.SettingConfig) string {
	f.updateArg = cfg
	return f.updateResult
}

func (f *fakeStore) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	return f.templates
}

func (f *fakeStore) GetPromptTemplatePage(query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	if f.pageErr != nil {
		return nil, f.pageErr
	}
	return f.page, nil
}

func (f *fakeStore) SavePromptTemplate(template models.PromptTemplate) string {
	f.savedTemplate = template
	return f.saveResult
}

func (f *fakeStore) DeletePromptTemplate(id uint) string {
	f.deletedID = id
	return f.deleteResult
}

func TestService_GetAiConfigsAndPromptsReturnStableDefaults(t *testing.T) {
	svc := NewService(&fakeStore{
		cfg:       nil,
		templates: nil,
	})

	cfg := svc.GetConfig()
	if cfg == nil {
		t.Fatalf("expected non-nil config")
	}
	if cfg.AiConfigs == nil {
		t.Fatalf("expected non-nil ai configs")
	}
	if len(cfg.AiConfigs) != 0 {
		t.Fatalf("expected empty ai configs, got %d", len(cfg.AiConfigs))
	}

	aiConfigs := svc.GetAiConfigs()
	if aiConfigs == nil {
		t.Fatalf("expected non-nil ai configs from GetAiConfigs")
	}
	if len(aiConfigs) != 0 {
		t.Fatalf("expected empty ai configs from GetAiConfigs, got %d", len(aiConfigs))
	}

	templates := svc.GetPromptTemplates("", "")
	if templates == nil {
		t.Fatalf("expected non-nil templates pointer")
	}
	if len(*templates) != 0 {
		t.Fatalf("expected empty prompt templates, got %d", len(*templates))
	}
}

func TestService_SaveLegacyPromptMapsToPromptTemplate(t *testing.T) {
	store := &fakeStore{
		saveResult: "ok",
	}
	svc := NewService(store)

	legacy := models.Prompt{
		ID:      12,
		Name:    "legacy",
		Content: "legacy-content",
		Type:    "模型系统Prompt",
	}
	msg := svc.SaveLegacyPrompt(legacy)
	if msg != "ok" {
		t.Fatalf("expected delegated save result, got %q", msg)
	}
	if store.savedTemplate.ID != legacy.ID {
		t.Fatalf("expected id mapped, got %d", store.savedTemplate.ID)
	}
	if store.savedTemplate.Name != legacy.Name {
		t.Fatalf("expected name mapped, got %q", store.savedTemplate.Name)
	}
	if store.savedTemplate.Content != legacy.Content {
		t.Fatalf("expected content mapped, got %q", store.savedTemplate.Content)
	}
	if store.savedTemplate.Type != legacy.Type {
		t.Fatalf("expected type mapped, got %q", store.savedTemplate.Type)
	}
}

func TestService_GetPromptTemplatePagePropagatesError(t *testing.T) {
	expectedErr := errors.New("query failed")
	svc := NewService(&fakeStore{
		pageErr: expectedErr,
	})

	page, err := svc.GetPromptTemplatePage(models.PromptTemplateQuery{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected propagated error, got %v", err)
	}
	if page != nil {
		t.Fatalf("expected nil page when error happens, got %#v", page)
	}
}

func TestService_UpdateConfigDelegates(t *testing.T) {
	store := &fakeStore{
		updateResult: "saved",
	}
	svc := NewService(store)

	input := &data.SettingConfig{
		Settings: &data.Settings{
			RefreshInterval: 30,
		},
		AiConfigs: []*data.AIConfig{},
	}
	msg := svc.UpdateConfig(input)
	if msg != "saved" {
		t.Fatalf("expected delegated update result, got %q", msg)
	}
	if store.updateArg != input {
		t.Fatalf("expected same config pointer delegated")
	}
}
