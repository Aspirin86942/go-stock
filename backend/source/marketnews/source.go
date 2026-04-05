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
