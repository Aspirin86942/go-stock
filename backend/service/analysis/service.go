package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"go-stock/backend/logger"
	"go-stock/backend/models"
)

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

// NewService enforces that all dependencies are wired to avoid silent failures.
func NewService(streams StreamSource, results ResultSource, prompts PromptSource) *Service {
	if streams == nil {
		panic("analysis: streams dependency is required")
	}
	if results == nil {
		panic("analysis: results dependency is required")
	}
	if prompts == nil {
		panic("analysis: prompts dependency is required")
	}
	return &Service{
		streams: streams,
		results: results,
		prompts: prompts,
	}
}

// StartStockAnalysis 委托 StreamSource 开始股票分析流并归一化事件。
func (s *Service) StartStockAnalysis(ctx context.Context, request StockRequest) <-chan StreamChunk {
	return mapChunks(s.streams.StockStream(ctx, request))
}

// StartMarketSummary 解析历史后委托 StreamSource 推送市场总结流。
func (s *Service) StartMarketSummary(ctx context.Context, request MarketSummaryRequest) <-chan StreamChunk {
	history, err := parseHistory(request.HistoryJSON)
	if err != nil {
		logHistoryParseError(err, request.HistoryJSON)
	}
	return mapChunks(s.streams.MarketSummaryStream(ctx, request, history))
}

func (s *Service) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	s.results.SaveResult(ctx, stockCode, stockName, result, chatID, question, aiConfigID)
}

func (s *Service) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	result := s.results.GetLatestResult(ctx, stockCode)
	if result == nil {
		return &models.AIResponseResult{}
	}
	return result
}

func (s *Service) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
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
	return s.results.DeleteResult(ctx, id)
}

func (s *Service) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return s.results.BatchDeleteResults(ctx, ids)
}

func (s *Service) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	result := s.prompts.GetPromptTemplates(ctx, name, promptType)
	if result == nil {
		return &empty
	}
	return result
}

func (s *Service) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
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
	return s.prompts.SavePromptTemplate(ctx, template)
}

func (s *Service) DeletePromptTemplate(ctx context.Context, id uint) string {
	return s.prompts.DeletePromptTemplate(ctx, id)
}

// parseHistory 把 JSON 字符串解码成用于回放的历史条目。
func parseHistory(raw string) ([]map[string]interface{}, error) {
	if raw == "" {
		return nil, nil
	}
	var items []models.AiAssistantMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
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
	return result, nil
}

// mapChunks 把原始 map 事件流转换成 StreamChunk。
func mapChunks(raw <-chan map[string]any) <-chan StreamChunk {
	if raw == nil {
		ch := make(chan StreamChunk)
		close(ch)
		return ch
	}
	out := make(chan StreamChunk, 128)
	go func() {
		defer close(out)
		for item := range raw {
			out <- StreamChunk{
				ChatID:           asString(item["chatId"]),
				Question:         asString(item["question"]),
				Content:          asString(item["content"]),
				ExtraContent:     asString(item["extraContent"]),
				Model:            asString(item["model"]),
				Time:             asString(item["time"]),
				ReasoningContent: asString(item["reasoning_content"]),
				ToolCalls:        asToolCalls(item["tool_calls"]),
			}
		}
	}()
	return out
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

func asToolCalls(value any) []map[string]any {
	if value == nil {
		return nil
	}
	if entries, ok := value.([]map[string]any); ok {
		copied := make([]map[string]any, len(entries))
		copy(copied, entries)
		return copied
	}
	rawSlice, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(rawSlice))
	for _, item := range rawSlice {
		if cast, ok := item.(map[string]any); ok {
			result = append(result, cast)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func logHistoryParseError(err error, raw string) {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return
	}
	const maxRawLength = 2048
	if len(raw) > maxRawLength {
		raw = raw[:maxRawLength]
	}
	runtimeLogger.ForSink(logger.SinkError, "analysis.service").Error(
		"analysis.market_history.parse_failed",
		"无法解析市场历史 JSON",
		logger.Err(err),
		logger.String("history_json", raw),
	)
}
