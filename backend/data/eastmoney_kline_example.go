package data

import (
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"go-stock/backend/logger"
	"strings"
)

// @Author spark
// @Date 2026/3/15
// @Desc 东方财富 K 线数据 API 使用示例

var eastMoneyExampleLog = dataModuleLogger(logger.SinkApp, "data.eastmoney_kline_example")

func logEastMoneyExample(event string, lines ...string) {
	eastMoneyExampleLog.Info(event, strings.Join(lines, "\n"))
}

// Example_GetDayKLine 获取日 K 线数据示例
func Example_GetDayKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	// 获取平安银行最近 30 天的日 K 线数据
	stockCode := "000001.SZ"
	kLines := api.GetDayKLine(stockCode, 30)

	if len(*kLines) == 0 {
		eastMoneyExampleLog.Error("data.eastmoney_kline_example.day_failed", "获取数据失败")
		return
	}

	lines := []string{
		fmt.Sprintf("获取到 %d 条日 K 线数据", len(*kLines)),
		strings.Repeat("-", 80),
		fmt.Sprintf("%-12s %-10s %-10s %-10s %-10s %-12s", "日期", "开盘", "收盘", "最高", "最低", "成交量 (手)"),
		strings.Repeat("-", 80),
	}

	for i := len(*kLines) - 1; i >= 0; i-- {
		kline := (*kLines)[i]
		lines = append(lines, fmt.Sprintf("%-12s %-10s %-10s %-10s %-10s %-12s",
			kline.Day, kline.Open, kline.Close, kline.High, kline.Low, kline.Volume))
	}
	logEastMoneyExample("data.eastmoney_kline_example.day_output", lines...)
}

// Example_GetWeekKLine 获取周 K 线数据示例
func Example_GetWeekKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	// 获取贵州茅台最近 20 周的周 K 线数据
	stockCode := "600519.SH"
	kLines := api.GetWeekKLine(stockCode, 20)

	if len(*kLines) == 0 {
		eastMoneyExampleLog.Error("data.eastmoney_kline_example.week_failed", "获取数据失败")
		return
	}

	lines := []string{
		fmt.Sprintf("获取到 %d 条周 K 线数据", len(*kLines)),
		strings.Repeat("-", 80),
		fmt.Sprintf("%-12s %-10s %-10s %-10s %-10s", "日期", "开盘", "收盘", "最高", "最低"),
		strings.Repeat("-", 80),
	}

	for i := len(*kLines) - 1; i >= 0; i-- {
		kline := (*kLines)[i]
		lines = append(lines, fmt.Sprintf("%-12s %-10s %-10s %-10s %-10s",
			kline.Day, kline.Open, kline.Close, kline.High, kline.Low))
	}
	logEastMoneyExample("data.eastmoney_kline_example.week_output", lines...)
}

// Example_GetAdjustedKLine 获取复权 K 线数据示例
func Example_GetAdjustedKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	// 获取宁德时代前复权日 K 线数据
	stockCode := "300750.SZ"

	qfqKLines := api.GetAdjustedKLine(stockCode, "qfq", 10)
	lines := []string{"=== 前复权数据 ==="}
	for _, kline := range *qfqKLines {
		lines = append(lines, fmt.Sprintf("日期:%s 开盘:%s 收盘:%s 最高:%s 最低:%s",
			kline.Day, kline.Open, kline.Close, kline.High, kline.Low))
	}

	// 获取后复权日 K 线数据
	lines = append(lines, "=== 后复权数据 ===")
	hfqKLines := api.GetAdjustedKLine(stockCode, "hfq", 10)
	for _, kline := range *hfqKLines {
		lines = append(lines, fmt.Sprintf("日期:%s 开盘:%s 收盘:%s 最高:%s 最低:%s",
			kline.Day, kline.Open, kline.Close, kline.High, kline.Low))
	}

	// 获取不复权数据
	lines = append(lines, "=== 不复权数据 ===")
	noAdjKLines := api.GetDayKLine(stockCode, 10)
	for _, kline := range *noAdjKLines {
		lines = append(lines, fmt.Sprintf("日期:%s 开盘:%s 收盘:%s 最高:%s 最低:%s",
			kline.Day, kline.Open, kline.Close, kline.High, kline.Low))
	}
	logEastMoneyExample("data.eastmoney_kline_example.adjusted_output", lines...)
}

// Example_GetMinuteKLine 获取分钟 K 线数据示例
func Example_GetMinuteKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	// 获取 5 分钟 K 线数据
	stockCode := "000001.SZ"

	kLines5Min := api.GetMinuteKLine(stockCode, KLineType5Min, 50)
	lines := []string{
		"=== 5 分钟 K 线 ===",
		fmt.Sprintf("获取到 %d 条 5 分钟 K 线数据", len(*kLines5Min)),
	}

	// 获取 15 分钟 K 线数据
	kLines15Min := api.GetMinuteKLine(stockCode, KLineType15Min, 50)
	lines = append(lines, "=== 15 分钟 K 线 ===")
	lines = append(lines, fmt.Sprintf("获取到 %d 条 15 分钟 K 线数据", len(*kLines15Min)))
	logEastMoneyExample("data.eastmoney_kline_example.minute_output", lines...)
}

// Example_BatchGetKLine 批量获取 K 线数据示例
func Example_BatchGetKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	// 批量获取多只股票的 K 线数据
	stockCodes := []string{
		"000001.SZ", // 平安银行
		"600519.SH", // 贵州茅台
		"300750.SZ", // 宁德时代
		"00700.HK",  // 腾讯控股
	}

	result := api.GetBatchKLineData(stockCodes, "101", 5)
	lines := []string{"=== 批量获取日 K 线数据 ==="}

	for code, kLines := range result {
		lines = append(lines, fmt.Sprintf("股票：%s, 获取到 %d 条数据", code, len(*kLines)))
		if len(*kLines) > 0 {
			latest := (*kLines)[len(*kLines)-1]
			lines = append(lines, fmt.Sprintf("最新数据 - 日期:%s, 收盘价:%s, 涨跌幅:%.2f%%",
				latest.Day, latest.Close, calculateChangePercent(latest.Open, latest.Close)))
		}
	}
	logEastMoneyExample("data.eastmoney_kline_example.batch_output", lines...)
}

// Example_AnalyzeKLine 分析 K 线数据示例
func Example_AnalyzeKLine() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	stockCode := "600519.SH"
	kLines := api.GetDayKLine(stockCode, 60)

	if len(*kLines) == 0 {
		eastMoneyExampleLog.Error("data.eastmoney_kline_example.analysis_failed", "获取数据失败")
		return
	}

	// 分析最近 30 天的数据
	recentDays := 30
	startIndex := len(*kLines) - recentDays
	if startIndex < 0 {
		startIndex = 0
	}

	var totalVolume float64
	var avgClose float64
	var highestPrice float64
	var lowestPrice float64

	for i := startIndex; i < len(*kLines); i++ {
		kline := (*kLines)[i]

		volume, _ := convertor.ToFloat(kline.Volume)
		totalVolume += volume

		close, _ := convertor.ToFloat(kline.Close)
		avgClose += close

		high, _ := convertor.ToFloat(kline.High)
		if high > highestPrice {
			highestPrice = high
		}

		low, _ := convertor.ToFloat(kline.Low)
		if lowestPrice == 0 || low < lowestPrice {
			lowestPrice = low
		}
	}

	count := len(*kLines) - startIndex
	avgClose /= float64(count)
	totalVolume /= float64(count)

	lines := []string{
		"=== K 线技术分析 ===",
		fmt.Sprintf("统计周期：%d天", count),
		fmt.Sprintf("平均收盘价：%.2f", avgClose),
		fmt.Sprintf("最高价：%.2f", highestPrice),
		fmt.Sprintf("最低价：%.2f", lowestPrice),
		fmt.Sprintf("平均成交量：%.2f 手", totalVolume),
	}

	// 判断趋势
	firstClose, _ := convertor.ToFloat((*kLines)[startIndex].Close)
	lastClose, _ := convertor.ToFloat((*kLines)[len(*kLines)-1].Close)
	changePercent := (lastClose - firstClose) / firstClose * 100

	lines = append(lines, fmt.Sprintf("期间涨跌幅：%.2f%%", changePercent))
	if changePercent > 0 {
		lines = append(lines, "趋势：上涨 ↗")
	} else if changePercent < 0 {
		lines = append(lines, "趋势：下跌 ↘")
	} else {
		lines = append(lines, "趋势：持平 →")
	}
	logEastMoneyExample("data.eastmoney_kline_example.analysis_output", lines...)
}

// Example_ExportKLineToJSON 导出 K 线数据为 JSON 示例
func Example_ExportKLineToJSON() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	stockCode := "000001.SZ"
	kLines := api.GetDayKLine(stockCode, 30)

	// 转换为 JSON
	jsonData, err := json.MarshalIndent(kLines, "", "  ")
	if err != nil {
		eastMoneyExampleLog.Errorf("data.eastmoney_kline_example.json_failed", "JSON 转换失败：%v", err)
		return
	}

	logEastMoneyExample("data.eastmoney_kline_example.json_output", "=== JSON 格式数据 ===", string(jsonData))
}

// Example_GetLatestMarketInfo 获取最新行情信息示例
func Example_GetLatestMarketInfo() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	stockCodes := []string{
		"000001.SZ", // 平安银行
		"600519.SH", // 贵州茅台
		"300750.SZ", // 宁德时代
	}

	lines := []string{"=== 最新行情 ==="}
	for _, code := range stockCodes {
		latestKLine := api.GetLatestKLine(code, "101")
		if latestKLine != nil {
			open, _ := convertor.ToFloat(latestKLine.Open)
			_, _ = convertor.ToFloat(latestKLine.Close) // closePrice - 暂时不使用
			high, _ := convertor.ToFloat(latestKLine.High)
			low, _ := convertor.ToFloat(latestKLine.Low)

			changePercent := calculateChangePercent(latestKLine.Open, latestKLine.Close)

			lines = append(lines, fmt.Sprintf("%-10s 日期:%-12s 收盘价:%-8s 涨跌幅:%+6.2f%% 振幅:%.2f%%",
				code, latestKLine.Day, latestKLine.Close, changePercent,
				(high-low)/open*100))
		}
	}
	logEastMoneyExample("data.eastmoney_kline_example.latest_output", lines...)
}

// calculateChangePercent 计算涨跌幅
func calculateChangePercent(open, close string) float64 {
	o, _ := convertor.ToFloat(open)
	c, _ := convertor.ToFloat(close)
	if o == 0 {
		return 0
	}
	return (c - o) / o * 100
}

// Example_ValidateStockCodes 批量验证股票代码示例
func Example_ValidateStockCodes() {
	config := GetSettingConfig()
	api := NewEastMoneyKLineApi(config)

	testCodes := []string{
		"000001.SZ",
		"600519.SH",
		"00700.HK",
		"invalid_code",
		"123456",
	}

	lines := []string{"=== 股票代码验证 ==="}
	for _, code := range testCodes {
		isValid := api.ValidateStockCode(code)
		status := "✓ 有效"
		if !isValid {
			status = "✗ 无效"
		}
		lines = append(lines, fmt.Sprintf("%-15s -> %s", code, status))
	}
	logEastMoneyExample("data.eastmoney_kline_example.validate_output", lines...)
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	logEastMoneyExample(
		"data.eastmoney_kline_example.run_start",
		strings.Repeat("=", 80),
		"东方财富 K 线数据 API 使用示例",
		strings.Repeat("=", 80),
	)

	Example_GetDayKLine()
	Example_GetWeekKLine()
	Example_GetAdjustedKLine()
	Example_GetMinuteKLine()
	Example_BatchGetKLine()
	Example_AnalyzeKLine()
	Example_ExportKLineToJSON()
	Example_GetLatestMarketInfo()
	Example_ValidateStockCodes()

	logEastMoneyExample(
		"data.eastmoney_kline_example.run_done",
		strings.Repeat("=", 80),
		"所有示例运行完成",
		strings.Repeat("=", 80),
	)
}
