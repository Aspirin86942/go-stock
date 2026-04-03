# Fix EastMoney Multi-KLine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the stock-detail "多周期K线" feature by fixing the EastMoney fetch path, adding one controlled cookie retry on retryable connection failures, and surfacing explicit failure text in the modal instead of the current misleading empty-data message.

**Architecture:** Keep the existing `GetStockEastMoneyKLine` / `GetStockEastMoneyKLinePage` APIs for compatibility, but add a detailed result path in the backend that records whether the request succeeded, failed, and whether cookie retry was used. The Vue chart modal switches to the detailed result path so it can distinguish "supported code but data-source request failed" from "truly empty data". The network hardening stays inside `backend/data/eastmoney_kline_api.go`, and the UI change stays inside `frontend/src/components/StockLightweightKlineChart.vue`.

**Tech Stack:** Go, Wails, Vue 3, Naive UI, Resty, chromedp, PowerShell, Go `testing`

---

### Task 1: Add deterministic backend tests for retryable EastMoney fetch behavior

**Files:**
- Create: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api_retry_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api.go`

- [ ] **Step 1: Write the failing backend tests**

```go
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
```

- [ ] **Step 2: Run the targeted backend tests to verify they fail**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/data -run "TestEastMoneyKLineApi_GetKLineDataBeforeResult_RetriesWithCookieOnEOF|TestEastMoneyKLineApi_GetKLineDataBeforeResult_ReportsRetryFailureMessage" -count=1
```

Expected:

- compile failure because `GetKLineDataBeforeResult`, `fetchHTTP`, and `cookieHeaderProvider` do not exist yet

- [ ] **Step 3: Implement the minimal backend retry and detailed-result path**

```go
type eastMoneyFetchResult struct {
	Data            []KLineData
	ErrorCode       string
	Message         string
	UsedCookieRetry bool
}

type eastMoneyFetchHTTPFunc func(reqURL string, cookieHeader string) ([]byte, error)
type eastMoneyCookieProvider func(config *SettingConfig) string

type EastMoneyKLineApi struct {
	client               *resty.Client
	config               *SettingConfig
	fetchHTTP            eastMoneyFetchHTTPFunc
	cookieHeaderProvider eastMoneyCookieProvider
}

func NewEastMoneyKLineApi(config *SettingConfig) *EastMoneyKLineApi {
	client := resty.New()
	if config == nil {
		config = &SettingConfig{Settings: &Settings{CrawlTimeOut: 30}}
	}
	client.SetTransport(newEastMoneyKLineTransport())
	if config.HttpProxyEnabled && strings.TrimSpace(config.HttpProxy) != "" {
		client.SetProxy(config.HttpProxy)
	}
	api := &EastMoneyKLineApi{
		client: client,
		config: config,
	}
	api.fetchHTTP = api.fetchKLineJSONBytesByHTTP
	api.cookieHeaderProvider = EastMoneyCookieHeaderForPush2his
	return api
}

func newEastMoneyKLineTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func shouldRetryEastMoneyWithCookie(err error, config *SettingConfig) bool {
	if err == nil || config == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	retryable := strings.Contains(msg, "eof") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "tls") ||
		strings.Contains(msg, "handshake")
	if !retryable {
		return false
	}
	if strings.TrimSpace(config.BrowserPath) != "" {
		return true
	}
	_, ok := CheckBrowser()
	return ok
}

func (receiver *EastMoneyKLineApi) fetchKLineJSONBytesByHTTP(reqURL string, cookieHeader string) ([]byte, error) {
	req := receiver.client.SetTimeout(time.Duration(receiver.config.CrawlTimeOut) * time.Second).R()
	setEastMoneyKlineBrowserHeaders(req, "https://quote.eastmoney.com")
	if strings.TrimSpace(cookieHeader) != "" {
		req.SetHeader("Cookie", cookieHeader)
	}
	resp, err := req.Get(reqURL)
	if err != nil {
		logger.SugaredLogger.Errorf("HTTP error: %v", err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		b := resp.Body()
		if len(b) > 500 {
			b = b[:500]
		}
		return nil, fmt.Errorf("HTTP %d: %q", resp.StatusCode(), string(b))
	}
	rawBody := resp.Body()
	if strings.EqualFold(resp.Header().Get("Content-Encoding"), "gzip") {
		reader, err := gzip.NewReader(bytes.NewReader(rawBody))
		if err != nil {
			return nil, fmt.Errorf("gzip.NewReader error: %w", err)
		}
		defer reader.Close()
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip decompress error: %w", err)
		}
		return decompressed, nil
	}
	return rawBody, nil
}

func (receiver *EastMoneyKLineApi) fetchKLineJSONBytes(reqURL string) ([]byte, bool, error) {
	body, err := receiver.fetchHTTP(reqURL, "")
	if err == nil {
		return body, false, nil
	}
	if !shouldRetryEastMoneyWithCookie(err, receiver.config) {
		return nil, false, err
	}
	cookieHeader := receiver.cookieHeaderProvider(receiver.config)
	if strings.TrimSpace(cookieHeader) == "" {
		return nil, false, err
	}
	body, retryErr := receiver.fetchHTTP(reqURL, cookieHeader)
	if retryErr != nil {
		return nil, true, fmt.Errorf("首次请求失败: %w; cookie 重试失败: %v", err, retryErr)
	}
	return body, true, nil
}

func (receiver *EastMoneyKLineApi) GetKLineDataBeforeResult(stockCode, kLineType, adjustFlag string, limit int, end string) *eastMoneyFetchResult {
	result := &eastMoneyFetchResult{Data: make([]KLineData, 0)}
	secid := receiver.convertStockCode(stockCode)
	if secid == "" {
		result.ErrorCode = "eastmoney_invalid_code"
		result.Message = fmt.Sprintf("东财 K 线不支持当前代码：%s", stockCode)
		return result
	}
	if limit <= 0 {
		result.ErrorCode = "eastmoney_invalid_limit"
		result.Message = "东财 K 线请求参数无效：limit 必须大于 0"
		return result
	}
	if strings.TrimSpace(end) == "" {
		end = "20500101"
	}
	params := url.Values{}
	params.Set("secid", secid)
	params.Set("klt", kLineType)
	params.Set("fqt", adjustFlag)
	params.Set("end", end)
	params.Set("lmt", convertor.ToString(limit))
	params.Set("fields1", "f1,f2,f3,f4,f5,f6")
	params.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f116")
	params.Set("wbp2u", "|0|0|0|web")
	params.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))
	reqURL := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?%s", params.Encode())

	body, usedCookieRetry, fetchErr := receiver.fetchKLineJSONBytes(reqURL)
	result.UsedCookieRetry = usedCookieRetry
	if fetchErr != nil {
		result.ErrorCode = "eastmoney_request_failed"
		result.Message = fmt.Sprintf("东财 K 线请求失败：%v", fetchErr)
		return result
	}
	var response EastMoneyKLineResponse
	if err := json.Unmarshal(body, &response); err != nil {
		result.ErrorCode = "eastmoney_decode_failed"
		result.Message = fmt.Sprintf("东财 K 线响应解析失败：%v", err)
		return result
	}
	if response.Rc != 0 || response.Code != 0 {
		result.ErrorCode = "eastmoney_api_error"
		result.Message = fmt.Sprintf("东财 K 线接口返回异常：rc=%d code=%d message=%s", response.Rc, response.Code, response.Message)
		return result
	}
	for _, klineStr := range response.Data.Klines {
		kline := receiver.parseKLine(klineStr, adjustFlag)
		if kline != nil {
			result.Data = append(result.Data, *kline)
		}
	}
	if len(result.Data) == 0 {
		result.ErrorCode = "eastmoney_empty_data"
		result.Message = "东财 K 线暂无可用数据"
	}
	return result
}

func (receiver *EastMoneyKLineApi) GetKLineDataBefore(stockCode, kLineType, adjustFlag string, limit int, end string) *[]KLineData {
	result := receiver.GetKLineDataBeforeResult(stockCode, kLineType, adjustFlag, limit, end)
	data := append([]KLineData(nil), result.Data...)
	return &data
}
```

- [ ] **Step 4: Run the targeted backend tests to verify they pass**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/data -run "TestEastMoneyKLineApi_GetKLineDataBeforeResult_RetriesWithCookieOnEOF|TestEastMoneyKLineApi_GetKLineDataBeforeResult_ReportsRetryFailureMessage" -count=1
```

Expected:

- command exits `0`
- both tests pass

- [ ] **Step 5: Review the backend retry slice in `git diff` (commit only if the user explicitly requests it)**

```powershell
git diff -- backend/data/eastmoney_kline_api.go backend/data/eastmoney_kline_api_retry_test.go
```

### Task 2: Expose detailed K-line status to the modal and remove the misleading empty-data text

**Files:**
- Modify: `D:\codex_work\go-stock\app.go`
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.d.ts`
- Generated: `D:\codex_work\go-stock\frontend\wailsjs\go\main\App.js`

- [ ] **Step 1: Write the failing frontend text and API usage check**

Run:

```powershell
rg -n "暂无 K 线数据（需东方财富支持的代码，如 600519.SH、000001.SZ）|GetStockEastMoneyKLine\\(|GetStockEastMoneyKLinePage\\(" frontend/src/components/StockLightweightKlineChart.vue
```

Expected:

- old misleading empty-data text is present
- the component still imports and calls the raw array-returning methods

- [ ] **Step 2: Add detailed Wails wrappers in `app.go`**

```go
func (a *App) GetStockEastMoneyKLineResult(stockCode, stockName string, klt string, limit int) map[string]any {
	return a.GetStockEastMoneyKLinePageResult(stockCode, stockName, klt, limit, "")
}

func (a *App) GetStockEastMoneyKLinePageResult(stockCode, stockName string, klt string, limit int, end string) map[string]any {
	api := data.NewEastMoneyKLineApi(data.GetSettingConfig())
	result := api.GetKLineDataBeforeResult(stockCode, klt, "", limit, end)
	return map[string]any{
		"ok":              len(result.Data) > 0 && result.ErrorCode == "",
		"data":            result.Data,
		"message":         result.Message,
		"errorCode":       result.ErrorCode,
		"usedCookieRetry": result.UsedCookieRetry,
	}
}
```

- [ ] **Step 3: Switch the multi-kline component to the detailed-result methods**

```vue
<script setup>
import { GetStockEastMoneyKLinePageResult, GetStockEastMoneyKLineResult } from '../../wailsjs/go/main/App'

function normalizeKlineEnvelope(raw) {
  return {
    ok: !!raw?.ok,
    list: Array.isArray(raw?.data) ? raw.data : [],
    message: String(raw?.message || ''),
    errorCode: String(raw?.errorCode || ''),
  }
}

async function loadOlderHistory() {
  if (
    loadingHistory.value ||
    !hasMoreOlder.value ||
    !mergedRawRows.length ||
    !props.code ||
    !chart ||
    !candleSeries
  ) {
    return
  }
  const kltSnap = activeKlt.value
  const codeSnap = props.code
  const oldest = mergedRawRows[0]
  const end = formatEastMoneyEndFromOldest(oldest.day, kltSnap)
  if (!end) {
    hasMoreOlder.value = false
    return
  }
  loadingHistory.value = true
  const logical = chart.timeScale().getVisibleLogicalRange()
  const beforeCount = mergedRawRows.length
  try {
    const envelope = normalizeKlineEnvelope(
      await GetStockEastMoneyKLinePageResult(codeSnap, props.stockName || '', kltSnap, HISTORY_PAGE_SIZE, end),
    )
    if (kltSnap !== activeKlt.value || codeSnap !== props.code) return
    const inc = envelope.list
    if (!inc.length) {
      if (!envelope.ok && envelope.message) {
        errorText.value = envelope.message
      }
      hasMoreOlder.value = false
      lastOlderHistoryEndTried = ''
      return
    }
    const merged = mergeKlineRows(mergedRawRows, inc)
    const added = merged.length - beforeCount
    if (added <= 0) {
      if (end === lastOlderHistoryEndTried) {
        hasMoreOlder.value = false
      } else {
        lastOlderHistoryEndTried = end
      }
      return
    }
    lastOlderHistoryEndTried = ''
    mergedRawRows = merged
    syncDefaultLatestPanelRow()
    withProgrammaticTimeRange(() => {
      applySeriesFromRaw()
      if (logical) {
        chart.timeScale().setVisibleLogicalRange({
          from: logical.from + added,
          to: logical.to + added,
        })
      }
    })
  } finally {
    loadingHistory.value = false
  }
}

async function loadData() {
  if (!props.code) {
    errorText.value = '未设置股票代码'
    mergedRawRows = []
    syncDefaultLatestPanelRow()
    hasMoreOlder.value = true
    lastOlderHistoryEndTried = ''
    candleSeries?.setData([])
    volSeries?.setData([])
    syncLongPositionPriceLines()
    return
  }
  loading.value = true
  errorText.value = ''
  mergedRawRows = []
  syncDefaultLatestPanelRow()
  hasMoreOlder.value = true
  lastOlderHistoryEndTried = ''
  try {
    const meta = INTERVALS.find((x) => x.klt === activeKlt.value) || INTERVALS[0]
    const envelope = normalizeKlineEnvelope(
      await GetStockEastMoneyKLineResult(props.code, props.stockName || '', meta.klt, meta.limit),
    )
    const list = envelope.list
    ensureChart()
    mergedRawRows = mergeKlineRows([], list)
    syncDefaultLatestPanelRow()
    const { candles } = toSeriesData(mergedRawRows)
    if (!candles.length) {
      errorText.value = envelope.message || '东财 K 线暂无可用数据，请稍后重试'
      candleSeries?.setData([])
      volSeries?.setData([])
      syncIndicators()
      syncLongPositionPriceLines()
      return
    }
    withProgrammaticTimeRange(() => {
      applySeriesFromRaw()
      applyDefaultVisibleRange()
    })
  } catch (e) {
    errorText.value = String(e?.message || e)
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 4: Regenerate bindings and run the frontend build**

Run:

```powershell
$env:PATH='C:\Program Files\Go\bin;' + $env:PATH
& 'C:\Users\Aspir\go\bin\wails.exe' build --platform windows/amd64
npm --prefix frontend run build
```

Expected:

- Wails build exits `0`
- `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/main/App.js` include the new `GetStockEastMoneyKLineResult` / `GetStockEastMoneyKLinePageResult` exports
- frontend build exits `0`

- [ ] **Step 5: Re-run the frontend text and API usage check**

Run:

```powershell
rg -n "暂无 K 线数据（需东方财富支持的代码，如 600519.SH、000001.SZ）|GetStockEastMoneyKLine\\(|GetStockEastMoneyKLinePage\\(" frontend/src/components/StockLightweightKlineChart.vue
```

Expected:

- no match for the old misleading text
- no direct raw-array calls remain in the multi-kline component

- [ ] **Step 6: Review the frontend and Wails wrapper slice in `git diff` (commit only if the user explicitly requests it)**

```powershell
git diff -- app.go frontend/src/components/StockLightweightKlineChart.vue frontend/wailsjs/go/main/App.d.ts frontend/wailsjs/go/main/App.js
```

### Task 3: Verify the repaired EastMoney path against logs and the packaged desktop build

**Files:**
- Verify only: `D:\codex_work\go-stock\backend\data\eastmoney_kline_api.go`
- Verify only: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
- Verify only: `D:\codex_work\go-stock\build\bin\logs\error.log`
- Verify only: `D:\codex_work\go-stock\build\bin\go-stock.exe`

- [ ] **Step 1: Run the targeted backend tests again as final evidence**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/data -run "TestEastMoneyKLineApi_GetKLineDataBeforeResult_RetriesWithCookieOnEOF|TestEastMoneyKLineApi_GetKLineDataBeforeResult_ReportsRetryFailureMessage" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 2: Run a desktop compile verification**

Run:

```powershell
& 'C:\Program Files\Go\bin\go.exe' build .
```

Expected:

- command exits `0`

- [ ] **Step 3: Launch the packaged app and manually verify the multi-kline modal**

Run:

```powershell
Start-Process 'D:\codex_work\go-stock\build\bin\go-stock.exe'
```

Manual verification checklist:

- open a stock detail card that previously reproduced the problem, such as `688188`
- click `多周期K线`
- confirm one of these is true:
  - candles render for at least `日K` and one minute-like interval, or
  - the modal shows an explicit EastMoney failure message instead of the old “需东方财富支持的代码” text

- [ ] **Step 4: Inspect the new error log lines for retry diagnostics**

Run:

```powershell
rg -n "东财 K 线请求失败|cookie 重试失败|UsedCookieRetry|EOF" build/bin/logs/error.log
```

Expected:

- if the data source still fails, the log now shows the new retry-aware error message rather than only a bare `EOF`
- if the retry succeeded, repeated `EOF` lines for the same interactive request should stop appearing

- [ ] **Step 5: Review any verification-driven adjustments and keep unrelated local changes untouched**

```powershell
git diff -- backend/data/eastmoney_kline_api.go frontend/src/components/StockLightweightKlineChart.vue app.go frontend/wailsjs/go/main/App.d.ts frontend/wailsjs/go/main/App.js
```
