package task

import (
	"context"
	"time"

	"go-stock/backend/agent"
	"go-stock/backend/models"

	"github.com/samber/lo"
)

type AgentStore struct{}

func NewStore() *AgentStore {
	return &AgentStore{}
}

func (s *AgentStore) Create(task *models.CronTask) error {
	return agent.NewCronTaskApi().Create(task)
}

func (s *AgentStore) Update(task *models.CronTask) error {
	return agent.NewCronTaskApi().Update(task)
}

func (s *AgentStore) Delete(id uint) error {
	return agent.NewCronTaskApi().Delete(id)
}

func (s *AgentStore) GetByID(id uint) (*models.CronTask, error) {
	return agent.NewCronTaskApi().GetByID(id)
}

func (s *AgentStore) List(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return agent.NewCronTaskApi().List(query)
}

func (s *AgentStore) EnableTask(id uint, enable bool) error {
	return agent.NewCronTaskApi().EnableTask(id, enable)
}

func (s *AgentStore) ExecuteTask(ctx context.Context, task *models.CronTask) error {
	return agent.NewCronTaskApi().ExecuteTask(ctx, task)
}

func (s *AgentStore) GetTaskTypes() []lo.Tuple2[string, string] {
	return agent.NewCronTaskApi().GetTaskTypes()
}

func (s *AgentStore) ValidateCronExpr(expr string) error {
	return agent.NewCronTaskApi().ValidateCronExpr(expr)
}

func (s *AgentStore) CalculateNextRunTime(expr string) time.Time {
	return agent.NewCronTaskApi().CalculateNextRunTime(expr)
}

func (s *AgentStore) CalculateNextRunTimes(expr string, count int) []time.Time {
	return agent.NewCronTaskApi().CalculateNextRunTimes(expr, count)
}

func (s *AgentStore) Search(keyword string) []models.CronTask {
	return agent.NewCronTaskApi().SearchTasks(keyword)
}
