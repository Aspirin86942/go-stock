# Stock Chip Distribution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a callable `GetStockChipDistribution` tool that computes stock chip distribution locally from existing EastMoney daily K-line data and returns both summary metrics and a price-distribution detail table.

**Architecture:** Keep the CYQ algorithm in a dedicated data-layer calculator so both tool-call paths can reuse the same logic. Add one backend tool handler in `backend/data`, expose the same tool schema in `backend/data/tools.go`, and mirror the tool in `backend/agent/tools/data_tools_wrapper.go` so OpenAI tool mode and agent wrapper mode stay aligned.

**Tech Stack:** Go, Resty, CloudWeGo Eino tools, PowerShell, Go `testing`, existing EastMoney K-line API helpers

---

### Task 1: Add deterministic CYQ calculator tests and the shared calculation core

**Files:**
- Create: `D:\codex_work\go-stock\backend\data\chip_distribution_calculator.go`
- Create: `D:\codex_work\go-stock\backend\data\chip_distribution_calculator_test.go`

- [ ] **Step 1: Write the failing calculator tests**

```go
package data

import (
	"strings"
	"testing"
)

func TestCalculateChipDistribution_ProducesSummaryAndDistribution(t *testing.T) {
	result, err := CalculateChipDistribution([]KLineData{
		{Day: "2026-03-31", Open: "10.00", Close: "10.20", High: "10.30", Low: "9.90", TurnoverRate: "3.20"},
		{Day: "2026-04-01", Open: "10.18", Close: "10.35", High: "10.40", Low: "10.10", TurnoverRate: "3.60"},
		{Day: "2026-04-02", Open: "10.32", Close: "10.48", High: "10.55", Low: "10.20", TurnoverRate: "4.10"},
		{Day: "2026-04-03", Open: "10.45", Close: "10.62", High: "10.80", Low: "10.30", TurnoverRate: "4.50"},
	}, 50)
	if err != nil {
		t.Fatalf("CalculateChipDistribution returned error: %v", err)
	}
	if result.LatestDate != "2026-04-03" {
		t.Fatalf("expected latest date 2026-04-03, got %q", result.LatestDate)
	}
	if result.SampleDays != 4 {
		t.Fatalf("expected sample days 4, got %d", result.SampleDays)
	}
	if result.BenefitPart < 0 || result.BenefitPart > 100 {
		t.Fatalf("expected benefit part in [0,100], got %f", result.BenefitPart)
	}
	if result.AvgCost < 9.90 || result.AvgCost > 10.80 {
		t.Fatalf("expected avg cost in [9.90,10.80], got %f", result.AvgCost)
	}
	if result.Cost70.LowerPrice < result.Cost90.LowerPrice {
		t.Fatalf("expected 70%% lower bound >= 90%% lower bound, got 70=%f 90=%f", result.Cost70.LowerPrice, result.Cost90.LowerPrice)
	}
	if result.Cost70.UpperPrice > result.Cost90.UpperPrice {
		t.Fatalf("expected 70%% upper bound <= 90%% upper bound, got 70=%f 90=%f", result.Cost70.UpperPrice, result.Cost90.UpperPrice)
	}
	if result.Cost70.Concentration < 0 || result.Cost90.Concentration < 0 {
		t.Fatalf("expected non-negative concentration, got 70=%f 90=%f", result.Cost70.Concentration, result.Cost90.Concentration)
	}
	if len(result.Distribution) == 0 {
		t.Fatalf("expected non-empty distribution")
	}
	if result.Distribution[0].Price > result.Distribution[len(result.Distribution)-1].Price {
		t.Fatalf("expected distribution sorted by price ascending")
	}
}

func TestCalculateChipDistribution_RejectsInvalidInput(t *testing.T) {
	_, err := CalculateChipDistribution(nil, 50)
	if err == nil || !strings.Contains(err.Error(), "K 线数据不能为空") {
		t.Fatalf("expected empty kline error, got %v", err)
	}

	_, err = CalculateChipDistribution([]KLineData{
		{Day: "2026-04-03", Open: "10.45", Close: "10.62", High: "10.80", Low: "10.30", TurnoverRate: "4.50"},
	}, 0)
	if err == nil || !strings.Contains(err.Error(), "accuracyFactor") {
		t.Fatalf("expected invalid accuracyFactor error, got %v", err)
	}
}
```

- [ ] **Step 2: Run the calculator tests to verify they fail**

Run:

```powershell
conda run -n test go test ./backend/data -run "TestCalculateChipDistribution_" -count=1
```

Expected:

- compile failure because `CalculateChipDistribution` and the result types do not exist yet

- [ ] **Step 3: Write the minimal shared calculator implementation**

```go
package data

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type ChipDistributionPoint struct {
	Price  float64 `json:"price"`
	Weight float64 `json:"weight"`
}

type ChipCostRange struct {
	LowerPrice    float64 `json:"lowerPrice"`
	UpperPrice    float64 `json:"upperPrice"`
	Concentration float64 `json:"concentration"`
}

type ChipDistributionResult struct {
	LatestDate   string                  `json:"latestDate"`
	LatestClose  float64                 `json:"latestClose"`
	SampleDays   int                     `json:"sampleDays"`
	BenefitPart  float64                 `json:"benefitPart"`
	AvgCost      float64                 `json:"avgCost"`
	Cost70       ChipCostRange           `json:"cost70"`
	Cost90       ChipCostRange           `json:"cost90"`
	Distribution []ChipDistributionPoint `json:"distribution"`
}

type chipParsedKLine struct {
	Day          string
	Open         float64
	Close        float64
	High         float64
	Low          float64
	TurnoverRate float64
}

func CalculateChipDistribution(klines []KLineData, accuracyFactor int) (*ChipDistributionResult, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("K 线数据不能为空")
	}
	if accuracyFactor <= 0 {
		return nil, fmt.Errorf("accuracyFactor 必须大于 0")
	}

	parsed, err := parseChipDistributionKLines(klines)
	if err != nil {
		return nil, err
	}

	priceStep := findChipPriceStep(parsed, accuracyFactor)
	chips := make(map[int]float64)
	for _, line := range parsed {
		turnover := math.Min(1, line.TurnoverRate/100)
		for idx, weight := range chips {
			chips[idx] = weight * (1 - turnover)
		}
		dailyWeights := buildDailyChipWeights(line, priceStep)
		for idx, weight := range dailyWeights {
			chips[idx] += weight * turnover
		}
	}

	distribution := normalizeChipDistribution(chips, priceStep)
	latest := parsed[len(parsed)-1]
	result := &ChipDistributionResult{
		LatestDate:   latest.Day,
		LatestClose:  roundChipValue(latest.Close, 4),
		SampleDays:   len(parsed),
		Distribution: distribution,
	}
	result.BenefitPart = roundChipValue(calcBenefitPart(distribution, latest.Close)*100, 2)
	result.AvgCost = roundChipValue(calcAvgCost(distribution), 4)
	result.Cost70 = calcChipCostRange(distribution, 0.70)
	result.Cost90 = calcChipCostRange(distribution, 0.90)
	return result, nil
}

func parseChipDistributionKLines(klines []KLineData) ([]chipParsedKLine, error) {
	parsed := make([]chipParsedKLine, 0, len(klines))
	for _, k := range klines {
		open, err := strconv.ParseFloat(k.Open, 64)
		if err != nil {
			return nil, fmt.Errorf("开盘价解析失败: %w", err)
		}
		closePrice, err := strconv.ParseFloat(k.Close, 64)
		if err != nil {
			return nil, fmt.Errorf("收盘价解析失败: %w", err)
		}
		high, err := strconv.ParseFloat(k.High, 64)
		if err != nil {
			return nil, fmt.Errorf("最高价解析失败: %w", err)
		}
		low, err := strconv.ParseFloat(k.Low, 64)
		if err != nil {
			return nil, fmt.Errorf("最低价解析失败: %w", err)
		}
		turnoverRate, err := strconv.ParseFloat(k.TurnoverRate, 64)
		if err != nil {
			return nil, fmt.Errorf("换手率解析失败: %w", err)
		}
		if high < low {
			return nil, fmt.Errorf("K 线价格区间无效: day=%s", k.Day)
		}
		parsed = append(parsed, chipParsedKLine{
			Day:          k.Day,
			Open:         open,
			Close:        closePrice,
			High:         high,
			Low:          low,
			TurnoverRate: turnoverRate,
		})
	}
	return parsed, nil
}

func findChipPriceStep(klines []chipParsedKLine, accuracyFactor int) float64 {
	minLow := klines[0].Low
	maxHigh := klines[0].High
	for _, k := range klines[1:] {
		if k.Low < minLow {
			minLow = k.Low
		}
		if k.High > maxHigh {
			maxHigh = k.High
		}
	}
	width := maxHigh - minLow
	if width <= 0 {
		return 0.01
	}
	step := width / float64(accuracyFactor)
	if step < 0.01 {
		return 0.01
	}
	return step
}

func buildDailyChipWeights(line chipParsedKLine, step float64) map[int]float64 {
	minIdx := priceToChipIndex(line.Low, step)
	maxIdx := priceToChipIndex(line.High, step)
	if maxIdx < minIdx {
		maxIdx = minIdx
	}
	centerPrice := (line.Open + line.Close + line.High + line.Low) / 4
	centerIdx := priceToChipIndex(centerPrice, step)
	if centerIdx < minIdx {
		centerIdx = minIdx
	}
	if centerIdx > maxIdx {
		centerIdx = maxIdx
	}
	if minIdx == maxIdx {
		return map[int]float64{minIdx: 1}
	}

	weights := make(map[int]float64, maxIdx-minIdx+1)
	total := 0.0
	span := float64(maxIdx - minIdx + 1)
	for idx := minIdx; idx <= maxIdx; idx++ {
		distance := math.Abs(float64(idx - centerIdx))
		weight := span - distance
		if weight < 1 {
			weight = 1
		}
		weights[idx] = weight
		total += weight
	}
	for idx, weight := range weights {
		weights[idx] = weight / total
	}
	return weights
}

func normalizeChipDistribution(chips map[int]float64, step float64) []ChipDistributionPoint {
	indexes := make([]int, 0, len(chips))
	total := 0.0
	for idx, weight := range chips {
		if weight <= 0 {
			continue
		}
		indexes = append(indexes, idx)
		total += weight
	}
	sort.Ints(indexes)
	if total == 0 {
		return nil
	}

	points := make([]ChipDistributionPoint, 0, len(indexes))
	for _, idx := range indexes {
		points = append(points, ChipDistributionPoint{
			Price:  roundChipValue(float64(idx)*step, 4),
			Weight: chips[idx] / total,
		})
	}
	return points
}

func calcBenefitPart(points []ChipDistributionPoint, latestClose float64) float64 {
	total := 0.0
	for _, point := range points {
		if point.Price <= latestClose {
			total += point.Weight
		}
	}
	return total
}

func calcAvgCost(points []ChipDistributionPoint) float64 {
	total := 0.0
	for _, point := range points {
		total += point.Price * point.Weight
	}
	return total
}

func calcChipCostRange(points []ChipDistributionPoint, target float64) ChipCostRange {
	if len(points) == 0 {
		return ChipCostRange{}
	}
	bestLeft, bestRight := 0, len(points)-1
	bestWidth := math.MaxFloat64
	left := 0
	sum := 0.0
	for right := 0; right < len(points); right++ {
		sum += points[right].Weight
		for sum >= target && left <= right {
			width := points[right].Price - points[left].Price
			if width < bestWidth {
				bestWidth = width
				bestLeft = left
				bestRight = right
			}
			sum -= points[left].Weight
			left++
		}
	}
	lower := points[bestLeft].Price
	upper := points[bestRight].Price
	concentration := 0.0
	if upper > 0 {
		concentration = (upper - lower) / upper * 100
	}
	return ChipCostRange{
		LowerPrice:    roundChipValue(lower, 4),
		UpperPrice:    roundChipValue(upper, 4),
		Concentration: roundChipValue(concentration, 2),
	}
}

func priceToChipIndex(price, step float64) int {
	return int(math.Round(price / step))
}

func roundChipValue(value float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(value*pow) / pow
}
```

- [ ] **Step 4: Run the calculator tests to verify they pass**

Run:

```powershell
conda run -n test go test ./backend/data -run "TestCalculateChipDistribution_" -count=1
```

Expected:

- command exits `0`
- both calculator tests pass

- [ ] **Step 5: Review the calculator diff before moving on**

```powershell
git diff -- backend/data/chip_distribution_calculator.go backend/data/chip_distribution_calculator_test.go
```

### Task 2: Add the backend OpenAI tool schema, handler, and markdown output

**Files:**
- Create: `D:\codex_work\go-stock\backend\data\tool_stock_chip_distribution.go`
- Create: `D:\codex_work\go-stock\backend\data\tool_stock_chip_distribution_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\tools.go`

- [ ] **Step 1: Write the failing backend tool tests**

```go
package data

import (
	"strings"
	"testing"
)

func TestTools_ContainsGetStockChipDistribution(t *testing.T) {
	tools := Tools(nil)
	found := false
	for _, tool := range tools {
		if tool.Function.Name != "GetStockChipDistribution" {
			continue
		}
		found = true
		if tool.Function.Parameters == nil {
			t.Fatalf("expected parameters for GetStockChipDistribution")
		}
		if _, ok := tool.Function.Parameters.Properties["accuracyFactor"]; !ok {
			t.Fatalf("expected accuracyFactor parameter in tool schema")
		}
	}
	if !found {
		t.Fatalf("expected GetStockChipDistribution in Tools(nil)")
	}
}

func TestChipDistributionSection_IncludesSummaryAndDetailTable(t *testing.T) {
	api := newStubEastMoneyKLineAPI(t, []string{
		"2026-03-31,10.00,10.20,10.30,9.90,100000,100000000,2.00,0.00,0.00,3.20",
		"2026-04-01,10.18,10.35,10.40,10.10,110000,110000000,2.10,1.47,0.15,3.60",
		"2026-04-02,10.32,10.48,10.55,10.20,120000,120000000,2.20,1.24,0.13,4.10",
		"2026-04-03,10.45,10.62,10.80,10.30,130000,130000000,2.30,1.34,0.14,4.50",
	})

	got := ChipDistributionSection(api, "600519.SH", "", 4, 50)

	for _, want := range []string{"获利比例", "平均成本", "90%成本区间", "70%成本区间", "价格", "筹码权重(%)", "2026-04-03"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, got)
		}
	}
}

func TestChipDistributionSection_RejectsInvalidCode(t *testing.T) {
	api := newStubEastMoneyKLineAPI(t, []string{
		"2026-04-03,10.45,10.62,10.80,10.30,130000,130000000,2.30,1.34,0.14,4.50",
	})

	got := ChipDistributionSection(api, "bad-code", "", 4, 50)
	if !strings.Contains(got, "股票代码无效") {
		t.Fatalf("expected invalid code message, got %q", got)
	}
}
```

- [ ] **Step 2: Run the backend tool tests to verify they fail**

Run:

```powershell
conda run -n test go test ./backend/data -run "TestTools_ContainsGetStockChipDistribution|TestChipDistributionSection_" -count=1
```

Expected:

- compile failure because `ChipDistributionSection` and the new schema entry do not exist yet

- [ ] **Step 3: Add the backend tool schema and handler**

```go
package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/tidwall/gjson"
)

func init() {
	registerToolHandler("GetStockChipDistribution", handleGetStockChipDistribution)
}

func ChipDistributionSection(api *EastMoneyKLineApi, stockCode, adjustFlag string, days, accuracyFactor int) string {
	if !api.ValidateStockCode(stockCode) {
		return stockCode + "：股票代码无效，请使用正确格式（如 000001.SZ、600000.SH、00700.HK）。"
	}
	if accuracyFactor <= 0 {
		return stockCode + "：参数 accuracyFactor 必须大于 0。"
	}

	adj := strings.TrimSpace(strings.ToLower(adjustFlag))
	if adj == "" {
		adj = "qfq"
	}
	if adj != "qfq" && adj != "hfq" {
		return stockCode + "：参数 adjustFlag 仅支持 qfq、hfq 或空值。"
	}
	if days <= 0 {
		days = 240
	}

	fetchResult := api.GetKLineDataBeforeResult(stockCode, "101", api.getAdjustType(adj), days, "20500101")
	if fetchResult.ErrorCode != "" {
		return stockCode + "：未获取到筹码分布所需日 K 数据，请检查股票代码与参数。"
	}

	result, err := CalculateChipDistribution(fetchResult.Data, accuracyFactor)
	if err != nil {
		return stockCode + "：筹码分布计算失败，" + err.Error()
	}

	summaryRows := []map[string]any{
		{
			"最新日期":      result.LatestDate,
			"样本天数":      result.SampleDays,
			"获利比例(%)":   fmt.Sprintf("%.2f", result.BenefitPart),
			"平均成本":      fmt.Sprintf("%.4f", result.AvgCost),
			"90%成本区间":  fmt.Sprintf("%.4f - %.4f", result.Cost90.LowerPrice, result.Cost90.UpperPrice),
			"90%集中度(%)": fmt.Sprintf("%.2f", result.Cost90.Concentration),
			"70%成本区间":  fmt.Sprintf("%.4f - %.4f", result.Cost70.LowerPrice, result.Cost70.UpperPrice),
			"70%集中度(%)": fmt.Sprintf("%.2f", result.Cost70.Concentration),
		},
	}
	detailRows := make([]map[string]any, 0, len(result.Distribution))
	for _, point := range result.Distribution {
		detailRows = append(detailRows, map[string]any{
			"价格":        fmt.Sprintf("%.4f", point.Price),
			"筹码权重(%)": fmt.Sprintf("%.4f", point.Weight*100),
		})
	}

	summaryJSON, _ := json.Marshal(summaryRows)
	summaryTable, _ := JSONToMarkdownTable(summaryJSON)
	detailJSON, _ := json.Marshal(detailRows)
	detailTable, _ := JSONToMarkdownTable(detailJSON)

	return "\r\n### " + stockCode + " 筹码分布（样本 " + convertor.ToString(result.SampleDays) + " 天）\r\n" +
		summaryTable + "\r\n\r\n#### 价格分布明细\r\n" + detailTable + "\r\n"
}

func handleGetStockChipDistribution(o *OpenAi, funcArguments string, ctx *ToolContext) error {
	days := int(gjson.Get(funcArguments, "days").Int())
	if days <= 0 {
		days = 240
	}
	adjustFlag := gjson.Get(funcArguments, "adjustFlag").String()

	accuracyFactor := 150
	if accuracy := gjson.Get(funcArguments, "accuracyFactor"); accuracy.Exists() {
		accuracyFactor = int(accuracy.Int())
		if accuracyFactor <= 0 {
			appendToolMessages(ctx.Messages, ctx.CurrentAIContent.String(), ctx.ReasoningContentText.String(),
				ctx.CurrentCallID, ctx.FuncName, funcArguments, "参数 accuracyFactor 必须大于 0。")
			return nil
		}
	}

	codes := parseStockCodesFromToolArgs(funcArguments, "stockCode")
	if len(codes) == 0 {
		appendToolMessages(ctx.Messages, ctx.CurrentAIContent.String(), ctx.ReasoningContentText.String(),
			ctx.CurrentCallID, ctx.FuncName, funcArguments, "参数 stockCode 或 stockCodes 不能为空，请传入股票代码（多只可用英文逗号分隔）。")
		return nil
	}

	ctx.Ch <- map[string]any{
		"code":     1,
		"question": ctx.Question,
		"chatId":   ctx.StreamResponseID,
		"model":    ctx.Model,
		"content":  "\r\n```\r\n开始调用工具：GetStockChipDistribution，\n参数：" + funcArguments + "\r\n```\r\n",
		"time":     time.Now().Format(time.DateTime),
	}

	res := parallelStockToolSections(codes, func(stockCode string) string {
		api := NewEastMoneyKLineApi(GetSettingConfig())
		return ChipDistributionSection(api, stockCode, adjustFlag, days, accuracyFactor)
	})
	appendToolMessages(ctx.Messages, ctx.CurrentAIContent.String(), ctx.ReasoningContentText.String(),
		ctx.CurrentCallID, ctx.FuncName, funcArguments, res)
	return nil
}
```

```go
tools = append(tools, Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "GetStockChipDistribution",
		Description: "获取股票筹码分布。基于东方财富日 K 线与换手率在本地计算最新筹码结构，返回获利比例、平均成本、70%/90%成本区间和价格分布明细表。支持一次查询多只，将并行请求后合并结果。",
		Parameters: &FunctionParameters{
			Type: "object",
			Properties: map[string]any{
				"stockCode": map[string]any{
					"type":        "string",
					"description": "股票代码。A股如 000001.SZ、600000.SH；港股如 00700.HK。多只时可用英文逗号分隔。",
				},
				"stockCodes": toolSchemaStockCodes,
				"days": map[string]any{
					"type":        "number",
					"description": "采样日 K 条数，默认 240。",
				},
				"adjustFlag": map[string]any{
					"type":        "string",
					"description": "复权类型：空值默认 qfq，也支持 qfq、hfq。",
				},
				"accuracyFactor": map[string]any{
					"type":        "number",
					"description": "价格桶精度，默认 150，必须大于 0。",
				},
			},
			Required: []string{"stockCode"},
		},
	},
})
```

- [ ] **Step 4: Run the backend calculator and tool tests to verify they pass**

Run:

```powershell
conda run -n test go test ./backend/data -run "TestCalculateChipDistribution_|TestTools_ContainsGetStockChipDistribution|TestChipDistributionSection_" -count=1
```

Expected:

- command exits `0`
- the calculator tests and backend tool tests all pass

- [ ] **Step 5: Review the backend tool diff before touching the agent wrapper**

```powershell
git diff -- backend/data/chip_distribution_calculator.go backend/data/chip_distribution_calculator_test.go backend/data/tool_stock_chip_distribution.go backend/data/tool_stock_chip_distribution_test.go backend/data/tools.go
```

### Task 3: Mirror the tool in the agent wrapper path

**Files:**
- Modify: `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper.go`
- Create: `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper_chip_distribution_test.go`

- [ ] **Step 1: Write the failing agent-wrapper test**

```go
package tools

import (
	"strings"
	"testing"
)

func TestGetAllDataTools_ContainsGetStockChipDistribution(t *testing.T) {
	allTools := GetAllDataTools()
	count := 0
	for _, tool := range allTools {
		info, err := tool.Info(nil)
		if err != nil {
			t.Fatalf("tool.Info returned error: %v", err)
		}
		if info.Name != "GetStockChipDistribution" {
			continue
		}
		count++
		if !strings.Contains(info.Desc, "筹码分布") {
			t.Fatalf("expected tool description to mention 筹码分布, got %q", info.Desc)
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one GetStockChipDistribution tool, got %d", count)
	}
}
```

- [ ] **Step 2: Run the wrapper test to verify it fails**

Run:

```powershell
conda run -n test go test ./backend/agent/tools -run "TestGetAllDataTools_ContainsGetStockChipDistribution" -count=1
```

Expected:

- test failure because `GetAllDataTools()` does not expose `GetStockChipDistribution` yet

- [ ] **Step 3: Add the wrapper entry that reuses the shared backend section builder**

```go
tools = append(tools, NewDataToolWrapper(
	"GetStockChipDistribution",
	"获取股票筹码分布。基于东方财富日 K 线与换手率在本地计算最新筹码结构，返回获利比例、平均成本、70%/90%成本区间和价格分布明细表。",
	map[string]*schema.ParameterInfo{
		"stockCode": {
			Type:     "string",
			Desc:     "股票代码。A股如 000001.SZ、600000.SH；港股如 00700.HK。多只时可用英文逗号分隔。",
			Required: true,
		},
		"stockCodes": {
			Type:     "array",
			Desc:     "可选，多只股票代码列表。",
			Required: false,
		},
		"days": {
			Type:     "integer",
			Desc:     "采样日 K 条数，默认 240。",
			Required: false,
		},
		"adjustFlag": {
			Type:     "string",
			Desc:     "复权类型：空值默认 qfq，也支持 qfq、hfq。",
			Required: false,
		},
		"accuracyFactor": {
			Type:     "integer",
			Desc:     "价格桶精度，默认 150，必须大于 0。",
			Required: false,
		},
	},
	func(args string) (string, error) {
		days := int(gjson.Get(args, "days").Int())
		if days <= 0 {
			days = 240
		}
		adjustFlag := gjson.Get(args, "adjustFlag").String()
		accuracyFactor := 150
		if accuracy := gjson.Get(args, "accuracyFactor"); accuracy.Exists() {
			accuracyFactor = int(accuracy.Int())
			if accuracyFactor <= 0 {
				return "参数 accuracyFactor 必须大于 0", nil
			}
		}

		codes := parseStockCodesFromArgs(args, "stockCode")
		if len(codes) == 0 {
			return "参数 stockCode 或 stockCodes 不能为空", nil
		}

		seen := make(map[string]struct{}, len(codes))
		results := make([]string, 0, len(codes))
		for _, code := range codes {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			api := data.NewEastMoneyKLineApi(data.GetSettingConfig())
			results = append(results, data.ChipDistributionSection(api, code, adjustFlag, days, accuracyFactor))
		}
		return strings.Join(results, "\n"), nil
	},
))
```

- [ ] **Step 4: Run the wrapper and duplicate-tool tests to verify they pass**

Run:

```powershell
conda run -n test go test ./backend/agent/tools ./backend/agent -run "TestGetAllDataTools_ContainsGetStockChipDistribution|TestGetAllDataTools|TestGetAllTools|TestAgentGoTools" -count=1
```

Expected:

- command exits `0`
- the new wrapper test passes
- existing duplicate-tool coverage still passes

- [ ] **Step 5: Review the wrapper diff**

```powershell
git diff -- backend/agent/tools/data_tools_wrapper.go backend/agent/tools/data_tools_wrapper_chip_distribution_test.go
```

### Task 4: Format files and run end-to-end verification for the touched packages

**Files:**
- Modify: `D:\codex_work\go-stock\backend\data\chip_distribution_calculator.go`
- Modify: `D:\codex_work\go-stock\backend\data\chip_distribution_calculator_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_stock_chip_distribution.go`
- Modify: `D:\codex_work\go-stock\backend\data\tool_stock_chip_distribution_test.go`
- Modify: `D:\codex_work\go-stock\backend\data\tools.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper.go`
- Modify: `D:\codex_work\go-stock\backend\agent\tools\data_tools_wrapper_chip_distribution_test.go`

- [ ] **Step 1: Format the touched Go files**

Run:

```powershell
conda run -n test gofmt -w `
  backend/data/chip_distribution_calculator.go `
  backend/data/chip_distribution_calculator_test.go `
  backend/data/tool_stock_chip_distribution.go `
  backend/data/tool_stock_chip_distribution_test.go `
  backend/data/tools.go `
  backend/agent/tools/data_tools_wrapper.go `
  backend/agent/tools/data_tools_wrapper_chip_distribution_test.go
```

Expected:

- command exits `0`
- no formatting errors

- [ ] **Step 2: Re-run the focused backend tests**

Run:

```powershell
conda run -n test go test ./backend/data -run "TestCalculateChipDistribution_|TestTools_ContainsGetStockChipDistribution|TestChipDistributionSection_" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 3: Re-run the focused agent-wrapper tests**

Run:

```powershell
conda run -n test go test ./backend/agent/tools ./backend/agent -run "TestGetAllDataTools_ContainsGetStockChipDistribution|TestGetAllDataTools|TestGetAllTools|TestAgentGoTools" -count=1
```

Expected:

- command exits `0`

- [ ] **Step 4: Run a broader compile smoke test for the touched backend package set**

Run:

```powershell
conda run -n test go test ./backend/data ./backend/agent ./backend/agent/tools ./backend/util ./backend/db ./backend/logger -run "^$" -count=1
```

Expected:

- command exits `0`
- all touched packages compile cleanly together

- [ ] **Step 5: Review the final diff and summarize the verification evidence**

Run:

```powershell
git diff -- backend/data/chip_distribution_calculator.go backend/data/chip_distribution_calculator_test.go backend/data/tool_stock_chip_distribution.go backend/data/tool_stock_chip_distribution_test.go backend/data/tools.go backend/agent/tools/data_tools_wrapper.go backend/agent/tools/data_tools_wrapper_chip_distribution_test.go
```

Expected summary:

- backend data layer contains one shared calculator and one tool handler
- OpenAI schema path contains `GetStockChipDistribution`
- agent wrapper path contains exactly one `GetStockChipDistribution`
- focused and broader verification commands have both passed
