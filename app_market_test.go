package main

import (
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
)

type marketReadServiceMock struct {
	loadFeedsResult         marketservice.FeedSet
	refreshFeedResult       marketservice.Feed
	globalIndexesResult     marketservice.IndexSet
	industryRanksResult     []marketservice.IndustryRankEntry
	lastGlobalIndexesTimout uint
	lastRefreshFeedSource   string
	lastIndustryRankSort    string
	lastIndustryRankCount   int
	loadFeedsCalled         int
	refreshFeedCalled       int
	loadGlobalIndexesCalled int
	loadIndustryRanksCalled int
}

func (m *marketReadServiceMock) LoadFeeds() marketservice.FeedSet {
	m.loadFeedsCalled++
	return m.loadFeedsResult
}

func (m *marketReadServiceMock) RefreshFeed(source string) marketservice.Feed {
	m.refreshFeedCalled++
	m.lastRefreshFeedSource = source
	return m.refreshFeedResult
}

func (m *marketReadServiceMock) RefreshAllFeeds() marketservice.FeedSet {
	return m.loadFeedsResult
}

func (m *marketReadServiceMock) LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet {
	m.loadGlobalIndexesCalled++
	m.lastGlobalIndexesTimout = crawlTimeout
	return m.globalIndexesResult
}

func (m *marketReadServiceMock) LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	m.loadIndustryRanksCalled++
	m.lastIndustryRankSort = sort
	m.lastIndustryRankCount = cnt
	return m.industryRanksResult
}

func (m *marketReadServiceMock) LoadStockList(keyword string) []data.StockBasic {
	return []data.StockBasic{}
}

func (m *marketReadServiceMock) SaveNtfyNews(news models.NtfyNews) (*models.Telegraph, bool) {
	return nil, false
}

func TestApp_MarketReadMethodsDelegateToService(t *testing.T) {
	telegraphItem := &models.Telegraph{Title: "telegraph"}
	sinaItem := &models.Telegraph{Title: "sina"}
	foreignItem := &models.Telegraph{Title: "foreign"}

	expectedFeeds := marketservice.FeedSet{
		Telegraph: []*models.Telegraph{telegraphItem},
		Sina:      []*models.Telegraph{sinaItem},
		Foreign:   []*models.Telegraph{foreignItem},
	}
	expectedRefreshFeed := marketservice.Feed{
		Source: "财联社电报",
		Items:  []*models.Telegraph{telegraphItem},
	}
	expectedIndexes := marketservice.IndexSet{
		Common: []marketservice.GlobalIndexEntry{
			{Code: "000001", Name: "上证指数", Region: "common"},
		},
		America: []marketservice.GlobalIndexEntry{
			{Code: "DJI", Name: "道琼斯", Region: "america"},
		},
	}
	expectedIndustryRanks := []marketservice.IndustryRankEntry{
		{BoardCode: "BK001", BoardName: "算力"},
		{BoardCode: "BK002", BoardName: "机器人"},
	}

	mockService := &marketReadServiceMock{
		loadFeedsResult:     expectedFeeds,
		refreshFeedResult:   expectedRefreshFeed,
		globalIndexesResult: expectedIndexes,
		industryRanksResult: expectedIndustryRanks,
	}
	app := &App{
		marketReadService: mockService,
	}

	gotFeeds := app.GetMarketFeeds()
	if !reflect.DeepEqual(gotFeeds, expectedFeeds) {
		t.Fatalf("GetMarketFeeds() mismatch, got=%#v want=%#v", gotFeeds, expectedFeeds)
	}

	gotRefreshFeed := app.RefreshMarketFeed("财联社电报")
	if !reflect.DeepEqual(gotRefreshFeed, expectedRefreshFeed) {
		t.Fatalf("RefreshMarketFeed() mismatch, got=%#v want=%#v", gotRefreshFeed, expectedRefreshFeed)
	}
	if mockService.lastRefreshFeedSource != "财联社电报" {
		t.Fatalf("RefreshMarketFeed() source mismatch, got=%q want=%q", mockService.lastRefreshFeedSource, "财联社电报")
	}

	gotIndexes := app.GetMarketGlobalIndexes()
	if !reflect.DeepEqual(gotIndexes, expectedIndexes) {
		t.Fatalf("GetMarketGlobalIndexes() mismatch, got=%#v want=%#v", gotIndexes, expectedIndexes)
	}
	if mockService.lastGlobalIndexesTimout != 30 {
		t.Fatalf("GetMarketGlobalIndexes() timeout mismatch, got=%d want=%d", mockService.lastGlobalIndexesTimout, 30)
	}

	gotRanks := app.GetMarketIndustryRanks("zdf", 30)
	if !reflect.DeepEqual(gotRanks, expectedIndustryRanks) {
		t.Fatalf("GetMarketIndustryRanks() mismatch, got=%#v want=%#v", gotRanks, expectedIndustryRanks)
	}
	if mockService.lastIndustryRankSort != "zdf" || mockService.lastIndustryRankCount != 30 {
		t.Fatalf("GetMarketIndustryRanks() args mismatch, got sort=%q cnt=%d", mockService.lastIndustryRankSort, mockService.lastIndustryRankCount)
	}

	if mockService.loadFeedsCalled != 1 {
		t.Fatalf("LoadFeeds should be called exactly once, got=%d", mockService.loadFeedsCalled)
	}
	if mockService.refreshFeedCalled != 1 {
		t.Fatalf("RefreshFeed should be called exactly once, got=%d", mockService.refreshFeedCalled)
	}
	if mockService.loadGlobalIndexesCalled != 1 {
		t.Fatalf("LoadGlobalIndexes should be called exactly once, got=%d", mockService.loadGlobalIndexesCalled)
	}
	if mockService.loadIndustryRanksCalled != 1 {
		t.Fatalf("LoadIndustryRanks should be called exactly once, got=%d", mockService.loadIndustryRanksCalled)
	}
}
