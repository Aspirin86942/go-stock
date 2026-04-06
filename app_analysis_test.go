package main

import (
	"context"
	"testing"

	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
)

type fakeAnalysisService struct {
	latest     *models.AIResponseResult
	page       *models.AIResponseResultPageData
	prompts    *[]models.PromptTemplate
	promptPage *models.PromptTemplatePageData

	savedStockCode string
	savedStockName string
	savedContent   string
	savedChatID    string
	savedQuestion  string
	savedAIConfig  int

	deletedResultID uint
	batchedIDs      []uint
	savedTemplate   models.PromptTemplate
	deletedPromptID uint
}

func (f *fakeAnalysisService) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (f *fakeAnalysisService) StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (f *fakeAnalysisService) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.savedStockCode = stockCode
	f.savedStockName = stockName
	f.savedContent = result
	f.savedChatID = chatID
	f.savedQuestion = question
	f.savedAIConfig = aiConfigID
}

func (f *fakeAnalysisService) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return f.latest
}

func (f *fakeAnalysisService) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return f.page, nil
}

func (f *fakeAnalysisService) DeleteResult(ctx context.Context, id uint) error {
	f.deletedResultID = id
	return nil
}

func (f *fakeAnalysisService) BatchDeleteResults(ctx context.Context, ids []uint) error {
	f.batchedIDs = append([]uint(nil), ids...)
	return nil
}

func (f *fakeAnalysisService) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.prompts
}

func (f *fakeAnalysisService) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return f.promptPage, nil
}

func (f *fakeAnalysisService) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTemplate = template
	return "模板已保存"
}

func (f *fakeAnalysisService) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedPromptID = id
	return "删除成功"
}

func TestApp_AnalysisReadMethodsDelegateToService(t *testing.T) {
	prompts := []models.PromptTemplate{{ID: 3, Name: "系统模板", Type: "模型系统Prompt"}}
	app := NewApp()
	fake := &fakeAnalysisService{
		latest: &models.AIResponseResult{
			StockCode: "000001.SZ",
			StockName: "平安银行",
			Content:   "分析完成",
		},
		page: &models.AIResponseResultPageData{
			List:       []models.AIResponseResult{{StockCode: "000001.SZ"}},
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	config := &fakeConfigService{
		promptsResult: &prompts,
		promptPage: &models.PromptTemplatePageData{
			List:       prompts,
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	app.analysisService = fake
	app.configService = config

	app.SaveAIResponseResult("000001.SZ", "平安银行", "新的分析", "chat-1", "怎么看", 12)
	if fake.savedStockCode != "000001.SZ" || fake.savedAIConfig != 12 {
		t.Fatalf("unexpected save args: %#v", fake)
	}

	if got := app.GetAIResponseResult("000001.SZ"); got.StockCode != "000001.SZ" {
		t.Fatalf("unexpected latest result: %#v", got)
	}

	if got := app.GetAIResponseResultList(models.AIResponseResultQuery{Page: 1, PageSize: 10}); got.Total != 1 {
		t.Fatalf("unexpected result page: %#v", got)
	}

	if msg := app.DeleteAIResponseResult(9); msg != "删除成功" || fake.deletedResultID != 9 {
		t.Fatalf("unexpected delete result state: msg=%q id=%d", msg, fake.deletedResultID)
	}

	if msg := app.BatchDeleteAIResponseResult([]uint{1, 2}); msg != "删除成功" || len(fake.batchedIDs) != 2 {
		t.Fatalf("unexpected batch delete state: msg=%q ids=%v", msg, fake.batchedIDs)
	}

	if got := app.GetPromptTemplates("", ""); len(*got) != 1 || (*got)[0].Name != "系统模板" {
		t.Fatalf("unexpected prompt templates: %#v", got)
	}

	if got := app.GetPromptTemplateList(models.PromptTemplateQuery{Page: 1, PageSize: 10}); got.Total != 1 {
		t.Fatalf("unexpected prompt page: %#v", got)
	}

	if msg := app.AddPromptTemplate(models.PromptTemplate{Name: "系统模板", Type: "模型系统Prompt", Content: "请先分析风险"}); msg != "模板已保存" {
		t.Fatalf("unexpected add prompt template message: %q", msg)
	}

	if msg := app.UpdatePromptTemplate(models.PromptTemplate{ID: 3, Name: "系统模板", Type: "模型系统Prompt", Content: "请更新"}); msg != "模板已保存" {
		t.Fatalf("unexpected update prompt template message: %q", msg)
	}

	if msg := app.DeletePromptTemplate(3); msg != "模板已删除" || config.lastDeletedTemplateID != 3 {
		t.Fatalf("unexpected delete prompt template state: msg=%q id=%d", msg, config.lastDeletedTemplateID)
	}

	if msg := app.AddPrompt(models.Prompt{Name: "用户模板", Type: "模型用户Prompt", Content: "请总结"}); msg != "旧版Prompt已保存" {
		t.Fatalf("unexpected add prompt message: %q", msg)
	}

	if msg := app.DelPrompt(11); msg != "旧版Prompt已删除" || config.lastDeletedLegacyPrompt != 11 {
		t.Fatalf("unexpected del prompt state: msg=%q id=%d", msg, config.lastDeletedLegacyPrompt)
	}
}
