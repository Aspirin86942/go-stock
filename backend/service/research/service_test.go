package research

import (
	"context"
	"errors"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
)

type fakeResearchStore struct {
	aiRecommendPage *models.AiRecommendStocksPageData
	aiRecommendErr  error
	addedStockInfo  models.AllStockInfo
	addStockInfoErr error
	deletedStockID  uint
	deleteStockErr  error
	batchDeleteIDs  []uint
	batchDeleteErr  error
}

func (f *fakeResearchStore) GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) (*models.AiRecommendStocksPageData, error) {
	return f.aiRecommendPage, f.aiRecommendErr
}

func (f *fakeResearchStore) GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	return &data.StockChangesResponse{}
}

func (f *fakeResearchStore) GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse {
	return &data.StockChangesResponse{}
}

func (f *fakeResearchStore) SaveStockChangesWithDedup(ctx context.Context, items []data.StockChangeItem) (int, error) {
	return 0, nil
}

func (f *fakeResearchStore) GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) (*models.StockChangeHistoryPageData, error) {
	return &models.StockChangeHistoryPageData{}, nil
}

func (f *fakeResearchStore) SaveStockChangesToHistory(ctx context.Context, items []data.StockChangeItem) error {
	return nil
}

func (f *fakeResearchStore) DeleteStockChangeHistory(ctx context.Context, days int) error {
	return nil
}

func (f *fakeResearchStore) DeleteAiRecommend(ctx context.Context, id uint) error {
	return nil
}

func (f *fakeResearchStore) SetAiRecommendAlert(ctx context.Context, id uint, enable bool) error {
	return nil
}

func (f *fakeResearchStore) ListAiRecommendAlertStocks(ctx context.Context) []models.AiRecommendStocks {
	return []models.AiRecommendStocks{}
}

func (f *fakeResearchStore) GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice {
	return []marketservice.RealtimePrice{}
}

func (f *fakeResearchStore) GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) (*data.AllStockInfoPageData, error) {
	return &data.AllStockInfoPageData{}, nil
}

func (f *fakeResearchStore) GetAllStockInfoByID(ctx context.Context, id uint) (*models.AllStockInfo, error) {
	return &models.AllStockInfo{}, nil
}

func (f *fakeResearchStore) AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) error {
	f.addedStockInfo = stock
	return f.addStockInfoErr
}

func (f *fakeResearchStore) DeleteAllStockInfo(ctx context.Context, id uint) error {
	f.deletedStockID = id
	return f.deleteStockErr
}

func (f *fakeResearchStore) BatchDeleteAllStockInfo(ctx context.Context, ids []uint) error {
	f.batchDeleteIDs = append([]uint(nil), ids...)
	return f.batchDeleteErr
}

func (f *fakeResearchStore) GetAllMarkets(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (f *fakeResearchStore) GetAllIndustries(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (f *fakeResearchStore) GetAllConcepts(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (f *fakeResearchStore) GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) (*data.TradingRecordPageData, error) {
	return &data.TradingRecordPageData{}, nil
}

func (f *fakeResearchStore) AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error) {
	return 0, nil
}

func (f *fakeResearchStore) GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error) {
	return &data.TradingRecord{}, nil
}

func (f *fakeResearchStore) GetTradingRecordStatistics(ctx context.Context) (*data.TradingRecordStatistics, error) {
	return &data.TradingRecordStatistics{}, nil
}

func (f *fakeResearchStore) UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error {
	return nil
}

func (f *fakeResearchStore) DeleteTradingRecord(ctx context.Context, id uint) error {
	return nil
}

func (f *fakeResearchStore) CheckFrequentTrading(ctx context.Context, stockCode string) (bool, string) {
	return true, ""
}

func TestResearchService_GetAiRecommendPageReturnsEmptyPageOnStoreError(t *testing.T) {
	svc := NewService(&fakeResearchStore{aiRecommendErr: errors.New("boom")})
	page := svc.GetAiRecommendPage(context.Background(), models.AiRecommendStocksQuery{})
	if page == nil || page.Total != 0 {
		t.Fatalf("expected empty page fallback")
	}
}

func TestResearchService_AllStockInfoMutationsDelegateToStore(t *testing.T) {
	store := &fakeResearchStore{}
	svc := NewService(store)

	added := models.AllStockInfo{SECUCODE: "000001.SZ"}
	if got := svc.AddAllStockInfo(context.Background(), added); got != "操作成功" {
		t.Fatalf("expected add success message, got %q", got)
	}
	if store.addedStockInfo.SECUCODE != "000001.SZ" {
		t.Fatalf("expected stock info add delegation, got %#v", store.addedStockInfo)
	}

	if got := svc.DeleteAllStockInfo(context.Background(), 7); got != "删除成功" {
		t.Fatalf("expected delete success message, got %q", got)
	}
	if store.deletedStockID != 7 {
		t.Fatalf("expected delete delegation, got %d", store.deletedStockID)
	}

	if got := svc.BatchDeleteAllStockInfo(context.Background(), []uint{1, 2, 3}); got != "批量删除成功" {
		t.Fatalf("expected batch delete success message, got %q", got)
	}
	if len(store.batchDeleteIDs) != 3 || store.batchDeleteIDs[2] != 3 {
		t.Fatalf("expected batch delete delegation, got %#v", store.batchDeleteIDs)
	}
}
