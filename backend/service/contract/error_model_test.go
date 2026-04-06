package contract

import "testing"

func TestUserVisibleError_ErrorUsesMessage(t *testing.T) {
	err := UserVisibleError{
		Code:      "analysis.result_missing",
		Message:   "分析结果不存在",
		Retryable: false,
		Stage:     StageService,
	}

	if err.Error() != "分析结果不存在" {
		t.Fatalf("expected Error() to return message, got %q", err.Error())
	}
}

func TestNewUserVisibleErrorBuildsStablePayload(t *testing.T) {
	err := NewUserVisibleError("market.fetch.failed", "市场数据获取失败", true, StageSource)
	if err.Code != "market.fetch.failed" || err.Message != "市场数据获取失败" || !err.Retryable || err.Stage != StageSource {
		t.Fatalf("unexpected user visible error: %#v", err)
	}
}
