I'm using the writing-plans skill to create the implementation plan.

# Analysis Service Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the `backend/service/analysis` package so the analysis flow has explicit contracts and the provided tests pass.

**Architecture:** The service wraps three clearly scoped contracts (stream, result, prompt) and delegates every method with nil guards, default responses, and helper utilities to parse history and normalize stream events.

**Tech Stack:** Go (backend models, fmt/encoding/json), standard Go test tooling.

---

### Task 1: Analysis service boundary

**Files:**
- Create: `backend/service/analysis/contracts.go`
- Create: `backend/service/analysis/service.go`
- Create: `backend/service/analysis/service_test.go`

- [ ] **Step 1: Write the failing tests**

```go
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
```

- [ ] **Step 2: Run the analysis service tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis -run "TestService_ReadAndPromptMethodsDelegateToStores|TestService_StartStreamsDelegateAndParseHistory" -count=1
```

Expected:

- command exits non-zero
- failure mentions missing package/files under `backend/service/analysis`

- [ ] **Step 3: Write the analysis contracts and minimal service implementation**

```go
package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"go-stock/backend/models"
)

type StockRequest struct {
	StockName   string `json:"stockName"`
	StockCode   string `json:"stockCode"`
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
}

type MarketSummaryRequest struct {
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
	HistoryJSON string `json:"historyJson"`
}

type StreamChunk struct {
	ChatID       string `json:"chatId"`
	Question     string `json:"question"`
	Content      string `json:"content"`
	ExtraContent string `json:"extraContent"`
	Model        string `json:"model"`
	Time         string `json:"time"`
}

type StreamSource interface {
	StockStream(ctx context.Context, request StockRequest) <-chan map[string]any
	MarketSummaryStream(ctx context.Context, request MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any
}

type ResultSource interface {
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
	GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult
	GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error)
	DeleteResult(ctx context.Context, id uint) error
	BatchDeleteResults(ctx context.Context, ids []uint) error
}

type PromptSource interface {
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
}

type Service struct {
	streams StreamSource
	results ResultSource
	prompts PromptSource
}

func NewService(streams StreamSource, results ResultSource, prompts PromptSource) *Service {
	return &Service{
		streams: streams,
		results: results,
		prompts: prompts,
	}
}

func (s *Service) StartStockAnalysis(ctx context.Context, request StockRequest) <-chan StreamChunk {
	if s.streams == nil {
		return closedChunks()
	}
	return mapChunks(s.streams.StockStream(ctx, request))
}

func (s *Service) StartMarketSummary(ctx context.Context, request MarketSummaryRequest) <-chan StreamChunk {
	if s.streams == nil {
		return closedChunks()
	}
	return mapChunks(s.streams.MarketSummaryStream(ctx, request, parseHistory(request.HistoryJSON)))
}

func (s *Service) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	if s.results == nil {
		return
	}
	s.results.SaveResult(ctx, stockCode, stockName, result, chatID, question, aiConfigID)
}

func (s *Service) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	if s.results == nil {
		return &models.AIResponseResult{}
	}
	result := s.results.GetLatestResult(ctx, stockCode)
	if result == nil {
		return &models.AIResponseResult{}
	}
	return result
}

func (s *Service) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	if s.results == nil {
		return &models.AIResponseResultPageData{}, nil
	}
	page, err := s.results.GetResultPage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.AIResponseResultPageData{}, nil
	}
	return page, nil
}

func (s *Service) DeleteResult(ctx context.Context, id uint) error {
	if s.results == nil {
		return nil
	}
	return s.results.DeleteResult(ctx, id)
}

func (s *Service) BatchDeleteResults(ctx context.Context, ids []uint) error {
	if s.results == nil {
		return nil
	}
	return s.results.BatchDeleteResults(ctx, ids)
}

func (s *Service) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	if s.prompts == nil {
		return &empty
	}
	result := s.prompts.GetPromptTemplates(ctx, name, promptType)
	if result == nil {
		return &empty
	}
	return result
}

func (s *Service) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	if s.prompts == nil {
		return &models.PromptTemplatePageData{}, nil
	}
	page, err := s.prompts.GetPromptTemplatePage(ctx, query)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return &models.PromptTemplatePageData{}, nil
	}
	return page, nil
}

func (s *Service) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	if s.prompts == nil {
		return ""
	}
	return s.prompts.SavePromptTemplate(ctx, template)
}

func (s *Service) DeletePromptTemplate(ctx context.Context, id uint) string {
	if s.prompts == nil {
		return ""
	}
	return s.prompts.DeletePromptTemplate(ctx, id)
}

func parseHistory(raw string) []map[string]interface{} {
	if raw == "" {
		return nil
	}
	var items []models.AiAssistantMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		entry := map[string]interface{}{
			"role":    item.Role,
			"content": item.Content,
		}
		if item.Role == "assistant" && item.Reasoning != "" {
			entry["reasoning_content"] = item.Reasoning
		}
		result = append(result, entry)
	}
	return result
}

func mapChunks(raw <-chan map[string]any) <-chan StreamChunk {
	out := make(chan StreamChunk, 128)
	go func() {
		defer close(out)
		for item := range raw {
			out <- StreamChunk{
				ChatID:       asString(item["chatId"]),
				Question:     asString(item["question"]),
				Content:      asString(item["content"]),
				ExtraContent: asString(item["extraContent"]),
				Model:        asString(item["model"]),
				Time:         asString(item["time"]),
			}
		}
	}()
	return out
}

func closedChunks() <-chan StreamChunk {
	ch := make(chan StreamChunk)
	close(ch)
	return ch
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
```

- [ ] **Step 4: Run the analysis service tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/analysis -run "TestService_ReadAndPromptMethodsDelegateToStores|TestService_StartStreamsDelegateAndParseHistory" -count=1
```

Expected:

- command exits `0`
- both analysis service tests pass

- [ ] **Step 5: Commit**

```powershell
git add backend/service/analysis/contracts.go backend/service/analysis/service.go backend/service/analysis/service_test.go
git commit -m "refactor: add analysis service boundary"
```
