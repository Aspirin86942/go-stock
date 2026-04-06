package watchlist

import (
	"context"
	"testing"
	"time"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"
)

type fakeStore struct {
	savedCronText string
	savedCode     string
	follow        data.FollowedStock
	list          []data.FollowedStock
	follows       []data.FollowedStock
	quotes        []marketservice.RealtimePrice

	lastTradingStockCode      string
	lastTradingEntryPrice     float64
	lastTradingTakeProfit     float64
	lastTradingStopLoss       float64
	lastTradingCostPrice      float64
	setTradingPriceReturnText string
}

func (f *fakeStore) SaveStockAICron(ctx context.Context, cronText, stockCode string) {
	f.savedCronText = cronText
	f.savedCode = stockCode
}

func (f *fakeStore) GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock {
	if f.follow.StockCode == stockCode {
		return f.follow
	}
	return data.FollowedStock{}
}

func (f *fakeStore) ListFollowedStocks(ctx context.Context) []data.FollowedStock {
	if f.follows != nil {
		return append([]data.FollowedStock(nil), f.follows...)
	}
	return append([]data.FollowedStock(nil), f.list...)
}

func (f *fakeStore) ListFollowedStocksByGroup(ctx context.Context, groupID int) []data.FollowedStock {
	return f.ListFollowedStocks(ctx)
}

func (f *fakeStore) Follow(ctx context.Context, stockCode string) string {
	return "关注成功"
}

func (f *fakeStore) Unfollow(ctx context.Context, stockCode string) string {
	return "取消关注成功"
}

func (f *fakeStore) ListGroups(ctx context.Context) []data.Group {
	return []data.Group{}
}

func (f *fakeStore) AddGroup(ctx context.Context, group data.Group) bool {
	return true
}

func (f *fakeStore) UpdateGroupSort(ctx context.Context, id int, newSort int) bool {
	return true
}

func (f *fakeStore) InitializeGroupSort(ctx context.Context) bool {
	return true
}

func (f *fakeStore) ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock {
	return []data.GroupStock{}
}

func (f *fakeStore) AddGroupStock(ctx context.Context, groupID int, stockCode string) bool {
	return true
}

func (f *fakeStore) RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) bool {
	return true
}

func (f *fakeStore) RemoveGroup(ctx context.Context, groupID int) bool {
	return true
}

func (f *fakeStore) SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string {
	return "设置成功"
}

func (f *fakeStore) SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	f.lastTradingStockCode = stockCode
	f.lastTradingEntryPrice = entryPrice
	f.lastTradingTakeProfit = takeProfitPrice
	f.lastTradingStopLoss = stopLossPrice
	f.lastTradingCostPrice = costPrice
	if f.setTradingPriceReturnText != "" {
		return f.setTradingPriceReturnText
	}
	return "设置成功"
}

func (f *fakeStore) GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice {
	return append([]marketservice.RealtimePrice(nil), f.quotes...)
}

func (f *fakeStore) SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string {
	return "设置成功"
}

func (f *fakeStore) SetStockSort(ctx context.Context, stockCode string, sort int64) {}

type fakeAnalyzer struct {
	request  analysisservice.StockRequest
	saveCode string
	saveName string
	saveBody string
	saveChat string
	saveQ    string
	saveAIID int
	stream   []analysisservice.StreamChunk
}

func (f *fakeAnalyzer) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	f.request = request
	ch := make(chan analysisservice.StreamChunk, len(f.stream))
	for _, item := range f.stream {
		ch <- item
	}
	close(ch)
	return ch
}

func (f *fakeAnalyzer) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
	f.saveCode = stockCode
	f.saveName = stockName
	f.saveBody = result
	f.saveChat = chatID
	f.saveQ = question
	f.saveAIID = aiConfigID
}

func TestService_SaveStockAICronNormalizesUsCode(t *testing.T) {
	store := &fakeStore{
		follow: data.FollowedStock{StockCode: "usaapl", Name: "Apple", AiConfigId: 7},
	}
	svc := NewService(store, &fakeAnalyzer{})

	result, userErr := svc.SaveStockAICron(context.Background(), "0 */5 * * * *", "gb_aapl")
	if userErr != nil {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
	if store.savedCode != "usaapl" {
		t.Fatalf("expected normalized stock code, got %q", store.savedCode)
	}
	if result.StockCode != "usaapl" || result.Name != "Apple" || result.Cron != "0 */5 * * * *" {
		t.Fatalf("unexpected save result: %#v", result)
	}
}

func TestService_ListScheduledStocksFiltersBlankCron(t *testing.T) {
	cronText := "0 */10 * * * *"
	store := &fakeStore{
		list: []data.FollowedStock{
			{StockCode: "sz000001", Name: "平安银行", Cron: &cronText, AiConfigId: 1},
			{StockCode: "sh600519", Name: "贵州茅台", Cron: nil, AiConfigId: 2},
		},
	}
	svc := NewService(store, &fakeAnalyzer{})

	result := svc.ListScheduledStocks(context.Background())
	if len(result) != 1 || result[0].StockCode != "sz000001" {
		t.Fatalf("unexpected scheduled stock list: %#v", result)
	}
}

func TestService_RunScheduledAnalysisSavesCollectedResult(t *testing.T) {
	store := &fakeStore{
		follow: data.FollowedStock{StockCode: "sz000001", Name: "平安银行", AiConfigId: 3},
	}
	analyzer := &fakeAnalyzer{
		stream: []analysisservice.StreamChunk{
			{ExtraContent: "第一段", ChatID: "chat-1", Question: "默认问题"},
			{Content: "第二段"},
		},
	}
	svc := NewService(store, analyzer)

	result, userErr := svc.RunScheduledAnalysis(context.Background(), "sz000001")
	if userErr != nil {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
	if result.StockCode != "sz000001" || result.Name != "平安银行" {
		t.Fatalf("unexpected run result: %#v", result)
	}
	if analyzer.saveCode != "sz000001" || analyzer.saveAIID != 3 || analyzer.saveBody != "第一段\n第二段" {
		t.Fatalf("unexpected save delegation: %#v", analyzer)
	}
}

func TestService_RunScheduledAnalysisReturnsUserVisibleErrorWhenStockMissing(t *testing.T) {
	svc := NewService(&fakeStore{}, &fakeAnalyzer{})

	_, userErr := svc.RunScheduledAnalysis(context.Background(), "gb_aapl")
	if userErr == nil {
		t.Fatal("expected user visible error for missing followed stock")
	}
	if userErr.Code != "watchlist.stock_not_followed" || userErr.Stage != contractservice.StageService {
		t.Fatalf("unexpected user visible error: %#v", userErr)
	}
}

func TestWatchlistService_SetTradingPriceDelegatesToStore(t *testing.T) {
	store := &fakeStore{
		setTradingPriceReturnText: "操作成功",
	}
	svc := NewService(store, nil)

	result := svc.SetTradingPrice(context.Background(), "sz000001", 10, 12, 9, 10.5)

	if result != "操作成功" {
		t.Fatalf("expected success message, got %q", result)
	}
	if store.lastTradingStockCode != "sz000001" {
		t.Fatalf("expected stock code to be forwarded")
	}
}

func TestWatchlistService_EvaluateCostAlertsBuildsDeliveries(t *testing.T) {
	store := &fakeStore{
		follows: []data.FollowedStock{
			{StockCode: "sz000001", Name: "平安银行", CostPrice: 12},
		},
		quotes: []marketservice.RealtimePrice{
			{StockCode: "sz000001", StockName: "平安银行", Price: "11.5"},
		},
	}
	svc := NewService(store, nil)

	deliveries := svc.EvaluateCostAlerts(context.Background(), time.Now())

	if len(deliveries) != 1 {
		t.Fatalf("expected one alert delivery, got %d", len(deliveries))
	}
}
