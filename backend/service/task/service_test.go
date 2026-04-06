package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type fakeStore struct {
	task        *models.CronTask
	enabled     []models.CronTask
	page        *models.CronTaskPageResp
	createErr   error
	updateErr   error
	deleteErr   error
	enableErr   error
	getErr      error
	runErr      error
	validateErr error
	createTask  *models.CronTask
	updateTask  *models.CronTask
	deleteID    uint
	enableID    uint
	enableValue bool
	runID       uint
	search      []models.CronTask
}

func (f *fakeStore) Create(task *models.CronTask) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createTask = task
	if task.ID == 0 {
		task.ID = 7
	}
	return nil
}

func (f *fakeStore) Update(task *models.CronTask) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updateTask = task
	return nil
}

func (f *fakeStore) Delete(id uint) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeStore) GetByID(id uint) (*models.CronTask, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.task != nil {
		return f.task, nil
	}
	return &models.CronTask{ID: id, Name: "市场分析", CronExpr: "0 */5 * * * *", Enable: true}, nil
}

func (f *fakeStore) List(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return f.page
}

func (f *fakeStore) EnableTask(id uint, enable bool) error {
	f.enableID = id
	f.enableValue = enable
	return f.enableErr
}

func (f *fakeStore) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	f.runID = task.ID
	return f.runErr
}

func (f *fakeStore) GetTaskTypes() []lo.Tuple2[string, string] {
	return []lo.Tuple2[string, string]{{A: "market_analysis", B: "市场分析"}}
}

func (f *fakeStore) ValidateCronExpr(expr string) error {
	return f.validateErr
}

func (f *fakeStore) CalculateNextRunTime(expr string) time.Time {
	return time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local)
}

func (f *fakeStore) CalculateNextRunTimes(expr string, count int) []time.Time {
	return []time.Time{
		time.Date(2026, 4, 6, 10, 0, 0, 0, time.Local),
		time.Date(2026, 4, 6, 10, 5, 0, 0, time.Local),
	}
}

func (f *fakeStore) Search(keyword string) []models.CronTask {
	return f.search
}

func (f *fakeStore) GetAllEnabled() []models.CronTask {
	return f.enabled
}

type fakeScheduler struct {
	registeredKey  string
	registeredSpec string
	registered     []string
	unregistered   []string
	registerErr    error
}

func (f *fakeScheduler) Register(key, spec string, job func()) error {
	f.registeredKey = key
	f.registeredSpec = spec
	f.registered = append(f.registered, key)
	return f.registerErr
}

func (f *fakeScheduler) Unregister(key string) {
	f.unregistered = append(f.unregistered, key)
}

func TestService_CreateUsesStableSchedulerKey(t *testing.T) {
	store := &fakeStore{}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)

	msg := svc.Create(context.Background(), &models.CronTask{Name: "市场分析", CronExpr: "0 */5 * * * *", Enable: true})

	if msg != "创建成功" || scheduler.registeredKey != "cron-task-7" || scheduler.registeredSpec != "0 */5 * * * *" {
		t.Fatalf("unexpected create scheduling: msg=%q key=%q spec=%q", msg, scheduler.registeredKey, scheduler.registeredSpec)
	}
}

func TestService_UpdateReplacesExistingRegistration(t *testing.T) {
	store := &fakeStore{}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)

	msg := svc.Update(context.Background(), &models.CronTask{ID: 9, Name: "市场分析", CronExpr: "0 */10 * * * *", Enable: true})

	if msg != "更新成功" || len(scheduler.unregistered) != 1 || scheduler.unregistered[0] != "cron-task-9" || scheduler.registeredKey != "cron-task-9" {
		t.Fatalf("unexpected update scheduling: msg=%q unregister=%#v key=%q", msg, scheduler.unregistered, scheduler.registeredKey)
	}
}

func TestService_EnableAndDisableUseStableKey(t *testing.T) {
	store := &fakeStore{task: &models.CronTask{ID: 11, Name: "任务", CronExpr: "0 */15 * * * *", Enable: true}}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)

	if msg := svc.Enable(context.Background(), 11, false); msg != "操作成功" {
		t.Fatalf("unexpected disable message: %q", msg)
	}
	if len(scheduler.unregistered) == 0 || scheduler.unregistered[0] != "cron-task-11" {
		t.Fatalf("unexpected unregister list after disable: %#v", scheduler.unregistered)
	}

	scheduler.unregistered = nil
	if msg := svc.Enable(context.Background(), 11, true); msg != "操作成功" || scheduler.registeredKey != "cron-task-11" {
		t.Fatalf("unexpected enable scheduling: msg=%q key=%q", msg, scheduler.registeredKey)
	}
}

func TestService_RunNowDelegatesToStore(t *testing.T) {
	store := &fakeStore{task: &models.CronTask{ID: 12, Name: "执行任务"}}
	svc := NewService(store, &fakeScheduler{}, context.Background)

	if err := svc.RunNow(context.Background(), 12); err != nil || store.runID != 12 {
		t.Fatalf("unexpected run delegation: err=%v id=%d", err, store.runID)
	}
}

func TestService_ValidateCronExprPropagatesFailure(t *testing.T) {
	store := &fakeStore{validateErr: errors.New("bad cron")}
	svc := NewService(store, &fakeScheduler{}, context.Background)

	if msg := svc.ValidateCronExpr(context.Background(), "*"); msg == "有效表达式" {
		t.Fatalf("expected validation failure, got %q", msg)
	}
}

func TestService_RestoreSchedulesUsesStableKeysForEnabledTasks(t *testing.T) {
	store := &fakeStore{
		enabled: []models.CronTask{
			{ID: 21, Name: "市场分析", CronExpr: "0 */5 * * * *", Enable: true},
			{ID: 22, Name: "个股分析", CronExpr: "0 */10 * * * *", Enable: true},
		},
	}
	scheduler := &fakeScheduler{}
	svc := NewService(store, scheduler, context.Background)

	if err := svc.RestoreSchedules(context.Background()); err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}
	if len(scheduler.registered) != 2 {
		t.Fatalf("expected 2 registered tasks, got %#v", scheduler.registered)
	}
	if scheduler.registered[0] != "cron-task-21" || scheduler.registered[1] != "cron-task-22" {
		t.Fatalf("unexpected registered keys: %#v", scheduler.registered)
	}
}
