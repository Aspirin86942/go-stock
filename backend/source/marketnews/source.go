package marketnews

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

type Source struct {
	api *data.MarketNewsApi
}

func NewSource() *Source {
	return &Source{
		api: data.NewMarketNewsApi(),
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
