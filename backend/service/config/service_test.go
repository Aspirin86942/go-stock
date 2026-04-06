package config

import (
	"context"
	"errors"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

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
}

func (f *fakeStore) GetConfig(ctx context.Context) *data.SettingConfig {
	f.getConfigCtx = ctx
	return f.cfg
}

func (f *fakeStore) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	f.updateCtx = ctx
	f.updateArg = cfg
	return f.updateResult
}

func (f *fakeStore) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	f.getPromptsCtx = ctx
	return f.templates
}

func (f *fakeStore) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	f.getPageCtx = ctx
	if f.pageErr != nil {
		return nil, f.pageErr
	}
	return f.page, nil
}

func (f *fakeStore) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.saveCtx = ctx
	f.savedTemplate = template
	return f.saveResult
}

func (f *fakeStore) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deleteCtx = ctx
	f.deletedID = id
	return f.deleteResult
}

func TestService_GetAiConfigsAndPromptsReturnStableDefaults(t *testing.T) {
	store := &fakeStore{
		cfg:       nil,
		templates: nil,
	}
	svc := NewService(store)
	bg := context.Background()

	cfg := svc.GetConfig(bg)
	if cfg == nil {
		t.Fatalf("expected non-nil config")
	}
	if cfg.AiConfigs == nil {
		t.Fatalf("expected non-nil ai configs")
	}
	if len(cfg.AiConfigs) != 0 {
		t.Fatalf("expected empty ai configs, got %d", len(cfg.AiConfigs))
	}
	if store.getConfigCtx != bg {
		t.Fatalf("expected background context passed to GetConfig")
	}

	aiConfigs := svc.GetAiConfigs(bg)
	if aiConfigs == nil {
		t.Fatalf("expected non-nil ai configs from GetAiConfigs")
	}
	if len(aiConfigs) != 0 {
		t.Fatalf("expected empty ai configs from GetAiConfigs, got %d", len(aiConfigs))
	}
	if store.getConfigCtx != bg {
		t.Fatalf("expected background context passed to GetAiConfigs->GetConfig")
	}

	templates := svc.GetPromptTemplates(bg, "", "")
	if templates == nil {
		t.Fatalf("expected non-nil templates pointer")
	}
	if len(*templates) != 0 {
		t.Fatalf("expected empty prompt templates, got %d", len(*templates))
	}
	if store.getPromptsCtx != bg {
		t.Fatalf("expected background context passed to GetPromptTemplates")
	}
}

func TestService_SaveLegacyPromptMapsToPromptTemplate(t *testing.T) {
	store := &fakeStore{
		saveResult: "ok",
	}
	svc := NewService(store)
	bg := context.Background()

	legacy := models.Prompt{
		ID:      12,
		Name:    "legacy",
		Content: "legacy-content",
		Type:    "模型系统Prompt",
	}
	msg := svc.SaveLegacyPrompt(bg, legacy)
	if msg != "ok" {
		t.Fatalf("expected delegated save result, got %q", msg)
	}
	if store.saveCtx != bg {
		t.Fatalf("expected background context passed to save")
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
	bg := context.Background()

	page, err := svc.GetPromptTemplatePage(bg, models.PromptTemplateQuery{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected propagated error, got %v", err)
	}
	if page != nil {
		t.Fatalf("expected nil page when error happens, got %#v", page)
	}
	fake := svc.store.(*fakeStore)
	if fake.getPageCtx != bg {
		t.Fatalf("expected background context passed to GetPromptTemplatePage")
	}
}

func TestService_UpdateConfigDelegates(t *testing.T) {
	store := &fakeStore{
		updateResult: "saved",
	}
	svc := NewService(store)
	bg := context.Background()

	input := &data.SettingConfig{
		Settings: &data.Settings{
			RefreshInterval: 30,
		},
		AiConfigs: []*data.AIConfig{},
	}
	msg := svc.UpdateConfig(bg, input)
	if msg != "saved" {
		t.Fatalf("expected delegated update result, got %q", msg)
	}
	if store.updateArg != input {
		t.Fatalf("expected same config pointer delegated")
	}
	if store.updateCtx != bg {
		t.Fatalf("expected background context passed to update")
	}
}
