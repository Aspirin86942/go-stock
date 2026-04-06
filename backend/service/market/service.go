package market

import (
	"go-stock/backend/models"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
)

type Service struct {
	source Source
}

func NewService(source Source) *Service {
	return &Service{source: source}
}

func (s *Service) LoadFeeds() FeedSet {
	telegraph := s.loadFeed("财联社电报")
	sina := s.loadFeed("新浪财经")
	foreign := s.loadFeed("外媒")

	return FeedSet{
		Telegraph: telegraph.Items,
		Sina:      sina.Items,
		Foreign:   foreign.Items,
	}
}

func (s *Service) RefreshFeed(source string) Feed {
	s.source.RefreshFeeds()
	return s.loadFeed(source)
}

func (s *Service) LoadGlobalIndexes(crawlTimeOut uint) IndexSet {
	raw := s.source.GlobalStockIndexes(crawlTimeOut)
	return IndexSet{
		Common:  mapGlobalIndexesByRegion(raw, "common"),
		America: mapGlobalIndexesByRegion(raw, "america"),
		Europe:  mapGlobalIndexesByRegion(raw, "europe"),
		Asia:    mapGlobalIndexesByRegion(raw, "asia"),
		Other:   mapGlobalIndexesByRegion(raw, "other"),
	}
}

func (s *Service) LoadIndustryRanks(sort string, cnt int) []IndustryRankEntry {
	raw := s.source.GetIndustryRank(sort, cnt)
	rows, ok := raw["data"].([]any)
	if !ok || len(rows) == 0 {
		return []IndustryRankEntry{}
	}

	result := make([]IndustryRankEntry, 0, len(rows))
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, IndustryRankEntry{
			BoardCode:             convertor.ToString(row["bd_code"]),
			BoardName:             convertor.ToString(row["bd_name"]),
			BoardChangePercent:    convertor.ToString(row["bd_zdf"]),
			BoardChangePercent5D:  convertor.ToString(row["bd_zdf5"]),
			BoardChangePercent20D: convertor.ToString(row["bd_zdf20"]),
			LeaderCode:            convertor.ToString(row["nzg_code"]),
			LeaderName:            convertor.ToString(row["nzg_name"]),
			LeaderChangePercent:   convertor.ToString(row["nzg_zdf"]),
			LeaderPrice:           convertor.ToString(row["nzg_zxj"]),
		})
	}
	return result
}

func (s *Service) LoadGlobalIndexesReadable(crawlTimeout uint) string {
	return strings.TrimSpace(s.source.GlobalStockIndexesReadable(crawlTimeout))
}

func (s *Service) LoadIndustryMoneyRanks(fenlei, sort string) []IndustryMoneyRankRow {
	raw := s.source.GetIndustryMoneyRankSina(fenlei, sort)
	if len(raw) == 0 {
		return []IndustryMoneyRankRow{}
	}

	result := make([]IndustryMoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, IndustryMoneyRankRow{
			Category:       convertor.ToString(row["category"]),
			Name:           convertor.ToString(row["name"]),
			AvgChangeRatio: toFloat(row["avg_changeratio"]),
			InAmount:       toFloat(row["inamount"]),
			OutAmount:      toFloat(row["outamount"]),
			NetAmount:      toFloat(row["netamount"]),
			RatioAmount:    toFloat(row["ratioamount"]),
			TSName:         convertor.ToString(row["ts_name"]),
			TSSymbol:       convertor.ToString(row["ts_symbol"]),
			TSChangeRatio:  toFloat(row["ts_changeratio"]),
			TSTrade:        toFloat(row["ts_trade"]),
			TSRatioAmount:  toFloat(row["ts_ratioamount"]),
		})
	}
	return result
}

func (s *Service) LoadMoneyRanks(sort string) []MoneyRankRow {
	raw := s.source.GetMoneyRankSina(sort)
	if len(raw) == 0 {
		return []MoneyRankRow{}
	}

	result := make([]MoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, MoneyRankRow{
			Symbol:      convertor.ToString(row["symbol"]),
			Name:        convertor.ToString(row["name"]),
			Trade:       toFloat(row["trade"]),
			ChangeRatio: toFloat(row["changeratio"]),
			Turnover:    toFloat(row["turnover"]),
			Amount:      toFloat(row["amount"]),
			OutAmount:   toFloat(row["outamount"]),
			InAmount:    toFloat(row["inamount"]),
			NetAmount:   toFloat(row["netamount"]),
			RatioAmount: toFloat(row["ratioamount"]),
			R0Out:       toFloat(row["r0_out"]),
			R0In:        toFloat(row["r0_in"]),
			R0Net:       toFloat(row["r0_net"]),
			R0Ratio:     toFloat(row["r0_ratio"]),
			R3Out:       toFloat(row["r3_out"]),
			R3In:        toFloat(row["r3_in"]),
			R3Net:       toFloat(row["r3_net"]),
			R3Ratio:     toFloat(row["r3_ratio"]),
		})
	}
	return result
}

func (s *Service) LoadStockMoneyTrend(stockCode string, days int) []StockMoneyTrendRow {
	raw := s.source.GetStockMoneyTrendByDay(stockCode, days)
	if len(raw) == 0 {
		return []StockMoneyTrendRow{}
	}

	result := make([]StockMoneyTrendRow, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		row := raw[i]
		result = append(result, StockMoneyTrendRow{
			OpenDate:  convertor.ToString(row["opendate"]),
			Trade:     toFloat(row["trade"]),
			NetAmount: toFloat(row["netamount"]),
			R0Net:     toFloat(row["r0_net"]),
		})
	}
	return result
}

func (s *Service) loadFeed(source string) Feed {
	items := s.source.GetTelegraphList(source)
	if items == nil {
		return Feed{Source: source, Items: []*models.Telegraph{}}
	}
	copied := make([]*models.Telegraph, len(*items))
	copy(copied, *items)
	return Feed{Source: source, Items: copied}
}

func mapGlobalIndexesByRegion(raw map[string]any, region string) []GlobalIndexEntry {
	rows, ok := raw[region].([]any)
	if !ok || len(rows) == 0 {
		return []GlobalIndexEntry{}
	}

	result := make([]GlobalIndexEntry, 0, len(rows))
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, GlobalIndexEntry{
			Code:     convertor.ToString(row["code"]),
			Name:     convertor.ToString(row["name"]),
			Location: convertor.ToString(row["location"]),
			Qtcode:   convertor.ToString(row["qtcode"]),
			State:    convertor.ToString(row["state"]),
			Zdf:      convertor.ToString(row["zdf"]),
			Zxj:      convertor.ToString(row["zxj"]),
			Img:      convertor.ToString(row["img"]),
			Region:   region,
		})
	}
	return result
}

func toFloat(value any) float64 {
	v, err := convertor.ToFloat(value)
	if err != nil {
		return 0
	}
	return v
}
