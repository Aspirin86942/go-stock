package marketnews

import (
	"go-stock/backend/data"
	"go-stock/backend/models"

	"github.com/coocood/freecache"
)

type Source struct {
	api   *data.MarketNewsApi
	cache *freecache.Cache
}

func NewSource() *Source {
	return &Source{
		api:   data.NewMarketNewsApi(),
		cache: freecache.NewCache(1 << 20),
	}
}

func (s *Source) GetTelegraphList(source string) *[]*models.Telegraph {
	return s.api.GetTelegraphList(source)
}

func (s *Source) RefreshFeeds() {
	go s.api.TelegraphList(30)
	go s.api.GetSinaNews(30)
	go s.api.TradingViewNews()
}

func (s *Source) GlobalStockIndexes(crawlTimeout uint) map[string]any {
	return s.api.GlobalStockIndexes(crawlTimeout)
}

func (s *Source) GetIndustryRank(sort string, cnt int) map[string]any {
	return s.api.GetIndustryRank(sort, cnt)
}

func (s *Source) GlobalStockIndexesReadable(crawlTimeout uint) string {
	return s.api.GlobalStockIndexesReadable(crawlTimeout)
}

func (s *Source) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return s.api.GetIndustryMoneyRankSina(fenlei, sort)
}

func (s *Source) GetMoneyRankSina(sort string) []map[string]any {
	return s.api.GetMoneyRankSina(sort)
}

func (s *Source) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	return s.api.GetStockMoneyTrendByDay(stockCode, days)
}

func (s *Source) LongTiger(date string) *[]models.LongTigerRankData {
	return s.api.LongTiger(date)
}

func (s *Source) StockResearchReport(stockCode string, days int) []any {
	return s.api.StockResearchReport(stockCode, days)
}

func (s *Source) StockNotice(stockCode string) []any {
	return s.api.StockNotice(stockCode)
}

func (s *Source) IndustryResearchReport(industryCode string, days int) []any {
	return s.api.IndustryResearchReport(industryCode, days)
}

func (s *Source) EMDictCode(code string) []any {
	return s.api.EMDictCode(code, s.cache)
}

func (s *Source) XueQiuHotStock(size int, marketType string) *[]models.HotItem {
	return s.api.XUEQIUHotStock(size, marketType)
}

func (s *Source) HotEvent(size int) *[]models.HotEvent {
	return s.api.HotEvent(size)
}

func (s *Source) HotTopic(size int) []any {
	return s.api.HotTopic(size)
}

func (s *Source) InvestCalendar(yearMonth string) []any {
	return s.api.InvestCalendar(yearMonth)
}

func (s *Source) ClsCalendar() []any {
	return s.api.ClsCalendar()
}

func (s *Source) SearchStock(words string, pageSize int) map[string]any {
	return data.NewSearchStockApi(words).SearchStock(pageSize)
}

func (s *Source) HotStrategy() map[string]any {
	return data.NewSearchStockApi("").HotStrategy()
}

func (s *Source) GetStockKLine(stockCode string, days int64) *[]data.KLineData {
	return data.NewStockDataApi().GetHK_KLineData(stockCode, "day", days)
}

func (s *Source) GetStockCommonKLine(stockCode string, days int64) *[]data.KLineData {
	return data.NewStockDataApi().GetCommonKLineData(stockCode, "day", days)
}

func (s *Source) GetStockMinutePriceData(stockCode string) (*[]data.MinuteData, string) {
	return data.NewStockDataApi().GetStockMinutePriceData(stockCode)
}

func (s *Source) GetStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) *[]data.KLineData {
	return data.NewEastMoneyKLineApi(data.GetSettingConfig()).GetKLineDataBefore(stockCode, klt, "", limit, end)
}

func (s *Source) GetStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) map[string]any {
	result := data.NewEastMoneyKLineApi(data.GetSettingConfig()).GetKLineDataBeforeResult(stockCode, klt, "", limit, end)
	return map[string]any{
		"ok":              len(result.Data) > 0 && result.ErrorCode == "",
		"data":            result.Data,
		"message":         result.Message,
		"errorCode":       result.ErrorCode,
		"usedCookieRetry": result.UsedCookieRetry,
	}
}

func (s *Source) GetStockRealtimePrice(stockCode string) *data.StockInfo {
	items, err := data.NewStockDataApi().GetStockCodeRealTimeData(stockCode)
	if err != nil || items == nil || len(*items) == 0 {
		return nil
	}
	item := (*items)[0]
	return &item
}
