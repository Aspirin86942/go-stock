package tools

import (
	"context"
	"strings"
	"testing"
)

func findStockChipDistributionTool(t *testing.T) *DataToolWrapper {
	t.Helper()
	allTools := GetAllDataTools()
	count := 0
	var target *DataToolWrapper
	for _, baseTool := range allTools {
		info, err := baseTool.Info(nil)
		if err != nil {
			t.Fatalf("tool.Info returned error: %v", err)
		}
		if info.Name != "GetStockChipDistribution" {
			continue
		}
		count++
		wrapper, ok := baseTool.(*DataToolWrapper)
		if !ok {
			t.Fatalf("GetStockChipDistribution type mismatch: %T", baseTool)
		}
		target = wrapper
	}
	if count != 1 {
		t.Fatalf("expected exactly one GetStockChipDistribution tool, got %d", count)
	}
	return target
}

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

func TestGetStockChipDistribution_InvalidAccuracyFactor_ReturnsReadableMessage(t *testing.T) {
	tool := findStockChipDistributionTool(t)
	out, err := tool.InvokableRun(context.Background(), `{"stockCode":"600000.SH","accuracyFactor":0}`)
	if err != nil {
		t.Fatalf("expected nil error for validation message, got %v", err)
	}
	if !strings.Contains(out, "accuracyFactor") {
		t.Fatalf("expected output to contain accuracyFactor, got %q", out)
	}
}

func TestGetStockChipDistribution_StockCodesOnly_IsAccepted(t *testing.T) {
	tool := findStockChipDistributionTool(t)
	stockCodeParam, ok := tool.params["stockCode"]
	if !ok {
		t.Fatalf("stockCode param not found in tool schema")
	}
	if stockCodeParam.Required {
		t.Fatalf("stockCode should not be required when stockCodes is supported")
	}

	codes := parseStockCodesFromArgs(`{"stockCodes":["BAD_CODE"]}`, "stockCode")
	if len(codes) != 1 {
		t.Fatalf("expected one code from stockCodes-only input, got %d", len(codes))
	}
	if codes[0] != "BAD_CODE" {
		t.Fatalf("expected code BAD_CODE from stockCodes-only input, got %q", codes[0])
	}
}
