package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go-stock/backend/data"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	analysisservice "go-stock/backend/service/analysis"
	configservice "go-stock/backend/service/config"
	contractservice "go-stock/backend/service/contract"
	fundservice "go-stock/backend/service/fund"
	marketservice "go-stock/backend/service/market"
	notificationservice "go-stock/backend/service/notification"
	researchservice "go-stock/backend/service/research"
	taskservice "go-stock/backend/service/task"
	watchlistservice "go-stock/backend/service/watchlist"
	analysissource "go-stock/backend/source/analysis"
	fundsource "go-stock/backend/source/fund"
	marketsource "go-stock/backend/source/marketnews"
	notificationsource "go-stock/backend/source/notification"
	researchsource "go-stock/backend/source/research"
	watchlistsource "go-stock/backend/source/watchlist"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/go-update"
	"github.com/samber/lo"

	"github.com/PuerkitoBio/goquery"
	"github.com/coocood/freecache"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// App struct
type App struct {
	ctx                   context.Context
	cache                 *freecache.Cache
	cron                  *cron.Cron
	cronEntrys            map[string]cron.EntryID
	cronEntrysMu          sync.Mutex
	AiTools               []data.Tool
	SponsorInfo           map[string]any
	VipLevel              int64
	summaryMu             sync.Mutex
	summaryCancel         context.CancelFunc
	agentMu               sync.Mutex
	agentCancel           context.CancelFunc
	stockAlertMu          sync.Mutex
	stockAlertLastSent    map[string]time.Time
	priceAtAlertReset     map[string]float64
	marketReadService     marketReadService
	analysisService       analysisService
	configService         configService
	taskService           taskService
	notificationService   notificationService
	watchlistService      watchlistService
	researchService       researchService
	fundService           fundService
	saveFileDialog        func(ctx context.Context, options runtime.SaveDialogOptions) (string, error)
	writeFile             func(name string, data []byte, perm os.FileMode) error
	shareAnalysisUploader func(artifact analysisservice.ResultArtifact) (string, error)
	emitEvent             func(ctx context.Context, name string, data ...interface{})
}

type marketReadService interface {
	LoadFeeds() marketservice.FeedSet
	RefreshFeed(source string) marketservice.Feed
	RefreshAllFeeds() marketservice.FeedSet
	LoadGlobalIndexes(crawlTimeout uint) marketservice.IndexSet
	LoadIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry
	LoadStockList(keyword string) []data.StockBasic
	SaveNtfyNews(news models.NtfyNews) (*models.Telegraph, bool)
}

type analysisService interface {
	StartStockAnalysis(ctx context.Context, request analysisservice.StockRequest) <-chan analysisservice.StreamChunk
	StartMarketSummary(ctx context.Context, request analysisservice.MarketSummaryRequest) <-chan analysisservice.StreamChunk
	SaveResult(ctx context.Context, stockCode, stockName, result, chatID, question string, aiConfigID int)
	GetLatestResult(ctx context.Context, stockCode string) *models.AIResponseResult
	GetResultPage(ctx context.Context, query models.AIResponseResultQuery) (*models.AIResponseResultPageData, error)
	DeleteResult(ctx context.Context, id uint) error
	BatchDeleteResults(ctx context.Context, ids []uint) error
}

type configService interface {
	GetConfig(ctx context.Context) *data.SettingConfig
	UpdateConfig(ctx context.Context, cfg *data.SettingConfig) string
	GetAiConfigs(ctx context.Context) []*data.AIConfig
	GetPromptTemplates(ctx context.Context, name, promptType string) *[]models.PromptTemplate
	GetPromptTemplatePage(ctx context.Context, query models.PromptTemplateQuery) (*models.PromptTemplatePageData, error)
	SavePromptTemplate(ctx context.Context, template models.PromptTemplate) string
	DeletePromptTemplate(ctx context.Context, id uint) string
	SaveLegacyPrompt(ctx context.Context, prompt models.Prompt) string
	DeleteLegacyPrompt(ctx context.Context, id uint) string
}

type taskService interface {
	Create(ctx context.Context, task *models.CronTask) string
	Update(ctx context.Context, task *models.CronTask) string
	Delete(ctx context.Context, id uint) string
	GetByID(ctx context.Context, id uint) (*models.CronTask, error)
	List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp
	Enable(ctx context.Context, id uint, enable bool) string
	RunNow(ctx context.Context, id uint) error
	GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string]
	ValidateCronExpr(ctx context.Context, expr string) string
	Search(ctx context.Context, keyword string) []models.CronTask
	CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time
	CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time
	RestoreSchedules(ctx context.Context) error
}

type marketResidualReadService interface {
	LoadGlobalIndexesReadable(crawlTimeout uint) string
	LoadIndustryMoneyRanks(fenlei, sort string) []marketservice.IndustryMoneyRankRow
	LoadMoneyRanks(sort string) []marketservice.MoneyRankRow
	LoadStockMoneyTrend(stockCode string, days int) []marketservice.StockMoneyTrendRow
}

type marketLegacyReadService interface {
	LoadLongTiger(date string) []models.LongTigerRankData
	LoadStockResearchReports(stockCode string) []marketservice.StockResearchReportEntry
	LoadStockNotices(stockCode string) []marketservice.StockNoticeEntry
	LoadIndustryResearchReports(industryCode string) []marketservice.IndustryResearchReportEntry
	LoadEMDictCodes(code string) []marketservice.EMDictCodeEntry
	LoadHotStocks(marketType string) []models.HotItem
	LoadHotEvents(size int) []models.HotEvent
	LoadHotTopics(size int) []marketservice.HotTopicEntry
	LoadInvestCalendar(yearMonth string) []marketservice.InvestCalendarDay
	LoadClsCalendar() []marketservice.ClsCalendarDay
	SearchStocks(words string) marketservice.SearchStockResponse
	LoadHotStrategies() models.HotStrategy
	LoadStockKLine(stockCode string, days int64) []data.KLineData
	LoadStockCommonKLine(stockCode string, days int64) []data.KLineData
	LoadStockMinutePriceLine(stockCode, stockName string) marketservice.MinutePriceLine
	LoadStockEastMoneyKLine(stockCode, klt string, limit int) []data.KLineData
	LoadStockEastMoneyKLineResult(stockCode, klt string, limit int) marketservice.EastMoneyKLinePageResult
	LoadStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) []data.KLineData
	LoadStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) marketservice.EastMoneyKLinePageResult
	LoadRealtimePrice(stockCode string) marketservice.RealtimePrice
}

type analysisArtifactService interface {
	GetResultArtifact(ctx context.Context, stockCode, stockName string) (analysisservice.ResultArtifact, *contractservice.UserVisibleError)
}

type configExportService interface {
	ExportConfig(ctx context.Context) string
}

type notificationService interface {
	SendDingTalk(message, stockCode string) string
	SendTyped(message, stockCode string, msgType int) notificationservice.Delivery
}

type watchlistService interface {
	SaveStockAICron(ctx context.Context, cronText, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError)
	ListScheduledStocks(ctx context.Context) []watchlistservice.ScheduledStock
	RunScheduledAnalysis(ctx context.Context, stockCode string) (watchlistservice.ScheduledStock, *contractservice.UserVisibleError)
	Follow(ctx context.Context, stockCode string) string
	Unfollow(ctx context.Context, stockCode string) string
	GetFollowList(ctx context.Context, groupID int) []data.FollowedStock
	GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock
	GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice
	SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string
	SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string
	SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string
	SetStockSort(ctx context.Context, stockCode string, sort int64)
	UpdateObservedPrice(ctx context.Context, stockCode string, price float64)
	ListGroups(ctx context.Context) []data.Group
	AddGroup(ctx context.Context, group data.Group) string
	UpdateGroupSort(ctx context.Context, id int, newSort int) bool
	InitializeGroupSort(ctx context.Context) bool
	ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock
	AddGroupStock(ctx context.Context, groupID int, stockCode string) string
	RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) string
	RemoveGroup(ctx context.Context, groupID int) string
	EvaluateCostAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery
}

type researchService interface {
	LoadAllStocks(ctx context.Context, page, pageSize int, name string, technicalIndicators models.TechnicalIndicators) *models.AllStocksResp
	SyncAllStockInfo(ctx context.Context) error
	RefreshStockBaseInfo(ctx context.Context) error
	GetStockChanges(ctx context.Context, changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse
	GetAllStockChangesWithPaging(ctx context.Context, pageSize int) *data.StockChangesResponse
	GetStockChangeHistory(ctx context.Context, query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData
	SaveStockChangesToHistory(ctx context.Context, changeTypes []int) string
	DeleteStockChangeHistory(ctx context.Context, days int) string
	GetAiRecommendPage(ctx context.Context, query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData
	DeleteAiRecommend(ctx context.Context, id uint) string
	SetAiRecommendAlert(ctx context.Context, id uint, enable bool) string
	GetAllStockInfoPage(ctx context.Context, query data.AllStockInfoQuery) *data.AllStockInfoPageData
	GetAllStockInfoByID(ctx context.Context, id uint) *models.AllStockInfo
	AddAllStockInfo(ctx context.Context, stock models.AllStockInfo) string
	DeleteAllStockInfo(ctx context.Context, id uint) string
	BatchDeleteAllStockInfo(ctx context.Context, ids []uint) string
	GetAllMarkets(ctx context.Context) []string
	GetAllIndustries(ctx context.Context) []string
	GetAllConcepts(ctx context.Context) []string
	GetTradingRecordList(ctx context.Context, query data.TradingRecordListQuery) *data.TradingRecordPageData
	AddTradingRecord(ctx context.Context, record data.TradingRecord) (uint, error)
	GetTradingRecordByID(ctx context.Context, id uint) (*data.TradingRecord, error)
	GetTradingRecordStatistics(ctx context.Context) *data.TradingRecordStatistics
	UpdateTradingRecord(ctx context.Context, record data.TradingRecord) error
	DeleteTradingRecord(ctx context.Context, id uint) error
	CheckFrequentTrading(ctx context.Context, stockCode string) researchservice.FrequentTradingCheck
	EvaluateAiRecommendAlerts(ctx context.Context, now time.Time) []notificationservice.Delivery
}

type fundService interface {
	LoadFundList(ctx context.Context, key string) []data.FundBasic
	LoadFollowedFunds(ctx context.Context) []data.FollowedFund
	FollowFund(ctx context.Context, fundCode string) string
	UnfollowFund(ctx context.Context, fundCode string) string
	RefreshFollowedFunds(ctx context.Context) error
	SyncAllFunds(ctx context.Context)
}

func (a *App) residualMarketReads() marketResidualReadService {
	if a.marketReadService == nil {
		return nil
	}
	service, _ := a.marketReadService.(marketResidualReadService)
	return service
}

func (a *App) legacyMarketReads() marketLegacyReadService {
	if a.marketReadService == nil {
		return nil
	}
	service, _ := a.marketReadService.(marketLegacyReadService)
	return service
}

func (a *App) artifactService() analysisArtifactService {
	if a.analysisService == nil {
		return nil
	}
	service, _ := a.analysisService.(analysisArtifactService)
	return service
}

func (a *App) exportConfigSource() configExportService {
	if a.configService == nil {
		return nil
	}
	service, _ := a.configService.(configExportService)
	return service
}

const (
	// 兼容保留：当前版本全部功能开放，但旧前端仍依赖 VIP 结构字段。
	unlockedVipLevel      = 2
	unlockedVipLevelText  = "2"
	unlockedVipStartTime  = "2024-01-01 00:00:00"
	unlockedVipEndTime    = "2099-12-31 23:59:59"
	unlockedSponsorNotice = "当前版本已开放全部功能，无需赞助码。"
)

// NewApp creates a new App application struct
func NewApp() *App {
	cacheSize := 512 * 1024
	cache := freecache.NewCache(cacheSize)
	c := cron.New(cron.WithSeconds())
	c.Start()
	var tools []data.Tool
	tools = data.Tools(tools)
	analysisProvider := analysissource.NewProvider(tools)
	analysisStore := analysissource.NewStore()
	analysisSvc := analysisservice.NewService(analysisProvider, analysisStore, analysisStore)
	configSvc := configservice.NewService(configservice.NewStore())
	researchSvc := researchservice.NewService(researchsource.NewStore())
	fundSvc := fundservice.NewService(fundsource.NewAdapter())
	app := &App{
		cache:                 cache,
		cron:                  c,
		cronEntrys:            make(map[string]cron.EntryID),
		AiTools:               tools,
		stockAlertLastSent:    make(map[string]time.Time),
		priceAtAlertReset:     make(map[string]float64),
		marketReadService:     marketservice.NewService(marketsource.NewSource()),
		analysisService:       analysisSvc,
		configService:         configSvc,
		notificationService:   notificationservice.NewService(cache, notificationsource.NewAdapter()),
		watchlistService:      watchlistservice.NewService(watchlistsource.NewStore(), analysisSvc),
		researchService:       researchSvc,
		fundService:           fundSvc,
		saveFileDialog:        runtime.SaveFileDialog,
		writeFile:             os.WriteFile,
		shareAnalysisUploader: uploadSharedAnalysis,
		emitEvent:             runtime.EventsEmit,
	}
	app.taskService = taskservice.NewService(taskservice.NewStore(), &appTaskScheduler{app: app}, func() context.Context {
		return app.ctx
	})
	return app
}

func appCoreLog() *logger.Logger {
	return appLifecycleLogger("app.core")
}

func appTrace(source string) logger.TraceContext {
	return appLifecycleTrace(source)
}

func appInfo(source, event, message string, fields ...zap.Field) {
	if log := appCoreLog(); log != nil {
		log.WithTrace(appTrace(source)).Info(event, message, fields...)
	}
}

func appWarn(source, event, message string, fields ...zap.Field) {
	if log := appCoreLog(); log != nil {
		log.WithTrace(appTrace(source)).Warn(event, message, fields...)
	}
}

func appError(source, event, message string, fields ...zap.Field) {
	if log := appCoreLog(); log != nil {
		log.WithTrace(appTrace(source)).Error(event, message, fields...)
	}
}

func (a *App) setCronEntry(key string, id cron.EntryID) {
	a.cronEntrysMu.Lock()
	a.cronEntrys[key] = id
	a.cronEntrysMu.Unlock()
}

func (a *App) getCronEntry(key string) (cron.EntryID, bool) {
	a.cronEntrysMu.Lock()
	id, exists := a.cronEntrys[key]
	a.cronEntrysMu.Unlock()
	return id, exists
}

func (a *App) removeCronEntry(key string) {
	a.cronEntrysMu.Lock()
	delete(a.cronEntrys, key)
	a.cronEntrysMu.Unlock()
}

func (a *App) GetSponsorInfo() map[string]any {
	// 兼容保留：返回稳定结构，避免旧前端读取赞助信息时出现空值分支。
	return map[string]any{
		"vipLevel":     unlockedVipLevelText,
		"vipStartTime": unlockedVipStartTime,
		"vipEndTime":   unlockedVipEndTime,
		"active":       true,
		"message":      unlockedSponsorNotice,
	}
}

// GetEffectiveSponsorVip 从本地配置解密赞助信息并判断当前是否在 VIP 有效期内（与 ai-assistant-web / data.EffectiveSponsorVipLevel 一致）。
func (a *App) GetEffectiveSponsorVip() map[string]any {
	return map[string]any{
		"vipLevel": unlockedVipLevel,
		"active":   true,
		"message":  unlockedSponsorNotice,
	}
}
func (a *App) CheckSponsorCode(sponsorCode string) map[string]any {
	return map[string]any{
		"code": 1,
		"msg":  unlockedSponsorNotice,
	}
}

func (a *App) CheckUpdate(flag int) {
	releaseVersion := &models.GitHubReleaseVersion{}
	_, err := resty.New().R().
		SetResult(releaseVersion).
		Get("https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest")
	if err != nil {
		appError("check-update", "update.release_fetch_failed", "get github release version failed", logger.Err(err))
		return
	}
	//logger.SugaredLogger.Infof("releaseVersion:%+v", releaseVersion.TagName)

	if _, vipLevel, ok := a.isVip("", "", releaseVersion); ok {
		level, _ := convertor.ToInt(vipLevel)
		a.VipLevel = level
		if level >= 2 {
			go a.syncNews()
		}
	}

	if releaseVersion.TagName != Version {
		tag := &models.Tag{}
		_, err = resty.New().R().
			SetResult(tag).
			Get("https://api.github.com/repos/ArvinLovegood/go-stock/git/ref/tags/" + releaseVersion.TagName)
		if err == nil {
			releaseVersion.Tag = *tag
		}

		commit := &models.Commit{}
		_, err = resty.New().R().
			SetResult(commit).
			Get(tag.Object.Url)
		if err == nil {
			releaseVersion.Commit = *commit
		}

		// 构建下载链接
		downloadUrl := fmt.Sprintf("https://github.com/ArvinLovegood/go-stock/releases/download/%s/go-stock-windows-amd64.exe", releaseVersion.TagName)
		if IsMacOS() {
			downloadUrl = fmt.Sprintf("https://github.com/ArvinLovegood/go-stock/releases/download/%s/go-stock-darwin-universal", releaseVersion.TagName)
		} else if IsLinux() {
			downloadUrl = fmt.Sprintf("https://github.com/ArvinLovegood/go-stock/releases/download/%s/go-stock-linux-amd64", releaseVersion.TagName)
		}
		downloadUrl, _, done := a.isVip("", downloadUrl, releaseVersion)
		if !done {
			return
		}
		go runtime.EventsEmit(a.ctx, "newsPush", map[string]any{
			"time":    "发现新版本：" + releaseVersion.TagName,
			"isRed":   true,
			"source":  "go-stock",
			"content": fmt.Sprintf("%s", commit.Message),
		})
		resp, err := resty.New().R().Get(downloadUrl)
		if err != nil {
			go runtime.EventsEmit(a.ctx, "newsPush", map[string]any{
				"time":    "新版本：" + releaseVersion.TagName,
				"isRed":   true,
				"source":  "go-stock",
				"content": commit.Message + "\n新版本下载失败,请稍后重试或请前往 https://github.com/ArvinLovegood/go-stock/releases 手动下载替换文件。",
			})
			return
		}
		body := resp.Body()

		if len(body) < 1024*500 {
			go runtime.EventsEmit(a.ctx, "newsPush", map[string]any{
				"time":    "新版本：" + releaseVersion.TagName,
				"isRed":   true,
				"source":  "go-stock",
				"content": commit.Message + "\n新版本下载失败,请稍后重试或请前往 https://github.com/ArvinLovegood/go-stock/releases 手动下载替换文件。",
			})
			return
		}

		err = update.Apply(bytes.NewReader(body), update.Options{})
		if err != nil {
			appError("check-update", "update.apply_failed", "apply update failed", logger.Err(err))
			go runtime.EventsEmit(a.ctx, "updateVersion", releaseVersion)
			return
		} else {
			go runtime.EventsEmit(a.ctx, "newsPush", map[string]any{
				"time":    "新版本：" + releaseVersion.TagName,
				"isRed":   true,
				"source":  "go-stock",
				"content": "版本更新完成,下次重启软件生效.",
			})
		}
	} else {
		if flag == 1 {
			go runtime.EventsEmit(a.ctx, "newsPush", map[string]any{
				"time":    "当前版本：" + Version,
				"isRed":   true,
				"source":  "go-stock",
				"content": "当前版本无更新",
			})
		}

	}
}

func (a *App) isVip(sponsorCode string, downloadUrl string, releaseVersion *models.GitHubReleaseVersion) (string, string, bool) {
	// 兼容保留：旧调用方仍会通过该方法读取 VIP 结果，但当前版本不再按赞助信息分流。
	return downloadUrl, unlockedVipLevelText, true
}

func (a *App) syncNews() {
	defer PanicHandler()
	client := resty.New()
	url := fmt.Sprintf("http://go-stock.sparkmemory.top:16666/FinancialNews/json?since=%d", time.Now().Add(-24*time.Hour).Unix())
	//logger.SugaredLogger.Infof("syncNews:%s", url)
	resp, err := client.R().SetDoNotParseResponse(true).Get(url)
	body := resp.RawBody()
	defer body.Close()
	if err != nil {
		appError("sync-news", "news.sync_failed", "sync news request failed", logger.Err(err))
	}
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		//line := scanner.Text()
		//logger.SugaredLogger.Infof("Received data: %s", line)
		news := &models.NtfyNews{}
		err := json.Unmarshal(scanner.Bytes(), news)
		if err != nil {
			return
		}
		dataTime := time.UnixMilli(int64(news.Time * 1000))

		if slice.ContainAny(news.Tags, []string{"外媒资讯", "财联社电报", "新浪财经", "外媒简讯", "外媒"}) {
			if a.marketReadService == nil {
				continue
			}
			telegraph, inserted := a.marketReadService.SaveNtfyNews(*news)
			if inserted && telegraph != nil && time.Since(dataTime) < 5*time.Minute {
				a.NewsPush(&[]models.Telegraph{*telegraph})
			}
		}
	}
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	defer PanicHandler()
	domReadyLog := appLifecycleLogger("app")
	domReadyTrace := appLifecycleTrace("wails-dom-ready")
	defer func() {
		// 增加延迟确保前端已准备好接收事件
		go func() {
			time.Sleep(2 * time.Second)
			runtime.EventsEmit(a.ctx, "loadingMsg", "done")
		}()
	}()
	domReadyLog.WithTrace(domReadyTrace).Info(
		"lifecycle.dom_ready",
		"frontend resources loaded",
	)

	//if stocksBin != nil && len(stocksBin) > 0 {
	//	go runtime.EventsEmit(a.ctx, "loadingMsg", "检查A股基础信息...")
	//	go initStockData(a.ctx)
	//}
	//
	//if stocksBinHK != nil && len(stocksBinHK) > 0 {
	//	go runtime.EventsEmit(a.ctx, "loadingMsg", "检查港股基础信息...")
	//	go initStockDataHK(a.ctx)
	//}
	//
	//if stocksBinUS != nil && len(stocksBinUS) > 0 {
	//	go runtime.EventsEmit(a.ctx, "loadingMsg", "检查美股基础信息...")
	//	go initStockDataUS(a.ctx)
	//}
	updateBasicInfo()

	// Add your action here
	//定时更新数据
	config := data.GetSettingConfig()
	go func() {
		if a.marketReadService != nil {
			go a.marketReadService.RefreshAllFeeds()
		}

		interval := config.RefreshInterval
		if interval <= 0 {
			interval = 1
		}
		//ticker := time.NewTicker(time.Second * time.Duration(interval))
		//defer ticker.Stop()
		//for range ticker.C {
		//	MonitorStockPrices(a)
		//}
		id, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", interval), func() {
			MonitorStockPrices(a)
		})
		if err != nil {
			appError("dom-ready", "cron.monitor_stock_prices_add_failed", "add MonitorStockPrices cron failed", logger.Err(err))
		} else {
			a.setCronEntry("MonitorStockPrices", id)
		}
		entryID, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", interval+10), func() {
			if a.marketReadService == nil {
				return
			}
			feed := a.marketReadService.RefreshFeed("财联社电报")
			news := copyTelegraphValues(feed.Items)
			if config.EnablePushNews && news != nil {
				go a.NewsPush(news)
			}
			go runtime.EventsEmit(a.ctx, "newTelegraph", news)
		})
		if err != nil {
			appError("dom-ready", "cron.telegraph_add_failed", "add GetNewTelegraph cron failed", logger.Err(err))
		} else {
			a.setCronEntry("GetNewTelegraph", entryID)
		}

		entryIDSina, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", interval+10), func() {
			if a.marketReadService == nil {
				return
			}
			feed := a.marketReadService.RefreshFeed("新浪财经")
			news := copyTelegraphValues(feed.Items)
			if config.EnablePushNews && news != nil {
				go a.NewsPush(news)
			}
			go runtime.EventsEmit(a.ctx, "newSinaNews", news)
		})
		if err != nil {
			appError("dom-ready", "cron.sina_news_add_failed", "add newSinaNews cron failed", logger.Err(err))
		} else {
			a.setCronEntry("newSinaNews", entryIDSina)
		}

		entryIDTradingViewNews, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", interval+10), func() {
			if a.marketReadService == nil {
				return
			}
			feed := a.marketReadService.RefreshFeed("外媒")
			news := copyTelegraphValues(feed.Items)
			if config.EnablePushNews && news != nil {
				go a.NewsPush(news)
			}
			go runtime.EventsEmit(a.ctx, "tradingViewNews", news)
		})
		if err != nil {
			appError("dom-ready", "cron.trading_view_add_failed", "add tradingViewNews cron failed", logger.Err(err))
		} else {
			a.setCronEntry("tradingViewNews", entryIDTradingViewNews)
		}
	}()

	//刷新基金净值信息
	go func() {
		//ticker := time.NewTicker(time.Second * time.Duration(60))
		//defer ticker.Stop()
		//for range ticker.C {
		//	MonitorFundPrices(a)
		//}
		if config.EnableFund {
			id, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", 60), func() {
				MonitorFundPrices(a)
			})
			if err != nil {
				appError("dom-ready", "cron.monitor_fund_add_failed", "add MonitorFundPrices cron failed", logger.Err(err))
			} else {
				a.setCronEntry("MonitorFundPrices", id)
			}
		}

		// AI 推荐股票价格监控定时器
		idAiStock, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", 60), func() {
			MonitorAiRecommendStockPrices(a)
		})
		if err != nil {
			appError("dom-ready", "cron.monitor_ai_stock_add_failed", "add MonitorAiRecommendStockPrices cron failed", logger.Err(err))
		} else {
			a.setCronEntry("MonitorAiRecommendStockPrices", idAiStock)
		}

		// 自选股成本价监控定时器
		idCostPrice, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", 60), func() {
			MonitorFollowedStockCostPrices(a)
		})
		if err != nil {
			appError("dom-ready", "cron.monitor_cost_price_add_failed", "add MonitorFollowedStockCostPrices cron failed", logger.Err(err))
		} else {
			a.setCronEntry("MonitorFollowedStockCostPrices", idCostPrice)
		}

	}()

	if config.EnableNews {
		//go func() {
		//	ticker := time.NewTicker(time.Second * time.Duration(60))
		//	defer ticker.Stop()
		//	for range ticker.C {
		//		telegraph := refreshTelegraphList()
		//		if telegraph != nil {
		//			go runtime.EventsEmit(a.ctx, "telegraph", telegraph)
		//		}
		//	}
		//
		//}()

		id, err := a.cron.AddFunc(fmt.Sprintf("@every %ds", 60), func() {
			telegraph := refreshTelegraphList()
			if telegraph != nil {
				go runtime.EventsEmit(a.ctx, "telegraph", telegraph)
			}
		})
		if err != nil {
			appError("dom-ready", "cron.refresh_telegraph_add_failed", "add refreshTelegraphList cron failed", logger.Err(err))
		} else {
			a.setCronEntry("refreshTelegraphList", id)
		}

		go runtime.EventsEmit(a.ctx, "telegraph", refreshTelegraphList())
	}
	go MonitorStockPrices(a)
	if config.EnableFund && a.fundService != nil {
		go MonitorFundPrices(a)
		go a.fundService.SyncAllFunds(a.ctx)
	}
	// AI 推荐股票价格监控
	go MonitorAiRecommendStockPrices(a)
	// 自选股成本价监控
	go MonitorFollowedStockCostPrices(a)
	//检查新版本
	go func() {
		a.CheckUpdate(0)
		go a.CheckStockBaseInfo(a.ctx)
		go a.syncAllStockInfo(a.ctx)

		a.cron.AddFunc("0 0 2 * * *", func() {
			appInfo("dom-ready", "cron.check_stock_base_info_started", "scheduled stock base info check started")
			a.CheckStockBaseInfo(a.ctx)
		})
		a.cron.AddFunc("30 05 8,12,20 * * *", func() {
			appInfo("dom-ready", "cron.check_update_started", "scheduled update check started")
			a.CheckUpdate(0)
		})
		a.cron.AddFunc("30 05 8,12,20 * * *", func() {
			a.syncAllStockInfo(a.ctx)
		})
	}()

	//检查谷歌浏览器
	//go func() {
	//	f := checkChromeOnWindows()
	//	if !f {
	//		go runtime.EventsEmit(a.ctx, "warnMsg", "谷歌浏览器未安装,ai分析功能可能无法使用")
	//	}
	//}()

	//检查Edge浏览器
	//go func() {
	//	path, e := checkEdgeOnWindows()
	//	if !e {
	//		go runtime.EventsEmit(a.ctx, "warnMsg", "Edge浏览器未安装,ai分析功能可能无法使用")
	//	} else {
	//		logger.SugaredLogger.Infof("Edge浏览器已安装，路径为: %s", path)
	//	}
	//}()
	a.restoreStockAICronSchedules()
	//logger.SugaredLogger.Infof("domReady-cronEntrys:%+v", a.cronEntrys)

}

func (a *App) syncAllStockInfo(ctx context.Context) {
	defer PanicHandler()
	defer func() {
		go runtime.EventsEmit(ctx, "loadingMsg", "done")
	}()
	if a.researchService == nil {
		return
	}
	if err := a.researchService.SyncAllStockInfo(ctx); err != nil {
		appError("sync-all-stock-info", "stock.sync_batch_insert_failed", "create all stock info batch failed", logger.Err(err))
	}
}

func (a *App) CheckStockBaseInfo(ctx context.Context) {
	defer PanicHandler()
	defer func() {
		go runtime.EventsEmit(ctx, "loadingMsg", "done")
	}()
	if a.researchService == nil {
		return
	}
	if err := a.researchService.RefreshStockBaseInfo(ctx); err != nil {
		appError("check-stock-base-info", "stock.base_info_save_failed", "save stock base info failed", logger.Err(err))
	}
}

func (a *App) NewsPush(news *[]models.Telegraph) {
	if news == nil || a.watchlistService == nil {
		return
	}
	follows := a.watchlistService.GetFollowList(a.ctx, 0)
	stockNames := slice.Map(follows, func(index int, item data.FollowedStock) string {
		return item.Name
	})

	for _, telegraph := range *news {
		if a.GetConfig().EnableOnlyPushRedNews {
			if telegraph.IsRed || strutil.ContainsAny(telegraph.Content, stockNames) {
				go runtime.EventsEmit(a.ctx, "newsPush", telegraph)
			}
		} else {
			go runtime.EventsEmit(a.ctx, "newsPush", telegraph)
		}
	}
}

func (a *App) restoreStockAICronSchedules() {
	if a.watchlistService == nil {
		return
	}
	for _, follow := range a.watchlistService.ListScheduledStocks(a.ctx) {
		a.registerStockAICron(follow.StockCode, follow.Cron)
	}
}

func (a *App) registerStockAICron(stockCode, cronText string) {
	if a.cron == nil {
		return
	}
	if entryID, exists := a.getCronEntry(stockCode); exists {
		a.cron.Remove(entryID)
	}
	if strings.TrimSpace(cronText) == "" {
		a.removeCronEntry(stockCode)
		return
	}
	id, err := a.cron.AddFunc(cronText, a.buildStockAICronJob(stockCode))
	if err != nil {
		appError("watchlist-cron", "watchlist.cron_add.failed", "add stock ai cron failed", logger.String("stock_code", stockCode), logger.String("cron_expr", cronText), logger.Err(err))
		return
	}
	a.setCronEntry(stockCode, id)
}

func (a *App) buildStockAICronJob(stockCode string) func() {
	return func() {
		if a.watchlistService == nil {
			return
		}
		displayName := stockCode
		for _, follow := range a.watchlistService.ListScheduledStocks(a.ctx) {
			if follow.StockCode == stockCode && strings.TrimSpace(follow.Name) != "" {
				displayName = follow.Name + "_" + follow.StockCode
				break
			}
		}
		if a.emitEvent != nil {
			a.emitEvent(a.ctx, "warnMsg", "开始自动分析"+displayName)
		}
		result, userErr := a.watchlistService.RunScheduledAnalysis(a.ctx, stockCode)
		if userErr != nil {
			if a.emitEvent != nil {
				a.emitEvent(a.ctx, "warnMsg", "AI分析失败："+userErr.Message)
			}
			return
		}
		if a.emitEvent != nil {
			a.emitEvent(a.ctx, "warnMsg", "AI分析完成："+result.Name+"_"+result.StockCode)
		}
	}
}

func refreshTelegraphList() *[]string {
	url := "https://www.cls.cn/telegraph"
	response, err := resty.New().R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.60").
		Get(url)
	if err != nil {
		return &[]string{}
	}
	//logger.SugaredLogger.Info(string(response.Body()))
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(response.Body())))
	if err != nil {
		return &[]string{}
	}
	var telegraph []string
	document.Find("div.telegraph-content-box").Each(func(i int, selection *goquery.Selection) {
		//logger.SugaredLogger.Info(selection.Text())
		telegraph = append(telegraph, selection.Text())
	})
	return &telegraph
}

// isTradingDay 判断是否是交易日
func isTradingDay(date time.Time) bool {
	weekday := date.Weekday()
	// 判断是否是周末
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	// 这里可以添加具体的节假日判断逻辑
	// 例如：判断是否是春节、国庆节等
	return true
}

// isTradingTime 判断是否是交易时间
func isTradingTime(date time.Time) bool {
	if !isTradingDay(date) {
		return false
	}

	hour, minute, _ := date.Clock()

	// 判断是否在9:15到11:30之间
	if (hour == 9 && minute >= 15) || (hour == 10) || (hour == 11 && minute <= 30) {
		return true
	}

	// 判断是否在13:00到15:00之间
	if (hour == 13) || (hour == 14) || (hour == 15 && minute <= 0) {
		return true
	}

	return false
}

// IsHKTradingTime 判断当前时间是否在港股交易时间内
func IsHKTradingTime(date time.Time) bool {
	hour, minute, _ := date.Clock()

	// 开市前竞价时段：09:00 - 09:30
	if (hour == 9 && minute >= 0) || (hour == 9 && minute <= 30) {
		return true
	}

	// 上午持续交易时段：09:30 - 12:00
	if (hour == 9 && minute > 30) || (hour >= 10 && hour < 12) || (hour == 12 && minute == 0) {
		return true
	}

	// 下午持续交易时段：13:00 - 16:00
	if (hour == 13 && minute >= 0) || (hour >= 14 && hour < 16) || (hour == 16 && minute == 0) {
		return true
	}

	// 收市竞价交易时段：16:00 - 16:10
	if (hour == 16 && minute >= 0) || (hour == 16 && minute <= 10) {
		return true
	}
	return false
}

// IsUSTradingTime 判断当前时间是否在美股交易时间内
func IsUSTradingTime(date time.Time) bool {
	// 获取美国东部时区
	est, err := time.LoadLocation("America/New_York")
	var estTime time.Time
	if err != nil {
		estTime = date.Add(time.Hour * -12)
	} else {
		// 将当前时间转换为美国东部时间
		estTime = date.In(est)
	}

	// 判断是否是周末
	weekday := estTime.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 获取小时和分钟
	hour, minute, _ := estTime.Clock()

	// 判断是否在4:00 AM到9:30 AM之间（盘前）
	if (hour == 4) || (hour == 5) || (hour == 6) || (hour == 7) || (hour == 8) || (hour == 9 && minute < 30) {
		return true
	}

	// 判断是否在9:30 AM到4:00 PM之间（盘中）
	if (hour == 9 && minute >= 30) || (hour >= 10 && hour < 16) || (hour == 16 && minute == 0) {
		return true
	}

	// 判断是否在4:00 PM到8:00 PM之间（盘后）
	if (hour == 16 && minute > 0) || (hour >= 17 && hour < 20) || (hour == 20 && minute == 0) {
		return true
	}

	return false
}
func MonitorFundPrices(a *App) {
	if a == nil || a.fundService == nil {
		return
	}

	// 检查 A 股是否开市（基金交易时间与 A 股一致）
	if !isTradingTime(time.Now()) {
		appInfo("monitor-fund-prices", "fund.monitor_skipped", "skip fund price monitor because A-share market is closed")
		return
	}

	appInfo("monitor-fund-prices", "fund.monitor_started", "start fund price monitor because A-share market is open")

	err := a.fundService.RefreshFollowedFunds(a.ctx)
	if err == nil {
		return
	}

	var refreshErr fundservice.RefreshError
	if errors.As(err, &refreshErr) && refreshErr.FundCode != "" {
		appError("monitor-fund-prices", "fund.basic_info_fetch_failed", "crawl fund basic info failed", logger.String("fund_code", refreshErr.FundCode), logger.Err(refreshErr.Err))
		return
	}
	appError("monitor-fund-prices", "fund.basic_info_fetch_failed", "crawl fund basic info failed", logger.Err(err))
}

// MonitorAiRecommendStockPrices 监控 AI 推荐股票的价格，当股价达到预警线时发送通知
func MonitorAiRecommendStockPrices(a *App) {
	if a.researchService == nil {
		return
	}

	isAStockOpen := isTradingTime(time.Now())
	isHKStockOpen := IsHKTradingTime(time.Now())
	isUSStockOpen := IsUSTradingTime(time.Now())

	if !isAStockOpen && !isHKStockOpen && !isUSStockOpen {
		appInfo("monitor-ai-recommend-stock-prices", "stock.ai_recommend_monitor_skipped", "skip ai recommend stock price monitor because all markets are closed")
		return
	}

	a.dispatchNotificationDeliveries(a.researchService.EvaluateAiRecommendAlerts(a.ctx, time.Now()))
}

// MonitorFollowedStockCostPrices 监控自选股的持仓成本价，当股价低于成本价时发送预警
func MonitorFollowedStockCostPrices(a *App) {
	if a.watchlistService == nil {
		return
	}

	isAStockOpen := isTradingTime(time.Now())
	isHKStockOpen := IsHKTradingTime(time.Now())
	isUSStockOpen := IsUSTradingTime(time.Now())

	if !isAStockOpen && !isHKStockOpen && !isUSStockOpen {
		appInfo("monitor-followed-stock-cost-prices", "stock.cost_monitor_skipped", "skip followed stock cost price monitor because all markets are closed")
		return
	}

	a.dispatchNotificationDeliveries(a.watchlistService.EvaluateCostAlerts(a.ctx, time.Now()))
}

func (a *App) GetStockInfos(follows ...data.FollowedStock) *[]data.StockInfo {
	stockInfos := make([]data.StockInfo, 0)
	if a == nil || a.watchlistService == nil {
		return &stockInfos
	}
	stockCodes := make([]string, 0)
	for _, follow := range follows {
		if strutil.HasPrefixAny(follow.StockCode, []string{"SZ", "SH", "sh", "sz"}) && (!isTradingTime(time.Now())) {
			continue
		}
		if strutil.HasPrefixAny(follow.StockCode, []string{"hk", "HK"}) && (!IsHKTradingTime(time.Now())) {
			continue
		}
		if strutil.HasPrefixAny(follow.StockCode, []string{"us", "US", "gb_"}) && (!IsUSTradingTime(time.Now())) {
			continue
		}
		stockCodes = append(stockCodes, follow.StockCode)
	}
	if len(stockCodes) == 0 {
		return &stockInfos
	}
	for _, quote := range a.watchlistService.GetRealtimePrices(a.ctx, stockCodes...) {
		info := realtimePriceToStockInfo(quote)
		v, ok := slice.FindBy(follows, func(idx int, follow data.FollowedStock) bool {
			if strutil.HasPrefixAny(follow.StockCode, []string{"US", "us"}) {
				return strings.ToLower(strings.Replace(follow.StockCode, "us", "gb_", 1)) == info.Code
			}

			return follow.StockCode == info.Code
		})
		if ok {
			a.addStockFollowData(v, &info)
			stockInfos = append(stockInfos, info)
		}
	}
	return &stockInfos
}

func (a *App) getStockInfo(follow data.FollowedStock) *data.StockInfo {
	if a == nil || a.watchlistService == nil {
		return &data.StockInfo{}
	}
	quotes := a.watchlistService.GetRealtimePrices(a.ctx, follow.StockCode)
	if len(quotes) == 0 {
		return &data.StockInfo{}
	}
	stockData := realtimePriceToStockInfo(quotes[0])
	a.addStockFollowData(follow, &stockData)
	return &stockData
}

func (a *App) addStockFollowData(follow data.FollowedStock, stockData *data.StockInfo) {
	stockData.PrePrice = follow.Price //上次当前价格
	stockData.Sort = follow.Sort
	stockData.CostPrice = follow.CostPrice //成本价
	stockData.CostVolume = follow.Volume   //成本量
	stockData.AlarmChangePercent = follow.AlarmChangePercent
	stockData.AlarmPrice = follow.AlarmPrice
	stockData.Groups = follow.Groups

	//当前价格
	price, _ := convertor.ToFloat(stockData.Price)
	//当前价格为0 时 使用卖一价格作为当前价格
	if price == 0 {
		price, _ = convertor.ToFloat(stockData.A1P)
	}
	//当前价格依然为0 时 使用买一报价作为当前价格
	if price == 0 {
		price, _ = convertor.ToFloat(stockData.B1P)
	}

	//昨日收盘价
	preClosePrice, _ := convertor.ToFloat(stockData.PreClose)

	//当前价格依然为0 时 使用昨日收盘价为当前价格
	if price == 0 {
		price = preClosePrice
	}

	//今日最高价
	highPrice, _ := convertor.ToFloat(stockData.High)
	if highPrice == 0 {
		highPrice, _ = convertor.ToFloat(stockData.Open)
	}

	//今日最低价
	lowPrice, _ := convertor.ToFloat(stockData.Low)
	if lowPrice == 0 {
		lowPrice, _ = convertor.ToFloat(stockData.Open)
	}
	//开盘价
	//openPrice, _ := convertor.ToFloat(stockData.Open)

	if price > 0 && preClosePrice > 0 {
		stockData.ChangePrice = mathutil.RoundToFloat(price-preClosePrice, 2)
		stockData.ChangePercent = mathutil.RoundToFloat(mathutil.Div(price-preClosePrice, preClosePrice)*100, 3)
	}
	if highPrice > 0 && preClosePrice > 0 {
		stockData.HighRate = mathutil.RoundToFloat(mathutil.Div(highPrice-preClosePrice, preClosePrice)*100, 3)
	}
	if lowPrice > 0 && preClosePrice > 0 {
		stockData.LowRate = mathutil.RoundToFloat(mathutil.Div(lowPrice-preClosePrice, preClosePrice)*100, 3)
	}
	if follow.CostPrice > 0 && follow.Volume > 0 {
		if price > 0 {
			stockData.Profit = mathutil.RoundToFloat(mathutil.Div(price-follow.CostPrice, follow.CostPrice)*100, 3)
			stockData.ProfitAmount = mathutil.RoundToFloat((price-follow.CostPrice)*float64(follow.Volume), 2)
			stockData.ProfitAmountToday = mathutil.RoundToFloat((price-preClosePrice)*float64(follow.Volume), 2)
		} else {
			//未开盘时当前价格为昨日收盘价
			stockData.Profit = mathutil.RoundToFloat(mathutil.Div(preClosePrice-follow.CostPrice, follow.CostPrice)*100, 3)
			stockData.ProfitAmount = mathutil.RoundToFloat((preClosePrice-follow.CostPrice)*float64(follow.Volume), 2)
			// 未开盘时，今日盈亏为 0
			stockData.ProfitAmountToday = 0
		}

	}

	//logger.SugaredLogger.Debugf("stockData:%+v", stockData)
	if follow.Price != price && price > 0 && a != nil && a.watchlistService != nil {
		go a.watchlistService.UpdateObservedPrice(a.ctx, follow.StockCode, price)
	}
}

func realtimePriceToStockInfo(quote marketservice.RealtimePrice) data.StockInfo {
	return data.StockInfo{
		Code:     quote.StockCode,
		Name:     quote.StockName,
		Price:    quote.Price,
		Bid:      quote.Bid,
		Ask:      quote.Ask,
		Open:     quote.Open,
		High:     quote.High,
		Low:      quote.Low,
		PreClose: quote.PreClose,
		Date:     quote.Date,
		Time:     quote.Time,
	}
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	defer PanicHandler()
	shutdownLog := appLifecycleLogger("app")
	shutdownTrace := appLifecycleTrace("wails-shutdown")
	// 记录当前窗口大小，供下次启动时还原
	if a.ctx != nil {
		if w, h := runtime.WindowGetSize(a.ctx); w > 0 && h > 0 {
			cfg := data.GetSettingConfig()
			cfg.WindowWidth = w
			cfg.WindowHeight = h
			data.UpdateConfig(cfg)
			shutdownLog.WithTrace(shutdownTrace).Info(
				"lifecycle.shutdown.window_saved",
				"saved window size during shutdown",
				logger.Int("width", w),
				logger.Int("height", h),
			)
		}
	}
	shutdownLog.WithTrace(shutdownTrace).Info(
		"lifecycle.shutdown",
		"application shutdown",
	)
}

// Greet returns a greeting for the given name
func (a *App) Greet(stockCode string) *data.StockInfo {
	if a.watchlistService == nil {
		return &data.StockInfo{}
	}
	follow := a.watchlistService.GetFollowedStock(a.ctx, stockCode)
	return a.getStockInfo(follow)
}

func (a *App) Follow(stockCode string) string {
	if a.watchlistService == nil {
		return "操作失败"
	}
	return a.watchlistService.Follow(a.ctx, stockCode)
}

func (a *App) UnFollow(stockCode string) string {
	if a.watchlistService == nil {
		return "操作失败"
	}
	return a.watchlistService.Unfollow(a.ctx, stockCode)
}

func (a *App) GetFollowList(groupId int) *[]data.FollowedStock {
	if a.watchlistService == nil {
		empty := []data.FollowedStock{}
		return &empty
	}
	list := a.watchlistService.GetFollowList(a.ctx, groupId)
	return &list
}

func (a *App) GetStockList(key string) []data.StockBasic {
	if a.marketReadService == nil {
		return []data.StockBasic{}
	}
	return a.marketReadService.LoadStockList(key)
}

func (a *App) SetCostPriceAndVolume(stockCode string, price float64, volume int64) string {
	if a.watchlistService == nil {
		return "操作失败"
	}
	return a.watchlistService.SetCostPriceAndVolume(a.ctx, stockCode, price, volume)
}

func (a *App) SetTradingPrice(stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	if a.watchlistService == nil {
		return "操作失败"
	}
	return a.watchlistService.SetTradingPrice(a.ctx, stockCode, entryPrice, takeProfitPrice, stopLossPrice, costPrice)
}

func (a *App) SetAlarmChangePercent(val, alarmPrice float64, stockCode string) string {
	if a.watchlistService == nil {
		return "操作失败"
	}
	return a.watchlistService.SetAlarmChangePercent(a.ctx, stockCode, val, alarmPrice)
}
func (a *App) SetStockSort(sort int64, stockCode string) {
	if a.watchlistService == nil {
		return
	}
	a.watchlistService.SetStockSort(a.ctx, stockCode, sort)
}
func (a *App) SendDingDingMessage(message string, stockCode string) string {
	if a.notificationService == nil {
		return ""
	}
	return a.notificationService.SendDingTalk(message, stockCode)
}

// SendDingDingMessageByType msgType 报警类型: 1 涨跌报警;2 股价报警 3 成本价报警
func (a *App) SendDingDingMessageByType(message string, stockCode string, msgType int) string {

	if strutil.HasPrefixAny(stockCode, []string{"SZ", "SH", "sh", "sz"}) && (!isTradingTime(time.Now())) {
		return "非A股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"hk", "HK"}) && (!IsHKTradingTime(time.Now())) {
		return "非港股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "US", "gb_"}) && (!IsUSTradingTime(time.Now())) {
		return "非美股交易时间"
	}
	if a.notificationService == nil {
		return ""
	}

	result := a.notificationService.SendTyped(message, stockCode, msgType)
	if a.emitEvent != nil && strings.TrimSpace(result.EventContent) != "" {
		a.emitEvent(a.ctx, "newsPush", map[string]any{
			"time":    "📈 " + result.EventTitle,
			"isRed":   true,
			"source":  "go-stock",
			"content": result.EventContent,
		})
	}
	return result.DingResult
}

func (a *App) dispatchNotificationDeliveries(deliveries []notificationservice.Delivery) {
	for _, delivery := range deliveries {
		msgType, err := convertor.ToInt(delivery.EventContent)
		if err != nil || msgType <= 0 {
			continue
		}
		a.SendDingDingMessageByType(delivery.DingResult, delivery.EventTitle, int(msgType))
	}
}

func (a *App) NewChatStream(stock, stockCode, question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool) {
	msgs := a.analysisService.StartStockAnalysis(a.ctx, analysisservice.StockRequest{
		StockName:   stock,
		StockCode:   stockCode,
		Question:    question,
		AIConfigID:  aiConfigId,
		SysPromptID: sysPromptId,
		EnableTools: enableTools,
		Think:       think,
	})
	for msg := range msgs {
		runtime.EventsEmit(a.ctx, "newChatStream", map[string]any{
			"chatId":            msg.ChatID,
			"question":          msg.Question,
			"content":           msg.Content,
			"extraContent":      msg.ExtraContent,
			"model":             msg.Model,
			"time":              msg.Time,
			"reasoning_content": msg.ReasoningContent,
			"tool_calls":        msg.ToolCalls,
		})
	}
	runtime.EventsEmit(a.ctx, "newChatStream", "DONE")
}

func (a *App) SaveAIResponseResult(stockCode, stockName, result, chatId, question string, aiConfigId int) {
	a.analysisService.SaveResult(a.ctx, stockCode, stockName, result, chatId, question, aiConfigId)
}
func (a *App) GetAIResponseResult(stock string) *models.AIResponseResult {
	return a.analysisService.GetLatestResult(a.ctx, stock)
}

func (a *App) GetVersionInfo() *models.VersionInfo {
	return &models.VersionInfo{
		Version:           Version,
		Icon:              GetImageBase(icon),
		Alipay:            GetImageBase(alipay),
		Wxpay:             GetImageBase(wxpay),
		Wxgzh:             GetImageBase(wxgzh),
		Content:           VersionCommit,
		OfficialStatement: OFFICIAL_STATEMENT,
	}
}

//// checkChromeOnWindows 在 Windows 系统上检查谷歌浏览器是否安装
//func checkChromeOnWindows() bool {
//	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe`, registry.QUERY_VALUE)
//	if err != nil {
//		// 尝试在 WOW6432Node 中查找（适用于 64 位系统上的 32 位程序）
//		key, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\App Paths\chrome.exe`, registry.QUERY_VALUE)
//		if err != nil {
//			return false
//		}
//		defer key.Close()
//	}
//	defer key.Close()
//	_, _, err = key.GetValue("Path", nil)
//	return err == nil
//}
//
//// checkEdgeOnWindows 在 Windows 系统上检查Edge浏览器是否安装，并返回安装路径
//func checkEdgeOnWindows() (string, bool) {
//	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\msedge.exe`, registry.QUERY_VALUE)
//	if err != nil {
//		// 尝试在 WOW6432Node 中查找（适用于 64 位系统上的 32 位程序）
//		key, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\App Paths\msedge.exe`, registry.QUERY_VALUE)
//		if err != nil {
//			return "", false
//		}
//		defer key.Close()
//	}
//	defer key.Close()
//	path, _, err := key.GetStringValue("Path")
//	if err != nil {
//		return "", false
//	}
//	return path, true
//}

func GetImageBase(bytes []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(bytes)
}

func GenNotificationMsg(stockInfo *data.StockInfo) string {
	Price, err := convertor.ToFloat(stockInfo.Price)
	if err != nil {
		Price = 0
	}
	PreClose, err := convertor.ToFloat(stockInfo.PreClose)
	if err != nil {
		PreClose = 0
	}
	var RF float64
	if PreClose > 0 {
		RF = mathutil.RoundToFloat(((Price-PreClose)/PreClose)*100, 2)
	}

	return "[" + stockInfo.Name + "] " + stockInfo.Price + " " + convertor.ToString(RF) + "% " + stockInfo.Date + " " + stockInfo.Time
}

// msgType : 1 涨跌报警(5分钟);2 股价报警(30分钟) 3 成本价报警(30分钟) 4 止盈报警(5分钟) 5 止损报警(5分钟)
func getMsgTypeTTL(msgType int) int {
	switch msgType {
	case 1:
		return 60 * 5
	case 2:
		return 60 * 30
	case 3:
		return 60 * 30
	case 4:
		return 60 * 5
	case 5:
		return 60 * 5
	default:
		return 60 * 5
	}
}

func getMsgTypeName(msgType int) string {
	switch msgType {
	case 1:
		return "涨跌报警"
	case 2:
		return "股价报警"
	case 3:
		return "成本价报警"
	case 4:
		return "止盈报警"
	case 5:
		return "止损报警"
	default:
		return "未知类型"
	}
}

func onExit(a *App) {
	// 清理操作
	//logger.SugaredLogger.Infof("systray onExit")
	//systray.Quit()
	//runtime.Quit(a.ctx)
}

func (a *App) UpdateConfig(settingConfig *data.SettingConfig) string {
	//s1, _ := json.Marshal(settingConfig)
	//logger.SugaredLogger.Infof("UpdateConfig:%s", s1)
	if settingConfig.RefreshInterval > 0 {
		if entryID, exists := a.getCronEntry("MonitorStockPrices"); exists {
			a.cron.Remove(entryID)
		}
		id, _ := a.cron.AddFunc(fmt.Sprintf("@every %ds", settingConfig.RefreshInterval), func() {
			MonitorStockPrices(a)
		})
		a.setCronEntry("MonitorStockPrices", id)
	}

	return a.configService.UpdateConfig(a.ctx, settingConfig)
}

func (a *App) GetConfig() *data.SettingConfig {
	return a.configService.GetConfig(a.ctx)
}

func uploadSharedAnalysis(artifact analysisservice.ResultArtifact) (string, error) {
	response, err := resty.New().SetHeader("ua-x", "go-stock").R().SetFormData(map[string]string{
		"text":         artifact.Content,
		"stockCode":    artifact.StockCode,
		"stockName":    artifact.StockName,
		"analysisTime": artifact.AnalysisDate,
	}).Post("http://go-stock.sparkmemory.top:16688/upload")
	if err != nil {
		return "", err
	}
	return response.String(), nil
}

func (a *App) ExportConfig() string {
	exporter := a.exportConfigSource()
	if exporter == nil {
		return "导出失败"
	}

	saveFileDialog := a.saveFileDialog
	if saveFileDialog == nil {
		saveFileDialog = runtime.SaveFileDialog
	}
	writeFile := a.writeFile
	if writeFile == nil {
		writeFile = os.WriteFile
	}

	config := exporter.ExportConfig(a.ctx)
	file, err := saveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "导出配置文件",
		CanCreateDirectories: true,
		DefaultFilename:      "config.json",
	})
	if err != nil {
		appError("export-config", "config.export_dialog_failed", "export config dialog failed", logger.Err(err))
		return err.Error()
	}
	if err := writeFile(file, []byte(config), os.ModePerm); err != nil {
		appError("export-config", "config.export_write_failed", "write exported config failed", logger.String("file", file), logger.Err(err))
		return err.Error()
	}
	return "导出成功:" + file
}

func (a *App) ShareAnalysis(stockCode, stockName string) string {
	artifacts := a.artifactService()
	if artifacts == nil {
		return "分析结果异常"
	}
	uploader := a.shareAnalysisUploader
	if uploader == nil {
		uploader = uploadSharedAnalysis
	}
	artifact, userErr := artifacts.GetResultArtifact(a.ctx, stockCode, stockName)
	if userErr != nil {
		return userErr.Message
	}
	msg, err := uploader(artifact)
	if err != nil {
		return err.Error()
	}
	return msg
}

// ShareText 直接把文本分享到社区（用于 AI 助手等非 AIResponseResult 场景）
func (a *App) ShareText(text, title string) string {
	text = strings.TrimSpace(text)
	title = strings.TrimSpace(title)
	if text == "" {
		return "内容为空"
	}
	if title == "" {
		title = "AI助手"
	}
	analysisTime := time.Now().Format("2006/01/02")
	response, err := resty.New().SetHeader("ua-x", "go-stock").R().SetFormData(map[string]string{
		"text":         text,
		"stockCode":    title,
		"stockName":    title,
		"analysisTime": analysisTime,
	}).Post("http://go-stock.sparkmemory.top:16688/upload")
	if err != nil {
		return err.Error()
	}
	return response.String()
}

func (a *App) GetfundList(key string) []data.FundBasic {
	if a.fundService == nil {
		return []data.FundBasic{}
	}
	return a.fundService.LoadFundList(a.ctx, key)
}
func (a *App) GetFollowedFund() []data.FollowedFund {
	if a.fundService == nil {
		return []data.FollowedFund{}
	}
	return a.fundService.LoadFollowedFunds(a.ctx)
}
func (a *App) FollowFund(fundCode string) string {
	if a.fundService == nil {
		return "关注失败"
	}
	return a.fundService.FollowFund(a.ctx, fundCode)
}
func (a *App) UnFollowFund(fundCode string) string {
	if a.fundService == nil {
		return "取消关注失败"
	}
	return a.fundService.UnfollowFund(a.ctx, fundCode)
}
func (a *App) SaveAsMarkdown(stockCode, stockName string) string {
	artifacts := a.artifactService()
	if artifacts == nil {
		return "分析结果异常,无法保存。"
	}
	saveFileDialog := a.saveFileDialog
	if saveFileDialog == nil {
		saveFileDialog = runtime.SaveFileDialog
	}
	writeFile := a.writeFile
	if writeFile == nil {
		writeFile = os.WriteFile
	}
	artifact, userErr := artifacts.GetResultArtifact(a.ctx, stockCode, stockName)
	if userErr != nil {
		return userErr.Message + ",无法保存。"
	}
	file, err := saveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存为Markdown",
		DefaultFilename: artifact.MarkdownFilename,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown",
				Pattern:     "*.md;*.markdown",
			},
		},
	})
	if err != nil {
		return err.Error()
	}
	if err := writeFile(file, []byte(artifact.Content), 0644); err != nil {
		return err.Error()
	}
	return "已保存至：" + file
}

func (a *App) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	if a.configService == nil {
		empty := []models.PromptTemplate{}
		return &empty
	}
	return a.configService.GetPromptTemplates(a.ctx, name, promptType)
}
func (a *App) AddPrompt(prompt models.Prompt) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SaveLegacyPrompt(a.ctx, prompt)
}
func (a *App) DelPrompt(id uint) string {
	if a.configService == nil {
		return "删除失败"
	}
	return a.configService.DeleteLegacyPrompt(a.ctx, id)
}
func (a *App) SetStockAICron(cronText, stockCode string) {
	if a.watchlistService == nil {
		return
	}
	result, userErr := a.watchlistService.SaveStockAICron(a.ctx, cronText, stockCode)
	if userErr != nil {
		if a.emitEvent != nil {
			a.emitEvent(a.ctx, "warnMsg", "AI分析任务保存失败："+userErr.Message)
		}
		return
	}
	a.registerStockAICron(result.StockCode, result.Cron)
}
func (a *App) AddGroup(group data.Group) string {
	if a.watchlistService == nil {
		return "添加失败"
	}
	return a.watchlistService.AddGroup(a.ctx, group)
}
func (a *App) GetGroupList() []data.Group {
	if a.watchlistService == nil {
		return []data.Group{}
	}
	return a.watchlistService.ListGroups(a.ctx)
}

func (a *App) UpdateGroupSort(id int, newSort int) bool {
	if a.watchlistService == nil {
		return false
	}
	return a.watchlistService.UpdateGroupSort(a.ctx, id, newSort)
}

func (a *App) InitializeGroupSort() bool {
	if a.watchlistService == nil {
		return false
	}
	return a.watchlistService.InitializeGroupSort(a.ctx)
}

func (a *App) GetGroupStockList(groupId int) []data.GroupStock {
	if a.watchlistService == nil {
		return []data.GroupStock{}
	}
	return a.watchlistService.ListGroupStocks(a.ctx, groupId)
}

func (a *App) AddStockGroup(groupId int, stockCode string) string {
	if a.watchlistService == nil {
		return "添加失败"
	}
	return a.watchlistService.AddGroupStock(a.ctx, groupId, stockCode)
}

func (a *App) RemoveStockGroup(code, name string, groupId int) string {
	if a.watchlistService == nil {
		return "移除失败"
	}
	return a.watchlistService.RemoveGroupStock(a.ctx, code, name, groupId)
}

func (a *App) RemoveGroup(groupId int) string {
	if a.watchlistService == nil {
		return "移除失败"
	}
	return a.watchlistService.RemoveGroup(a.ctx, groupId)
}

func (a *App) GetStockKLine(stockCode, stockName string, days int64) *[]data.KLineData {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadStockKLine(stockCode, days)
		return &items
	}
	items := []data.KLineData{}
	return &items
}

func (a *App) GetStockMinutePriceLineData(stockCode, stockName string) marketservice.MinutePriceLine {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadStockMinutePriceLine(stockCode, stockName)
	}
	return marketservice.MinutePriceLine{
		StockCode: stockCode,
		StockName: stockName,
		PriceData: []data.MinuteData{},
	}
}

func (a *App) GetStockCommonKLine(stockCode, stockName string, days int64) *[]data.KLineData {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadStockCommonKLine(stockCode, days)
		return &items
	}
	items := []data.KLineData{}
	return &items
}

// GetStockEastMoneyKLine 东方财富多周期 K 线（分钟：1/5/10/60/120；日 101、周 102、半年 105、年 106）。
// klt 与东方财富接口一致；10 分钟由 1 分钟数据聚合。limit 为根数上限（最大 5000）。
func (a *App) GetStockEastMoneyKLine(stockCode, stockName string, klt string, limit int) *[]data.KLineData {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadStockEastMoneyKLine(stockCode, klt, limit)
		return &items
	}
	items := []data.KLineData{}
	return &items
}

func (a *App) GetStockEastMoneyKLineResult(stockCode, stockName string, klt string, limit int) marketservice.EastMoneyKLinePageResult {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadStockEastMoneyKLineResult(stockCode, klt, limit)
	}
	return marketservice.EastMoneyKLinePageResult{Data: []data.KLineData{}}
}

// GetStockEastMoneyKLinePage 分页拉取 K 线：end 为东财 end 参数（YYYYMMDD 或 YYYYMMDDHHmmss），空字符串表示取最新一段（同 GetStockEastMoneyKLine）。
func (a *App) GetStockEastMoneyKLinePage(stockCode, stockName string, klt string, limit int, end string) *[]data.KLineData {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadStockEastMoneyKLinePage(stockCode, klt, limit, end)
		return &items
	}
	items := []data.KLineData{}
	return &items
}

func (a *App) GetStockEastMoneyKLinePageResult(stockCode, stockName string, klt string, limit int, end string) marketservice.EastMoneyKLinePageResult {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadStockEastMoneyKLinePageResult(stockCode, klt, limit, end)
	}
	return marketservice.EastMoneyKLinePageResult{Data: []data.KLineData{}}
}

func (a *App) GetMarketFeeds() marketservice.FeedSet {
	return a.marketReadService.LoadFeeds()
}

func (a *App) RefreshMarketFeed(source string) marketservice.Feed {
	return a.marketReadService.RefreshFeed(source)
}

func (a *App) GetMarketGlobalIndexes() marketservice.IndexSet {
	return a.marketReadService.LoadGlobalIndexes(30)
}

func (a *App) GetMarketIndustryRanks(sort string, cnt int) []marketservice.IndustryRankEntry {
	return a.marketReadService.LoadIndustryRanks(sort, cnt)
}

func feedItemsForSource(feeds marketservice.FeedSet, source string) []*models.Telegraph {
	switch source {
	case "财联社电报":
		return append([]*models.Telegraph(nil), feeds.Telegraph...)
	case "新浪财经":
		return append([]*models.Telegraph(nil), feeds.Sina...)
	case "外媒":
		return append([]*models.Telegraph(nil), feeds.Foreign...)
	default:
		return []*models.Telegraph{}
	}
}

func copyTelegraphValues(items []*models.Telegraph) *[]models.Telegraph {
	result := make([]models.Telegraph, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, *item)
	}
	return &result
}

func (a *App) GetTelegraphList(source string) *[]*models.Telegraph {
	if a.marketReadService == nil {
		empty := []*models.Telegraph{}
		return &empty
	}
	items := feedItemsForSource(a.marketReadService.LoadFeeds(), source)
	return &items
}

func (a *App) ReFleshTelegraphList(source string) *[]*models.Telegraph {
	if a.marketReadService == nil {
		empty := []*models.Telegraph{}
		return &empty
	}
	feed := a.marketReadService.RefreshFeed(source)
	items := append([]*models.Telegraph(nil), feed.Items...)
	return &items
}

func (a *App) GlobalStockIndexes() marketservice.IndexSet {
	if a.marketReadService == nil {
		return marketservice.IndexSet{}
	}
	return a.marketReadService.LoadGlobalIndexes(30)
}

// GlobalStockIndexesReadable 将全球指数 JSON 转为 AI 易读 Markdown 文本。
func (a *App) GlobalStockIndexesReadable() string {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadGlobalIndexesReadable(30)
	}
	return ""
}

func (a *App) SummaryStockNews(question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool, eventName string, historyJSON string) {
	ctx, cancel := context.WithCancel(a.ctx)

	// 保存当前会话的 cancel，用于前端中断
	a.summaryMu.Lock()
	if a.summaryCancel != nil {
		a.summaryCancel()
	}
	a.summaryCancel = cancel
	a.summaryMu.Unlock()

	// 允许前端自定义事件名，避免不同页面之间的事件冲突
	if strings.TrimSpace(eventName) == "" {
		eventName = "summaryStockNews"
	}

	msgs := a.analysisService.StartMarketSummary(ctx, analysisservice.MarketSummaryRequest{
		Question:    question,
		AIConfigID:  aiConfigId,
		SysPromptID: sysPromptId,
		EnableTools: enableTools,
		Think:       think,
		HistoryJSON: historyJSON,
	})

	for msg := range msgs {
		runtime.EventsEmit(a.ctx, eventName, map[string]any{
			"chatId":            msg.ChatID,
			"question":          msg.Question,
			"content":           msg.Content,
			"extraContent":      msg.ExtraContent,
			"model":             msg.Model,
			"time":              msg.Time,
			"reasoning_content": msg.ReasoningContent,
			"tool_calls":        msg.ToolCalls,
		})
	}

	a.summaryMu.Lock()
	a.summaryCancel = nil
	a.summaryMu.Unlock()

	runtime.EventsEmit(a.ctx, eventName, "DONE")
}
func (a *App) GetIndustryRank(sort string, cnt int) []marketservice.IndustryRankEntry {
	if a.marketReadService == nil {
		return []marketservice.IndustryRankEntry{}
	}
	return a.marketReadService.LoadIndustryRanks(sort, cnt)
}
func (a *App) GetIndustryMoneyRankSina(fenlei, sort string) []marketservice.IndustryMoneyRankRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadIndustryMoneyRanks(fenlei, sort)
	}
	return []marketservice.IndustryMoneyRankRow{}
}
func (a *App) GetMoneyRankSina(sort string) []marketservice.MoneyRankRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadMoneyRanks(sort)
	}
	return []marketservice.MoneyRankRow{}
}

func (a *App) GetStockMoneyTrendByDay(stockCode string, days int) []marketservice.StockMoneyTrendRow {
	if service := a.residualMarketReads(); service != nil {
		return service.LoadStockMoneyTrend(stockCode, days)
	}
	return []marketservice.StockMoneyTrendRow{}
}

// OpenURL
//
//	@Description:  跨平台打开默认浏览器
//	@receiver a
//	@param url
func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

// SaveImage
//
//	@Description: 跨平台保存图片
//	@receiver a
//	@param name
//	@param base64Data
//	@return error
func (a *App) SaveImage(name, base64Data string) string {
	// 打开保存文件对话框
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存图片",
		DefaultFilename: name + "AI分析.png",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PNG 图片",
				Pattern:     "*.png",
			},
		},
	})
	if err != nil || filePath == "" {
		return "文件路径,无法保存。"
	}

	// 解码并保存
	decodeString, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "文件内容异常,无法保存。"
	}

	err = os.WriteFile(filepath.Clean(filePath), decodeString, os.ModePerm)
	if err != nil {
		return "保存结果异常,无法保存。"
	}
	return filePath
}

// SaveWordFile
//
//	@Description: // 跨平台保存word
//	@receiver a
//	@param filename
//	@param base64Data
//	@return error
func (a *App) SaveWordFile(filename string, base64Data string) string {
	// 弹出保存文件对话框
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存 Word 文件",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{
			{DisplayName: "Word 文件", Pattern: "*.docx"},
		},
	})
	if err != nil || filePath == "" {
		return "文件路径,无法保存。"
	}

	// 解码 base64 内容
	decodeString, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "文件内容异常,无法保存。"
	}
	// 保存为文件
	err = os.WriteFile(filepath.Clean(filePath), decodeString, 0777)
	if err != nil {
		return "保存结果异常,无法保存。"
	}
	return filePath
}

// GetAiConfigs
//
//	@Description: // 获取 AiConfig 列表
//	@receiver a
//	@return error
func (a *App) GetAiConfigs() []*data.AIConfig {
	return a.configService.GetAiConfigs(a.ctx)
}

// GetAiAssistantSession 获取 AI 助手会话消息列表，sessionId 为空时获取最新的
// Phase-5 compatibility allowlist: multi-turn assistant orchestration remains bridge-owned until a dedicated assistant refactor.
func (a *App) GetAiAssistantSession(sessionId string) (*models.AiAssistantSessionResp, error) {
	return data.GetAiAssistantSession(sessionId)
}

// SaveAiAssistantSession 保存 AI 助手会话消息到数据库
// Phase-5 compatibility allowlist: multi-turn assistant orchestration remains bridge-owned until a dedicated assistant refactor.
func (a *App) SaveAiAssistantSession(sessionId string, messages []models.AiAssistantMessage) error {
	return data.SaveAiAssistantSession(sessionId, messages)
}

// FetchAiModels
//
//	@Description: 根据接口地址与 apiKey 自动获取支持的模型列表（OpenAI/DeepSeek 兼容 /models 接口）
//	@receiver a
//	@param baseUrl 接口地址（如 https://api.deepseek.com）
//	@param apiKey  鉴权令牌
//	@return []string 模型 ID 列表
func (a *App) FetchAiModels(baseUrl, apiKey string) []string {
	baseUrl = strutil.Trim(baseUrl)
	apiKey = strutil.Trim(apiKey)
	if baseUrl == "" || apiKey == "" {
		return []string{}
	}

	type modelItem struct {
		ID string `json:"id"`
	}
	var respData struct {
		Data []modelItem `json:"data"`
	}

	client := resty.New()
	client.SetBaseURL(baseUrl)
	client.SetHeader("Authorization", "Bearer "+apiKey)
	client.SetHeader("Content-Type", "application/json")

	resp, err := client.R().
		SetResult(&respData).
		Get("/models")
	if err != nil {
		appError("fetch-ai-models", "ai.models_request_failed", "fetch ai models failed", logger.Err(err))
		return []string{}
	}
	if resp.IsError() {
		appError("fetch-ai-models", "ai.models_http_error", "fetch ai models returned http error", logger.String("status", resp.Status()))
		return []string{}
	}

	modelsList := make([]string, 0, len(respData.Data))
	for _, m := range respData.Data {
		if strings.TrimSpace(m.ID) != "" {
			modelsList = append(modelsList, m.ID)
		}
	}
	return modelsList
}

// InitCronTasks 在应用启动时，自动为启用状态的定时任务创建调度
func (a *App) InitCronTasks() {
	if err := a.taskService.RestoreSchedules(a.ctx); err != nil {
		appError("init-cron-tasks", "cron.task_restore_failed", "restore cron tasks failed", logger.Err(err))
	}
}

// AbortSummaryStockNews 取消当前进行中的 SummaryStockNews 流式回答
func (a *App) AbortSummaryStockNews() {
	a.summaryMu.Lock()
	defer a.summaryMu.Unlock()
	if a.summaryCancel != nil {
		a.summaryCancel()
		a.summaryCancel = nil
	}
}

// CreateCronTask
//
//	@Description: 创建定时任务
//	@receiver a
//	@param task 定时任务信息
//	@return string 操作结果
func (a *App) CreateCronTask(task *models.CronTask) string {
	return a.taskService.Create(a.ctx, task)
}

func (a *App) UpdateCronTask(task *models.CronTask) string {
	return a.taskService.Update(a.ctx, task)
}

// DeleteCronTask
//
//	@Description: 删除定时任务
//	@receiver a
//	@param id 任务 ID
//	@return string 操作结果
func (a *App) DeleteCronTask(id uint) string {
	return a.taskService.Delete(a.ctx, id)
}

// GetCronTaskByID
//
//	@Description: 根据 ID 获取定时任务
//	@receiver a
//	@param id 任务 ID
//	@return *models.CronTask 任务信息
func (a *App) GetCronTaskByID(id uint) *models.CronTask {
	task, err := a.taskService.GetByID(a.ctx, id)
	if err != nil {
		return nil
	}
	return task
}

// GetCronTaskList
//
//	@Description: 获取定时任务列表
//	@receiver a
//	@param query 查询条件
//	@return *models.CronTaskPageResp 分页结果
func (a *App) GetCronTaskList(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return a.taskService.List(a.ctx, query)
}

// EnableCronTask
//
//	@Description: 启用/禁用定时任务
//	@receiver a
func (a *App) EnableCronTask(id uint, enable bool) string {
	return a.taskService.Enable(a.ctx, id, enable)
}

// ExecuteCronTaskNow
//
//	@Description: 立即执行定时任务
//	@receiver a
//	@param id 任务 ID
//	@return string 操作结果
func (a *App) ExecuteCronTaskNow(id uint) string {
	go func() {
		err := a.taskService.RunNow(a.ctx, id)
		if err != nil {
			appError("trigger-cron-task", "cron.task_execute_failed", "trigger cron task failed", logger.Uint("task_id", id), logger.Err(err))
		}
	}()

	return "任务执行中"
}

// GetCronTaskTypes
//
//	@Description: 获取所有任务类型
//	@receiver a
//	@return []lo.Tuple2[string, string] 任务类型列表
func (a *App) GetCronTaskTypes() []lo.Tuple2[string, string] {
	return a.taskService.GetTaskTypes(a.ctx)
}

// ValidateCronExpr
//
//	@Description: 验证 Cron 表达式
//	@receiver a
//	@param expr Cron 表达式
//	@return string 验证结果
func (a *App) ValidateCronExpr(expr string) string {
	return a.taskService.ValidateCronExpr(a.ctx, expr)
}

// SearchCronTasks
//
//	@Description: 搜索定时任务
//	@receiver a
//	@param keyword 搜索关键词
//	@return []models.CronTask 搜索结果
func (a *App) SearchCronTasks(keyword string) []models.CronTask {
	return a.taskService.Search(a.ctx, keyword)
}

// CalculateNextRunTime 根据 Cron 表达式计算下一次运行时间
// 参数:
//   - cron: Cron 表达式，用于定义任务调度的时间规则
//
// 返回值:
//   - string: 格式化为 "2006-01-02 15:04:05" 的下一次运行时间字符串
func (a *App) CalculateNextRunTime(cron string) string {
	return a.taskService.CalculateNextRunTime(a.ctx, cron).Format("2006-01-02 15:04:05")
}

// CalculateNextRunTimes 根据 Cron 表达式计算未来多次运行时间
// 参数:
//   - cron: Cron 表达式
//   - count: 需要计算的次数
//
// 返回值:
//   - []string: 按时间顺序排序的运行时间列表，格式为 "2006-01-02 15:04:05"
func (a *App) CalculateNextRunTimes(cron string, count int) []string {
	times := a.taskService.CalculateNextRunTimes(a.ctx, cron, count)
	result := make([]string, 0, len(times))
	for _, t := range times {
		result = append(result, t.Format("2006-01-02 15:04:05"))
	}
	return result
}

// AddTradingRecord 添加交易记录
// 参数:
//   - record: 交易记录结构体
//
// 返回值:
//   - uint: 新添加的交易记录ID
//   - error: 错误信息
func (a *App) AddTradingRecord(record data.TradingRecord) (uint, error) {
	if a.researchService == nil {
		return 0, nil
	}
	return a.researchService.AddTradingRecord(a.ctx, record)
}

// GetTradingRecordList 获取交易记录列表（分页与筛选，返回结构与 AI 推荐列表一致）
func (a *App) GetTradingRecordList(query data.TradingRecordListQuery) *data.TradingRecordPageData {
	if a.researchService == nil {
		return &data.TradingRecordPageData{}
	}
	return a.researchService.GetTradingRecordList(a.ctx, query)
}

// GetTradingRecordById 根据ID获取单个交易记录
// 参数:
//   - id: 交易记录ID
//
// 返回值:
//   - *data.TradingRecord: 交易记录指针
//   - error: 错误信息
func (a *App) GetTradingRecordById(id uint) (*data.TradingRecord, error) {
	if a.researchService == nil {
		return &data.TradingRecord{}, nil
	}
	return a.researchService.GetTradingRecordByID(a.ctx, id)
}

// GetTradingRecordStatistics 获取交易记录统计数据
//
// 返回值:
//   - *data.TradingRecordStatistics: 统计数据指针
func (a *App) GetTradingRecordStatistics() *data.TradingRecordStatistics {
	if a.researchService == nil {
		return &data.TradingRecordStatistics{}
	}
	return a.researchService.GetTradingRecordStatistics(a.ctx)
}

// UpdateTradingRecord 更新交易记录
// 参数:
//   - record: 交易记录结构体
//
// 返回值:
//   - error: 错误信息
func (a *App) UpdateTradingRecord(record data.TradingRecord) error {
	if a.researchService == nil {
		return nil
	}
	return a.researchService.UpdateTradingRecord(a.ctx, record)
}

// DeleteTradingRecord 删除交易记录
// 参数:
//   - id: 交易记录ID
//
// 返回值:
//   - error: 错误信息
func (a *App) DeleteTradingRecord(id uint) error {
	if a.researchService == nil {
		return nil
	}
	return a.researchService.DeleteTradingRecord(a.ctx, id)
}

// CheckFrequentTrading 检查是否频繁交易
// 参数:
//   - stockCode: 股票代码
//
// 返回值:
//   - map[string]any: 包含 canTrade (bool) 和 msg (string)
func (a *App) CheckFrequentTrading(stockCode string) map[string]any {
	if a.researchService == nil {
		return map[string]any{
			"canTrade": true,
			"msg":      "",
		}
	}
	result := a.researchService.CheckFrequentTrading(a.ctx, stockCode)
	return map[string]any{
		"canTrade": result.CanTrade,
		"msg":      result.Message,
	}
}
