package analysis

import (
	"context"
	"encoding/json"
	"fmt"

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

func NewService(streams StreamSource, results ResultSource, prompts PromptSource) *Service {
	return &Service{
		streams: streams,
		results: results,
		prompts: prompts,
	}
}

// StartStockAnalysis 委托 StreamSource 开始股票分析流并归一化事件。
func (s *Service) StartStockAnalysis(ctx context.Context, request StockRequest) <-chan StreamChunk {
	if s.streams == nil {
		return closedChunks()
	}
	return mapChunks(s.streams.StockStream(ctx, request))
}

// StartMarketSummary 解析历史后委托 StreamSource 推送市场总结流。
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

// parseHistory 把 JSON 字符串解码成用于回放的历史条目。
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

// mapChunks 把原始 map 事件流转换成 StreamChunk。
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

// closedChunks 生成一个立即关闭的空事件流。
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
