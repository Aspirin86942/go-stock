package main

import (
	"context"
	"testing"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type fakeTaskService struct {
	createResult string
	updateResult string
	deleteResult string
	enableResult string
	validateResp string
	runErr       error
	task         *models.CronTask
	page         *models.CronTaskPageResp
	types        []lo.Tuple2[string, string]
	search       []models.CronTask
	createTask   *models.CronTask
	updateTask   *models.CronTask
	deleteID     uint
	enableID     uint
	enableValue  bool
	runID        uint
	runCalled    chan uint
	validateExpr string
	searchKey    string
	nextRunExpr  string
	nextRunsExpr string
	nextRunsCnt  int
}

func (f *fakeTaskService) Create(ctx context.Context, task *models.CronTask) string {
	f.createTask = task
	return f.createResult
}

func (f *fakeTaskService) Update(ctx context.Context, task *models.CronTask) string {
	f.updateTask = task
	return f.updateResult
}

func (f *fakeTaskService) Delete(ctx context.Context, id uint) string {
	f.deleteID = id
	return f.deleteResult
}

func (f *fakeTaskService) GetByID(ctx context.Context, id uint) (*models.CronTask, error) {
	return f.task, nil
}

func (f *fakeTaskService) List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp {
	return f.page
}

func (f *fakeTaskService) Enable(ctx context.Context, id uint, enable bool) string {
	f.enableID = id
	f.enableValue = enable
	return f.enableResult
}

func (f *fakeTaskService) RunNow(ctx context.Context, id uint) error {
	f.runID = id
	if f.runCalled != nil {
		f.runCalled <- id
	}
	return f.runErr
}

func (f *fakeTaskService) GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string] {
	return f.types
}

func (f *fakeTaskService) ValidateCronExpr(ctx context.Context, expr string) string {
	f.validateExpr = expr
	return f.validateResp
}

func (f *fakeTaskService) Search(ctx context.Context, keyword string) []models.CronTask {
	f.searchKey = keyword
	return f.search
}

func (f *fakeTaskService) CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time {
	f.nextRunExpr = cronExpr
	return time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local)
}

func (f *fakeTaskService) CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time {
	f.nextRunsExpr = cronExpr
	f.nextRunsCnt = count
	return []time.Time{
		time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 6, 10, 5, 0, 0, time.Local),
	}
}

func TestApp_TaskMethodsDelegateToTaskService(t *testing.T) {
	app := NewApp()
	fake := &fakeTaskService{
		createResult: "创建成功",
		updateResult: "更新成功",
		deleteResult: "删除成功",
		enableResult: "操作成功",
		validateResp: "有效表达式",
		task:         &models.CronTask{ID: 8, Name: "市场分析"},
		page:         &models.CronTaskPageResp{Total: 1, Data: []models.CronTask{{ID: 8, Name: "市场分析"}}},
		types:        []lo.Tuple2[string, string]{{A: "market_analysis", B: "市场分析"}},
		search:       []models.CronTask{{ID: 8, Name: "市场分析"}},
		runCalled:    make(chan uint, 1),
	}
	app.taskService = fake

	if msg := app.CreateCronTask(&models.CronTask{Name: "创建任务"}); msg != "创建成功" {
		t.Fatalf("unexpected create message: %q", msg)
	}
	if fake.createTask == nil || fake.createTask.Name != "创建任务" {
		t.Fatalf("unexpected create payload: %#v", fake.createTask)
	}

	if msg := app.UpdateCronTask(&models.CronTask{ID: 8, Name: "更新任务"}); msg != "更新成功" {
		t.Fatalf("unexpected update message: %q", msg)
	}
	if fake.updateTask == nil || fake.updateTask.Name != "更新任务" {
		t.Fatalf("unexpected update payload: %#v", fake.updateTask)
	}

	if msg := app.DeleteCronTask(8); msg != "删除成功" || fake.deleteID != 8 {
		t.Fatalf("unexpected delete result: msg=%q id=%d", msg, fake.deleteID)
	}

	if task := app.GetCronTaskByID(8); task == nil || task.ID != 8 {
		t.Fatalf("unexpected get task result: %#v", task)
	}

	if page := app.GetCronTaskList(&models.CronTaskQuery{Page: 1, PageSize: 10}); page.Total != 1 {
		t.Fatalf("unexpected task page: %#v", page)
	}

	if msg := app.EnableCronTask(8, true); msg != "操作成功" || fake.enableID != 8 || !fake.enableValue {
		t.Fatalf("unexpected enable result: msg=%q id=%d enable=%v", msg, fake.enableID, fake.enableValue)
	}

	if msg := app.ExecuteCronTaskNow(8); msg != "任务执行中" {
		t.Fatalf("unexpected execute message: %q", msg)
	}
	select {
	case got := <-fake.runCalled:
		if got != 8 || fake.runID != 8 {
			t.Fatalf("unexpected run delegation: got=%d id=%d", got, fake.runID)
		}
	case <-time.After(time.Second):
		t.Fatal("ExecuteCronTaskNow() should call taskService.RunNow asynchronously")
	}

	if got := app.GetCronTaskTypes(); len(got) != 1 || got[0].A != "market_analysis" {
		t.Fatalf("unexpected task types: %#v", got)
	}

	if got := app.ValidateCronExpr("0 */5 * * * *"); got != "有效表达式" || fake.validateExpr != "0 */5 * * * *" {
		t.Fatalf("unexpected validate result: msg=%q expr=%q", got, fake.validateExpr)
	}

	if got := app.SearchCronTasks("市场"); len(got) != 1 || fake.searchKey != "市场" {
		t.Fatalf("unexpected search result: %#v keyword=%q", got, fake.searchKey)
	}

	if got := app.CalculateNextRunTime("0 */5 * * * *"); got != "2026-04-06 10:00:00" || fake.nextRunExpr != "0 */5 * * * *" {
		t.Fatalf("unexpected next run time: got=%q expr=%q", got, fake.nextRunExpr)
	}

	if got := app.CalculateNextRunTimes("0 */5 * * * *", 2); len(got) != 2 || got[1] != "2026-04-06 10:05:00" || fake.nextRunsExpr != "0 */5 * * * *" || fake.nextRunsCnt != 2 {
		t.Fatalf("unexpected next run times: got=%#v expr=%q count=%d", got, fake.nextRunsExpr, fake.nextRunsCnt)
	}
}
