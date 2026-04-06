package fund

import (
	"context"

	"go-stock/backend/data"
)

type Source interface {
	GetFundList(key string) []data.FundBasic
	GetFollowedFund() []data.FollowedFund
	FollowFund(fundCode string) string
	UnFollowFund(fundCode string) string
	CrawlFundBasic(fundCode string) (*data.FundBasic, error)
	CrawlFundNetEstimatedUnit(code string)
	CrawlFundNetUnitValue(code string)
	AllFund()
}

type RefreshError struct {
	FundCode string
	Err      error
}

func (e RefreshError) Error() string {
	if e.Err == nil {
		return "fund refresh failed"
	}
	return e.Err.Error()
}

func (e RefreshError) Unwrap() error {
	return e.Err
}

type Service struct {
	source Source
}

func NewService(source Source) *Service {
	if source == nil {
		panic("fund: source dependency is required")
	}
	return &Service{source: source}
}

func (s *Service) LoadFundList(ctx context.Context, key string) []data.FundBasic {
	_ = ctx
	return s.source.GetFundList(key)
}

func (s *Service) LoadFollowedFunds(ctx context.Context) []data.FollowedFund {
	_ = ctx
	result := s.source.GetFollowedFund()
	if result == nil {
		return []data.FollowedFund{}
	}
	return append([]data.FollowedFund(nil), result...)
}

func (s *Service) FollowFund(ctx context.Context, fundCode string) string {
	_ = ctx
	return s.source.FollowFund(fundCode)
}

func (s *Service) UnfollowFund(ctx context.Context, fundCode string) string {
	_ = ctx
	return s.source.UnFollowFund(fundCode)
}

func (s *Service) RefreshFollowedFunds(ctx context.Context) error {
	follows := s.LoadFollowedFunds(ctx)
	var firstErr error
	for _, follow := range follows {
		_, err := s.source.CrawlFundBasic(follow.Code)
		if err != nil {
			if firstErr == nil {
				firstErr = RefreshError{
					FundCode: follow.Code,
					Err:      err,
				}
			}
			continue
		}
		s.source.CrawlFundNetEstimatedUnit(follow.Code)
		s.source.CrawlFundNetUnitValue(follow.Code)
	}
	return firstErr
}

func (s *Service) SyncAllFunds(ctx context.Context) {
	_ = ctx
	s.source.AllFund()
}
