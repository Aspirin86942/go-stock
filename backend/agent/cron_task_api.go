package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type CronTaskApi struct{}

func NewCronTaskApi() *CronTaskApi {
	return &CronTaskApi{}
}

func (a *CronTaskApi) Create(task *models.CronTask) error {
	return db.Dao.Create(task).Error
}

func (a *CronTaskApi) Update(task *models.CronTask) error {
	if task == nil || task.ID == 0 {
		return fmt.Errorf("无效的任务ID")
	}

	updates := map[string]any{
		"name":        task.Name,
		"cron_expr":   task.CronExpr,
		"task_type":   task.TaskType,
		"target":      task.Target,
		"params":      task.Params,
		"enable":      task.Enable,
		"status":      task.Status,
		"description": task.Description,
	}

	return db.Dao.Model(&models.CronTask{}).
		Where("id = ?", task.ID).
		Updates(updates).Error
}

func (a *CronTaskApi) Delete(id uint) error {
	return db.Dao.Delete(&models.CronTask{}, id).Error
}

func (a *CronTaskApi) GetByID(id uint) (*models.CronTask, error) {
	var task models.CronTask
	err := db.Dao.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (a *CronTaskApi) List(query *models.CronTaskQuery) *models.CronTaskPageResp {
	var tasks []models.CronTask
	var total int64

	dbQuery := db.Dao.Model(&models.CronTask{})

	if query.Name != "" {
		dbQuery = dbQuery.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.TaskType != "" {
		dbQuery = dbQuery.Where("task_type = ?", query.TaskType)
	}
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Enable != nil {
		dbQuery = dbQuery.Where("enable = ?", *query.Enable)
	}

	dbQuery.Count(&total)

	page := query.Page
	pageSize := query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	err := dbQuery.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(moduleTrace("cron-task-list")).Error(
				"task.list_failed",
				"query cron task list failed",
				logger.Err(err),
			)
		}
		return nil
	}

	return &models.CronTaskPageResp{
		Total: int(total),
		Data:  tasks,
	}
}

func (a *CronTaskApi) GetAll() []models.CronTask {
	var tasks []models.CronTask
	db.Dao.Where("enable = ?", true).Order("created_at DESC").Find(&tasks)
	return tasks
}

func (a *CronTaskApi) EnableTask(id uint, enable bool) error {
	return db.Dao.Model(&models.CronTask{}).Where("id = ?", id).Updates(map[string]any{
		"enable": enable,
	}).Error
}

func (a *CronTaskApi) UpdateRunInfo(ctx context.Context, id uint, lastRunAt time.Time, nextRunAt *time.Time, lastRunResult string) error {
	dao := db.Dao
	if ctx != nil {
		dao = dao.WithContext(ctx)
	}
	return dao.Model(&models.CronTask{}).Where("id = ?", id).Updates(map[string]any{
		"last_run_at":     lastRunAt,
		"next_run_at":     nextRunAt,
		"run_count":       gorm.Expr("run_count + 1"),
		"last_run_result": lastRunResult,
	}).Error
}

func (a *CronTaskApi) GetTaskTypes() []lo.Tuple2[string, string] {
	return []lo.Tuple2[string, string]{
		{A: "stock_analysis", B: "股票分析"},
		{A: "market_analysis", B: "市场分析"},
		{A: "global_stock_index_cache", B: "全球指数缓存"},
		{A: "stock_change_save", B: "异动数据保存"},
	}
}

func (a *CronTaskApi) ValidateCronExpr(expr string) error {
	_, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(expr)
	return err
}

func (a *CronTaskApi) CalculateNextRunTimes(cronExpr string, count int) []time.Time {
	if count <= 0 {
		return []time.Time{}
	}

	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(cronExpr)
	if err != nil {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(moduleTrace("cron-task-parse")).Error(
				"task.cron_parse_failed",
				"parse cron expression failed",
				logger.String("cron_expr", cronExpr),
				logger.Err(err),
			)
		}
		return []time.Time{}
	}

	times := make([]time.Time, 0, count)
	next := time.Now()
	for i := 0; i < count; i++ {
		next = schedule.Next(next)
		times = append(times, next)
	}
	return times
}

func (a *CronTaskApi) SearchTasks(keyword string) []models.CronTask {
	var tasks []models.CronTask
	query := db.Dao.Model(&models.CronTask{})
	if keyword != "" {
		keyword = strings.TrimSpace(keyword)
		query = query.Where("name LIKE ? OR target LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Order("created_at DESC").Limit(20).Find(&tasks)
	return tasks
}

func (a *CronTaskApi) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	ctx, trace := ensureModuleTraceContext(ctx, "cron-task-execute")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.execute_started",
			"started cron task execution",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
			logger.String("task_type", task.TaskType),
		)
	}

	now := time.Now()
	nextRunAt := a.CalculateNextRunTime(task.CronExpr)

	var runResult string
	err := a.executeTaskByType(ctx, task)
	if err != nil {
		runResult = "失败: " + err.Error()
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(trace).Error(
				"task.execute_failed",
				"cron task execution failed",
				logger.Uint("task_id", task.ID),
				logger.String("task_name", task.Name),
				logger.String("error_class", "task_error"),
				logger.Err(err),
			)
		}
	} else {
		runResult = "成功"
	}

	err2 := a.UpdateRunInfo(ctx, task.ID, now, &nextRunAt, runResult)
	if err2 != nil {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(trace).Error(
				"task.update_run_info_failed",
				"update cron task run info failed",
				logger.Uint("task_id", task.ID),
				logger.Err(err2),
			)
		}
	}

	return err
}

func (a *CronTaskApi) executeTaskByType(ctx context.Context, task *models.CronTask) error {
	switch task.TaskType {
	case "stock_analysis":
		return a.executeStockAnalysis(ctx, task)
	case "market_analysis":
		return a.executeMarketAnalysis(ctx, task)
	case "global_stock_index_cache":
		return a.executeGlobalStockIndexCache(ctx, task)
	case "fund_analysis":
		return a.executeFundAnalysis(ctx, task)
	case "news_fetch":
		return a.executeNewsFetch(ctx, task)
	case "stock_monitor":
		return a.executeStockMonitor(ctx, task)
	case "stock_change_save":
		return a.executeStockChangeSave(ctx, task)
	case "custom":
		return a.executeCustomTask(ctx, task)
	default:
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(moduleTrace(ctx, "cron-task-dispatch")).Warn(
				"task.unknown_type",
				"unknown cron task type",
				logger.String("task_type", task.TaskType),
			)
		}
		return fmt.Errorf("未知任务类型：%s", task.TaskType)
	}
}

func (a *CronTaskApi) CalculateNextRunTime(cronExpr string) time.Time {
	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(cronExpr)
	if err != nil {
		return time.Now().Add(time.Hour)
	}
	return schedule.Next(time.Now())
}

func (a *CronTaskApi) executeStockAnalysis(ctx context.Context, task *models.CronTask) error {
	trace := moduleTrace(ctx, "cron-task-stock-analysis")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.stock_analysis_started",
			"started stock analysis task",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
		)
	}
	var params struct {
		PromptId    int    `json:"promptId"`
		AiConfigId  int    `json:"aiConfigId"`
		SysPromptId int    `json:"sysPromptId"`
		Thinking    bool   `json:"thinking"`
		StockCode   string `json:"stockCode"`
		StockName   string `json:"stockName"`
	}
	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Error(
					"task.stock_analysis_params_invalid",
					"parse stock analysis task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}

	prompt := fmt.Sprintf("分析总结市场资讯，针对%s[%s]，找出潜在投资机会", params.StockName, params.StockCode)
	prompt = data.NewPromptTemplateApi().GetPromptTemplateByID(params.PromptId)
	var tools []data.Tool
	tools = data.Tools(tools)
	msgs := data.NewDeepSeekOpenAi(ctx, params.AiConfigId).NewChatStream(params.StockName, data.ConvertTushareCodeToStockCode(params.StockCode), prompt, &params.SysPromptId, tools, params.Thinking)
	content := &strings.Builder{}
	for msg := range msgs {
		content.WriteString(msg["content"].(string))
	}
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.stock_analysis_completed",
			"stock analysis task produced content",
			logger.Uint("task_id", task.ID),
			logger.Int("content_length", content.Len()),
		)
	}
	data.NewDeepSeekOpenAi(ctx, params.AiConfigId).SaveAIResponseResult(params.StockCode, params.StockName, content.String(), "", prompt)
	return nil
}

func (a *CronTaskApi) executeFundAnalysis(ctx context.Context, task *models.CronTask) error {
	var params struct {
		FundCodes  []string `json:"fund_codes"`
		AiConfigId int      `json:"ai_config_id"`
	}

	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(moduleTrace(ctx, "cron-task-fund-analysis")).Error(
					"task.fund_analysis_params_invalid",
					"parse fund analysis task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}

	for _, fundCode := range params.FundCodes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(moduleTrace(ctx, "cron-task-fund-analysis")).Info(
					"task.fund_analysis_item",
					"analyzing fund",
					logger.Uint("task_id", task.ID),
					logger.String("fund_code", fundCode),
				)
			}
		}
	}

	return nil
}

func (a *CronTaskApi) executeNewsFetch(ctx context.Context, task *models.CronTask) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		data.NewMarketNewsApi().TelegraphList(30)
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(moduleTrace(ctx, "cron-task-news-fetch")).Info(
				"task.news_fetch_completed",
				"news fetch task completed",
				logger.Uint("task_id", task.ID),
			)
		}
		return nil
	}
}

func (a *CronTaskApi) executeStockMonitor(ctx context.Context, task *models.CronTask) error {
	var params struct {
		StockCodes      []string `json:"stock_codes"`
		PriceThreshold  float64  `json:"price_threshold"`
		ChangeThreshold float64  `json:"change_threshold"`
	}

	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(moduleTrace(ctx, "cron-task-stock-monitor")).Error(
					"task.stock_monitor_params_invalid",
					"parse stock monitor task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}

	for _, stockCode := range params.StockCodes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(moduleTrace(ctx, "cron-task-stock-monitor")).Info(
					"task.stock_monitor_item",
					"monitoring stock",
					logger.Uint("task_id", task.ID),
					logger.String("stock_code", stockCode),
				)
			}
		}
	}

	return nil
}

func (a *CronTaskApi) executeCustomTask(ctx context.Context, task *models.CronTask) error {
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(moduleTrace(ctx, "cron-task-custom")).Info(
			"task.custom_started",
			"started custom task",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
		)
	}
	return nil
}

func (a *CronTaskApi) executeMarketAnalysis(ctx context.Context, task *models.CronTask) error {
	trace := moduleTrace(ctx, "cron-task-market-analysis")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.market_analysis_started",
			"started market analysis task",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
		)
	}
	var params struct {
		PromptId    int  `json:"promptId"`
		AiConfigId  int  `json:"aiConfigId"`
		SysPromptId int  `json:"sysPromptId"`
		Thinking    bool `json:"thinking"`
	}
	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Error(
					"task.market_analysis_params_invalid",
					"parse market analysis task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}

	prompt := "分析总结市场资讯，找出潜在投资机会"
	prompt = data.NewPromptTemplateApi().GetPromptTemplateByID(params.PromptId)
	content := &strings.Builder{}

	ch := NewStockAiAgentApi().ChatWithContext(ctx, prompt, params.AiConfigId, &params.SysPromptId, false, 0, false)
	for msg := range ch {
		if msg.ReasoningContent != "" {
			content.WriteString(msg.ReasoningContent)
		}
		content.WriteString(msg.Content)
	}
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.market_analysis_completed",
			"market analysis task produced content",
			logger.Uint("task_id", task.ID),
			logger.Int("content_length", content.Len()),
		)
	}
	data.NewDeepSeekOpenAi(ctx, params.AiConfigId).SaveAIResponseResult("市场分析", "市场分析", content.String(), "", prompt)
	return nil
}

func (a *CronTaskApi) executeGlobalStockIndexCache(ctx context.Context, task *models.CronTask) error {
	trace := moduleTrace(ctx, "cron-task-global-index-cache")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.global_index_cache_started",
			"started global stock index cache task",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
		)
	}
	var params struct {
		CrawlTimeOut uint `json:"crawlTimeOut"`
	}
	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Error(
					"task.global_index_cache_params_invalid",
					"parse global stock index cache task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}
	if params.CrawlTimeOut == 0 {
		params.CrawlTimeOut = 30
	}
	return data.NewMarketNewsApi().CacheGlobalStockIndexes(params.CrawlTimeOut)
}

func (a *CronTaskApi) executeStockChangeSave(ctx context.Context, task *models.CronTask) error {
	trace := moduleTrace(ctx, "cron-task-stock-change-save")
	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.stock_change_save_started",
			"started stock change save task",
			logger.Uint("task_id", task.ID),
			logger.String("task_name", task.Name),
		)
	}

	if !isTradingTime() {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(trace).Info(
				"task.stock_change_save_skipped",
				"skip stock change save outside trading hours",
				logger.Uint("task_id", task.ID),
			)
		}
		return nil
	}

	var params struct {
		ChangeTypes []int `json:"changeTypes"`
		DeleteDays  int   `json:"deleteDays"`
	}

	if task.Params != "" {
		err := json.Unmarshal([]byte(task.Params), &params)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Error(
					"task.stock_change_save_params_invalid",
					"parse stock change save task params failed",
					logger.Uint("task_id", task.ID),
					logger.Err(err),
				)
			}
			return err
		}
	}

	if len(params.ChangeTypes) == 0 {
		params.ChangeTypes = []int{8201, 8202, 8193, 4, 8204, 8203, 8194, 8}
	}

	api := data.NewStockChangesApi()
	result := api.GetStockChanges(params.ChangeTypes, 0, 500)
	if result == nil || len(result.Data) == 0 {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(trace).Info(
				"task.stock_change_save_empty",
				"no stock change data fetched",
				logger.Uint("task_id", task.ID),
			)
		}
		return nil
	}

	savedCount, err := data.NewStockChangeHistoryService().SaveStockChangesWithDedup(result.Data)
	if err != nil {
		if log := taskLogger("agent.cron_task"); log != nil {
			log.WithTrace(trace).Error(
				"task.stock_change_save_failed",
				"save stock change data failed",
				logger.Uint("task_id", task.ID),
				logger.Err(err),
			)
		}
		return err
	}

	if log := taskLogger("agent.cron_task"); log != nil {
		log.WithTrace(trace).Info(
			"task.stock_change_save_completed",
			"saved stock change data successfully",
			logger.Uint("task_id", task.ID),
			logger.Int("saved_count", savedCount),
		)
	}

	if params.DeleteDays > 0 {
		err = data.NewStockChangeHistoryService().DeleteOldData(params.DeleteDays)
		if err != nil {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Warn(
					"task.stock_change_cleanup_failed",
					"delete old stock change data failed",
					logger.Uint("task_id", task.ID),
					logger.Int("delete_days", params.DeleteDays),
					logger.Err(err),
				)
			}
		} else {
			if log := taskLogger("agent.cron_task"); log != nil {
				log.WithTrace(trace).Info(
					"task.stock_change_cleanup_completed",
					"deleted old stock change history data",
					logger.Uint("task_id", task.ID),
					logger.Int("delete_days", params.DeleteDays),
				)
			}
		}
	}

	return nil
}

func isTradingTime() bool {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	hour, minute := now.Hour(), now.Minute()
	currentTime := hour*100 + minute

	morningStart := 915
	morningEnd := 1130
	afternoonStart := 1300
	afternoonEnd := 1500

	isMorning := currentTime >= morningStart && currentTime <= morningEnd
	isAfternoon := currentTime >= afternoonStart && currentTime <= afternoonEnd

	return isMorning || isAfternoon
}
