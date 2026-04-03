package data

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestEastMoneyKLineApi_GetKLineDataBeforeResult_RetriesWithCookieOnEOF(t *testing.T) {
	api := NewEastMoneyKLineApi(&SettingConfig{
		Settings: &Settings{
			CrawlTimeOut: 15,
			BrowserPath:  `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		},
	})

	var cookieCalls []string
	api.fetchHTTP = func(reqURL string, cookieHeader string) ([]byte, error) {
		cookieCalls = append(cookieCalls, cookieHeader)
		if len(cookieCalls) == 1 {
			return nil, io.EOF
		}
		return []byte(`{"rc":0,"code":0,"data":{"klines":["2026-04-03,10,11,12,9,100,1000,2.0,1.1,1.0,0.5"]}}`), nil
	}
	api.cookieHeaderProvider = func(config *SettingConfig) string {
		return "st_si=abc123"
	}

	result := api.GetKLineDataBeforeResult("600519.SH", "101", "", 5, "20500101")

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 kline after cookie retry, got %d", len(result.Data))
	}
	if !result.UsedCookieRetry {
		t.Fatalf("expected UsedCookieRetry=true")
	}
	if result.ErrorCode != "" {
		t.Fatalf("expected empty ErrorCode, got %q", result.ErrorCode)
	}
	if len(cookieCalls) != 2 {
		t.Fatalf("expected 2 fetch attempts, got %d", len(cookieCalls))
	}
	if cookieCalls[0] != "" {
		t.Fatalf("expected first request without cookie, got %q", cookieCalls[0])
	}
	if cookieCalls[1] != "st_si=abc123" {
		t.Fatalf("expected second request with cookie header, got %q", cookieCalls[1])
	}
}

func TestEastMoneyKLineApi_GetKLineDataBeforeResult_ReportsRetryFailureMessage(t *testing.T) {
	api := NewEastMoneyKLineApi(&SettingConfig{
		Settings: &Settings{
			CrawlTimeOut: 15,
			BrowserPath:  `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		},
	})

	api.fetchHTTP = func(reqURL string, cookieHeader string) ([]byte, error) {
		if cookieHeader == "" {
			return nil, io.EOF
		}
		return nil, errors.New("tls: handshake failure")
	}
	api.cookieHeaderProvider = func(config *SettingConfig) string {
		return "st_si=abc123"
	}

	result := api.GetKLineDataBeforeResult("600519.SH", "101", "", 5, "20500101")

	if len(result.Data) != 0 {
		t.Fatalf("expected empty data on double failure, got %d", len(result.Data))
	}
	if result.ErrorCode != "eastmoney_request_failed" {
		t.Fatalf("expected eastmoney_request_failed, got %q", result.ErrorCode)
	}
	if !result.UsedCookieRetry {
		t.Fatalf("expected UsedCookieRetry=true")
	}
	if !strings.Contains(result.Message, "cookie 重试失败") {
		t.Fatalf("expected retry failure message, got %q", result.Message)
	}
}
