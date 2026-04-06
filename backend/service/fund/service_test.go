package fund

import (
	"context"
	"testing"

	"go-stock/backend/data"
)

type fakeSource struct {
	fundList       []data.FundBasic
	followedFunds  []data.FollowedFund
	followResult   string
	unfollowResult string

	lastQuery        string
	lastFollowCode   string
	lastUnfollowCode string
}

func (f *fakeSource) GetFundList(key string) []data.FundBasic {
	f.lastQuery = key
	return append([]data.FundBasic(nil), f.fundList...)
}

func (f *fakeSource) GetFollowedFund() []data.FollowedFund {
	if f.followedFunds == nil {
		return nil
	}
	return append([]data.FollowedFund(nil), f.followedFunds...)
}

func (f *fakeSource) FollowFund(fundCode string) string {
	f.lastFollowCode = fundCode
	return f.followResult
}

func (f *fakeSource) UnFollowFund(fundCode string) string {
	f.lastUnfollowCode = fundCode
	return f.unfollowResult
}

func (f *fakeSource) CrawlFundBasic(fundCode string) (*data.FundBasic, error) {
	return nil, nil
}

func (f *fakeSource) CrawlFundNetEstimatedUnit(code string) {}

func (f *fakeSource) CrawlFundNetUnitValue(code string) {}

func (f *fakeSource) AllFund() {}

func TestFundService_LoadFollowedFundsReturnsEmptySliceOnNil(t *testing.T) {
	svc := NewService(&fakeSource{})

	result := svc.LoadFollowedFunds(context.Background())
	if result == nil {
		t.Fatalf("expected non-nil empty slice, got nil")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got len=%d", len(result))
	}
}

func TestFundService_LoadFundListAndFollowDelegates(t *testing.T) {
	source := &fakeSource{
		fundList: []data.FundBasic{
			{Code: "000001", Name: "示例基金"},
		},
		followedFunds: []data.FollowedFund{
			{Code: "000001", Name: "示例基金"},
		},
		followResult:   "关注成功",
		unfollowResult: "取消关注成功",
	}
	svc := NewService(source)

	funds := svc.LoadFundList(context.Background(), "示例")
	if len(funds) != 1 || funds[0].Code != "000001" {
		t.Fatalf("expected delegated fund list, got=%#v", funds)
	}
	if source.lastQuery != "示例" {
		t.Fatalf("expected query delegation, got=%q", source.lastQuery)
	}

	followed := svc.LoadFollowedFunds(context.Background())
	if len(followed) != 1 || followed[0].Code != "000001" {
		t.Fatalf("expected delegated followed funds, got=%#v", followed)
	}

	if got := svc.FollowFund(context.Background(), "000001"); got != "关注成功" {
		t.Fatalf("expected follow result delegation, got=%q", got)
	}
	if source.lastFollowCode != "000001" {
		t.Fatalf("expected follow code delegation, got=%q", source.lastFollowCode)
	}

	if got := svc.UnfollowFund(context.Background(), "000001"); got != "取消关注成功" {
		t.Fatalf("expected unfollow result delegation, got=%q", got)
	}
	if source.lastUnfollowCode != "000001" {
		t.Fatalf("expected unfollow code delegation, got=%q", source.lastUnfollowCode)
	}
}
