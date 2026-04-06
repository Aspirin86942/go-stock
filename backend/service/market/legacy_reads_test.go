package market

import (
	"testing"

	"go-stock/backend/models"
)

type residualSourceStub struct{}

func (s *residualSourceStub) GetTelegraphList(source string) *[]*models.Telegraph { return nil }
func (s *residualSourceStub) RefreshFeeds()                                       {}
func (s *residualSourceStub) GlobalStockIndexes(crawlTimeout uint) map[string]any {
	return map[string]any{}
}
func (s *residualSourceStub) GetIndustryRank(sort string, cnt int) map[string]any {
	return map[string]any{}
}
func (s *residualSourceStub) GlobalStockIndexesReadable(crawlTimeout uint) string {
	return "  亚洲市场：上证指数 +1.23%\n"
}
func (s *residualSourceStub) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return []map[string]any{{
		"category":        "board-1",
		"name":            "机器人",
		"avg_changeratio": 0.0312,
		"inamount":        2500000.0,
		"outamount":       1250000.0,
		"netamount":       1250000.0,
		"ratioamount":     0.22,
		"ts_name":         "机器人龙头",
		"ts_symbol":       "300024",
		"ts_changeratio":  0.0415,
		"ts_trade":        15.66,
		"ts_ratioamount":  0.19,
	}}
}
func (s *residualSourceStub) GetMoneyRankSina(sort string) []map[string]any {
	return []map[string]any{{
		"symbol":      "600519",
		"name":        "贵州茅台",
		"trade":       1688.88,
		"changeratio": 0.015,
		"turnover":    0.031,
		"amount":      5530000.0,
		"outamount":   1650000.0,
		"inamount":    3880000.0,
		"netamount":   2230000.0,
		"ratioamount": 0.015,
		"r0_out":      730000.0,
		"r0_in":       910000.0,
		"r0_net":      180000.0,
		"r0_ratio":    0.07,
		"r3_out":      300000.0,
		"r3_in":       210000.0,
		"r3_net":      -90000.0,
		"r3_ratio":    -0.03,
	}}
}
func (s *residualSourceStub) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	return []map[string]any{
		{"opendate": "2026-04-02", "trade": 18.90, "netamount": 260000.0, "r0_net": 135000.0},
		{"opendate": "2026-04-01", "trade": 18.32, "netamount": 250000.0, "r0_net": 120000.0},
	}
}

func TestService_LoadResidualMarketReads_NormalizesTypedContracts(t *testing.T) {
	svc := NewService(&residualSourceStub{})

	if got := svc.LoadGlobalIndexesReadable(30); got != "亚洲市场：上证指数 +1.23%" {
		t.Fatalf("unexpected readable indexes: %q", got)
	}

	industry := svc.LoadIndustryMoneyRanks("0", "netamount")
	if len(industry) != 1 || industry[0].Name != "机器人" || industry[0].TSSymbol != "300024" || industry[0].NetAmount != 1250000 {
		t.Fatalf("unexpected industry money rank: %#v", industry)
	}

	moneyRanks := svc.LoadMoneyRanks("netamount")
	if len(moneyRanks) != 1 || moneyRanks[0].Symbol != "600519" || moneyRanks[0].R0Net != 180000 {
		t.Fatalf("unexpected money rank data: %#v", moneyRanks)
	}

	trend := svc.LoadStockMoneyTrend("600519", 20)
	if len(trend) != 2 || trend[0].OpenDate != "2026-04-01" || trend[1].OpenDate != "2026-04-02" {
		t.Fatalf("unexpected stock money trend order: %#v", trend)
	}
}
