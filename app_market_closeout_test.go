package main

import (
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
)

type marketLegacyReadServiceStub struct {
	longTigerRanks              []models.LongTigerRankData
	stockResearchReports        []marketservice.StockResearchReportEntry
	stockNotices                []marketservice.StockNoticeEntry
	industryResearchReports     []marketservice.IndustryResearchReportEntry
	dictCodes                   []marketservice.EMDictCodeEntry
	hotStocks                   []models.HotItem
	hotEvents                   []models.HotEvent
	hotTopics                   []marketservice.HotTopicEntry
	investCalendar              []marketservice.InvestCalendarDay
	clsCalendar                 []marketservice.ClsCalendarDay
	searchStockResult           marketservice.SearchStockResponse
	hotStrategyResult           models.HotStrategy
	stockKLine                  []data.KLineData
	stockCommonKLine            []data.KLineData
	minutePriceLine             marketservice.MinutePriceLine
	eastMoneyKLine              []data.KLineData
	eastMoneyKLinePage          []data.KLineData
	eastMoneyKLineResult        marketservice.EastMoneyKLinePageResult
	eastMoneyKLinePageResult    marketservice.EastMoneyKLinePageResult
	realtimePrice               marketservice.RealtimePrice
	lastLongTigerDate           string
	lastStockResearchCode       string
	lastStockNoticeCode         string
	lastIndustryResearchCode    string
	lastDictCode                string
	lastHotStockMarketType      string
	lastHotEventSize            int
	lastHotTopicSize            int
	lastInvestCalendarYearMonth string
	lastSearchStockWords        string
	lastStockKLineCode          string
	lastStockKLineDays          int64
	lastStockCommonKLineCode    string
	lastStockCommonKLineDays    int64
	lastMinutePriceStockCode    string
	lastMinutePriceStockName    string
	lastEastMoneyStockCode      string
	lastEastMoneyKlt            string
	lastEastMoneyLimit          int
	lastEastMoneyEnd            string
	lastRealtimePriceStockCode  string
}

func (s *marketLegacyReadServiceStub) LoadFeeds() marketservice.FeedSet {
	return marketservice.FeedSet{}
}

func (s *marketLegacyReadServiceStub) RefreshFeed(source string) marketservice.Feed {
	return marketservice.Feed{}
}

func (s *marketLegacyReadServiceStub) LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet {
	return marketservice.IndexSet{}
}

func (s *marketLegacyReadServiceStub) LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return []marketservice.IndustryRankEntry{}
}

func (s *marketLegacyReadServiceStub) LoadLongTiger(date string) []models.LongTigerRankData {
	s.lastLongTigerDate = date
	return s.longTigerRanks
}

func (s *marketLegacyReadServiceStub) LoadStockResearchReports(stockCode string) []marketservice.StockResearchReportEntry {
	s.lastStockResearchCode = stockCode
	return s.stockResearchReports
}

func (s *marketLegacyReadServiceStub) LoadStockNotices(stockCode string) []marketservice.StockNoticeEntry {
	s.lastStockNoticeCode = stockCode
	return s.stockNotices
}

func (s *marketLegacyReadServiceStub) LoadIndustryResearchReports(industryCode string) []marketservice.IndustryResearchReportEntry {
	s.lastIndustryResearchCode = industryCode
	return s.industryResearchReports
}

func (s *marketLegacyReadServiceStub) LoadEMDictCodes(code string) []marketservice.EMDictCodeEntry {
	s.lastDictCode = code
	return s.dictCodes
}

func (s *marketLegacyReadServiceStub) LoadHotStocks(marketType string) []models.HotItem {
	s.lastHotStockMarketType = marketType
	return s.hotStocks
}

func (s *marketLegacyReadServiceStub) LoadHotEvents(size int) []models.HotEvent {
	s.lastHotEventSize = size
	return s.hotEvents
}

func (s *marketLegacyReadServiceStub) LoadHotTopics(size int) []marketservice.HotTopicEntry {
	s.lastHotTopicSize = size
	return s.hotTopics
}

func (s *marketLegacyReadServiceStub) LoadInvestCalendar(yearMonth string) []marketservice.InvestCalendarDay {
	s.lastInvestCalendarYearMonth = yearMonth
	return s.investCalendar
}

func (s *marketLegacyReadServiceStub) LoadClsCalendar() []marketservice.ClsCalendarDay {
	return s.clsCalendar
}

func (s *marketLegacyReadServiceStub) SearchStocks(words string) marketservice.SearchStockResponse {
	s.lastSearchStockWords = words
	return s.searchStockResult
}

func (s *marketLegacyReadServiceStub) LoadHotStrategies() models.HotStrategy {
	return s.hotStrategyResult
}

func (s *marketLegacyReadServiceStub) LoadStockKLine(stockCode string, days int64) []data.KLineData {
	s.lastStockKLineCode = stockCode
	s.lastStockKLineDays = days
	return s.stockKLine
}

func (s *marketLegacyReadServiceStub) LoadStockCommonKLine(stockCode string, days int64) []data.KLineData {
	s.lastStockCommonKLineCode = stockCode
	s.lastStockCommonKLineDays = days
	return s.stockCommonKLine
}

func (s *marketLegacyReadServiceStub) LoadStockMinutePriceLine(stockCode, stockName string) marketservice.MinutePriceLine {
	s.lastMinutePriceStockCode = stockCode
	s.lastMinutePriceStockName = stockName
	return s.minutePriceLine
}

func (s *marketLegacyReadServiceStub) LoadStockEastMoneyKLine(stockCode, klt string, limit int) []data.KLineData {
	s.lastEastMoneyStockCode = stockCode
	s.lastEastMoneyKlt = klt
	s.lastEastMoneyLimit = limit
	s.lastEastMoneyEnd = ""
	return s.eastMoneyKLine
}

func (s *marketLegacyReadServiceStub) LoadStockEastMoneyKLineResult(stockCode, klt string, limit int) marketservice.EastMoneyKLinePageResult {
	s.lastEastMoneyStockCode = stockCode
	s.lastEastMoneyKlt = klt
	s.lastEastMoneyLimit = limit
	s.lastEastMoneyEnd = ""
	return s.eastMoneyKLineResult
}

func (s *marketLegacyReadServiceStub) LoadStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) []data.KLineData {
	s.lastEastMoneyStockCode = stockCode
	s.lastEastMoneyKlt = klt
	s.lastEastMoneyLimit = limit
	s.lastEastMoneyEnd = end
	return s.eastMoneyKLinePage
}

func (s *marketLegacyReadServiceStub) LoadStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) marketservice.EastMoneyKLinePageResult {
	s.lastEastMoneyStockCode = stockCode
	s.lastEastMoneyKlt = klt
	s.lastEastMoneyLimit = limit
	s.lastEastMoneyEnd = end
	return s.eastMoneyKLinePageResult
}

func (s *marketLegacyReadServiceStub) LoadRealtimePrice(stockCode string) marketservice.RealtimePrice {
	s.lastRealtimePriceStockCode = stockCode
	return s.realtimePrice
}

func TestApp_MarketLegacyReadsDelegateToService(t *testing.T) {
	legacy := &marketLegacyReadServiceStub{
		longTigerRanks: []models.LongTigerRankData{{SECUCODE: "000001.SZ"}},
		stockResearchReports: []marketservice.StockResearchReportEntry{
			{InfoCode: "REPORT-1", StockCode: "000001"},
		},
		stockNotices: []marketservice.StockNoticeEntry{
			{ArtCode: "NOTICE-1", Title: "公告"},
		},
		industryResearchReports: []marketservice.IndustryResearchReportEntry{
			{InfoCode: "IND-1", IndustryName: "半导体"},
		},
		dictCodes: []marketservice.EMDictCodeEntry{
			{BKCode: "016001", BKName: "半导体"},
		},
		hotStocks:         []models.HotItem{{Code: "000001", Name: "平安银行"}},
		hotEvents:         []models.HotEvent{{Tag: "热点", Content: "内容"}},
		hotTopics:         []marketservice.HotTopicEntry{{Nickname: "股吧热议", HTID: "htid-1"}},
		investCalendar:    []marketservice.InvestCalendarDay{{Date: "2026-04-01"}},
		clsCalendar:       []marketservice.ClsCalendarDay{{CalendarDay: "2026-04-01"}},
		searchStockResult: marketservice.SearchStockResponse{Code: 100, Msg: "ok"},
		hotStrategyResult: models.HotStrategy{
			Code: 1,
			Data: []*models.HotStrategyData{{Question: "量价齐升", Rank: 1}},
		},
		stockKLine:       []data.KLineData{{Day: "2026-04-01", Close: "12.34"}},
		stockCommonKLine: []data.KLineData{{Day: "2026-04-02", Close: "22.34"}},
		minutePriceLine: marketservice.MinutePriceLine{
			StockCode: "sz000001",
			StockName: "平安银行",
			Date:      "2026-04-01",
			PriceData: []data.MinuteData{{Time: "09:30", Price: 12.34}},
		},
		eastMoneyKLine:     []data.KLineData{{Day: "2026-04-03", Close: "32.34"}},
		eastMoneyKLinePage: []data.KLineData{{Day: "2026-03-31", Close: "31.34"}},
		eastMoneyKLineResult: marketservice.EastMoneyKLinePageResult{
			OK:        true,
			Data:      []data.KLineData{{Day: "2026-04-03", Close: "32.34"}},
			ErrorCode: "",
		},
		eastMoneyKLinePageResult: marketservice.EastMoneyKLinePageResult{
			OK:              true,
			Data:            []data.KLineData{{Day: "2026-03-31", Close: "31.34"}},
			UsedCookieRetry: true,
		},
		realtimePrice: marketservice.RealtimePrice{
			StockCode: "sz000001",
			StockName: "平安银行",
			Price:     "12.34",
		},
	}
	app := &App{marketReadService: legacy}

	if got := app.LongTigerRank("2026-04-01"); !reflect.DeepEqual(*got, legacy.longTigerRanks) {
		t.Fatalf("LongTigerRank() mismatch, got=%#v want=%#v", got, legacy.longTigerRanks)
	}
	if got := app.StockResearchReport("sz000001"); !reflect.DeepEqual(got, legacy.stockResearchReports) {
		t.Fatalf("StockResearchReport() mismatch, got=%#v want=%#v", got, legacy.stockResearchReports)
	}
	if got := app.StockNotice("sz000001"); !reflect.DeepEqual(got, legacy.stockNotices) {
		t.Fatalf("StockNotice() mismatch, got=%#v want=%#v", got, legacy.stockNotices)
	}
	if got := app.IndustryResearchReport("016001"); !reflect.DeepEqual(got, legacy.industryResearchReports) {
		t.Fatalf("IndustryResearchReport() mismatch, got=%#v want=%#v", got, legacy.industryResearchReports)
	}
	if got := app.EMDictCode("016"); !reflect.DeepEqual(got, legacy.dictCodes) {
		t.Fatalf("EMDictCode() mismatch, got=%#v want=%#v", got, legacy.dictCodes)
	}
	if got := app.HotStock("10"); !reflect.DeepEqual(*got, legacy.hotStocks) {
		t.Fatalf("HotStock() mismatch, got=%#v want=%#v", got, legacy.hotStocks)
	}
	if got := app.HotEvent(50); !reflect.DeepEqual(*got, legacy.hotEvents) {
		t.Fatalf("HotEvent() mismatch, got=%#v want=%#v", got, legacy.hotEvents)
	}
	if got := app.HotTopic(10); !reflect.DeepEqual(got, legacy.hotTopics) {
		t.Fatalf("HotTopic() mismatch, got=%#v want=%#v", got, legacy.hotTopics)
	}
	if got := app.InvestCalendarTimeLine("2026-04"); !reflect.DeepEqual(got, legacy.investCalendar) {
		t.Fatalf("InvestCalendarTimeLine() mismatch, got=%#v want=%#v", got, legacy.investCalendar)
	}
	if got := app.ClsCalendar(); !reflect.DeepEqual(got, legacy.clsCalendar) {
		t.Fatalf("ClsCalendar() mismatch, got=%#v want=%#v", got, legacy.clsCalendar)
	}
	if got := app.SearchStock("量价齐升"); !reflect.DeepEqual(got, legacy.searchStockResult) {
		t.Fatalf("SearchStock() mismatch, got=%#v want=%#v", got, legacy.searchStockResult)
	}
	if got := app.GetHotStrategy(); !reflect.DeepEqual(got, legacy.hotStrategyResult) {
		t.Fatalf("GetHotStrategy() mismatch, got=%#v want=%#v", got, legacy.hotStrategyResult)
	}
	if got := app.GetStockKLine("sz000001", "平安银行", 365); !reflect.DeepEqual(*got, legacy.stockKLine) {
		t.Fatalf("GetStockKLine() mismatch, got=%#v want=%#v", got, legacy.stockKLine)
	}
	if got := app.GetStockCommonKLine("sz000001", "平安银行", 240); !reflect.DeepEqual(*got, legacy.stockCommonKLine) {
		t.Fatalf("GetStockCommonKLine() mismatch, got=%#v want=%#v", got, legacy.stockCommonKLine)
	}
	if got := app.GetStockMinutePriceLineData("sz000001", "平安银行"); !reflect.DeepEqual(got, legacy.minutePriceLine) {
		t.Fatalf("GetStockMinutePriceLineData() mismatch, got=%#v want=%#v", got, legacy.minutePriceLine)
	}
	if got := app.GetStockEastMoneyKLine("sz000001", "平安银行", "101", 500); !reflect.DeepEqual(*got, legacy.eastMoneyKLine) {
		t.Fatalf("GetStockEastMoneyKLine() mismatch, got=%#v want=%#v", got, legacy.eastMoneyKLine)
	}
	if got := app.GetStockEastMoneyKLineResult("sz000001", "平安银行", "101", 500); !reflect.DeepEqual(got, legacy.eastMoneyKLineResult) {
		t.Fatalf("GetStockEastMoneyKLineResult() mismatch, got=%#v want=%#v", got, legacy.eastMoneyKLineResult)
	}
	if got := app.GetStockEastMoneyKLinePage("sz000001", "平安银行", "101", 500, "20260401"); !reflect.DeepEqual(*got, legacy.eastMoneyKLinePage) {
		t.Fatalf("GetStockEastMoneyKLinePage() mismatch, got=%#v want=%#v", got, legacy.eastMoneyKLinePage)
	}
	if got := app.GetStockEastMoneyKLinePageResult("sz000001", "平安银行", "101", 500, "20260401"); !reflect.DeepEqual(got, legacy.eastMoneyKLinePageResult) {
		t.Fatalf("GetStockEastMoneyKLinePageResult() mismatch, got=%#v want=%#v", got, legacy.eastMoneyKLinePageResult)
	}
	if got := app.GetStockRealTimePrice("sz000001"); !reflect.DeepEqual(got, map[string]any{
		"code":    0,
		"message": "success",
		"price":   12.34,
		"name":    "平安银行",
	}) {
		t.Fatalf("GetStockRealTimePrice() mismatch, got=%#v", got)
	}
	if legacy.lastLongTigerDate != "2026-04-01" {
		t.Fatalf("LongTigerRank() arg mismatch, got=%q", legacy.lastLongTigerDate)
	}
	if legacy.lastStockResearchCode != "sz000001" || legacy.lastStockNoticeCode != "sz000001" {
		t.Fatalf("report args mismatch: %#v", legacy)
	}
	if legacy.lastIndustryResearchCode != "016001" || legacy.lastDictCode != "016" {
		t.Fatalf("industry args mismatch: %#v", legacy)
	}
	if legacy.lastHotStockMarketType != "10" || legacy.lastHotEventSize != 50 || legacy.lastHotTopicSize != 10 {
		t.Fatalf("hot args mismatch: %#v", legacy)
	}
	if legacy.lastInvestCalendarYearMonth != "2026-04" || legacy.lastSearchStockWords != "量价齐升" {
		t.Fatalf("calendar/search args mismatch: %#v", legacy)
	}
	if legacy.lastStockKLineCode != "sz000001" || legacy.lastStockKLineDays != 365 {
		t.Fatalf("stock kline args mismatch: %#v", legacy)
	}
	if legacy.lastStockCommonKLineCode != "sz000001" || legacy.lastStockCommonKLineDays != 240 {
		t.Fatalf("stock common kline args mismatch: %#v", legacy)
	}
	if legacy.lastMinutePriceStockCode != "sz000001" || legacy.lastMinutePriceStockName != "平安银行" {
		t.Fatalf("minute price args mismatch: %#v", legacy)
	}
	if legacy.lastEastMoneyStockCode != "sz000001" || legacy.lastEastMoneyKlt != "101" || legacy.lastEastMoneyLimit != 500 || legacy.lastEastMoneyEnd != "20260401" {
		t.Fatalf("eastmoney args mismatch: %#v", legacy)
	}
	if legacy.lastRealtimePriceStockCode != "sz000001" {
		t.Fatalf("realtime price args mismatch: %#v", legacy)
	}
}

func TestApp_GetStockRealTimePriceReturnsFailureWhenServiceHasNoUsablePrice(t *testing.T) {
	app := &App{
		marketReadService: &marketLegacyReadServiceStub{
			realtimePrice: marketservice.RealtimePrice{
				StockCode: "sz000001",
				StockName: "平安银行",
			},
		},
	}

	if got := app.GetStockRealTimePrice("sz000001"); !reflect.DeepEqual(got, map[string]any{
		"code":    -1,
		"message": "获取股票价格失败",
		"price":   0,
	}) {
		t.Fatalf("GetStockRealTimePrice() should keep legacy failure contract, got=%#v", got)
	}
}
