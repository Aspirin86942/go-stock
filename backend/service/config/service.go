package config

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Store interface {
	GetConfig() *data.SettingConfig
	UpdateConfig(cfg *data.SettingConfig) string
	GetPromptTemplates(name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(template models.PromptTemplate) string
	DeletePromptTemplate(id uint) string
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

func (s *Service) GetConfig() *data.SettingConfig {
	cfg := s.store.GetConfig()
	if cfg == nil {
		cfg = &data.SettingConfig{}
	}
	if cfg.AiConfigs == nil {
		cfg.AiConfigs = []*data.AIConfig{}
	}
	return cfg
}

func (s *Service) UpdateConfig(cfg *data.SettingConfig) string {
	return s.store.UpdateConfig(cfg)
}

func (s *Service) GetAiConfigs() []*data.AIConfig {
	return s.GetConfig().AiConfigs
}

func (s *Service) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	templates := s.store.GetPromptTemplates(name, promptType)
	if templates == nil {
		empty := []models.PromptTemplate{}
		return &empty
	}
	return templates
}

func (s *Service) GetPromptTemplatePage(query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	page, err := s.store.GetPromptTemplatePage(query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.PromptTemplatePageData{}, nil
	}
	return page, nil
}

func (s *Service) SavePromptTemplate(template models.PromptTemplate) string {
	return s.store.SavePromptTemplate(template)
}

func (s *Service) DeletePromptTemplate(id uint) string {
	return s.store.DeletePromptTemplate(id)
}

func (s *Service) SaveLegacyPrompt(prompt models.Prompt) string {
	return s.SavePromptTemplate(models.PromptTemplate{
		ID:      prompt.ID,
		Name:    prompt.Name,
		Content: prompt.Content,
		Type:    prompt.Type,
	})
}

func (s *Service) DeleteLegacyPrompt(id uint) string {
	return s.DeletePromptTemplate(id)
}
