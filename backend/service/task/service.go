package task

import (
	"context"
	"fmt"
	"time"

	"go-stock/backend/models"

	"github.com/samber/lo"
)

type Store interface {
	Create(task *models.CronTask) error
	Update(task *models.CronTask) error
	Delete(id uint) error
	GetByID(id uint) (*models.CronTask, error)
	List(query *models.CronTaskQuery) *models.CronTaskPageResp
	EnableTask(id uint, enable bool) error
	ExecuteTask(ctx context.Context, task *models.CronTask) error
	GetTaskTypes() []lo.Tuple2[string, string]
	ValidateCronExpr(expr string) error
	CalculateNextRunTime(expr string) time.Time
	CalculateNextRunTimes(expr string, count int) []time.Time
	Search(keyword string) []models.CronTask
}

type Scheduler interface {
	Register(key, spec string, job func()) error
	Unregister(key string)
}

type Service struct {
	store       Store
	scheduler   Scheduler
	contextFunc func() context.Context
}

func NewService(store Store, scheduler Scheduler, contextFunc func() context.Context) *Service {
	if store == nil {
		panic("task: store dependency is required")
	}
	if scheduler == nil {
		panic("task: scheduler dependency is required")
	}
	if contextFunc == nil {
		panic("task: context func is required")
	}
	return &Service{
		store:       store,
		scheduler:   scheduler,
		contextFunc: contextFunc,
	}
}

func scheduleKey(id uint) string {
	return fmt.Sprintf("cron-task-%d", id)
}

func (s *Service) syncSchedule(task *models.CronTask) error {
	if task == nil {
		return nil
	}

	key := scheduleKey(task.ID)
	s.scheduler.Unregister(key)
	if !task.Enable {
		return nil
	}

	return s.scheduler.Register(key, task.CronExpr, func() {
		latest, err := s.store.GetByID(task.ID)
		if err != nil || latest == nil {
			return
		}
		_ = s.store.ExecuteTask(s.contextFunc(), latest)
	})
}

func (s *Service) Create(ctx context.Context, task *models.CronTask) string {
	_ = ctx
	if err := s.store.Create(task); err != nil {
		return fmt.Sprintf("创建失败：%v", err)
	}
	if err := s.syncSchedule(task); err != nil {
		return "任务创建成功,但定时失败"
	}
	return "创建成功"
}

func (s *Service) Update(ctx context.Context, task *models.CronTask) string {
	_ = ctx
	if err := s.store.Update(task); err != nil {
		return fmt.Sprintf("更新失败：%v", err)
	}
	if err := s.syncSchedule(task); err != nil {
		return "更新成功,但定时失败"
	}
	return "更新成功"
}

func (s *Service) Delete(ctx context.Context, id uint) string {
	_ = ctx
	s.scheduler.Unregister(scheduleKey(id))
	if err := s.store.Delete(id); err != nil {
		return fmt.Sprintf("删除失败：%v", err)
	}
	return "删除成功"
}

func (s *Service) GetByID(ctx context.Context, id uint) (*models.CronTask, error) {
	_ = ctx
	return s.store.GetByID(id)
}

func (s *Service) List(ctx context.Context, query *models.CronTaskQuery) *models.CronTaskPageResp {
	_ = ctx
	return s.store.List(query)
}

func (s *Service) Enable(ctx context.Context, id uint, enable bool) string {
	_ = ctx
	if err := s.store.EnableTask(id, enable); err != nil {
		return fmt.Sprintf("操作失败：%v", err)
	}

	key := scheduleKey(id)
	s.scheduler.Unregister(key)
	if !enable {
		return "操作成功"
	}

	task, err := s.store.GetByID(id)
	if err != nil {
		return fmt.Sprintf("操作失败：%v", err)
	}
	if task == nil {
		return "操作失败：任务不存在"
	}

	task.Enable = enable
	if err := s.syncSchedule(task); err != nil {
		return "操作成功,但定时失败"
	}
	return "操作成功"
}

func (s *Service) RunNow(ctx context.Context, id uint) error {
	task, err := s.store.GetByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("任务不存在")
	}
	return s.store.ExecuteTask(ctx, task)
}

func (s *Service) GetTaskTypes(ctx context.Context) []lo.Tuple2[string, string] {
	_ = ctx
	return s.store.GetTaskTypes()
}

func (s *Service) ValidateCronExpr(ctx context.Context, expr string) string {
	_ = ctx
	if err := s.store.ValidateCronExpr(expr); err != nil {
		return fmt.Sprintf("无效表达式：%v", err)
	}
	return "有效表达式"
}

func (s *Service) Search(ctx context.Context, keyword string) []models.CronTask {
	_ = ctx
	return s.store.Search(keyword)
}

func (s *Service) CalculateNextRunTime(ctx context.Context, cronExpr string) time.Time {
	_ = ctx
	return s.store.CalculateNextRunTime(cronExpr)
}

func (s *Service) CalculateNextRunTimes(ctx context.Context, cronExpr string, count int) []time.Time {
	_ = ctx
	return s.store.CalculateNextRunTimes(cronExpr, count)
}
