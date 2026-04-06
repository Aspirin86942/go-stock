package main

import (
	"context"
	"testing"

	"go-stock/backend/data"
)

type stubFundService struct {
	funds            []data.FundBasic
	followedFunds    []data.FollowedFund
	lastFollowCode   string
	lastUnfollowCode string
	refreshCalls     int
}

func (s *stubFundService) LoadFundList(ctx context.Context, key string) []data.FundBasic {
	_ = ctx
	return append([]data.FundBasic(nil), s.funds...)
}

func (s *stubFundService) LoadFollowedFunds(ctx context.Context) []data.FollowedFund {
	_ = ctx
	return append([]data.FollowedFund(nil), s.followedFunds...)
}

func (s *stubFundService) FollowFund(ctx context.Context, fundCode string) string {
	_ = ctx
	s.lastFollowCode = fundCode
	return "关注成功"
}

func (s *stubFundService) UnfollowFund(ctx context.Context, fundCode string) string {
	_ = ctx
	s.lastUnfollowCode = fundCode
	return "取消关注成功"
}

func (s *stubFundService) RefreshFollowedFunds(ctx context.Context) error {
	_ = ctx
	s.refreshCalls++
	return nil
}

func (s *stubFundService) SyncAllFunds(ctx context.Context) {
	_ = ctx
}

func TestApp_FundMethodsDelegateToService(t *testing.T) {
	app := &App{
		ctx: context.Background(),
		fundService: &stubFundService{
			funds:         []data.FundBasic{{Code: "510300"}},
			followedFunds: []data.FollowedFund{{Code: "510300"}},
		},
	}

	if got := app.GetFollowedFund(); len(got) != 1 {
		t.Fatalf("expected delegated followed fund list")
	}
	if got := app.GetfundList("510300"); len(got) != 1 {
		t.Fatalf("expected delegated fund list")
	}
	if got := app.FollowFund("510300"); got != "关注成功" {
		t.Fatalf("expected delegated follow result, got %q", got)
	}
	if got := app.UnFollowFund("510300"); got != "取消关注成功" {
		t.Fatalf("expected delegated unfollow result, got %q", got)
	}
}
