package notification

import (
	"testing"

	"go-stock/backend/data"
)

type fakeCache struct {
	ttl    map[string]uint32
	setTTL map[string]int
}

func (f *fakeCache) TTL(key []byte) (uint32, error) {
	if f.ttl == nil {
		return 0, nil
	}
	return f.ttl[string(key)], nil
}

func (f *fakeCache) Set(key []byte, value []byte, expireSeconds int) error {
	if f.setTTL == nil {
		f.setTTL = map[string]int{}
	}
	f.setTTL[string(key)] = expireSeconds
	return nil
}

type fakeAdapter struct {
	dingResult   string
	localTitle   string
	localContent string
	stockInfo    *data.StockInfo
}

func (f *fakeAdapter) SendDingTalk(message string) string {
	return f.dingResult
}

func (f *fakeAdapter) SendLocal(title, content string) bool {
	f.localTitle = title
	f.localContent = content
	return true
}

func (f *fakeAdapter) LoadStockInfo(stockCode string) *data.StockInfo {
	return f.stockInfo
}

func TestService_SendDingTalkSkipsWhileTTLActive(t *testing.T) {
	cache := &fakeCache{ttl: map[string]uint32{"sh600519": 30}}
	adapter := &fakeAdapter{dingResult: "发送成功"}
	svc := NewService(cache, adapter)

	if got := svc.SendDingTalk("body", "sh600519"); got != "" {
		t.Fatalf("expected empty result when ttl is active, got %q", got)
	}
	if _, exists := cache.setTTL["sh600519"]; exists {
		t.Fatalf("did not expect cache reset when ttl is already active")
	}
}

func TestService_SendTypedDeliversLocalAndDingTalk(t *testing.T) {
	cache := &fakeCache{}
	adapter := &fakeAdapter{
		dingResult: "发送钉钉消息成功",
		stockInfo: &data.StockInfo{
			Name:     "平安银行",
			Price:    "12.34",
			PreClose: "12.00",
			Date:     "2026-04-06",
			Time:     "10:00:00",
		},
	}
	svc := NewService(cache, adapter)

	result := svc.SendTyped("body", "sz000001", 1)
	if result.DingResult != "发送钉钉消息成功" {
		t.Fatalf("unexpected ding result: %#v", result)
	}
	if adapter.localTitle != "涨跌报警" {
		t.Fatalf("unexpected local title: %q", adapter.localTitle)
	}
	if adapter.localContent == "" || result.EventContent == "" {
		t.Fatalf("expected non-empty local/event content: adapter=%q result=%#v", adapter.localContent, result)
	}
	if cache.setTTL["sz000001"] != 300 {
		t.Fatalf("unexpected ttl seconds: %d", cache.setTTL["sz000001"])
	}
}
