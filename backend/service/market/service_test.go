package market

import (
	"testing"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/models"
)

type fakeSource struct {
	telegraphsBySource map[string]*[]*models.Telegraph
	globalIndexes      map[string]any
	industryRanks      map[string]any
	longTigerRanks     *[]models.LongTigerRankData
	hotTopics          []any
	realtimePrice      *data.StockInfo
	events             []string
}

func (f *fakeSource) GetTelegraphList(source string) *[]*models.Telegraph {
	f.events = append(f.events, "load:"+source)
	return f.telegraphsBySource[source]
}

func (f *fakeSource) RefreshFeeds() {
	f.events = append(f.events, "refresh")
}

func (f *fakeSource) GlobalStockIndexes(crawlTimeOut uint) map[string]any {
	f.events = append(f.events, "global")
	return f.globalIndexes
}

func (f *fakeSource) GetIndustryRank(sort string, cnt int) map[string]any {
	f.events = append(f.events, "industry")
	return f.industryRanks
}

func (f *fakeSource) GlobalStockIndexesReadable(crawlTimeout uint) string {
	return ""
}

func (f *fakeSource) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return []map[string]any{}
}

func (f *fakeSource) GetMoneyRankSina(sort string) []map[string]any {
	return []map[string]any{}
}

func (f *fakeSource) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	return []map[string]any{}
}

func (f *fakeSource) LongTiger(date string) *[]models.LongTigerRankData {
	return f.longTigerRanks
}

func (f *fakeSource) StockResearchReport(stockCode string, days int) []any {
	return []any{}
}

func (f *fakeSource) StockNotice(stockCode string) []any {
	return []any{}
}

func (f *fakeSource) IndustryResearchReport(industryCode string, days int) []any {
	return []any{}
}

func (f *fakeSource) EMDictCode(code string) []any {
	return []any{}
}

func (f *fakeSource) XueQiuHotStock(size int, marketType string) *[]models.HotItem {
	items := []models.HotItem{}
	return &items
}

func (f *fakeSource) HotEvent(size int) *[]models.HotEvent {
	items := []models.HotEvent{}
	return &items
}

func (f *fakeSource) HotTopic(size int) []any {
	return f.hotTopics
}

func (f *fakeSource) InvestCalendar(yearMonth string) []any {
	return []any{}
}

func (f *fakeSource) ClsCalendar() []any {
	return []any{}
}

func (f *fakeSource) SearchStock(words string, pageSize int) map[string]any {
	return map[string]any{}
}

func (f *fakeSource) HotStrategy() map[string]any {
	return map[string]any{}
}

func (f *fakeSource) GetStockKLine(stockCode string, days int64) *[]data.KLineData {
	rows := []data.KLineData{}
	return &rows
}

func (f *fakeSource) GetStockCommonKLine(stockCode string, days int64) *[]data.KLineData {
	rows := []data.KLineData{}
	return &rows
}

func (f *fakeSource) GetStockMinutePriceData(stockCode string) (*[]data.MinuteData, string) {
	rows := []data.MinuteData{}
	return &rows, ""
}

func (f *fakeSource) GetStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) *[]data.KLineData {
	rows := []data.KLineData{}
	return &rows
}

func (f *fakeSource) GetStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) map[string]any {
	return map[string]any{}
}

func (f *fakeSource) GetStockRealtimePrice(stockCode string) *data.StockInfo {
	return f.realtimePrice
}

func TestService_LoadReadModel_NormalizesTypedContracts(t *testing.T) {
	now := time.Now()
	telegraphItems := []*models.Telegraph{
		{
			Title:    "telegraph-1",
			Source:   "财联社电报",
			DataTime: &now,
		},
	}
	sinaItems := []*models.Telegraph{
		{
			Title:    "sina-1",
			Source:   "新浪财经",
			DataTime: &now,
		},
	}
	foreignItems := []*models.Telegraph{
		{
			Title:    "foreign-1",
			Source:   "外媒",
			DataTime: &now,
		},
	}

	src := &fakeSource{
		telegraphsBySource: map[string]*[]*models.Telegraph{
			"财联社电报": &telegraphItems,
			"新浪财经":  &sinaItems,
			"外媒":    &foreignItems,
		},
		globalIndexes: map[string]any{
			"common": []any{
				map[string]any{
					"code":     "csi300",
					"name":     "沪深300",
					"location": "中国",
					"qtcode":   "CN000300",
					"state":    "open",
					"zdf":      "0.88",
					"zxj":      "3550.11",
					"img":      "https://example.com/common.png",
				},
			},
			"america": []any{
				map[string]any{
					"code":     "dji",
					"name":     "道琼斯",
					"location": "美国",
					"qtcode":   "USDJI",
					"state":    "close",
					"zdf":      "-0.22",
					"zxj":      "39100.20",
					"img":      "https://example.com/america.png",
				},
			},
			"europe": []any{},
			"asia": []any{
				map[string]any{
					"code":     "n225",
					"name":     "日经225",
					"location": "日本",
					"qtcode":   "JPN225",
					"state":    "open",
					"zdf":      "1.02",
					"zxj":      "40123.05",
					"img":      "https://example.com/asia.png",
				},
			},
			"other": []any{},
		},
		industryRanks: map[string]any{
			"data": []any{
				map[string]any{
					"bd_code":  "BK1234",
					"bd_name":  "半导体",
					"bd_zdf":   "2.31",
					"bd_zdf5":  "6.12",
					"bd_zdf20": "13.57",
					"nzg_code": "688981",
					"nzg_name": "中芯国际",
					"nzg_zdf":  "4.88",
					"nzg_zxj":  "48.02",
				},
			},
		},
	}

	svc := NewService(src)

	feeds := svc.LoadFeeds()
	if len(feeds.Telegraph) != 1 || feeds.Telegraph[0].Source != "财联社电报" {
		t.Fatalf("telegraph feed mapping mismatch: %+v", feeds.Telegraph)
	}
	if len(feeds.Sina) != 1 || feeds.Sina[0].Source != "新浪财经" {
		t.Fatalf("sina feed mapping mismatch: %+v", feeds.Sina)
	}
	if len(feeds.Foreign) != 1 || feeds.Foreign[0].Source != "外媒" {
		t.Fatalf("foreign feed mapping mismatch: %+v", feeds.Foreign)
	}

	indexes := svc.LoadGlobalIndexes(30)
	if len(indexes.Common) != 1 || indexes.Common[0].Region != "common" {
		t.Fatalf("common indexes mapping mismatch: %+v", indexes.Common)
	}
	if len(indexes.America) != 1 || indexes.America[0].Region != "america" {
		t.Fatalf("america indexes mapping mismatch: %+v", indexes.America)
	}
	if len(indexes.Asia) != 1 || indexes.Asia[0].Region != "asia" {
		t.Fatalf("asia indexes mapping mismatch: %+v", indexes.Asia)
	}

	industry := svc.LoadIndustryRanks("0", 10)
	if len(industry) != 1 {
		t.Fatalf("industry rank mapping mismatch, got %d", len(industry))
	}
	if industry[0].BoardCode != "BK1234" || industry[0].LeaderCode != "688981" {
		t.Fatalf("industry rank fields mismatch: %+v", industry[0])
	}
	if industry[0].BoardChangePercent != "2.31" || industry[0].LeaderChangePercent != "4.88" {
		t.Fatalf("industry rank normalized fields mismatch: %+v", industry[0])
	}
}

func TestService_RefreshFeed_RefreshesBeforeReload(t *testing.T) {
	telegraphItems := []*models.Telegraph{{Title: "telegraph"}}
	sinaItems := []*models.Telegraph{{Title: "sina"}}
	foreignItems := []*models.Telegraph{{Title: "foreign"}}

	src := &fakeSource{
		telegraphsBySource: map[string]*[]*models.Telegraph{
			"财联社电报": &telegraphItems,
			"新浪财经":  &sinaItems,
			"外媒":    &foreignItems,
		},
	}
	svc := NewService(src)

	feed := svc.RefreshFeed("新浪财经")
	if feed.Source != "新浪财经" || len(feed.Items) != 1 {
		t.Fatalf("refresh should reload requested source only: %+v", feed)
	}
	if len(src.events) != 2 {
		t.Fatalf("unexpected event count: %+v", src.events)
	}
	if src.events[0] != "refresh" || src.events[1] != "load:新浪财经" {
		t.Fatalf("refresh must happen before reload, got events: %+v", src.events)
	}
}

func TestService_FeedResults_DoNotAliasSourceSlices(t *testing.T) {
	srcTelegraph := []*models.Telegraph{{Title: "t-1", Source: "财联社电报"}}
	srcSina := []*models.Telegraph{{Title: "s-1", Source: "新浪财经"}}
	srcForeign := []*models.Telegraph{{Title: "f-1", Source: "外媒"}}
	src := &fakeSource{
		telegraphsBySource: map[string]*[]*models.Telegraph{
			"财联社电报": &srcTelegraph,
			"新浪财经":  &srcSina,
			"外媒":    &srcForeign,
		},
	}
	svc := NewService(src)

	feeds := svc.LoadFeeds()
	feeds.Telegraph[0] = &models.Telegraph{Title: "mutated"}
	feeds.Telegraph = append(feeds.Telegraph, &models.Telegraph{Title: "extra"})

	if len(srcTelegraph) != 1 {
		t.Fatalf("source telegraph slice length should stay unchanged, got %d", len(srcTelegraph))
	}
	if srcTelegraph[0].Title != "t-1" {
		t.Fatalf("source telegraph item should stay unchanged, got %+v", srcTelegraph[0])
	}

	refreshed := svc.RefreshFeed("新浪财经")
	refreshed.Items[0] = &models.Telegraph{Title: "mutated-sina"}
	refreshed.Items = append(refreshed.Items, &models.Telegraph{Title: "extra-sina"})

	if len(srcSina) != 1 {
		t.Fatalf("source sina slice length should stay unchanged, got %d", len(srcSina))
	}
	if srcSina[0].Title != "s-1" {
		t.Fatalf("source sina item should stay unchanged, got %+v", srcSina[0])
	}
}

func TestService_GracefulFallbacks_AreStable(t *testing.T) {
	src := &fakeSource{
		telegraphsBySource: map[string]*[]*models.Telegraph{
			"财联社电报": nil,
			"新浪财经":  nil,
			"外媒":    nil,
		},
		globalIndexes: map[string]any{
			"common": "bad-type",
			"america": []any{
				"bad-row",
				map[string]any{"code": "dji", "name": "道琼斯"},
			},
			"europe": nil,
			"asia":   []any{},
		},
		industryRanks: map[string]any{
			"data": []any{
				"bad-row",
				map[string]any{
					"bd_code":  "BK1",
					"nzg_code": "000001",
				},
			},
		},
	}
	svc := NewService(src)

	feeds := svc.LoadFeeds()
	if feeds.Telegraph == nil || len(feeds.Telegraph) != 0 {
		t.Fatalf("nil telegraph feed should degrade to empty slice: %+v", feeds.Telegraph)
	}
	if feeds.Sina == nil || len(feeds.Sina) != 0 {
		t.Fatalf("nil sina feed should degrade to empty slice: %+v", feeds.Sina)
	}
	if feeds.Foreign == nil || len(feeds.Foreign) != 0 {
		t.Fatalf("nil foreign feed should degrade to empty slice: %+v", feeds.Foreign)
	}

	indexes := svc.LoadGlobalIndexes(30)
	if indexes.Common == nil || len(indexes.Common) != 0 {
		t.Fatalf("bad common payload should degrade to empty slice: %+v", indexes.Common)
	}
	if len(indexes.America) != 1 || indexes.America[0].Code != "dji" {
		t.Fatalf("america payload should skip bad row and keep good row: %+v", indexes.America)
	}
	if indexes.Europe == nil || len(indexes.Europe) != 0 {
		t.Fatalf("nil europe payload should degrade to empty slice: %+v", indexes.Europe)
	}

	industry := svc.LoadIndustryRanks("0", 10)
	if len(industry) != 1 || industry[0].BoardCode != "BK1" || industry[0].LeaderCode != "000001" {
		t.Fatalf("industry payload should skip bad row and keep good row: %+v", industry)
	}

	emptyIndustry := NewService(&fakeSource{industryRanks: map[string]any{"data": []any{}}}).LoadIndustryRanks("0", 10)
	if emptyIndustry == nil || len(emptyIndustry) != 0 {
		t.Fatalf("empty industry data should degrade to empty slice: %+v", emptyIndustry)
	}

	missingIndustry := NewService(&fakeSource{industryRanks: map[string]any{}}).LoadIndustryRanks("0", 10)
	if missingIndustry == nil || len(missingIndustry) != 0 {
		t.Fatalf("missing industry data should degrade to empty slice: %+v", missingIndustry)
	}
}

func TestService_LoadRealtimePriceFallsBackThroughBidLevels(t *testing.T) {
	a1First := NewService(&fakeSource{
		realtimePrice: &data.StockInfo{
			Code:     "sz000001",
			Name:     "平安银行",
			A1P:      "12.35",
			B1P:      "12.34",
			PreClose: "12.33",
		},
	})

	a1Price := a1First.LoadRealtimePrice("sz000001")
	if a1Price.StockCode != "sz000001" {
		t.Fatalf("unexpected stock code: %+v", a1Price)
	}
	if a1Price.Price != "12.35" {
		t.Fatalf("expected A1P to win before B1P, got %+v", a1Price)
	}

	b1BeforePreClose := NewService(&fakeSource{
		realtimePrice: &data.StockInfo{
			Code:     "sz000002",
			Name:     "万科A",
			B1P:      "21.09",
			PreClose: "21.08",
		},
	})

	b1Price := b1BeforePreClose.LoadRealtimePrice("sz000002")
	if b1Price.Price != "21.09" {
		t.Fatalf("expected B1P to win before pre-close, got %+v", b1Price)
	}

	preCloseSvc := NewService(&fakeSource{
		realtimePrice: &data.StockInfo{
			Code:     "sz000003",
			Name:     "招商银行",
			PreClose: "31.08",
		},
	})

	preClosePrice := preCloseSvc.LoadRealtimePrice("sz000003")
	if preClosePrice.Price != "31.08" {
		t.Fatalf("expected fallback price from pre-close, got %+v", preClosePrice)
	}
}

func TestService_LoadHotTopicsReturnsEmptySliceWhenSourceReturnsNil(t *testing.T) {
	svc := NewService(&fakeSource{hotTopics: nil})

	topics := svc.LoadHotTopics(10)
	if topics == nil {
		t.Fatal("expected empty slice instead of nil")
	}
	if len(topics) != 0 {
		t.Fatalf("expected no topics, got %+v", topics)
	}
}
