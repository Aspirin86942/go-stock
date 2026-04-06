package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/agent"
	"go-stock/backend/data"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	marketservice "go-stock/backend/service/market"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// @Author spark
// @Date 2025/6/8 20:45
// @Desc
//--------------------------------------------------------------------------------

var ShanghaiTimezone = time.FixedZone("CST", 8*60*60)

func appLifecycleLogger(module string) *logger.Logger {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return nil
	}
	return runtimeLogger.ForSink(logger.SinkApp, module)
}

func appLifecycleTrace(source string) logger.TraceContext {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return logger.TraceContext{Source: source}
	}
	return runtimeLogger.NewTrace(source)
}

// logFrontendRuntimeError 统一记录前端运行时错误，保留页面、路由和堆栈等关键上下文。
func logFrontendRuntimeError(optionalData []interface{}) {
	runtimeLogger := logger.Default()
	if runtimeLogger == nil {
		return
	}

	payload := logger.NormalizeFrontendError(optionalData)
	frontendLog := runtimeLogger.ForSink(logger.SinkFrontend, "frontend")
	trace := logger.TraceContext{
		TraceID: strings.TrimSpace(payload.TraceID),
		Source:  "wails-frontend",
	}
	if trace.TraceID == "" {
		trace = runtimeLogger.NewTrace("wails-frontend")
	}

	if payload.Extra != nil {
		frontendLog.WithTrace(trace).Error(
			"frontend.error",
			"frontend runtime error",
			logger.String("error_class", "frontend_error"),
			logger.String("page", payload.Page),
			logger.String("route", payload.Route),
			logger.String("error_message", payload.Message),
			logger.String("source", payload.Source),
			logger.Int("lineno", payload.Line),
			logger.Int("colno", payload.Column),
			logger.String("stack", payload.Stack),
			logger.Any("extra", payload.Extra),
		)
		return
	}

	frontendLog.WithTrace(trace).Error(
		"frontend.error",
		"frontend runtime error",
		logger.String("error_class", "frontend_error"),
		logger.String("page", payload.Page),
		logger.String("route", payload.Route),
		logger.String("error_message", payload.Message),
		logger.String("source", payload.Source),
		logger.Int("lineno", payload.Line),
		logger.Int("colno", payload.Column),
		logger.String("stack", payload.Stack),
	)
}

func GetShanghaiTime() time.Time {
	return time.Now().In(ShanghaiTimezone)
}

func FormatShanghaiTime(t time.Time) string {
	return t.In(ShanghaiTimezone).Format("2006-01-02 15:04:05")
}

func (a *App) GetTimezone() map[string]any {
	return map[string]any{
		"offset":   8 * 60 * 60,
		"location": "Asia/Shanghai",
	}
}

func (a *App) LongTigerRank(date string) *[]models.LongTigerRankData {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadLongTiger(date)
		return &items
	}
	items := []models.LongTigerRankData{}
	return &items
}

func (a *App) StockResearchReport(stockCode string) []marketservice.StockResearchReportEntry {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadStockResearchReports(stockCode)
	}
	return []marketservice.StockResearchReportEntry{}
}
func (a *App) StockNotice(stockCode string) []marketservice.StockNoticeEntry {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadStockNotices(stockCode)
	}
	return []marketservice.StockNoticeEntry{}
}

func (a *App) IndustryResearchReport(industryCode string) []marketservice.IndustryResearchReportEntry {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadIndustryResearchReports(industryCode)
	}
	return []marketservice.IndustryResearchReportEntry{}
}
func (a *App) EMDictCode(code string) []marketservice.EMDictCodeEntry {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadEMDictCodes(code)
	}
	return []marketservice.EMDictCodeEntry{}
}

func (a *App) AnalyzeSentiment(text string) models.SentimentResult {
	return data.AnalyzeSentiment(text)
}

func (a *App) HotStock(marketType string) *[]models.HotItem {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadHotStocks(marketType)
		return &items
	}
	items := []models.HotItem{}
	return &items
}

func (a *App) HotEvent(size int) *[]models.HotEvent {
	if service := a.legacyMarketReads(); service != nil {
		items := service.LoadHotEvents(size)
		return &items
	}
	items := []models.HotEvent{}
	return &items
}
func (a *App) HotTopic(size int) []marketservice.HotTopicEntry {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadHotTopics(size)
	}
	return []marketservice.HotTopicEntry{}
}

func (a *App) InvestCalendarTimeLine(yearMonth string) []marketservice.InvestCalendarDay {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadInvestCalendar(yearMonth)
	}
	return []marketservice.InvestCalendarDay{}
}
func (a *App) ClsCalendar() []marketservice.ClsCalendarDay {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadClsCalendar()
	}
	return []marketservice.ClsCalendarDay{}
}

func (a *App) SearchStock(words string) marketservice.SearchStockResponse {
	if service := a.legacyMarketReads(); service != nil {
		return service.SearchStocks(words)
	}
	return marketservice.SearchStockResponse{}
}
func (a *App) GetHotStrategy() models.HotStrategy {
	if service := a.legacyMarketReads(); service != nil {
		return service.LoadHotStrategies()
	}
	return models.HotStrategy{Data: []*models.HotStrategyData{}}
}

func (a *App) GetAllStocks(page int, pageSize int, name string, technicalIndicators models.TechnicalIndicators) *models.AllStocksResp {
	return data.NewStockDataApi().GetAllStocks(page, pageSize, name, technicalIndicators)
}

func (a *App) ChatWithAgent(question string, aiConfigId int, sysPromptId *int, memoryMode bool, memoryCount int, thinkingMode bool) {
	defer func() {
		if r := recover(); r != nil {
			if log := appLifecycleLogger("app.agent"); log != nil {
				log.WithTrace(appLifecycleTrace("chat-with-agent")).Error(
					"agent.chat.panic",
					"chat with agent panicked",
					logger.Any("panic_value", r),
				)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	a.agentMu.Lock()
	if a.agentCancel != nil {
		a.agentCancel()
	}
	a.agentCancel = cancel
	a.agentMu.Unlock()

	defer func() {
		a.agentMu.Lock()
		a.agentCancel = nil
		a.agentMu.Unlock()
	}()

	ch := agent.NewStockAiAgentApi().ChatWithContext(ctx, question, aiConfigId, sysPromptId, memoryMode, memoryCount, thinkingMode)
	for msg := range ch {
		runtime.EventsEmit(a.ctx, "agent-message", agentMessageToFrontendMap(msg))
	}
	runtime.EventsEmit(a.ctx, "agent-message", agentMessageToFrontendMap(&schema.Message{
		Role:    schema.Assistant,
		Content: "agent-DONE",
	}))
}

// agentMessageToFrontendMap 用标准 JSON 将 schema.Message 转为 map 再 EventsEmit，
// 保证与 json 标签一致（如 reasoning_content、extra），避免 Wails 直接传结构体时前端字段名不一致。
func agentMessageToFrontendMap(msg *schema.Message) map[string]any {
	if msg == nil {
		return map[string]any{}
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return map[string]any{
			"role":              string(msg.Role),
			"content":           msg.Content,
			"reasoning_content": msg.ReasoningContent,
		}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{
			"role":              string(msg.Role),
			"content":           msg.Content,
			"reasoning_content": msg.ReasoningContent,
		}
	}
	return m
}

func (a *App) AbortChatWithAgent() {
	a.agentMu.Lock()
	defer a.agentMu.Unlock()
	if a.agentCancel != nil {
		a.agentCancel()
		a.agentCancel = nil
	}
}

func (a *App) AnalyzeSentimentWithFreqWeight(text string) map[string]any {
	result, cleanFrequencies := data.NewsAnalyze(text, false)
	return map[string]any{
		"result":      result,
		"frequencies": cleanFrequencies,
	}
}

func (a *App) GetAIResponseResultList(query models.AIResponseResultQuery) *models.AIResponseResultPageData {
	page, err := a.analysisService.GetResultPage(a.ctx, query)
	if err != nil {
		return &models.AIResponseResultPageData{}
	}
	return page
}
func (a *App) DeleteAIResponseResult(id uint) string {
	if err := a.analysisService.DeleteResult(a.ctx, id); err != nil {
		return "删除失败"
	}
	return "删除成功"
}
func (a *App) BatchDeleteAIResponseResult(ids []uint) string {
	if err := a.analysisService.BatchDeleteResults(a.ctx, ids); err != nil {
		return "删除失败"
	}
	return "删除成功"
}

func (a *App) GetStockChanges(changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	return data.NewStockChangesApi().GetStockChanges(changeTypes, pageIndex, pageSize)
}

func (a *App) GetAllStockChangesWithPaging(pageSize int) *data.StockChangesResponse {
	all := data.NewStockChangesApi().GetAllStockChangesWithPaging(pageSize)
	historyService := data.NewStockChangeHistoryService()
	_, _ = historyService.SaveStockChangesWithDedup(all.Data)
	return all
}

func (a *App) GetStockChangeHistory(query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	result, err := data.NewStockChangeHistoryService().GetHistoryList(query)
	if err != nil {
		return &models.StockChangeHistoryPageData{}
	}
	return result
}

func (a *App) SaveStockChangesToHistory(changeTypes []int) string {
	api := data.NewStockChangesApi()
	result := api.GetStockChanges(changeTypes, 0, 500)
	if result == nil || len(result.Data) == 0 {
		return "没有获取到异动数据"
	}

	err := data.NewStockChangeHistoryService().SaveStockChanges(result.Data)
	if err != nil {
		return "保存失败: " + err.Error()
	}
	return fmt.Sprintf("成功保存 %d 条异动数据", len(result.Data))
}

func (a *App) DeleteStockChangeHistory(days int) string {
	err := data.NewStockChangeHistoryService().DeleteOldData(days)
	if err != nil {
		return "删除失败: " + err.Error()
	}
	return fmt.Sprintf("已删除 %d 天前的历史数据", days)
}

func (a *App) GetAiRecommendStocksList(query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	page, err := data.NewAiRecommendStocksService().GetAiRecommendStocksList(&query)
	if err != nil {
		return &models.AiRecommendStocksPageData{}
	}
	return page
}
func (a *App) DeleteAiRecommendStocks(id uint) string {
	err := data.NewAiRecommendStocksService().DeleteAiRecommendStocks(id)
	if err != nil {
		return "删除失败"
	}
	return "删除成功"
}

func (a *App) UpdateAiRecommendStocksAlert(id uint, enableAlert bool) string {
	err := data.NewAiRecommendStocksService().UpdateAiRecommendStocksAlert(id, enableAlert)
	if err != nil {
		return "更新预警状态失败"
	}
	return "更新预警状态成功"
}

func (a *App) GetPromptTemplateList(query models.PromptTemplateQuery) *models.PromptTemplatePageData {
	if a.configService == nil {
		return &models.PromptTemplatePageData{}
	}
	page, err := a.configService.GetPromptTemplatePage(a.ctx, query)
	if err != nil {
		return &models.PromptTemplatePageData{}
	}
	return page
}

func (a *App) AddPromptTemplate(template models.PromptTemplate) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) UpdatePromptTemplate(template models.PromptTemplate) string {
	if a.configService == nil {
		return "保存失败"
	}
	return a.configService.SavePromptTemplate(a.ctx, template)
}

func (a *App) DeletePromptTemplate(id uint) string {
	if a.configService == nil {
		return "删除失败"
	}
	return a.configService.DeletePromptTemplate(a.ctx, id)
}

func (a *App) GetAllStockInfoList(query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	page, err := data.NewStockDataApi().GetAllStockInfoList(&query)
	if err != nil {
		return &data.AllStockInfoPageData{}
	}
	return page
}

func (a *App) GetAllStockInfoById(id uint) *models.AllStockInfo {
	stock, err := data.NewStockDataApi().GetAllStockInfoById(id)
	if err != nil {
		return &models.AllStockInfo{}
	}
	return stock
}

func (a *App) AddAllStockInfo(stock models.AllStockInfo) string {
	err := data.NewStockDataApi().AddAllStockInfo(stock)
	if err != nil {
		return "操作失败: " + err.Error()
	}
	return "操作成功"
}

func (a *App) DeleteAllStockInfo(id uint) string {
	err := data.NewStockDataApi().DeleteAllStockInfo(id)
	if err != nil {
		return "删除失败: " + err.Error()
	}
	return "删除成功"
}

func (a *App) BatchDeleteAllStockInfo(ids []uint) string {
	err := data.NewStockDataApi().BatchDeleteAllStockInfo(ids)
	if err != nil {
		return "批量删除失败: " + err.Error()
	}
	return "批量删除成功"
}

func (a *App) GetAllMarkets() []string {
	markets, err := data.NewStockDataApi().GetAllMarkets()
	if err != nil {
		return []string{}
	}
	return markets
}

func (a *App) GetAllIndustries() []string {
	industries, err := data.NewStockDataApi().GetAllIndustries()
	if err != nil {
		return []string{}
	}
	return industries
}

func (a *App) GetAllConcepts() []string {
	concepts, err := data.NewStockDataApi().GetAllConcepts()
	if err != nil {
		return []string{}
	}
	return concepts
}

func (a *App) GetStockRealTimePrice(stockCode string) map[string]any {
	service := a.legacyMarketReads()
	if service == nil {
		return map[string]any{
			"code":    -1,
			"message": "获取股票价格失败",
			"price":   0,
		}
	}
	stock := service.LoadRealtimePrice(stockCode)
	price, _ := convertor.ToFloat(stock.Price)
	return map[string]any{
		"code":    0,
		"message": "success",
		"price":   price,
		"name":    stock.StockName,
	}
}
