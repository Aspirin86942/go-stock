package main

import (
	"context"
	"testing"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/models"
	contractservice "go-stock/backend/service/contract"
	notificationservice "go-stock/backend/service/notification"
	researchservice "go-stock/backend/service/research"
	watchlistservice "go-stock/backend/service/watchlist"
)

type stubWatchlistService struct {
	groups         []data.Group
	follows        []data.FollowedStock
	lastFollowCode string
}

func (s *stubWatchlistService) SaveStockAICron(ctx context.Context, cronText, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	return watchlistservice.ScheduledStock{}, nil
}

func (s *stubWatchlistService) ListScheduledStocks(ctx context.Context) []watchlistservice.ScheduledStock {
	return []watchlistservice.ScheduledStock{}
}

func (s *stubWatchlistService) RunScheduledAnalysis(ctx context.Context, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	return watchlistservice.ScheduledStock{}, nil
}

func (s *stubWatchlistService) ListGroups(ctx context.Context) []data.Group {
	return append([]data.Group(nil), s.groups...)
}

func (s *stubWatchlistService) Follow(ctx context.Context, stockCode string) string {
	s.lastFollowCode = stockCode
	return "关注成功"
}

func (s *stubWatchlistService) Unfollow(ctx context.Context, stockCode string) string {
	s.lastFollowCode = stockCode
	return "取消关注成功"
}

func (s *stubWatchlistService) GetFollowList(ctx context.Context, groupID int) []data.FollowedStock {
	return append([]data.FollowedStock(nil), s.follows...)
}

func (s *stubWatchlistService) SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string {
	return "设置成功"
}

func (s *stubWatchlistService) SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	s.lastFollowCode = stockCode
	return "设置成功"
}

func (s *stubWatchlistService) SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string {
	return "设置成功"
}

func (s *stubWatchlistService) SetStockSort(ctx context.Context, stockCode string, sort int64) {}

func (s *stubWatchlistService) AddGroup(ctx context.Context, group data.Group) string {
	return "添加成功"
}

func (s *stubWatchlistService) UpdateGroupSort(ctx context.Context, id int, newSort int) bool {
	return true
}

func (s *stubWatchlistService) InitializeGroupSort(ctx context.Context) bool {
	return true
}

func (s *stubWatchlistService) ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock {
	return []data.GroupStock{{StockCode: "sz000001"}}
}

func (s *stubWatchlistService) AddGroupStock(ctx context.Context, groupID int, stockCode string) string {
	return "添加成功"
}

func (s *stubWatchlistService) RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) string {
	return "移除成功"
}

func (s *stubWatchlistService) RemoveGroup(ctx context.Context, groupID int) string {
	return "移除成功"
}

func (s *stubWatchlistService) EvaluateCostAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	return []notificationservice.Delivery{}
}

type stubResearchService struct {
	markets            []string
	page               *models.AiRecommendStocksPageData
	recordID           uint
	addedAllStockInfo  models.AllStockInfo
	deletedAllStockID  uint
	batchDeletedStockIDs []uint
}

func (s *stubResearchService) GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	return &data.StockChangesResponse{Data: []data.StockChangeItem{{Code: "sz000001"}}}
}

func (s *stubResearchService) GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse {
	return &data.StockChangesResponse{Data: []data.StockChangeItem{{Code: "sz000001"}}}
}

func (s *stubResearchService) GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	return &models.StockChangeHistoryPageData{}
}

func (s *stubResearchService) SaveStockChangesToHistory(ctx context.Context, changeTypes []int) string {
	return ""
}

func (s *stubResearchService) DeleteStockChangeHistory(ctx context.Context, days int) string {
	return ""
}

func (s *stubResearchService) GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	if s.page != nil {
		return s.page
	}
	return &models.AiRecommendStocksPageData{List: []models.AiRecommendStocks{}}
}

func (s *stubResearchService) DeleteAiRecommend(ctx context.Context, id uint) string {
	return ""
}

func (s *stubResearchService) SetAiRecommendAlert(ctx context.Context, id uint, enable bool) string {
	return ""
}

func (s *stubResearchService) GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	return &data.AllStockInfoPageData{}
}

func (s *stubResearchService) GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) *data.TradingRecordPageData {
	return &data.TradingRecordPageData{}
}

func (s *stubResearchService) AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error) {
	return s.recordID, nil
}

func (s *stubResearchService) GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error) {
	return &data.TradingRecord{ID: id}, nil
}

func (s *stubResearchService) GetTradingRecordStatistics(ctx context.Context) *data.TradingRecordStatistics {
	return &data.TradingRecordStatistics{StockCount: 1}
}

func (s *stubResearchService) UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error {
	return nil
}

func (s *stubResearchService) DeleteTradingRecord(ctx context.Context, id uint) error {
	return nil
}

func (s *stubResearchService) CheckFrequentTrading(ctx context.Context, stockCode string) researchservice.FrequentTradingCheck {
	return researchservice.FrequentTradingCheck{CanTrade: true, Message: "ok"}
}

func (s *stubResearchService) EvaluateAiRecommendAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	return []notificationservice.Delivery{}
}

func (s *stubResearchService) GetAllStockInfoByID(ctx context.Context, id uint) *models.AllStockInfo {
	return &models.AllStockInfo{SECUCODE: "000001.SZ"}
}

func (s *stubResearchService) AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) string {
	s.addedAllStockInfo = stock
	return "操作成功"
}

func (s *stubResearchService) DeleteAllStockInfo(ctx context.Context, id uint) string {
	s.deletedAllStockID = id
	return "删除成功"
}

func (s *stubResearchService) BatchDeleteAllStockInfo(ctx context.Context, ids []uint) string {
	s.batchDeletedStockIDs = append([]uint(nil), ids...)
	return "批量删除成功"
}

func (s *stubResearchService) GetAllMarkets(ctx context.Context) []string {
	return append([]string(nil), s.markets...)
}

func (s *stubResearchService) GetAllIndustries(ctx context.Context) []string {
	return []string{"银行"}
}

func (s *stubResearchService) GetAllConcepts(ctx context.Context) []string {
	return []string{"中特估"}
}

func TestApp_ResearchAndWatchlistMethodsDelegateToServices(t *testing.T) {
	watchlistStub := &stubWatchlistService{
		groups:  []data.Group{{Name: "全部"}},
		follows: []data.FollowedStock{{StockCode: "sz000001"}},
	}
	researchStub := &stubResearchService{
		markets: []string{"沪市"},
		page:    &models.AiRecommendStocksPageData{List: []models.AiRecommendStocks{{StockCode: "000001.SZ"}}},
		recordID: 9,
	}
	app := &App{
		ctx:              context.Background(),
		watchlistService: watchlistStub,
		researchService:  researchStub,
	}
	if got := app.GetGroupList(); len(got) != 1 {
		t.Fatalf("expected delegated group list")
	}
	if got := app.GetFollowList(0); len(*got) != 1 {
		t.Fatalf("expected delegated follow list")
	}
	if got := app.Follow("sz000001"); got != "关注成功" || watchlistStub.lastFollowCode != "sz000001" {
		t.Fatalf("expected delegated follow call")
	}
	if got := app.GetStockChanges([]int{8201}, 0, 20); len(got.Data) != 1 {
		t.Fatalf("expected delegated stock changes")
	}
	if got := app.GetAiRecommendStocksList(models.AiRecommendStocksQuery{}); len(got.List) != 1 {
		t.Fatalf("expected delegated ai recommend page")
	}
	if got := app.GetTradingRecordList(data.TradingRecordListQuery{}); got == nil {
		t.Fatalf("expected delegated trading record list")
	}
	if id, err := app.AddTradingRecord(data.TradingRecord{}); err != nil || id != 9 {
		t.Fatalf("expected delegated add trading record")
	}
	if got, err := app.GetTradingRecordById(3); err != nil || got.ID != 3 {
		t.Fatalf("expected delegated get trading record by id")
	}
	if got := app.GetTradingRecordStatistics(); got.StockCount != 1 {
		t.Fatalf("expected delegated trading record statistics")
	}
	if got := app.CheckFrequentTrading("sz000001"); got["canTrade"] != true || got["msg"] != "ok" {
		t.Fatalf("expected delegated frequent trading check")
	}
	if got := app.GetAllMarkets(); len(got) != 1 {
		t.Fatalf("expected delegated market list")
	}
	if got := app.AddAllStockInfo(models.AllStockInfo{SECUCODE: "000001.SZ"}); got != "操作成功" || researchStub.addedAllStockInfo.SECUCODE != "000001.SZ" {
		t.Fatalf("expected delegated add all stock info")
	}
	if got := app.DeleteAllStockInfo(12); got != "删除成功" || researchStub.deletedAllStockID != 12 {
		t.Fatalf("expected delegated delete all stock info")
	}
	if got := app.BatchDeleteAllStockInfo([]uint{3, 4}); got != "批量删除成功" || len(researchStub.batchDeletedStockIDs) != 2 {
		t.Fatalf("expected delegated batch delete all stock info")
	}
}
