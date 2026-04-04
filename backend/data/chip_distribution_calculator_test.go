package data

import (
	"math"
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
	sum := 0.0
	for _, bucket := range result.Distribution {
		sum += bucket.Weight
	}
	if math.Abs(sum-1) > 1e-6 {
		t.Fatalf("expected distribution weights sum to 1, got %f", sum)
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

func TestCalculateChipDistribution_InvalidFieldValues(t *testing.T) {
	_, err := CalculateChipDistribution([]KLineData{
		{Day: "2026-04-04", Open: "10.45", Close: "10.62", High: "10.30", Low: "10.45", TurnoverRate: "4.50"},
	}, 50)
	if err == nil || !strings.Contains(err.Error(), "High 必须大于 Low") {
		t.Fatalf("expected high/low validation error, got %v", err)
	}

	_, err = CalculateChipDistribution([]KLineData{
		{Day: "2026-04-04", Open: "10.45", Close: "10.62", High: "10.80", Low: "10.30", TurnoverRate: "-1"},
	}, 50)
	if err == nil || !strings.Contains(err.Error(), "TurnoverRate") {
		t.Fatalf("expected turnover rate validation error, got %v", err)
	}
}

func TestCalculateChipDistribution_FlatDay(t *testing.T) {
	result, err := CalculateChipDistribution([]KLineData{
		{Day: "2026-04-01", Open: "10.10", Close: "10.10", High: "10.10", Low: "10.10", TurnoverRate: "2.50"},
	}, 20)
	if err != nil {
		t.Fatalf("expected flat day to succeed, got %v", err)
	}
	nonZero := 0
	for _, bucket := range result.Distribution {
		if bucket.Weight > 1e-6 {
			nonZero++
		}
	}
	if nonZero != 1 {
		t.Fatalf("expected only one bucket for flat day, got %d", nonZero)
	}
}

func TestCalculateChipDistribution_SortsByDate(t *testing.T) {
	result, err := CalculateChipDistribution([]KLineData{
		{Day: "2026-04-03", Open: "10.45", Close: "10.62", High: "10.80", Low: "10.30", TurnoverRate: "4.50"},
		{Day: "2026-04-02", Open: "10.32", Close: "10.48", High: "10.55", Low: "10.20", TurnoverRate: "4.10"},
	}, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LatestDate != "2026-04-03" {
		t.Fatalf("expected latest date 2026-04-03 but got %q", result.LatestDate)
	}
	if result.LatestClose != 10.62 {
		t.Fatalf("expected latest close 10.62 but got %f", result.LatestClose)
	}
}

func TestCalculateChipDistribution_TurnoverFullReplacement(t *testing.T) {
	result, err := CalculateChipDistribution([]KLineData{
		{Day: "2026-04-01", Open: "10.45", Close: "10.62", High: "10.80", Low: "10.30", TurnoverRate: "5.00"},
		{Day: "2026-04-02", Open: "12.50", Close: "13.00", High: "13.00", Low: "12.40", TurnoverRate: "100"},
	}, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(result.BenefitPart-100) > 1e-6 {
		t.Fatalf("expected benefit part 100 after full turnover, got %f", result.BenefitPart)
	}
}
