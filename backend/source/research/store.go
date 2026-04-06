package research

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	_ = ctx
	return data.NewStockChangesApi().GetStockChanges(changeTypes, pageIndex, pageSize)
}

func (s *Store) GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse {
	_ = ctx
	return data.NewStockChangesApi().GetAllStockChangesWithPaging(pageSize)
}

func (s *Store) SaveStockChangesWithDedup(ctx context.Context, items []data.StockChangeItem) (int, error) {
	_ = ctx
	return data.NewStockChangeHistoryService().SaveStockChangesWithDedup(items)
}

func (s *Store) GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) (*models.StockChangeHistoryPageData, error) {
	_ = ctx
	return data.NewStockChangeHistoryService().GetHistoryList(query)
}

func (s *Store) SaveStockChangesToHistory(ctx context.Context, items []data.StockChangeItem) error {
	_ = ctx
	return data.NewStockChangeHistoryService().SaveStockChanges(items)
}

func (s *Store) DeleteStockChangeHistory(ctx context.Context, days int) error {
	_ = ctx
	return data.NewStockChangeHistoryService().DeleteOldData(days)
}

func (s *Store) GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) (*models.AiRecommendStocksPageData, error) {
	_ = ctx
	return data.NewAiRecommendStocksService().GetAiRecommendStocksList(&query)
}

func (s *Store) DeleteAiRecommend(ctx context.Context, id uint) error {
	_ = ctx
	return data.NewAiRecommendStocksService().DeleteAiRecommendStocks(id)
}

func (s *Store) SetAiRecommendAlert(ctx context.Context, id uint, enable bool) error {
	_ = ctx
	return data.NewAiRecommendStocksService().UpdateAiRecommendStocksAlert(id, enable)
}

func (s *Store) ListAiRecommendAlertStocks(ctx context.Context) []models.AiRecommendStocks {
	_ = ctx
	list := make([]models.AiRecommendStocks, 0)
	db.Dao.Model(&models.AiRecommendStocks{}).Where("enable_alert = ?", true).Find(&list)
	return list
}

func (s *Store) GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice {
	_ = ctx
	if len(stockCodes) == 0 {
		return []marketservice.RealtimePrice{}
	}

	quotes, err := data.NewStockDataApi().GetStockCodeRealTimeData(stockCodes...)
	if err != nil || quotes == nil {
		return []marketservice.RealtimePrice{}
	}

	result := make([]marketservice.RealtimePrice, 0, len(*quotes))
	for _, item := range *quotes {
		result = append(result, marketservice.RealtimePrice{
			StockCode: item.Code,
			StockName: item.Name,
			Price:     item.Price,
			Bid:       item.Bid,
			Ask:       item.Ask,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			PreClose:  item.PreClose,
			Date:      item.Date,
			Time:      item.Time,
		})
	}
	return result
}

func (s *Store) GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) (*data.AllStockInfoPageData, error) {
	_ = ctx
	return data.NewStockDataApi().GetAllStockInfoList(&query)
}

func (s *Store) GetAllStockInfoByID(ctx context.Context, id uint) (*models.AllStockInfo, error) {
	_ = ctx
	return data.NewStockDataApi().GetAllStockInfoById(id)
}

func (s *Store) AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) error {
	_ = ctx
	return data.NewStockDataApi().AddAllStockInfo(stock)
}

func (s *Store) DeleteAllStockInfo(ctx context.Context, id uint) error {
	_ = ctx
	return data.NewStockDataApi().DeleteAllStockInfo(id)
}

func (s *Store) BatchDeleteAllStockInfo(ctx context.Context, ids []uint) error {
	_ = ctx
	return data.NewStockDataApi().BatchDeleteAllStockInfo(ids)
}

func (s *Store) GetAllMarkets(ctx context.Context) ([]string, error) {
	_ = ctx
	return data.NewStockDataApi().GetAllMarkets()
}

func (s *Store) GetAllIndustries(ctx context.Context) ([]string, error) {
	_ = ctx
	return data.NewStockDataApi().GetAllIndustries()
}

func (s *Store) GetAllConcepts(ctx context.Context) ([]string, error) {
	_ = ctx
	return data.NewStockDataApi().GetAllConcepts()
}

func (s *Store) GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) (*data.TradingRecordPageData, error) {
	_ = ctx
	return data.NewStockDataApi().GetTradingRecordList(query)
}

func (s *Store) AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error) {
	_ = ctx
	return data.NewStockDataApi().AddTradingRecord(record)
}

func (s *Store) GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error) {
	_ = ctx
	return data.NewStockDataApi().GetTradingRecordById(id)
}

func (s *Store) GetTradingRecordStatistics(ctx context.Context) (*data.TradingRecordStatistics, error) {
	_ = ctx
	return data.NewStockDataApi().GetTradingRecordStatistics()
}

func (s *Store) UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error {
	_ = ctx
	return data.NewStockDataApi().UpdateTradingRecord(record)
}

func (s *Store) DeleteTradingRecord(ctx context.Context, id uint) error {
	_ = ctx
	return data.NewStockDataApi().DeleteTradingRecord(id)
}

func (s *Store) CheckFrequentTrading(ctx context.Context, stockCode string) (bool, string) {
	_ = ctx
	return data.NewStockDataApi().CheckFrequentTrading(stockCode)
}
