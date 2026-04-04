package data

import (
	"fmt"
	"strings"
	"testing"
)

func newStubChipEastMoneyKLineAPI(t *testing.T, klines []string) *EastMoneyKLineApi {
	t.Helper()
	api := NewEastMoneyKLineApi(&SettingConfig{Settings: &Settings{CrawlTimeOut: 5}})
	response := fmt.Sprintf(`{"rc":0,"code":0,"data":{"code":"600519","name":"贵州茅台","klines":["%s"]}}`, strings.Join(klines, `","`))
	api.fetchHTTP = func(reqURL string, cookieHeader string) ([]byte, error) {
		return []byte(response), nil
	}
	api.cookieHeaderProvider = nil
	return api
}

func TestTools_ContainsGetStockChipDistribution(t *testing.T) {
	tools := Tools(nil)
	for _, tool := range tools {
		if tool.Function.Name != "GetStockChipDistribution" {
			continue
		}
		if tool.Function.Parameters == nil {
			t.Fatalf("GetStockChipDistribution parameters should not be nil")
		}
		if _, ok := tool.Function.Parameters.Properties["accuracyFactor"]; !ok {
			t.Fatalf("GetStockChipDistribution should contain accuracyFactor parameter")
		}
		if tool.Function.Parameters.Properties["days"].(map[string]any)["type"] != "integer" {
			t.Fatalf("days should be declared as integer")
		}
		if tool.Function.Parameters.Properties["accuracyFactor"].(map[string]any)["type"] != "integer" {
			t.Fatalf("accuracyFactor should be declared as integer")
		}
		for _, required := range tool.Function.Parameters.Required {
			if required == "stockCode" {
				t.Fatalf("stockCode should not be required when stockCodes is supported")
			}
		}
		return
	}
	t.Fatalf("Tools should contain GetStockChipDistribution")
}

func TestParseChipDistributionToolArgs_StockCodesOnlyAndDefaults(t *testing.T) {
	req, err := ParseChipDistributionToolArgs(`{"stockCodes":["600000.SH","600000.SH","000001.SZ"]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(req.StockCodes) != 2 {
		t.Fatalf("expected deduped stock codes, got %+v", req.StockCodes)
	}
	if req.StockCodes[0] != "600000.SH" || req.StockCodes[1] != "000001.SZ" {
		t.Fatalf("unexpected stock code order: %+v", req.StockCodes)
	}
	if req.Days != 240 {
		t.Fatalf("expected default days=240, got %d", req.Days)
	}
	if req.AdjustFlag != "qfq" {
		t.Fatalf("expected default adjustFlag=qfq, got %q", req.AdjustFlag)
	}
	if req.AccuracyFactor != 150 {
		t.Fatalf("expected default accuracyFactor=150, got %d", req.AccuracyFactor)
	}
}

func TestParseChipDistributionToolArgs_InvalidAccuracyFactor(t *testing.T) {
	_, err := ParseChipDistributionToolArgs(`{"stockCode":"600000.SH","accuracyFactor":0}`)
	if err == nil || !strings.Contains(err.Error(), "accuracyFactor") {
		t.Fatalf("expected accuracyFactor validation error, got %v", err)
	}
}

func TestHandleGetStockChipDistribution_InvalidAccuracyFactor_AppendsToolMessage(t *testing.T) {
	messages := []map[string]any{}
	err := handleGetStockChipDistribution(nil, `{"stockCode":"600000.SH","accuracyFactor":0}`, &ToolContext{
		Messages:             &messages,
		CurrentAIContent:     &strings.Builder{},
		ReasoningContentText: &strings.Builder{},
		CurrentCallID:        "call_1",
		FuncName:             "GetStockChipDistribution",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected assistant/tool messages to be appended, got %d", len(messages))
	}
	content, _ := messages[1]["content"].(string)
	if !strings.Contains(content, "accuracyFactor") {
		t.Fatalf("expected tool message to contain accuracyFactor error, got %q", content)
	}
}

func TestChipDistributionSection_IncludesSummaryAndDetailTable(t *testing.T) {
	api := newStubChipEastMoneyKLineAPI(t, []string{
		"2026-04-01,1500.00,1510.00,1520.00,1490.00,100000,150000000,2.00,0.66,10.00,2.50",
		"2026-04-02,1510.00,1520.00,1535.00,1505.00,110000,167000000,1.98,0.66,10.00,2.80",
		"2026-04-03,1520.00,1530.00,1540.00,1510.00,120000,183000000,1.97,0.66,10.00,3.10",
	})

	section := ChipDistributionSection(api, "600519.SH", "qfq", 240, 150)
	requiredParts := []string{
		"样本天数",
		"获利比例",
		"平均成本",
		"90%成本区间",
		"90%集中度",
		"70%成本区间",
		"70%集中度",
		"价格",
		"筹码权重(%)",
		"2026-04-03",
	}
	for _, part := range requiredParts {
		if !strings.Contains(section, part) {
			t.Fatalf("section should contain %q, got: %s", part, section)
		}
	}
	if strings.Index(section, "获利比例") > strings.Index(section, "筹码权重(%)") {
		t.Fatalf("summary should be before detail table, got: %s", section)
	}
}

func TestChipDistributionSection_RejectsInvalidCode(t *testing.T) {
	api := newStubChipEastMoneyKLineAPI(t, []string{
		"2026-04-03,10.00,10.20,10.30,9.90,1000,10000,2.00,1.00,0.10,2.10",
	})
	section := ChipDistributionSection(api, "invalid-code", "", 240, 150)
	if !strings.Contains(section, "股票代码无效") {
		t.Fatalf("expected invalid code message, got: %s", section)
	}
}

func TestChipDistributionSection_RejectsBareFiveDigitCode(t *testing.T) {
	api := newStubChipEastMoneyKLineAPI(t, []string{
		"2026-04-03,10.00,10.20,10.30,9.90,1000,10000,2.00,1.00,0.10,2.10",
	})
	section := ChipDistributionSection(api, "00700", "qfq", 240, 150)
	if !strings.Contains(section, "股票代码无效") {
		t.Fatalf("expected bare 5-digit code to be invalid, got: %s", section)
	}
}
