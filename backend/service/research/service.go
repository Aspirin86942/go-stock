package research

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/go-resty/resty/v2"
)

type Store interface {
	GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse
	GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse
	SaveStockChangesWithDedup(ctx context.Context, items []data.StockChangeItem) (int, error)
	GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) (*models.StockChangeHistoryPageData, error)
	SaveStockChangesToHistory(ctx context.Context, items []data.StockChangeItem) error
	DeleteStockChangeHistory(ctx context.Context, days int) error
	GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) (*models.AiRecommendStocksPageData, error)
	DeleteAiRecommend(ctx context.Context, id uint) error
	SetAiRecommendAlert(ctx context.Context, id uint, enable bool) error
	ListAiRecommendAlertStocks(ctx context.Context) []models.AiRecommendStocks
	GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice
	GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) (*data.AllStockInfoPageData, error)
	GetAllStockInfoByID(ctx context.Context, id uint) (*models.AllStockInfo, error)
	AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) error
	DeleteAllStockInfo(ctx context.Context, id uint) error
	BatchDeleteAllStockInfo(ctx context.Context, ids []uint) error
	GetAllMarkets(ctx context.Context) ([]string, error)
	GetAllIndustries(ctx context.Context) ([]string, error)
	GetAllConcepts(ctx context.Context) ([]string, error)
	GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) (*data.TradingRecordPageData, error)
	AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error)
	GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error)
	GetTradingRecordStatistics(ctx context.Context) (*data.TradingRecordStatistics, error)
	UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error
	DeleteTradingRecord(ctx context.Context, id uint) error
	CheckFrequentTrading(ctx context.Context, stockCode string) (bool, string)
}

type Service struct {
	store             Store
	alertMu           sync.Mutex
	alertLastSent     map[string]time.Time
	priceAtAlertReset map[string]float64
}

func NewService(store Store) *Service {
	if store == nil {
		panic("research: store dependency is required")
	}
	return &Service{
		store:             store,
		alertLastSent:     make(map[string]time.Time),
		priceAtAlertReset: make(map[string]float64),
	}
}

func (s *Service) LoadAllStocks(ctx context.Context, page, pageSize int, name string, technicalIndicators models.TechnicalIndicators) *models.AllStocksResp {
	_ = ctx
	result := data.NewStockDataApi().GetAllStocks(page, pageSize, name, technicalIndicators)
	if result == nil {
		return &models.AllStocksResp{
			Result: struct {
				Nextpage    bool               `json:"nextpage"`
				Currentpage int                `json:"currentpage"`
				Data        []models.StockInfo `json:"data"`
				Config      []interface{}      `json:"config"`
				Count       int                `json:"count"`
			}{
				Data:   []models.StockInfo{},
				Config: []interface{}{},
			},
		}
	}
	if result.Result.Data == nil {
		result.Result.Data = []models.StockInfo{}
	}
	if result.Result.Config == nil {
		result.Result.Config = []interface{}{}
	}
	return result
}

func (s *Service) SyncAllStockInfo(ctx context.Context) error {
	_ = ctx
	db.Dao.Unscoped().Model(&models.AllStockInfo{}).Where("1 = 1").Delete(&models.AllStockInfo{})
	for page := 1; page < 3; page++ {
		result := s.LoadAllStocks(ctx, page, 3000, "", models.TechnicalIndicators{})
		datas := make([]models.AllStockInfo, 0, len(result.Result.Data))
		for _, item := range result.Result.Data {
			datas = append(datas, item.ToAllStockInfo())
		}
		if len(datas) == 0 {
			continue
		}
		if err := db.Dao.CreateInBatches(&datas, 1000).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RefreshStockBaseInfo(ctx context.Context) error {
	_ = ctx

	stockBasics := &[]data.StockBasic{}
	if _, err := resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockBasics).
		Get("http://8.134.249.145:18080/go-stock/stock_basic.json"); err != nil {
		return err
	}
	db.Dao.Unscoped().Model(&data.StockBasic{}).Where("1 = 1").Delete(&data.StockBasic{})
	if err := db.Dao.CreateInBatches(stockBasics, 400).Error; err != nil {
		return err
	}

	stockHKBasics := &[]models.StockInfoHK{}
	if _, err := resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockHKBasics).
		Get("http://8.134.249.145:18080/go-stock/stock_base_info_hk.json"); err != nil {
		return err
	}
	db.Dao.Unscoped().Model(&models.StockInfoHK{}).Where("1 = 1").Delete(&models.StockInfoHK{})
	if err := db.Dao.CreateInBatches(stockHKBasics, 400).Error; err != nil {
		return err
	}

	stockUSBasics := &[]models.StockInfoUS{}
	if _, err := resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockUSBasics).
		Get("http://8.134.249.145:18080/go-stock/stock_base_info_us.json"); err != nil {
		return err
	}
	db.Dao.Unscoped().Model(&models.StockInfoUS{}).Where("1 = 1").Delete(&models.StockInfoUS{})
	if err := db.Dao.CreateInBatches(stockUSBasics, 400).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	return normalizeStockChangesPage(s.store.GetStockChanges(ctx, changeTypes, pageIndex, pageSize))
}

func (s *Service) GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse {
	page := normalizeStockChangesPage(s.store.GetAllStockChangesWithPaging(ctx, pageSize))
	if len(page.Data) > 0 {
		_, _ = s.store.SaveStockChangesWithDedup(ctx, page.Data)
	}
	return page
}

func (s *Service) GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	page, err := s.store.GetStockChangeHistory(ctx, query)
	if err != nil {
		return emptyStockChangeHistoryPage(query)
	}
	return normalizeStockChangeHistoryPage(page, query)
}

func (s *Service) SaveStockChangesToHistory(ctx context.Context, changeTypes []int) string {
	result := normalizeStockChangesPage(s.store.GetStockChanges(ctx, changeTypes, 0, 500))
	if len(result.Data) == 0 {
		return "没有获取到异动数据"
	}
	if err := s.store.SaveStockChangesToHistory(ctx, result.Data); err != nil {
		return "保存失败: " + err.Error()
	}
	return fmt.Sprintf("成功保存 %d 条异动数据", len(result.Data))
}

func (s *Service) DeleteStockChangeHistory(ctx context.Context, days int) string {
	if err := s.store.DeleteStockChangeHistory(ctx, days); err != nil {
		return "删除失败: " + err.Error()
	}
	return fmt.Sprintf("已删除 %d 天前的历史数据", days)
}

func (s *Service) GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	page, err := s.store.GetAiRecommendPage(ctx, query)
	if err != nil {
		return emptyAiRecommendPage(query)
	}
	return normalizeAiRecommendPage(page, query)
}

func (s *Service) DeleteAiRecommend(ctx context.Context, id uint) string {
	if err := s.store.DeleteAiRecommend(ctx, id); err != nil {
		return "删除失败"
	}
	return "删除成功"
}

func (s *Service) SetAiRecommendAlert(ctx context.Context, id uint, enable bool) string {
	if err := s.store.SetAiRecommendAlert(ctx, id, enable); err != nil {
		return "更新预警状态失败"
	}
	return "更新预警状态成功"
}

func (s *Service) GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	page, err := s.store.GetAllStockInfoPage(ctx, query)
	if err != nil {
		return emptyAllStockInfoPage(query)
	}
	return normalizeAllStockInfoPage(page, query)
}

func (s *Service) GetAllStockInfoByID(ctx context.Context, id uint) *models.AllStockInfo {
	stock, err := s.store.GetAllStockInfoByID(ctx, id)
	if err != nil || stock == nil {
		return &models.AllStockInfo{}
	}
	return stock
}

func (s *Service) AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) string {
	if err := s.store.AddAllStockInfo(ctx, stock); err != nil {
		return "操作失败: " + err.Error()
	}
	return "操作成功"
}

func (s *Service) DeleteAllStockInfo(ctx context.Context, id uint) string {
	if err := s.store.DeleteAllStockInfo(ctx, id); err != nil {
		return "删除失败: " + err.Error()
	}
	return "删除成功"
}

func (s *Service) BatchDeleteAllStockInfo(ctx context.Context, ids []uint) string {
	if err := s.store.BatchDeleteAllStockInfo(ctx, ids); err != nil {
		return "批量删除失败: " + err.Error()
	}
	return "批量删除成功"
}

func (s *Service) GetAllMarkets(ctx context.Context) []string {
	values, err := s.store.GetAllMarkets(ctx)
	if err != nil || values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}

func (s *Service) GetAllIndustries(ctx context.Context) []string {
	values, err := s.store.GetAllIndustries(ctx)
	if err != nil || values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}

func (s *Service) GetAllConcepts(ctx context.Context) []string {
	values, err := s.store.GetAllConcepts(ctx)
	if err != nil || values == nil {
		return []string{}
	}
	return append([]string(nil), values...)
}

func (s *Service) GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) *data.TradingRecordPageData {
	page, err := s.store.GetTradingRecordList(ctx, query)
	if err != nil {
		return emptyTradingRecordPage(query)
	}
	return normalizeTradingRecordPage(page, query)
}

func (s *Service) AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error) {
	return s.store.AddTradingRecord(ctx, record)
}

func (s *Service) GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error) {
	record, err := s.store.GetTradingRecordByID(ctx, id)
	if err != nil || record == nil {
		return &data.TradingRecord{}, err
	}
	return record, nil
}

func (s *Service) GetTradingRecordStatistics(ctx context.Context) *data.TradingRecordStatistics {
	stats, err := s.store.GetTradingRecordStatistics(ctx)
	if err != nil || stats == nil {
		return &data.TradingRecordStatistics{}
	}
	return stats
}

func (s *Service) UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error {
	return s.store.UpdateTradingRecord(ctx, record)
}

func (s *Service) DeleteTradingRecord(ctx context.Context, id uint) error {
	return s.store.DeleteTradingRecord(ctx, id)
}

func (s *Service) CheckFrequentTrading(ctx context.Context, stockCode string) FrequentTradingCheck {
	canTrade, msg := s.store.CheckFrequentTrading(ctx, stockCode)
	return FrequentTradingCheck{
		CanTrade: canTrade,
		Message:  msg,
	}
}

func (s *Service) EvaluateAiRecommendAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery {
	_ = now

	aiRecommendStocks := s.store.ListAiRecommendAlertStocks(ctx)
	if len(aiRecommendStocks) == 0 {
		return []notificationservice.Delivery{}
	}

	stockCodes := make([]string, 0, len(aiRecommendStocks))
	stockCodeMap := make(map[string]*models.AiRecommendStocks, len(aiRecommendStocks))
	for i := range aiRecommendStocks {
		stock := &aiRecommendStocks[i]
		stopLossPrice, _ := convertor.ToFloat(stock.RecommendStopLossPrice)
		if stock.RecommendBuyPriceMin <= 0 && stock.RecommendStopProfitPriceMin <= 0 && stopLossPrice <= 0 {
			continue
		}
		code := normalizeAiRecommendStockCode(stock.StockCode)
		if code == "" {
			continue
		}
		stockCodes = append(stockCodes, code)
		stockCodeMap[code] = stock
	}
	if len(stockCodes) == 0 {
		return []notificationservice.Delivery{}
	}

	quotes := s.store.GetRealtimePrices(ctx, stockCodes...)
	deliveries := make([]notificationservice.Delivery, 0)
	for _, quote := range quotes {
		aiStock, ok := stockCodeMap[normalizeAiRecommendStockCode(quote.StockCode)]
		if !ok {
			continue
		}

		currentPrice, err := convertor.ToFloat(quote.Price)
		if err != nil || currentPrice <= 0 {
			continue
		}

		baseAlertKey := fmt.Sprintf("%s:%s", aiStock.StockCode, aiStock.DataTime.Format("20060102"))
		deliveries = append(deliveries, s.evaluateAiBuyAlert(*aiStock, currentPrice, baseAlertKey)...)
		deliveries = append(deliveries, s.evaluateAiProfitAlert(*aiStock, currentPrice, baseAlertKey)...)
		deliveries = append(deliveries, s.evaluateAiLossAlert(*aiStock, currentPrice, baseAlertKey)...)
	}
	return deliveries
}

func (s *Service) evaluateAiBuyAlert(stock models.AiRecommendStocks, currentPrice float64, baseAlertKey string) []notificationservice.Delivery {
	if stock.RecommendBuyPriceMin <= 0 {
		return []notificationservice.Delivery{}
	}

	alertKey := baseAlertKey + ":BUY"
	if currentPrice <= stock.RecommendBuyPriceMin {
		priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
		if priceSinceLastAlert == 0 || priceSinceLastAlert > stock.RecommendBuyPriceMin {
			if s.canSendAlert(alertKey, 5*time.Minute) {
				s.updateAlertSentTime(alertKey)
				s.updatePriceAtAlertReset(alertKey, currentPrice)
				message := fmt.Sprintf(
					"### 【买入预警】%s\n\n- 股票代码: %s\n- 当前价格: %.2f\n- 建议买入价: %.2f - %.2f\n- 推荐时间: %s",
					stock.StockName,
					stock.StockCode,
					currentPrice,
					stock.RecommendBuyPriceMin,
					stock.RecommendBuyPriceMax,
					stock.DataTime.Format("2006-01-02 15:04:05"),
				)
				return []notificationservice.Delivery{buildTypedDelivery(message, normalizeAiRecommendStockCode(stock.StockCode), 2)}
			}
		} else {
			s.updatePriceAtAlertReset(alertKey, currentPrice)
		}
		return []notificationservice.Delivery{}
	}

	priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
	if priceSinceLastAlert == 0 || currentPrice > priceSinceLastAlert {
		s.updatePriceAtAlertReset(alertKey, currentPrice)
	}
	return []notificationservice.Delivery{}
}

func (s *Service) evaluateAiProfitAlert(stock models.AiRecommendStocks, currentPrice float64, baseAlertKey string) []notificationservice.Delivery {
	if stock.RecommendStopProfitPriceMin <= 0 {
		return []notificationservice.Delivery{}
	}

	alertKey := baseAlertKey + ":PROFIT"
	if currentPrice >= stock.RecommendStopProfitPriceMin {
		priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
		if priceSinceLastAlert == 0 || priceSinceLastAlert < stock.RecommendStopProfitPriceMin {
			if s.canSendAlert(alertKey, 5*time.Minute) {
				s.updateAlertSentTime(alertKey)
				s.updatePriceAtAlertReset(alertKey, currentPrice)
				message := fmt.Sprintf(
					"### 【止盈预警】%s\n\n- 股票代码: %s\n- 当前价格: %.2f\n- 建议止盈价: %.2f - %.2f\n- 推荐时间: %s",
					stock.StockName,
					stock.StockCode,
					currentPrice,
					stock.RecommendStopProfitPriceMin,
					stock.RecommendStopProfitPriceMax,
					stock.DataTime.Format("2006-01-02 15:04:05"),
				)
				return []notificationservice.Delivery{buildTypedDelivery(message, normalizeAiRecommendStockCode(stock.StockCode), 4)}
			}
		} else {
			s.updatePriceAtAlertReset(alertKey, currentPrice)
		}
		return []notificationservice.Delivery{}
	}

	priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
	if priceSinceLastAlert == 0 || currentPrice < priceSinceLastAlert {
		s.updatePriceAtAlertReset(alertKey, currentPrice)
	}
	return []notificationservice.Delivery{}
}

func (s *Service) evaluateAiLossAlert(stock models.AiRecommendStocks, currentPrice float64, baseAlertKey string) []notificationservice.Delivery {
	stopLossPrice, err := convertor.ToFloat(stock.RecommendStopLossPrice)
	if err != nil || stopLossPrice <= 0 {
		return []notificationservice.Delivery{}
	}

	alertKey := baseAlertKey + ":LOSS"
	if currentPrice <= stopLossPrice {
		priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
		if priceSinceLastAlert == 0 || priceSinceLastAlert > stopLossPrice {
			if s.canSendAlert(alertKey, 5*time.Minute) {
				s.updateAlertSentTime(alertKey)
				s.updatePriceAtAlertReset(alertKey, currentPrice)
				message := fmt.Sprintf(
					"### 【止损预警】%s\n\n- 股票代码: %s\n- 当前价格: %.2f\n- 建议止损价: %s\n- 推荐时间: %s",
					stock.StockName,
					stock.StockCode,
					currentPrice,
					stock.RecommendStopLossPrice,
					stock.DataTime.Format("2006-01-02 15:04:05"),
				)
				return []notificationservice.Delivery{buildTypedDelivery(message, normalizeAiRecommendStockCode(stock.StockCode), 5)}
			}
		} else {
			s.updatePriceAtAlertReset(alertKey, currentPrice)
		}
		return []notificationservice.Delivery{}
	}

	priceSinceLastAlert := s.getPriceAtAlertReset(alertKey)
	if priceSinceLastAlert == 0 || currentPrice > priceSinceLastAlert {
		s.updatePriceAtAlertReset(alertKey, currentPrice)
	}
	return []notificationservice.Delivery{}
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

func normalizeStockChangesPage(page *data.StockChangesResponse) *data.StockChangesResponse {
	if page == nil {
		return &data.StockChangesResponse{Data: []data.StockChangeItem{}}
	}
	if page.Data == nil {
		page.Data = []data.StockChangeItem{}
	}
	return page
}

func emptyStockChangeHistoryPage(query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}
	return &models.StockChangeHistoryPageData{
		List:       []models.StockChangeHistory{},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}
}

func normalizeStockChangeHistoryPage(page *models.StockChangeHistoryPageData, query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	if page == nil {
		return emptyStockChangeHistoryPage(query)
	}
	if page.List == nil {
		page.List = []models.StockChangeHistory{}
	}
	return page
}

func emptyAiRecommendPage(query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	return &models.AiRecommendStocksPageData{
		List:       []models.AiRecommendStocks{},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}
}

func normalizeAiRecommendPage(page *models.AiRecommendStocksPageData, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	if page == nil {
		return emptyAiRecommendPage(query)
	}
	if page.List == nil {
		page.List = []models.AiRecommendStocks{}
	}
	return page
}

func emptyAllStockInfoPage(query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return &data.AllStockInfoPageData{
		List:       []models.AllStockInfo{},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}
}

func normalizeAllStockInfoPage(page *data.AllStockInfoPageData, query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	if page == nil {
		return emptyAllStockInfoPage(query)
	}
	if page.List == nil {
		page.List = []models.AllStockInfo{}
	}
	return page
}

func emptyTradingRecordPage(query data.TradingRecordListQuery) *data.TradingRecordPageData {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return &data.TradingRecordPageData{
		List:       []data.TradingRecordItem{},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}
}

func normalizeTradingRecordPage(page *data.TradingRecordPageData, query data.TradingRecordListQuery) *data.TradingRecordPageData {
	if page == nil {
		return emptyTradingRecordPage(query)
	}
	if page.List == nil {
		page.List = []data.TradingRecordItem{}
	}
	return page
}

func normalizeAiRecommendStockCode(stockCode string) string {
	code := strings.TrimSpace(stockCode)
	if code == "" {
		return ""
	}
	if strings.Contains(code, ".") {
		return data.ConvertTushareCodeToStockCode(code)
	}
	return strings.ToLower(code)
}

func buildTypedDelivery(message, stockCode string, msgType int) notificationservice.Delivery {
	return notificationservice.Delivery{
		DingResult:   strings.TrimSpace(message),
		EventTitle:   strings.TrimSpace(stockCode),
		EventContent: strconv.Itoa(msgType),
	}
}
