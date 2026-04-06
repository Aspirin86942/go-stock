package config

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

type DataStore struct{}

func NewStore() *DataStore {
	return &DataStore{}
}

func (s *DataStore) GetConfig() *data.SettingConfig {
	return data.GetSettingConfig()
}

func (s *DataStore) UpdateConfig(cfg *data.SettingConfig) string {
	return data.UpdateConfig(cfg)
}

func (s *DataStore) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	return data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
}

func (s *DataStore) GetPromptTemplatePage(query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return data.NewPromptTemplateApi().GetPromptTemplateList(&query)
}

func (s *DataStore) SavePromptTemplate(template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (s *DataStore) DeletePromptTemplate(id uint) string {
	return data.NewPromptTemplateApi().DelPrompt(id)
}
