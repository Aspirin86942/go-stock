package data

import (
	"fmt"
	"strings"
	"testing"
)

func newStubEastMoneyKLineAPI(t *testing.T, klines []string) *EastMoneyKLineApi {
	t.Helper()

	api := NewEastMoneyKLineApi(&SettingConfig{Settings: &Settings{CrawlTimeOut: 5}})
	response := fmt.Sprintf(`{"rc":0,"data":{"code":"600519","name":"贵州茅台","klines":["%s"]}}`, strings.Join(klines, `","`))
	api.fetchHTTP = func(reqURL string, cookieHeader string) ([]byte, error) {
		return []byte(response), nil
	}
	api.cookieHeaderProvider = nil
	return api
}

func TestEastMoneyKLineSection_IncludesAmount(t *testing.T) {
	api := newStubEastMoneyKLineAPI(t, []string{
		"2026-04-03,10.01,10.10,10.20,9.90,100000,123456789,2.50,1.20,0.12,3.45",
	})

	got := EastMoneyKLineSection(api, "600519.SH", "day", "", 1)

	if !strings.Contains(got, "成交额") {
		t.Fatalf("expected markdown to include 成交额 column, got:\n%s", got)
	}
	if !strings.Contains(got, "123456789") {
		t.Fatalf("expected markdown to include 成交额 value, got:\n%s", got)
	}
	if !strings.Contains(got, "换手率") {
		t.Fatalf("expected markdown to keep 换手率 column, got:\n%s", got)
	}
}

func TestEastMoneyKLineWithMASection_IncludesAmount(t *testing.T) {
	api := newStubEastMoneyKLineAPI(t, []string{
		"2026-03-28,10.00,10.00,10.10,9.90,100000,100000000,2.00,0.00,0.00,3.10",
		"2026-03-29,10.10,10.20,10.30,10.00,110000,110000000,2.10,2.00,0.20,3.20",
		"2026-03-30,10.20,10.30,10.40,10.10,120000,120000000,2.20,0.98,0.10,3.30",
		"2026-03-31,10.30,10.40,10.50,10.20,130000,130000000,2.30,0.97,0.10,3.40",
		"2026-04-01,10.40,10.50,10.60,10.30,140000,140000000,2.40,0.96,0.10,3.50",
		"2026-04-02,10.50,10.60,10.70,10.40,150000,150000000,2.50,0.95,0.10,3.60",
	})

	got := EastMoneyKLineWithMASection(api, "600519.SH", "day", 1, "5")

	if !strings.Contains(got, "成交额") {
		t.Fatalf("expected markdown to include 成交额 column, got:\n%s", got)
	}
	if !strings.Contains(got, "150000000") {
		t.Fatalf("expected markdown to include latest 成交额 value, got:\n%s", got)
	}
	if !strings.Contains(got, "MA5") {
		t.Fatalf("expected markdown to keep MA column, got:\n%s", got)
	}
}
