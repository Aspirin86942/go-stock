package data

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func init() {
	registerToolHandler("GetStockChipDistribution", handleGetStockChipDistribution)
}

func normalizeChipAdjustFlag(adjustFlag string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(adjustFlag))
	if normalized == "" {
		return "qfq", nil
	}
	if normalized == "qfq" || normalized == "hfq" {
		return normalized, nil
	}
	return "", fmt.Errorf("参数 adjustFlag 仅支持空值、qfq、hfq")
}

func parseChipAccuracyFactor(root gjson.Result) (int, error) {
	value := root.Get("accuracyFactor")
	if !value.Exists() {
		return 150, nil
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return 0, fmt.Errorf("参数 accuracyFactor 必须为正整数")
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("参数 accuracyFactor 必须为正整数")
	}
	return parsed, nil
}

type ChipDistributionToolArgs struct {
	StockCodes     []string
	Days           int
	AdjustFlag     string
	AccuracyFactor int
}

// ParseChipDistributionToolArgs 统一解析筹码分布工具参数，保证 data tool 和 agent wrapper 使用同一套契约。
func ParseChipDistributionToolArgs(funcArguments string) (*ChipDistributionToolArgs, error) {
	codes := parseStockCodesFromToolArgs(funcArguments, "stockCode")
	if len(codes) == 0 {
		return nil, fmt.Errorf("参数 stockCode 或 stockCodes 不能为空，请传入股票代码（多只可用英文逗号分隔）。")
	}

	root := gjson.Parse(funcArguments)
	days := int(root.Get("days").Int())
	if days <= 0 {
		days = 240
	}
	adjustFlag, err := normalizeChipAdjustFlag(root.Get("adjustFlag").String())
	if err != nil {
		return nil, err
	}
	accuracyFactor, err := parseChipAccuracyFactor(root)
	if err != nil {
		return nil, err
	}

	return &ChipDistributionToolArgs{
		StockCodes:     codes,
		Days:           days,
		AdjustFlag:     adjustFlag,
		AccuracyFactor: accuracyFactor,
	}, nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isSupportedChipStockCode(stockCode string) bool {
	stockCode = strings.TrimSpace(stockCode)
	if stockCode == "" {
		return false
	}
	upper := strings.ToUpper(stockCode)
	if strings.Contains(upper, ".") {
		parts := strings.Split(upper, ".")
		if len(parts) != 2 {
			return false
		}
		code, market := parts[0], parts[1]
		if !allDigits(code) {
			return false
		}
		switch market {
		case "SH", "SZ", "BJ", "SS":
			return len(code) == 6
		case "HK":
			return len(code) == 5
		default:
			return false
		}
	}
	lower := strings.ToLower(stockCode)
	switch {
	case strings.HasPrefix(lower, "sh"), strings.HasPrefix(lower, "sz"), strings.HasPrefix(lower, "bj"):
		return len(lower) == 8 && allDigits(lower[2:])
	case strings.HasPrefix(lower, "hk"):
		return len(lower) == 7 && allDigits(lower[2:])
	default:
		return allDigits(lower) && len(lower) == 6
	}
}

func ChipDistributionSection(api *EastMoneyKLineApi, stockCode, adjustFlag string, days, accuracyFactor int) string {
	stockCode = strings.TrimSpace(stockCode)
	if !isSupportedChipStockCode(stockCode) || !api.ValidateStockCode(stockCode) {
		return stockCode + "：股票代码无效，请使用正确格式（如 000001.SZ、600000.SH、00700.HK）。"
	}
	if days <= 0 {
		days = 240
	}
	if accuracyFactor <= 0 {
		return stockCode + "：参数 accuracyFactor 必须为正整数。"
	}

	normalizedAdjust, err := normalizeChipAdjustFlag(adjustFlag)
	if err != nil {
		return stockCode + "：" + err.Error()
	}

	var list *[]KLineData
	if normalizedAdjust == "qfq" || normalizedAdjust == "hfq" {
		list = api.GetAdjustedKLine(stockCode, normalizedAdjust, days)
	} else {
		list = api.GetKLineData(stockCode, "101", strings.TrimSpace(normalizedAdjust), days)
	}
	if list == nil || len(*list) == 0 {
		return stockCode + "：未获取到 K 线数据，请检查股票代码与参数。"
	}

	result, err := CalculateChipDistribution(*list, accuracyFactor)
	if err != nil {
		return stockCode + "：筹码分布计算失败：" + err.Error()
	}

	rows := make([]map[string]any, 0, len(result.Distribution))
	for _, bucket := range result.Distribution {
		rows = append(rows, map[string]any{
			"价格":      fmt.Sprintf("%.2f", bucket.Price),
			"筹码权重(%)": fmt.Sprintf("%.4f", bucket.Weight*100),
		})
	}
	jsonData, _ := json.Marshal(rows)
	markdownTable, mdErr := JSONToMarkdownTable(jsonData)
	if mdErr != nil {
		markdownTable = string(jsonData)
	}

	return fmt.Sprintf(
		"\r\n### %s 筹码分布\r\n- 最新日期：%s\r\n- 样本天数：%d\r\n- 获利比例：%.2f%%\r\n- 平均成本：%.2f\r\n- 90%%成本区间：%.2f ~ %.2f\r\n- 90%%集中度：%.2f%%\r\n- 70%%成本区间：%.2f ~ %.2f\r\n- 70%%集中度：%.2f%%\r\n\r\n%s\r\n",
		stockCode,
		result.LatestDate,
		result.SampleDays,
		result.BenefitPart,
		result.AvgCost,
		result.Cost90.LowerPrice,
		result.Cost90.UpperPrice,
		result.Cost90.Concentration,
		result.Cost70.LowerPrice,
		result.Cost70.UpperPrice,
		result.Cost70.Concentration,
		markdownTable,
	)
}

func handleGetStockChipDistribution(o *OpenAi, funcArguments string, ctx *ToolContext) error {
	req, err := ParseChipDistributionToolArgs(funcArguments)
	if err != nil {
		appendToolMessages(
			ctx.Messages,
			ctx.CurrentAIContent.String(),
			ctx.ReasoningContentText.String(),
			ctx.CurrentCallID,
			ctx.FuncName,
			funcArguments,
			err.Error(),
		)
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

	content := parallelStockToolSections(req.StockCodes, func(stockCode string) string {
		api := NewEastMoneyKLineApi(GetSettingConfig())
		return ChipDistributionSection(api, stockCode, req.AdjustFlag, req.Days, req.AccuracyFactor)
	})

	appendToolMessages(
		ctx.Messages,
		ctx.CurrentAIContent.String(),
		ctx.ReasoningContentText.String(),
		ctx.CurrentCallID,
		ctx.FuncName,
		funcArguments,
		content,
	)
	return nil
}
