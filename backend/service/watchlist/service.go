package watchlist

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"

	"github.com/duke-git/lancet/v2/convertor"
)

type Store interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string)
	GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock
	ListFollowedStocks(ctx context.Context) []data.FollowedStock
	ListFollowedStocksByGroup(ctx context.Context, groupID int) []data.FollowedStock
	Follow(ctx context.Context, stockCode string) string
	Unfollow(ctx context.Context, stockCode string) string
	ListGroups(ctx context.Context) []data.Group
	AddGroup(ctx context.Context, group data.Group) bool
	UpdateGroupSort(ctx context.Context, id int, newSort int) bool
	InitializeGroupSort(ctx context.Context) bool
	ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock
	AddGroupStock(ctx context.Context, groupID int, stockCode string) bool
	RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) bool
	RemoveGroup(ctx context.Context, groupID int) bool
	SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string
	SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string
	SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string
	SetStockSort(ctx context.Context, stockCode string, sort int64)
	GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice
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
	store             Store
	analyzer          Analyzer
	alertMu           sync.Mutex
	alertLastSent     map[string]time.Time
	priceAtAlertReset map[string]float64
}

func NewService(store Store, analyzer Analyzer) *Service {
	if store == nil {
		panic("watchlist: store dependency is required")
	}
	return &Service{
		store:             store,
		analyzer:          analyzer,
		alertLastSent:     make(map[string]time.Time),
		priceAtAlertReset: make(map[string]float64),
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
	if s.analyzer == nil {
		err := contractservice.NewUserVisibleError("watchlist.analyzer_unavailable", "AI分析服务不可用", false, contractservice.StageService)
		return ScheduledStock{}, &err
	}

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

func (s *Service) Follow(ctx context.Context, stockCode string) string {
	return s.store.Follow(ctx, NormalizeStockCode(stockCode))
}

func (s *Service) Unfollow(ctx context.Context, stockCode string) string {
	return s.store.Unfollow(ctx, NormalizeStockCode(stockCode))
}

func (s *Service) GetFollowList(ctx context.Context, groupID int) []data.FollowedStock {
	if groupID <= 0 {
		return s.store.ListFollowedStocks(ctx)
	}
	return s.store.ListFollowedStocksByGroup(ctx, groupID)
}

func (s *Service) SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string {
	return s.store.SetCostPriceAndVolume(ctx, NormalizeStockCode(stockCode), price, volume)
}

func (s *Service) SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	return s.store.SetTradingPrice(ctx, NormalizeStockCode(stockCode), entryPrice, takeProfitPrice, stopLossPrice, costPrice)
}

func (s *Service) SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string {
	return s.store.SetAlarmChangePercent(ctx, NormalizeStockCode(stockCode), val, alarmPrice)
}

func (s *Service) SetStockSort(ctx context.Context, stockCode string, sort int64) {
	s.store.SetStockSort(ctx, NormalizeStockCode(stockCode), sort)
}

func (s *Service) ListGroups(ctx context.Context) []data.Group {
	return s.store.ListGroups(ctx)
}

func (s *Service) AddGroup(ctx context.Context, group data.Group) string {
	if s.store.AddGroup(ctx, group) {
		return "添加成功"
	}
	return "添加失败"
}

func (s *Service) UpdateGroupSort(ctx context.Context, id int, newSort int) bool {
	return s.store.UpdateGroupSort(ctx, id, newSort)
}

func (s *Service) InitializeGroupSort(ctx context.Context) bool {
	return s.store.InitializeGroupSort(ctx)
}

func (s *Service) ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock {
	return s.store.ListGroupStocks(ctx, groupID)
}

func (s *Service) AddGroupStock(ctx context.Context, groupID int, stockCode string) string {
	if s.store.AddGroupStock(ctx, groupID, NormalizeStockCode(stockCode)) {
		return "添加成功"
	}
	return "添加失败"
}

func (s *Service) RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) string {
	if s.store.RemoveGroupStock(ctx, NormalizeStockCode(stockCode), name, groupID) {
		return "移除成功"
	}
	return "移除失败"
}

func (s *Service) RemoveGroup(ctx context.Context, groupID int) string {
	if s.store.RemoveGroup(ctx, groupID) {
		return "移除成功"
	}
	return "移除失败"
}

func (s *Service) EvaluateCostAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	_ = now

	follows := s.store.ListFollowedStocks(ctx)
	costFollows := make([]data.FollowedStock, 0, len(follows))
	stockCodes := make([]string, 0, len(follows))
	for _, follow := range follows {
		if strings.TrimSpace(follow.StockCode) == "" || follow.CostPrice <= 0 {
			continue
		}
		costFollows = append(costFollows, follow)
		stockCodes = append(stockCodes, NormalizeStockCode(follow.StockCode))
	}
	if len(costFollows) == 0 {
		return []notificationservice.Delivery{}
	}

	quotes := s.store.GetRealtimePrices(ctx, stockCodes...)
	priceByCode := make(map[string]float64, len(quotes))
	for _, quote := range quotes {
		price, err := convertor.ToFloat(quote.Price)
		if err != nil || price <= 0 {
			continue
		}
		priceByCode[NormalizeStockCode(quote.StockCode)] = price
	}

	deliveries := make([]notificationservice.Delivery, 0, len(costFollows))
	for _, follow := range costFollows {
		currentPrice, ok := priceByCode[NormalizeStockCode(follow.StockCode)]
		if !ok {
			continue
		}
		alertKey := fmt.Sprintf("COST:%s:%s", follow.StockCode, follow.Time.Format("20060102"))
		if currentPrice >= follow.CostPrice {
			priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
			if priceSinceLastAlert == 0 || currentPrice < priceSinceLastAlert {
				s.updatePriceAtAlertReset(alertKey, currentPrice)
			}
			continue
		}

		priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
		if priceSinceLastAlert != 0 && priceSinceLastAlert < follow.CostPrice {
			s.updatePriceAtAlertReset(alertKey, currentPrice)
			continue
		}
		if !s.canSendAlert(alertKey, 5*time.Minute) {
			continue
		}

		dropPercent := ((follow.CostPrice - currentPrice) / follow.CostPrice) * 100
		message := fmt.Sprintf(
			"【成本价预警】%s(%s)\n当前价格: %.2f\n成本价: %.2f\n亏损: %.2f%%",
			follow.Name,
			follow.StockCode,
			currentPrice,
			follow.CostPrice,
			dropPercent,
		)
		s.updateAlertSentTime(alertKey)
		s.updatePriceAtAlertReset(alertKey, currentPrice)
		deliveries = append(deliveries, buildTypedDelivery(message, follow.StockCode, 3))
	}
	return deliveries
}

func (s *Service) canSendAlert(alertKey string, interval time.Duration) bool {
	s.alertMu.Lock()
	defer s.alertMu.Unlock()

	lastSent, exists := s.alertLastSent[alertKey]
	if !exists {
		return true
	}
	return time.Since(lastSent) >= interval
}

func (s *Service) updateAlertSentTime(alertKey string) {
	s.alertMu.Lock()
	defer s.alertMu.Unlock()
	s.alertLastSent[alertKey] = time.Now()
}

func (s *Service) getPriceAtAlertReset(alertKey string) float64 {
	s.alertMu.Lock()
	defer s.alertMu.Unlock()
	return s.priceAtAlertReset[alertKey]
}

func (s *Service) updatePriceAtAlertReset(alertKey string, price float64) {
	s.alertMu.Lock()
	defer s.alertMu.Unlock()
	s.priceAtAlertReset[alertKey] = price
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func buildTypedDelivery(message, stockCode string, msgType int) notificationservice.Delivery {
	return notificationservice.Delivery{
		DingResult:   strings.TrimSpace(message),
		EventTitle:   strings.TrimSpace(stockCode),
		EventContent: strconv.Itoa(msgType),
	}
}
