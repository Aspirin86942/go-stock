package config

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type DataStore struct{}

func NewStore() *DataStore {
	return &DataStore{}
}

func (s *DataStore) GetConfig(ctx context.Context) *data.SettingConfig {
	_ = ctx
	return data.GetSettingConfig()
}

func (s *DataStore) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	_ = ctx
	return data.UpdateConfig(cfg)
}

func (s *DataStore) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	_ = ctx
	return data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
}

func (s *DataStore) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	_ = ctx
	return data.NewPromptTemplateApi().GetPromptTemplateList(&query)
}

func (s *DataStore) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	_ = ctx
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (s *DataStore) DeletePromptTemplate(ctx context.Context, id uint) string {
	_ = ctx
	return data.NewPromptTemplateApi().DelPrompt(id)
}
