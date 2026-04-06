package analysis

import (
	"context"
	"testing"

	"go-stock/backend/models"
)

type fakeStreams struct {
	stockRequests  []StockRequest
	marketRequests []MarketSummaryRequest
	marketHistory  []map[string]interface{}
	stockEvents    []map[string]any
	marketEvents   []map[string]any
}

func (f *fakeStreams) StockStream(ctx context.Context, request StockRequest) <-chan map[string]any {
	f.stockRequests = append(f.stockRequests, request)
	return toEventChannel(f.stockEvents)
}

func (f *fakeStreams) MarketSummaryStream(ctx context.Context, request MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any {
	f.marketRequests = append(f.marketRequests, request)
	f.marketHistory = append([]map[string]interface{}(nil), history...)
	return toEventChannel(f.marketEvents)
}

type fakeResults struct {
	latest *models.AIResponseResult
	page   *models.AIResponseResultPageData

	savedStockCode string
	savedStockName string
	savedResult    string
	savedChatID    string
	savedQuestion  string
	savedAIConfig  int

	deletedID uint
	batched   []uint
}

func (f *fakeResults) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.savedStockCode = stockCode
	f.savedStockName = stockName
	f.savedResult = result
	f.savedChatID = chatID
	f.savedQuestion = question
	f.savedAIConfig = aiConfigID
}

func (f *fakeResults) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return f.latest
}

func (f *fakeResults) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return f.page, nil
}

func (f *fakeResults) DeleteResult(ctx context.Context, id uint) error {
	f.deletedID = id
	return nil
}

func (f *fakeResults) BatchDeleteResults(ctx context.Context, ids []uint) error {
	f.batched = append([]uint(nil), ids...)
	return nil
}

type fakePrompts struct {
	templates     *[]models.PromptTemplate
	page          *models.PromptTemplatePageData
	savedTemplate models.PromptTemplate
	deletedID     uint
	saveResult    string
	deleteResult  string
}

func (f *fakePrompts) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	return f.templates
}

func (f *fakePrompts) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return f.page, nil
}

func (f *fakePrompts) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	f.savedTemplate = template
	return f.saveResult
}

func (f *fakePrompts) DeletePromptTemplate(ctx context.Context, id uint) string {
	f.deletedID = id
	return f.deleteResult
}

func TestService_ReadAndPromptMethodsDelegateToStores(t *testing.T) {
	templates := []models.PromptTemplate{
		{ID: 7, Name: "系统模板", Type: "模型系统Prompt", Content: "请先分析风险"},
	}

	results := &fakeResults{
		latest: &models.AIResponseResult{
			StockCode: "000001.SZ",
			StockName: "平安银行",
			Question:  "怎么看银行股",
			Content:   "分析完成",
		},
		page: &models.AIResponseResultPageData{
			List:       []models.AIResponseResult{{StockCode: "000001.SZ", Content: "page-item"}},
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		},
	}
	prompts := &fakePrompts{
		templates:    &templates,
		page:         &models.PromptTemplatePageData{List: templates, Total: 1, Page: 1, PageSize: 10, TotalPages: 1},
		saveResult:   "模板已保存",
		deleteResult: "删除成功",
	}
	svc := NewService(&fakeStreams{}, results, prompts)

	latest := svc.GetLatestResult(context.Background(), "000001.SZ")
	if latest.StockCode != "000001.SZ" || latest.Content != "分析完成" {
		t.Fatalf("unexpected latest result: %#v", latest)
	}

	page, err := svc.GetResultPage(context.Background(), models.AIResponseResultQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected page error: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 {
		t.Fatalf("unexpected result page: %#v", page)
	}

	gotTemplates := svc.GetPromptTemplates(context.Background(), "", "")
	if len(*gotTemplates) != 1 || (*gotTemplates)[0].Name != "系统模板" {
		t.Fatalf("unexpected prompt templates: %#v", gotTemplates)
	}

	promptPage, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected prompt page error: %v", err)
	}
	if promptPage.Total != 1 || len(promptPage.List) != 1 {
		t.Fatalf("unexpected prompt page: %#v", promptPage)
	}

	svc.SaveResult(context.Background(), "000001.SZ", "平安银行", "新的分析", "chat-1", "更新一下", 12)
	if results.savedStockCode != "000001.SZ" || results.savedAIConfig != 12 {
		t.Fatalf("unexpected save args: %#v", results)
	}

	if msg := svc.SavePromptTemplate(context.Background(), models.PromptTemplate{Name: "用户模板", Type: "模型用户Prompt", Content: "请总结"}); msg != "模板已保存" {
		t.Fatalf("unexpected save prompt message: %q", msg)
	}

	if err := svc.DeleteResult(context.Background(), 9); err != nil || results.deletedID != 9 {
		t.Fatalf("unexpected delete result state: err=%v id=%d", err, results.deletedID)
	}

	if err := svc.BatchDeleteResults(context.Background(), []uint{1, 2}); err != nil || len(results.batched) != 2 {
		t.Fatalf("unexpected batch delete state: err=%v ids=%v", err, results.batched)
	}

	if msg := svc.DeletePromptTemplate(context.Background(), 7); msg != "删除成功" || prompts.deletedID != 7 {
		t.Fatalf("unexpected prompt delete state: msg=%q id=%d", msg, prompts.deletedID)
	}
}

func TestService_StartStreamsDelegateAndParseHistory(t *testing.T) {
	streams := &fakeStreams{
		stockEvents: []map[string]any{
			{"chatId": "stock-1", "question": "怎么看", "content": "第一段"},
		},
		marketEvents: []map[string]any{
			{"chatId": "summary-1", "question": "总结市场", "content": "市场第一段"},
			{"extraContent": "市场第二段", "model": "deepseek-chat", "time": "2026-04-06 10:00:00"},
		},
	}
	svc := NewService(streams, &fakeResults{}, &fakePrompts{})

	stockChunks := collectChunks(svc.StartStockAnalysis(context.Background(), StockRequest{
		StockName:   "平安银行",
		StockCode:   "000001.SZ",
		Question:    "怎么看",
		AIConfigID:  3,
		EnableTools: true,
		Think:       true,
	}))
	if len(stockChunks) != 1 || stockChunks[0].ChatID != "stock-1" || stockChunks[0].Content != "第一段" {
		t.Fatalf("unexpected stock chunks: %#v", stockChunks)
	}
	if len(streams.stockRequests) != 1 || streams.stockRequests[0].StockCode != "000001.SZ" {
		t.Fatalf("unexpected stock request capture: %#v", streams.stockRequests)
	}

	marketChunks := collectChunks(svc.StartMarketSummary(context.Background(), MarketSummaryRequest{
		Question:    "总结市场",
		AIConfigID:  8,
		EnableTools: true,
		Think:       false,
		HistoryJSON: `[{"role":"assistant","content":"旧回答","reasoning":"旧推理"}]`,
	}))
	if len(marketChunks) != 2 || marketChunks[1].ExtraContent != "市场第二段" {
		t.Fatalf("unexpected market chunks: %#v", marketChunks)
	}
	if len(streams.marketHistory) != 1 {
		t.Fatalf("unexpected market history: %#v", streams.marketHistory)
	}
	if got := streams.marketHistory[0]["reasoning_content"]; got != "旧推理" {
		t.Fatalf("unexpected reasoning_content: %#v", got)
	}
}

func toEventChannel(items []map[string]any) <-chan map[string]any {
	ch := make(chan map[string]any, len(items))
	for _, item := range items {
		ch <- item
	}
	close(ch)
	return ch
}

func collectChunks(ch <-chan StreamChunk) []StreamChunk {
	var chunks []StreamChunk
	for item := range ch {
		chunks = append(chunks, item)
	}
	return chunks
}

func TestNewService_requiresDependencies(t *testing.T) {
	validStreams := &fakeStreams{}
	validResults := &fakeResults{}
	validPrompts := &fakePrompts{}

	cases := []struct {
		name    string
		streams StreamSource
		results ResultSource
		prompts PromptSource
	}{
		{name: "missingStreams", streams: nil, results: validResults, prompts: validPrompts},
		{name: "missingResults", streams: validStreams, results: nil, prompts: validPrompts},
		{name: "missingPrompts", streams: validStreams, results: validResults, prompts: nil},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for %s", tc.name)
				}
			}()
			NewService(tc.streams, tc.results, tc.prompts)
		})
	}
}

func TestService_GetResultPageHandlesNilPage(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{page: nil}, &fakePrompts{})
	page, err := svc.GetResultPage(context.Background(), models.AIResponseResultQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatalf("expected non-nil page result")
	}
	if page.Total != 0 || len(page.List) != 0 {
		t.Fatalf("expected empty page when source returns nil, got %#v", page)
	}
}

func TestService_GetPromptTemplatePageHandlesNilPage(t *testing.T) {
	svc := NewService(&fakeStreams{}, &fakeResults{}, &fakePrompts{page: nil})
	page, err := svc.GetPromptTemplatePage(context.Background(), models.PromptTemplateQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatalf("expected non-nil prompt page result")
	}
	if page.Total != 0 || len(page.List) != 0 {
		t.Fatalf("expected empty prompt page when source returns nil, got %#v", page)
	}
}
