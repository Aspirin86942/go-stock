package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/go-resty/resty/v2"
	uaFake "github.com/lib4u/fake-useragent"
)

var eastMoneyKlineLog = dataModuleLogger(logger.SinkHTTP, "data.eastmoney_kline")

// 模拟 Windows 上 Chrome 从 quote.eastmoney.com 请求 push2his 行情接口（与 DevTools Network 常见字段对齐）。
// 不显式设置 Accept-Encoding：由 net/http 默认协商 gzip 并自动解压；若声明 br/zstd 而 Transport 不解压会导致乱码/失败。
// getRandomUA 随机返回一个 User-Agent（使用 fake-useragent 库）
func getRandomUA() string {
	ua, _ := uaFake.New()
	if ua != nil {
		randomUA := ua.Filter().Platform("desktop").Get()
		//logger.SugaredLogger.Infof("User-Agent: %s", randomUA)
		return randomUA
	}
	// 如果库获取失败，返回备用 UA
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
}

// Enhanced headers with more realistic browser characteristics
func setEastMoneyKlineBrowserHeaders(r *resty.Request, referer string) {
	r.SetHeader("User-Agent", getRandomUA())
	r.SetHeader("Accept", "*/*")
	r.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	r.SetHeader("Connection", "keep-alive")
	r.SetHeader("Referer", referer)
}

// fetchKLineJSONBytesByHTTP 每次调用均发起真实 GET，不缓存 K 线响应。
// 这里允许注入 cookieHeader，是因为东财在部分环境下会对无 cookie 的直连请求直接断流。
func (receiver *EastMoneyKLineApi) fetchKLineJSONBytesByHTTP(reqURL string, cookieHeader string) ([]byte, error) {
	req := receiver.client.SetTimeout(time.Duration(receiver.config.CrawlTimeOut) * time.Second).R()
	setEastMoneyKlineBrowserHeaders(req, "https://quote.eastmoney.com")
	if strings.TrimSpace(cookieHeader) != "" {
		req.SetHeader("Cookie", cookieHeader)
	}

	resp, err := req.Get(reqURL)
	if err != nil {
		eastMoneyKlineLog.Errorf("data.eastmoney_kline.http_failed", "HTTP error: %v", err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		b := resp.Body()
		if len(b) > 500 {
			b = b[:500]
		}
		return nil, fmt.Errorf("HTTP %d: %q", resp.StatusCode(), string(b))
	}

	// 读取响应体
	rawBody := resp.Body()

	// 检查 Content-Encoding 并处理 gzip 压缩
	contentEncoding := resp.Header().Get("Content-Encoding")
	if strings.ToLower(contentEncoding) == "gzip" {
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

// Helper functions for error classification
func isNetworkError(err error) bool {
	return strings.Contains(err.Error(), "network") ||
		strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "timeout")
}

func isTimeoutError(err error) bool {
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline")
}

func isBotDetected(resp *resty.Response) bool {
	// Check for common bot detection indicators
	body := string(resp.Body())

	botIndicators := []string{
		"bot", "robot", "crawler", "spider",
		"验证", "验证码", "安全验证",
		"access denied", "forbidden",
		"请开启JavaScript", "enable javascript",
	}

	for _, indicator := range botIndicators {
		if strings.Contains(strings.ToLower(body), strings.ToLower(indicator)) {
			return true
		}
	}

	// Check for CAPTCHA pages
	if strings.Contains(body, "captcha") || strings.Contains(body, "recaptcha") {
		return true
	}

	return false
}

func isValidResponse(body []byte) bool {
	if len(body) == 0 {
		return false
	}

	// Check if response contains valid JSON structure
	var temp interface{}
	if err := json.Unmarshal(body, &temp); err != nil {
		// Not valid JSON, might be HTML error page
		bodyStr := string(body)
		return !(strings.Contains(bodyStr, "<html") || strings.Contains(bodyStr, "<!DOCTYPE"))
	}

	return true
}

// @Author spark
// @Date 2026/3/15
// @Desc 东方财富 K 线数据 API 工具

// EastMoneyKLineApi 东方财富 K 线 API 结构体
type EastMoneyKLineApi struct {
	client               *resty.Client
	config               *SettingConfig
	fetchHTTP            eastMoneyFetchHTTPFunc
	cookieHeaderProvider eastMoneyCookieProvider
}

type eastMoneyFetchResult struct {
	Data            []KLineData
	ErrorCode       string
	Message         string
	UsedCookieRetry bool
}

type eastMoneyFetchHTTPFunc func(reqURL string, cookieHeader string) ([]byte, error)
type eastMoneyCookieProvider func(config *SettingConfig) string

// KLineType K 线类型枚举
type KLineType string

const (
	KLineType1Min     KLineType = "1"   // 1 分钟
	KLineType5Min     KLineType = "5"   // 5 分钟
	KLineType15Min    KLineType = "15"  // 15 分钟
	KLineType30Min    KLineType = "30"  // 30 分钟
	KLineType60Min    KLineType = "60"  // 60 分钟
	KLineType120Min   KLineType = "120" // 120 分钟
	KLineTypeDay      KLineType = "101" // 日 K
	KLineTypeWeek     KLineType = "102" // 周 K
	KLineTypeMonth    KLineType = "103" // 月 K
	KLineTypeQuarter  KLineType = "104" // 季 K
	KLineTypeHalfYear KLineType = "105" // 半年 K
	KLineTypeYear     KLineType = "106" // 年 K
)

// EastMoneyKLineResponse 东方财富 K 线响应结构
type EastMoneyKLineResponse struct {
	Rc      int    `json:"rc"` // 接口实际返回 rc，成功为 0
	Version string `json:"version"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID       int      `json:"id"`
		Klines   []string `json:"klines"` // K 线数据数组，逗号分隔的字符串
		Name     string   `json:"name"`
		Code     string   `json:"code"`
		Market   any      `json:"market"`
		Period   string   `json:"period"`
		Pair     int      `json:"pair"`
		PrePrice string   `json:"prePrice"`
	} `json:"data"`
}

// CallAuctionData 竞价数据结构
type CallAuctionData struct {
	Time         string // 时间 (HH:MM:SS)
	Price        string // 撮合价格
	Volume       string // 撮合数量 (手)
	Amount       string // 撮合金额 (元)
	ChangeNum    string // 增减量
	ChangeRatio  string // 增减比例 (%)
	MatchedVol   string // 匹配量
	UnmatchedVol string // 未匹配量
	AskPrice1    string // 卖一价
	AskVol1      string // 卖一量
	BidPrice1    string // 买一价
	BidVol1      string // 买一量
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

// NewEastMoneyKLineApi 创建东方财富 K 线 API 实例
func NewEastMoneyKLineApi(config *SettingConfig) *EastMoneyKLineApi {
	client := resty.New()
	if config == nil {
		config = &SettingConfig{Settings: &Settings{}}
	}
	if config.Settings == nil {
		config.Settings = &Settings{}
	}
	if config.CrawlTimeOut <= 0 {
		config.CrawlTimeOut = 30
	}
	client.SetTransport(newEastMoneyKLineTransport())
	if config.HttpProxyEnabled && strings.TrimSpace(config.HttpProxy) != "" {
		client.SetProxy(config.HttpProxy)
	}

	//// 配置强制 IPv4 优先的 Transport，解决 IPv6 连接问题
	//dialer := &net.Dialer{
	//	Timeout:       10 * time.Second,
	//	KeepAlive:     30 * time.Second,
	//	DualStack:     false, // 禁用双栈
	//	FallbackDelay: -1,    // 禁用 Happy Eyeballs
	//}
	//
	//client.SetTransport(&http.Transport{
	//	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
	//		// 强制只使用 IPv4
	//		host, port, err := net.SplitHostPort(addr)
	//		if err != nil {
	//			return nil, err
	//		}
	//		// 解析 A 记录（IPv4）
	//		ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	//		if err != nil {
	//			return nil, err
	//		}
	//		if len(ips) == 0 {
	//			return nil, fmt.Errorf("no IPv4 address found for %s", host)
	//		}
	//		ipv4 := ips[0].String()
	//		return dialer.DialContext(ctx, "tcp4", net.JoinHostPort(ipv4, port))
	//	},
	//	TLSClientConfig: &tls.Config{
	//		MinVersion: tls.VersionTLS12,
	//		ServerName: "push2his.eastmoney.com",
	//	},
	//	DisableCompression:  true, // 禁用自动压缩，手动处理 gzip
	//	MaxIdleConns:        100,
	//	MaxIdleConnsPerHost: 10,
	//	IdleConnTimeout:     90 * time.Second,
	//	ForceAttemptHTTP2:   false, // 强制使用 HTTP/1.1
	//})
	//
	//client.SetTimeout(time.Duration(config.CrawlTimeOut) * time.Second)

	api := &EastMoneyKLineApi{
		client: client,
		config: config,
	}
	api.fetchHTTP = api.fetchKLineJSONBytesByHTTP
	api.cookieHeaderProvider = EastMoneyCookieHeaderForPush2his
	return api
}

func (receiver *EastMoneyKLineApi) fetchKLineJSONBytes(reqURL string) ([]byte, bool, error) {
	body, err := receiver.fetchHTTP(reqURL, "")
	if err == nil {
		return body, false, nil
	}
	if !shouldRetryEastMoneyWithCookie(err, receiver.config) {
		return nil, false, err
	}
	if receiver.cookieHeaderProvider == nil {
		return nil, false, err
	}
	// 只在连接类失败时走一次 cookie 重试，避免把真正的业务错误放大成无限重试。
	cookieHeader := receiver.cookieHeaderProvider(receiver.config)
	if strings.TrimSpace(cookieHeader) == "" {
		return nil, false, err
	}
	body, retryErr := receiver.fetchHTTP(reqURL, cookieHeader)
	if retryErr != nil {
		return nil, true, fmt.Errorf("首次请求失败: %v; cookie 重试失败: %w", err, retryErr)
	}
	return body, true, nil
}

// GetKLineData 获取 K 线数据（最新一段，等价于 end=20500101）
func (receiver *EastMoneyKLineApi) GetKLineData(stockCode, kLineType, adjustFlag string, days int) *[]KLineData {
	return receiver.GetKLineDataBefore(stockCode, kLineType, adjustFlag, days, "20500101")
}

// GetKLineData2 与 GetKLineData 相同，保留给历史测试/调用方。
func (receiver *EastMoneyKLineApi) GetKLineData2(stockCode, kLineType, adjustFlag string, days int) *[]KLineData {
	return receiver.GetKLineDataBefore(stockCode, kLineType, adjustFlag, days, "20500101")
}

// GetKLineDataBefore 获取 end 时间点之前的 limit 根 K 线。
// end 为空或 "20500101" 表示取到最新；否则为东方财富格式：日/周等为 YYYYMMDD，分钟线多为 YYYYMMDDHHmmss（与 f51 字段一致即可）。
func (receiver *EastMoneyKLineApi) GetKLineDataBefore(stockCode, kLineType, adjustFlag string, limit int, end string) *[]KLineData {
	result := receiver.GetKLineDataBeforeResult(stockCode, kLineType, adjustFlag, limit, end)
	if result.ErrorCode != "" {
		eastMoneyKlineLog.Errorf(
			"data.eastmoney_kline.fetch_failed",
			"GetKLineDataBefore failed stock=%s klt=%s limit=%d end=%s retry=%t code=%s msg=%s",
			stockCode,
			kLineType,
			limit,
			end,
			result.UsedCookieRetry,
			result.ErrorCode,
			result.Message,
		)
	}
	data := append([]KLineData(nil), result.Data...)
	return &data
}

func (receiver *EastMoneyKLineApi) GetKLineDataBeforeResult(stockCode, kLineType, adjustFlag string, limit int, end string) *eastMoneyFetchResult {
	result := &eastMoneyFetchResult{
		Data: make([]KLineData, 0),
	}

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

	fields := "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f116"
	if adjustFlag != "" {
		fields = "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f116,f113,f114,f115"
	}

	params := url.Values{}
	params.Set("secid", secid)
	params.Set("klt", kLineType)
	params.Set("fqt", adjustFlag)
	params.Set("end", end)
	params.Set("lmt", convertor.ToString(limit))
	params.Set("fields1", "f1,f2,f3,f4,f5,f6")
	params.Set("fields2", fields)
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

// GetMinuteKLine 获取分时 K 线数据 (1 分钟、5 分钟等)
func (receiver *EastMoneyKLineApi) GetMinuteKLine(stockCode string, minuteType KLineType, days int) *[]KLineData {
	return receiver.GetKLineData(stockCode, string(minuteType), "", days)
}

// GetDayKLine 获取日 K 线数据
func (receiver *EastMoneyKLineApi) GetDayKLine(stockCode string, days int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "101", "", days)
}

// GetWeekKLine 获取周 K 线数据
func (receiver *EastMoneyKLineApi) GetWeekKLine(stockCode string, weeks int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "102", "", weeks)
}

// GetMonthKLine 获取月 K 线数据
func (receiver *EastMoneyKLineApi) GetMonthKLine(stockCode string, months int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "103", "", months)
}

// GetQuarterKLine 获取季 K 线数据
func (receiver *EastMoneyKLineApi) GetQuarterKLine(stockCode string, quarters int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "104", "", quarters)
}

// GetYearKLine 获取年 K 线数据
func (receiver *EastMoneyKLineApi) GetYearKLine(stockCode string, years int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "106", "", years)
}

// GetAdjustedKLine 获取复权 K 线数据
// adjustType: qfq=前复权，hfq=后复权
func (receiver *EastMoneyKLineApi) GetAdjustedKLine(stockCode, adjustType string, days int) *[]KLineData {
	return receiver.GetKLineData(stockCode, "101", adjustType, days)
}

func isNumericStockCode(code string) bool {
	if code == "" {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// convertStockCode 转换股票代码为东方财富格式
// 输入：000001 或 sz000001 或 000001.SZ
// 输出：0.000001 或 1.600000
func (receiver *EastMoneyKLineApi) convertStockCode(stockCode string) string {
	stockCode = strings.ToUpper(strings.TrimSpace(stockCode))
	if stockCode == "" {
		return ""
	}

	// 如果已经包含点号，说明是标准格式
	if strings.Contains(stockCode, ".") {
		parts := strings.Split(stockCode, ".")
		if len(parts) == 2 {
			code := parts[0]
			market := parts[1]
			if !isNumericStockCode(code) {
				return ""
			}

			switch market {
			case "SH", "SS":
				return "1." + code
			case "SZ":
				return "0." + code
			case "BJ":
				return "0." + code
			case "HK":
				return "128." + code
			case "BK":
				return "90." + code

			default:
				return ""
			}
		}
		return ""
	}

	// 处理带市场前缀的代码
	if len(stockCode) > 2 && (strings.HasPrefix(stockCode, "SH") || strings.HasPrefix(stockCode, "SZ") || strings.HasPrefix(stockCode, "BJ") || strings.HasPrefix(stockCode, "HK") || strings.HasPrefix(stockCode, "BK")) {
		market := stockCode[:2]
		code := stockCode[2:]
		if !isNumericStockCode(code) {
			return ""
		}

		switch market {
		case "SH":
			return "1." + code
		case "SZ":
			return "0." + code
		case "BJ":
			return "0." + code
		case "HK":
			return "128." + code
		case "BK":
			return "90." + code
		default:
			return ""
		}
	}

	// 纯数字代码仅对 6 位 A 股代码做市场推断，避免将 5 位港股代码误判为有效代码。
	if len(stockCode) == 6 && isNumericStockCode(stockCode) {
		firstChar := stockCode[0:1]
		switch firstChar {
		case "6": // 沪市主板
			return "1." + stockCode
		case "8", "9": // 北交所
			return "0." + stockCode
		case "0", "3": // 深市
			return "0." + stockCode
		default:
			return ""
		}
	}

	return ""
}

// getAdjustType 获取复权类型对应的数字
func (receiver *EastMoneyKLineApi) getAdjustType(adjustFlag string) string {
	switch strings.ToLower(adjustFlag) {
	case "qfq":
		return "1" // 前复权
	case "hfq":
		return "2" // 后复权
	default:
		return "0" // 不复权
	}
}

// parseKLine 解析单条 K 线数据
// K 线数据格式：日期，开盘价，收盘价，最高价，最低价，成交量，成交额，振幅，涨跌幅，涨跌额，换手率，市盈率 TTM
func (receiver *EastMoneyKLineApi) parseKLine(klineStr, adjustFlag string) *KLineData {
	//logger.SugaredLogger.Debugf("parseKLine: %s", klineStr)
	parts := strings.Split(klineStr, ",")
	if len(parts) < 11 {
		eastMoneyKlineLog.Warnf("data.eastmoney_kline.invalid_format", "invalid kline format: %s", klineStr)
		return nil
	}

	kline := &KLineData{
		Day:           parts[0],                       // 日期
		Open:          receiver.parseFloat(parts[1]),  // 开盘价
		Close:         receiver.parseFloat(parts[2]),  // 收盘价
		High:          receiver.parseFloat(parts[3]),  // 最高价
		Low:           receiver.parseFloat(parts[4]),  // 最低价
		Volume:        receiver.parseFloat(parts[5]),  // 成交量 (手)
		Amount:        receiver.parseFloat(parts[6]),  // 成交额 (元)
		Amplitude:     receiver.parseFloat(parts[7]),  // 振幅 (%)
		ChangePercent: receiver.parseFloat(parts[8]),  // 涨跌幅 (%)
		ChangeValue:   receiver.parseFloat(parts[9]),  // 涨跌额 (元)
		TurnoverRate:  receiver.parseFloat(parts[10]), // 换手率 (%)
	}

	// 如果有复权数据，解析额外字段
	if adjustFlag != "" && len(parts) >= 14 {
		// f113: 前复权开盘价
		// f114: 前复权收盘价
		// f115: 前复权最高价
		// f116: 前复权最低价
		// 这些字段可以根据需要使用
	}

	return kline
}

// parseFloat 安全转换浮点数字符串
func (receiver *EastMoneyKLineApi) parseFloat(s string) string {
	if s == "" || s == "-" || s == "null" {
		return "0"
	}
	return s
}

// GetBatchKLineData 批量获取多只股票的 K 线数据
func (receiver *EastMoneyKLineApi) GetBatchKLineData(stockCodes []string, kLineType string, days int) map[string]*[]KLineData {
	result := make(map[string]*[]KLineData)

	for i, stockCode := range stockCodes {
		kLines := receiver.GetKLineData(stockCode, kLineType, "", days)
		result[stockCode] = kLines

		// 使用更智能的延迟策略
		if i < len(stockCodes)-1 { // 最后一个不需要延迟
			// 随机延迟 200-800ms，模拟人类操作
			delay := 200 + rand.Intn(600)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	return result
}

// GetLatestKLine 获取最新一条 K 线数据
func (receiver *EastMoneyKLineApi) GetLatestKLine(stockCode string, kLineType string) *KLineData {
	kLines := receiver.GetKLineData(stockCode, kLineType, "", 1)
	if len(*kLines) > 0 {
		return &(*kLines)[0]
	}
	return nil
}

// GetKLineWithMA 获取带均线的 K 线数据，支持任意周期的简单移动平均（SMA，以收盘价计算）。
// maPeriods 为均线周期，如 5,10,20,60,120；若未传则默认 5,10,20,60。
func (receiver *EastMoneyKLineApi) GetKLineWithMA(stockCode string, kLineType string, days int, maPeriods ...int) (*[]KLineData, error) {
	periods := maPeriods
	if len(periods) == 0 {
		periods = []int{5, 10, 20, 60}
	}
	maxPeriod := getMaxPeriod(periods)
	fetchDays := days + maxPeriod
	if fetchDays < 1 {
		fetchDays = maxPeriod + 60
	}
	full := receiver.GetKLineData(stockCode, kLineType, "", fetchDays)
	if full == nil || len(*full) == 0 {
		return full, nil
	}
	total := len(*full)
	// 收盘价序列（完整长度，用于计算均线）
	closes := make([]float64, total)
	for i, k := range *full {
		v, _ := parseFloatToFloat(k.Close)
		closes[i] = v
	}
	// 只返回最后 days 条
	if total > days {
		*full = (*full)[total-days:]
		total = len(*full)
	}
	offset := len(closes) - total // 截取后在 closes 中的起始下标
	// 对每个周期计算 SMA，并写回每条 K 线的 MA
	for _, p := range periods {
		if p <= 0 {
			continue
		}
		for i := 0; i < total; i++ {
			idx := offset + i
			ma := computeSMA(closes, idx, p)
			if ma < 0 {
				continue
			}
			if (*full)[i].MA == nil {
				(*full)[i].MA = make(map[string]string)
			}
			(*full)[i].MA[fmt.Sprintf("%d", p)] = fmt.Sprintf("%.4f", ma)
		}
	}
	return full, nil
}

// computeSMA 计算 closes 从 idx 往前 period 根的收盘价简单移动平均；不足 period 根返回 -1。
func computeSMA(closes []float64, idx, period int) float64 {
	start := idx - period + 1
	if start < 0 {
		return -1
	}
	sum := 0.0
	for j := start; j <= idx; j++ {
		sum += closes[j]
	}
	return sum / float64(period)
}

// parseFloatToFloat 将 K 线价格字符串转为 float64
func parseFloatToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" || s == "null" {
		return 0, nil
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func getMaxPeriod(periods []int) int {
	max := 0
	for _, p := range periods {
		if p > max {
			max = p
		}
	}
	return max
}

// AggregateKLineEveryN 将连续 n 根 1 分钟 K 线合并为一根（用于东方财富无原生 10 分钟周期时）。
// 假定 src 已按时间正序排列。
func AggregateKLineEveryN(src *[]KLineData, n int) *[]KLineData {
	if src == nil || n < 2 {
		return src
	}
	arr := *src
	if len(arr) == 0 {
		return src
	}
	out := make([]KLineData, 0, len(arr)/n+1)
	for i := 0; i < len(arr); {
		end := i + n
		if end > len(arr) {
			end = len(arr)
		}
		chunk := arr[i:end]
		first := chunk[0]
		last := chunk[len(chunk)-1]
		highF := -1e18
		lowF := 1e18
		volSum := 0.0
		amtSum := 0.0
		for _, c := range chunk {
			h, _ := parseFloatToFloat(c.High)
			l, _ := parseFloatToFloat(c.Low)
			v, _ := parseFloatToFloat(c.Volume)
			a, _ := parseFloatToFloat(c.Amount)
			if h > highF {
				highF = h
			}
			if l < lowF {
				lowF = l
			}
			volSum += v
			amtSum += a
		}
		openS := first.Open
		closeS := last.Close
		highS := fmt.Sprintf("%.4f", highF)
		lowS := fmt.Sprintf("%.4f", lowF)
		if highF <= -1e17 {
			highS = first.High
		}
		if lowF >= 1e17 {
			lowS = first.Low
		}
		out = append(out, KLineData{
			Day:           last.Day,
			Open:          openS,
			Close:         closeS,
			High:          highS,
			Low:           lowS,
			Volume:        fmt.Sprintf("%.0f", volSum),
			Amount:        fmt.Sprintf("%.2f", amtSum),
			ChangePercent: last.ChangePercent,
			ChangeValue:   last.ChangeValue,
			TurnoverRate:  last.TurnoverRate,
			Amplitude:     last.Amplitude,
		})
		i = end
		if len(chunk) < n {
			break
		}
	}
	return &out
}

// ValidateStockCode 验证股票代码是否有效
func (receiver *EastMoneyKLineApi) ValidateStockCode(stockCode string) bool {
	secid := receiver.convertStockCode(stockCode)
	return secid != ""
}

// GetKLineCount 获取指定时间段内的 K 线数量
func (receiver *EastMoneyKLineApi) GetKLineCount(stockCode string, kLineType string, startDate, endDate string) int {
	// 这里可以实现获取指定时间范围内的 K 线数量
	// 实际实现需要根据具体的 API 参数来调整
	kLines := receiver.GetKLineData(stockCode, kLineType, "", 365) // 默认获取一年数据
	return len(*kLines)
}

// init 函数添加随机数种子初始化
func init() {
	rand.Seed(time.Now().UnixNano())
}
