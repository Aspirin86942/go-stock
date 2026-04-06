package analysis

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	data.NewDeepSeekOpenAi(ctx, aiConfigID).SaveAIResponseResult(stockCode, stockName, result, chatID, question)
}

func (s *Store) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return data.NewDeepSeekOpenAi(ctx, 0).GetAIResponseResult(stockCode)
}

func (s *Store) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return data.NewAIResponseResultService().GetAIResponseResultList(query)
}

func (s *Store) DeleteResult(ctx context.Context, id uint) error {
	return data.NewAIResponseResultService().DeleteAIResponseResult(id)
}

func (s *Store) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return data.NewAIResponseResultService().BatchDeleteAIResponseResult(ids)
}

func (s *Store) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
}

func (s *Store) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return data.NewPromptTemplateApi().GetPromptTemplateList(&query)
}

func (s *Store) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (s *Store) DeletePromptTemplate(ctx context.Context, id uint) string {
	return data.NewPromptTemplateApi().DelPrompt(id)
}
