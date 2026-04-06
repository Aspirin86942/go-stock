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

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	if store == nil {
		panic("config: store dependency is required")
	}
	return &Service{store: store}
}

func (s *Service) GetConfig(ctx context.Context) *data.SettingConfig {
	cfg := s.store.GetConfig(ctx)
	if cfg == nil {
		cfg = &data.SettingConfig{}
	}
	if cfg.Settings == nil {
		cfg.Settings = &data.Settings{}
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
	templates := s.store.GetPromptTemplates(ctx, name, promptType)
	if templates == nil {
		empty := []models.PromptTemplate{}
		return &empty
	}
	return templates
}

func (s *Service) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	page, err := s.store.GetPromptTemplatePage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.PromptTemplatePageData{
			List: []models.PromptTemplate{},
		}, nil
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
		ID:      prompt.ID,
		Name:    prompt.Name,
		Content: prompt.Content,
		Type:    prompt.Type,
	})
}

func (s *Service) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	return s.DeletePromptTemplate(ctx, id)
}
