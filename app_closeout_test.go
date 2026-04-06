package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
	contractservice "go-stock/backend/service/contract"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"
	watchlistservice "go-stock/backend/service/watchlist"

	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type closeoutMarketServiceStub struct {
	loadFeedsResult             marketservice.FeedSet
	refreshFeedResult           marketservice.Feed
	globalIndexesResult         marketservice.IndexSet
	industryRanksResult         []marketservice.IndustryRankEntry
	globalIndexesReadableResult string
	industryMoneyRanksResult    []marketservice.IndustryMoneyRankRow
	moneyRanksResult            []marketservice.MoneyRankRow
	stockMoneyTrendResult       []marketservice.StockMoneyTrendRow
}

func (s *closeoutMarketServiceStub) LoadFeeds() marketservice.FeedSet {
	return s.loadFeedsResult
}

func (s *closeoutMarketServiceStub) RefreshFeed(source string) marketservice.Feed {
	return s.refreshFeedResult
}

func (s *closeoutMarketServiceStub) LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet {
	return s.globalIndexesResult
}

func (s *closeoutMarketServiceStub) LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return s.industryRanksResult
}

func (s *closeoutMarketServiceStub) LoadGlobalIndexesReadable(crawlTimeout uint) string {
	return s.globalIndexesReadableResult
}

func (s *closeoutMarketServiceStub) LoadIndustryMoneyRanks(fenlei, sort string) []marketservice.IndustryMoneyRankRow {
	return s.industryMoneyRanksResult
}

func (s *closeoutMarketServiceStub) LoadMoneyRanks(sort string) []marketservice.MoneyRankRow {
	return s.moneyRanksResult
}

func (s *closeoutMarketServiceStub) LoadStockMoneyTrend(stockCode string, days int) []marketservice.StockMoneyTrendRow {
	return s.stockMoneyTrendResult
}

type closeoutAnalysisServiceStub struct {
	artifact    analysisservice.ResultArtifact
	artifactErr *contractservice.UserVisibleError
}

func (s *closeoutAnalysisServiceStub) StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (s *closeoutAnalysisServiceStub) StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk {
	ch := make(chan analysisservice.StreamChunk)
	close(ch)
	return ch
}

func (s *closeoutAnalysisServiceStub) SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int) {
}

func (s *closeoutAnalysisServiceStub) GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult {
	return nil
}

func (s *closeoutAnalysisServiceStub) GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error) {
	return &models.AIResponseResultPageData{}, nil
}

func (s *closeoutAnalysisServiceStub) DeleteResult(ctx context.Context, id uint) error {
	return nil
}

func (s *closeoutAnalysisServiceStub) BatchDeleteResults(ctx context.Context, ids []uint) error {
	return nil
}

func (s *closeoutAnalysisServiceStub) GetResultArtifact(ctx context.Context, stockCode, stockName string) (analysisservice.ResultArtifact, *contractservice.UserVisibleError) {
	return s.artifact, s.artifactErr
}

type closeoutConfigServiceStub struct {
	exportResult string
}

func (s *closeoutConfigServiceStub) GetConfig(ctx context.Context) *data.SettingConfig {
	return &data.SettingConfig{Settings: &data.Settings{}, AiConfigs: []*data.AIConfig{}}
}

func (s *closeoutConfigServiceStub) UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string {
	return "更新成功"
}

func (s *closeoutConfigServiceStub) GetAiConfigs(ctx context.Context) []*data.AIConfig {
	return []*data.AIConfig{}
}

func (s *closeoutConfigServiceStub) GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate {
	empty := []models.PromptTemplate{}
	return &empty
}

func (s *closeoutConfigServiceStub) GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error) {
	return &models.PromptTemplatePageData{}, nil
}

func (s *closeoutConfigServiceStub) SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string {
	return "模板已保存"
}

func (s *closeoutConfigServiceStub) DeletePromptTemplate(ctx context.Context, id uint) string {
	return "模板已删除"
}

func (s *closeoutConfigServiceStub) SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string {
	return "旧版Prompt已保存"
}

func (s *closeoutConfigServiceStub) DeleteLegacyPrompt(ctx context.Context, id uint) string {
	return "旧版Prompt已删除"
}

func (s *closeoutConfigServiceStub) ExportConfig(ctx context.Context) string {
	return s.exportResult
}

type closeoutNotificationServiceStub struct {
	sendResult    string
	typedResult   notificationservice.Delivery
	lastMessage   string
	lastStockCode string
	lastType      int
}

func (s *closeoutNotificationServiceStub) SendDingTalk(message, stockCode string) string {
	s.lastMessage = message
	s.lastStockCode = stockCode
	return s.sendResult
}

func (s *closeoutNotificationServiceStub) SendTyped(message, stockCode string, msgType int) notificationservice.Delivery {
	s.lastMessage = message
	s.lastStockCode = stockCode
	s.lastType = msgType
	return s.typedResult
}

type closeoutWatchlistServiceStub struct {
	saveCronText  string
	saveStockCode string
	saveResult    watchlistservice.ScheduledStock
	listResult    []watchlistservice.ScheduledStock
	runStockCode  string
	runResult     watchlistservice.ScheduledStock
}

func (s *closeoutWatchlistServiceStub) SaveStockAICron(ctx context.Context, cronText, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	s.saveCronText = cronText
	s.saveStockCode = stockCode
	return s.saveResult, nil
}

func (s *closeoutWatchlistServiceStub) ListScheduledStocks(ctx context.Context) []watchlistservice.ScheduledStock {
	return append([]watchlistservice.ScheduledStock(nil), s.listResult...)
}

func (s *closeoutWatchlistServiceStub) RunScheduledAnalysis(ctx context.Context, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError) {
	s.runStockCode = stockCode
	return s.runResult, nil
}

func TestApp_CloseoutMarketHelpersDelegateToMarketService(t *testing.T) {
	telegraphItem := &models.Telegraph{Title: "telegraph"}
	market := &closeoutMarketServiceStub{
		loadFeedsResult: marketservice.FeedSet{
			Telegraph: []*models.Telegraph{telegraphItem},
		},
		refreshFeedResult: marketservice.Feed{
			Source: "新浪财经",
			Items:  []*models.Telegraph{{Title: "refreshed"}},
		},
		globalIndexesResult: marketservice.IndexSet{
			Common: []marketservice.GlobalIndexEntry{{Code: "000001"}},
		},
		industryRanksResult:         []marketservice.IndustryRankEntry{{BoardName: "机器人"}},
		globalIndexesReadableResult: "全球指数文本",
		industryMoneyRanksResult:    []marketservice.IndustryMoneyRankRow{{Name: "机器人", NetAmount: 1200}},
		moneyRanksResult:            []marketservice.MoneyRankRow{{Symbol: "600519"}},
		stockMoneyTrendResult:       []marketservice.StockMoneyTrendRow{{OpenDate: "2026-04-06"}},
	}
	app := &App{marketReadService: market}

	if got := app.GetTelegraphList("财联社电报"); len(*got) != 1 || (*got)[0] != telegraphItem {
		t.Fatalf("unexpected telegraph delegation: %#v", got)
	}
	if got := app.ReFleshTelegraphList("新浪财经"); len(*got) != 1 || (*got)[0].Title != "refreshed" {
		t.Fatalf("unexpected refreshed feed: %#v", got)
	}
	if got := app.GlobalStockIndexes(); !reflect.DeepEqual(got, market.globalIndexesResult) {
		t.Fatalf("unexpected global indexes: %#v", got)
	}
	if got := app.GlobalStockIndexesReadable(); got != "全球指数文本" {
		t.Fatalf("unexpected readable indexes: %q", got)
	}
	if got := app.GetIndustryRank("netamount", 10); !reflect.DeepEqual(got, market.industryRanksResult) {
		t.Fatalf("unexpected industry ranks: %#v", got)
	}
	if got := app.GetIndustryMoneyRankSina("0", "netamount"); !reflect.DeepEqual(got, market.industryMoneyRanksResult) {
		t.Fatalf("unexpected industry money ranks: %#v", got)
	}
	if got := app.GetMoneyRankSina("netamount"); !reflect.DeepEqual(got, market.moneyRanksResult) {
		t.Fatalf("unexpected money ranks: %#v", got)
	}
	if got := app.GetStockMoneyTrendByDay("600519", 20); !reflect.DeepEqual(got, market.stockMoneyTrendResult) {
		t.Fatalf("unexpected stock money trend: %#v", got)
	}
}

func TestApp_ShareAnalysisUsesArtifactUploader(t *testing.T) {
	var uploaded analysisservice.ResultArtifact
	app := &App{
		ctx: context.Background(),
		analysisService: &closeoutAnalysisServiceStub{
			artifact: analysisservice.ResultArtifact{
				StockCode:    "000001.SZ",
				StockName:    "平安银行",
				Content:      "analysis content",
				AnalysisDate: "2026/04/06",
			},
		},
		shareAnalysisUploader: func(artifact analysisservice.ResultArtifact) (string, error) {
			uploaded = artifact
			return "分享成功", nil
		},
	}

	if got := app.ShareAnalysis("000001.SZ", "平安银行"); got != "分享成功" {
		t.Fatalf("unexpected share result: %q", got)
	}
	if uploaded.StockCode != "000001.SZ" || uploaded.StockName != "平安银行" || uploaded.AnalysisDate != "2026/04/06" {
		t.Fatalf("unexpected uploaded artifact: %#v", uploaded)
	}
}

func TestApp_SaveAsMarkdownUsesArtifactAndWriter(t *testing.T) {
	var savedPath string
	var savedContent string
	var savedPerm os.FileMode
	app := &App{
		ctx: context.Background(),
		analysisService: &closeoutAnalysisServiceStub{
			artifact: analysisservice.ResultArtifact{
				StockCode:        "000001.SZ",
				StockName:        "平安银行",
				Content:          "markdown body",
				MarkdownFilename: "平安银行[000001.SZ]AI分析结果_2026-04-06_09_30_00.md",
			},
		},
		saveFileDialog: func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
			return "D:\\temp\\analysis.md", nil
		},
		writeFile: func(name string, data []byte, perm os.FileMode) error {
			savedPath = name
			savedContent = string(data)
			savedPerm = perm
			return nil
		},
	}

	if got := app.SaveAsMarkdown("000001.SZ", "平安银行"); got != "已保存至：D:\\temp\\analysis.md" {
		t.Fatalf("unexpected markdown save result: %q", got)
	}
	if savedPath != "D:\\temp\\analysis.md" || savedContent != "markdown body" || savedPerm != 0644 {
		t.Fatalf("unexpected markdown write: path=%q content=%q perm=%v", savedPath, savedContent, savedPerm)
	}
}

func TestApp_ExportConfigUsesConfigServiceExport(t *testing.T) {
	var savedPath string
	var savedContent string
	app := &App{
		ctx: context.Background(),
		configService: &closeoutConfigServiceStub{
			exportResult: "{\"darkTheme\":true}",
		},
		saveFileDialog: func(ctx context.Context, options runtime.SaveDialogOptions) (string, error) {
			return "D:\\temp\\config.json", nil
		},
		writeFile: func(name string, data []byte, perm os.FileMode) error {
			savedPath = name
			savedContent = string(data)
			return nil
		},
	}

	if got := app.ExportConfig(); got != "导出成功:D:\\temp\\config.json" {
		t.Fatalf("unexpected export result: %q", got)
	}
	if savedPath != "D:\\temp\\config.json" || savedContent != "{\"darkTheme\":true}" {
		t.Fatalf("unexpected export write: path=%q content=%q", savedPath, savedContent)
	}
}

func TestApp_NotificationHelpersDelegateToService(t *testing.T) {
	notification := &closeoutNotificationServiceStub{
		sendResult: "普通通知成功",
		typedResult: notificationservice.Delivery{
			DingResult:   "类型通知成功",
			EventTitle:   "涨跌报警",
			EventContent: "[平安银行] 12.34 2.83% 2026-04-06 10:00:00",
		},
	}
	var eventName string
	var eventPayload any
	app := &App{
		ctx:                 context.Background(),
		notificationService: notification,
		emitEvent: func(ctx context.Context, name string, data ...interface{}) {
			eventName = name
			if len(data) > 0 {
				eventPayload = data[0]
			}
		},
	}

	if got := app.SendDingDingMessage("body-1", "sz000001"); got != "普通通知成功" {
		t.Fatalf("unexpected normal notification result: %q", got)
	}
	if notification.lastMessage != "body-1" || notification.lastStockCode != "sz000001" {
		t.Fatalf("unexpected normal notification args: %#v", notification)
	}

	if got := app.SendDingDingMessageByType("body-2", "demo001", 1); got != "类型通知成功" {
		t.Fatalf("unexpected typed notification result: %q", got)
	}
	if eventName != "newsPush" {
		t.Fatalf("expected newsPush event, got %q", eventName)
	}
	payload, ok := eventPayload.(map[string]any)
	if !ok || payload["content"] != "[平安银行] 12.34 2.83% 2026-04-06 10:00:00" {
		t.Fatalf("unexpected event payload: %#v", eventPayload)
	}
}

func TestApp_SetStockAICronRegistersAndRestoresJobsThroughWatchlistService(t *testing.T) {
	cronScheduler := cron.New(cron.WithSeconds())
	cronScheduler.Start()
	t.Cleanup(func() { cronScheduler.Stop() })

	watchlist := &closeoutWatchlistServiceStub{
		saveResult: watchlistservice.ScheduledStock{
			StockCode:  "usaapl",
			Name:       "Apple",
			Cron:       "0 */5 * * * *",
			AIConfigID: 7,
		},
		listResult: []watchlistservice.ScheduledStock{
			{StockCode: "sz000001", Name: "平安银行", Cron: "0 */10 * * * *", AIConfigID: 3},
		},
		runResult: watchlistservice.ScheduledStock{
			StockCode:  "usaapl",
			Name:       "Apple",
			Cron:       "0 */5 * * * *",
			AIConfigID: 7,
		},
	}
	var events []string
	app := &App{
		ctx:              context.Background(),
		cron:             cronScheduler,
		cronEntrys:       make(map[string]cron.EntryID),
		watchlistService: watchlist,
		emitEvent: func(ctx context.Context, name string, data ...interface{}) {
			if len(data) == 0 {
				return
			}
			if text, ok := data[0].(string); ok {
				events = append(events, text)
			}
		},
	}

	app.SetStockAICron("0 */5 * * * *", "gb_aapl")
	if watchlist.saveCronText != "0 */5 * * * *" || watchlist.saveStockCode != "gb_aapl" {
		t.Fatalf("unexpected watchlist save args: %#v", watchlist)
	}
	if _, exists := app.getCronEntry("usaapl"); !exists {
		t.Fatal("expected usaapl cron entry after SetStockAICron")
	}

	app.restoreStockAICronSchedules()
	if _, exists := app.getCronEntry("sz000001"); !exists {
		t.Fatal("expected restored sz000001 cron entry")
	}

	job := app.buildStockAICronJob("usaapl")
	job()
	if watchlist.runStockCode != "usaapl" {
		t.Fatalf("unexpected scheduled run stock code: %q", watchlist.runStockCode)
	}
	if len(events) < 2 || events[0] != "开始自动分析Apple_usaapl" || events[1] != "AI分析完成：Apple_usaapl" {
		t.Fatalf("unexpected event log: %#v", events)
	}
}
