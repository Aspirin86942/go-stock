package market

import (
	"go-stock/backend/models"

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
