package watchlist

import (
	"context"
	"strings"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
)

type Store interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string)
	GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock
	ListFollowedStocks(ctx context.Context) []data.FollowedStock
}

type Analyzer interface {
	StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
}

type ScheduledStock struct {
	StockCode  string `json:"stockCode"`
	Name       string `json:"name"`
	Cron       string `json:"cron"`
	AIConfigID int    `json:"aiConfigId"`
}

type Service struct {
	store    Store
	analyzer Analyzer
}

func NewService(store Store, analyzer Analyzer) *Service {
	if store == nil {
		panic("watchlist: store dependency is required")
	}
	if analyzer == nil {
		panic("watchlist: analyzer dependency is required")
	}
	return &Service{
		store:    store,
		analyzer: analyzer,
	}
}

func NormalizeStockCode(stockCode string) string {
	code := strings.TrimSpace(stockCode)
	upper := strings.ToUpper(code)
	if strings.HasPrefix(upper, "GB_") {
		return strings.ToLower(strings.Replace(upper, "GB_", "us", 1))
	}
	return strings.ToLower(code)
}

func (s *Service) SaveStockAICron(ctx context.Context, cronText, stockCode string) (ScheduledStock, *contractservice.UserVisibleError) {
	normalized := NormalizeStockCode(stockCode)
	s.store.SaveStockAICron(ctx, cronText, normalized)

	follow := s.store.GetFollowedStock(ctx, normalized)
	if strings.TrimSpace(follow.StockCode) == "" {
		err := contractservice.NewUserVisibleError("watchlist.stock_not_followed", "股票未关注", false, contractservice.StageService)
		return ScheduledStock{}, &err
	}
	return ScheduledStock{
		StockCode:  follow.StockCode,
		Name:       follow.Name,
		Cron:       cronText,
		AIConfigID: follow.AiConfigId,
	}, nil
}

func (s *Service) ListScheduledStocks(ctx context.Context) []ScheduledStock {
	follows := s.store.ListFollowedStocks(ctx)
	result := make([]ScheduledStock, 0, len(follows))
	for _, follow := range follows {
		if follow.Cron == nil || strings.TrimSpace(*follow.Cron) == "" {
			continue
		}
		result = append(result, ScheduledStock{
			StockCode:  follow.StockCode,
			Name:       follow.Name,
			Cron:       *follow.Cron,
			AIConfigID: follow.AiConfigId,
		})
	}
	return result
}

func (s *Service) RunScheduledAnalysis(ctx context.Context, stockCode string) (ScheduledStock, *contractservice.UserVisibleError) {
	follow := s.store.GetFollowedStock(ctx, NormalizeStockCode(stockCode))
	if strings.TrimSpace(follow.StockCode) == "" {
		err := contractservice.NewUserVisibleError("watchlist.stock_not_followed", "股票未关注", false, contractservice.StageService)
		return ScheduledStock{}, &err
	}

	stream := s.analyzer.StartStockAnalysis(ctx, analysisservice.StockRequest{
		StockName:   follow.Name,
		StockCode:   follow.StockCode,
		Question:    "",
		AIConfigID:  follow.AiConfigId,
		EnableTools: true,
		Think:       true,
	})

	var content strings.Builder
	chatID := ""
	question := ""
	for chunk := range stream {
		if chunk.ExtraContent != "" {
			content.WriteString(chunk.ExtraContent)
			content.WriteString("\n")
		}
		if chunk.Content != "" {
			content.WriteString(chunk.Content)
		}
		if chunk.ChatID != "" {
			chatID = chunk.ChatID
		}
		if chunk.Question != "" {
			question = chunk.Question
		}
	}

	s.analyzer.SaveResult(ctx, follow.StockCode, follow.Name, content.String(), chatID, question, follow.AiConfigId)
	return ScheduledStock{
		StockCode:  follow.StockCode,
		Name:       follow.Name,
		Cron:       valueOrEmpty(follow.Cron),
		AIConfigID: follow.AiConfigId,
	}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
