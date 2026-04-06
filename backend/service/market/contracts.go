package market

import "go-stock/backend/models"

type FeedSet struct {
	Telegraph []*models.Telegraph `json:"telegraph"`
	Sina      []*models.Telegraph `json:"sina"`
	Foreign   []*models.Telegraph `json:"foreign"`
}

type Feed struct {
	Source string              `json:"source"`
	Items  []*models.Telegraph `json:"items"`
}

type IndexSet struct {
	Common  []GlobalIndexEntry `json:"common"`
	America []GlobalIndexEntry `json:"america"`
	Europe  []GlobalIndexEntry `json:"europe"`
	Asia    []GlobalIndexEntry `json:"asia"`
	Other   []GlobalIndexEntry `json:"other"`
}

type GlobalIndexEntry struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Qtcode   string `json:"qtcode"`
	State    string `json:"state"`
	Zdf      string `json:"zdf"`
	Zxj      string `json:"zxj"`
	Img      string `json:"img"`
	Region   string `json:"region"`
}

type IndustryRankEntry struct {
	BoardCode             string `json:"boardCode"`
	BoardName             string `json:"boardName"`
	BoardChangePercent    string `json:"boardChangePercent"`
	BoardChangePercent5D  string `json:"boardChangePercent5D"`
	BoardChangePercent20D string `json:"boardChangePercent20D"`
	LeaderCode            string `json:"leaderCode"`
	LeaderName            string `json:"leaderName"`
	LeaderChangePercent   string `json:"leaderChangePercent"`
	LeaderPrice           string `json:"leaderPrice"`
}

type IndustryMoneyRankRow struct {
	Category       string  `json:"category"`
	Name           string  `json:"name"`
	AvgChangeRatio float64 `json:"avg_changeratio"`
	InAmount       float64 `json:"inamount"`
	OutAmount      float64 `json:"outamount"`
	NetAmount      float64 `json:"netamount"`
	RatioAmount    float64 `json:"ratioamount"`
	TSName         string  `json:"ts_name"`
	TSSymbol       string  `json:"ts_symbol"`
	TSChangeRatio  float64 `json:"ts_changeratio"`
	TSTrade        float64 `json:"ts_trade"`
	TSRatioAmount  float64 `json:"ts_ratioamount"`
}

type MoneyRankRow struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Trade       float64 `json:"trade"`
	ChangeRatio float64 `json:"changeratio"`
	Turnover    float64 `json:"turnover"`
	Amount      float64 `json:"amount"`
	OutAmount   float64 `json:"outamount"`
	InAmount    float64 `json:"inamount"`
	NetAmount   float64 `json:"netamount"`
	RatioAmount float64 `json:"ratioamount"`
	R0Out       float64 `json:"r0_out"`
	R0In        float64 `json:"r0_in"`
	R0Net       float64 `json:"r0_net"`
	R0Ratio     float64 `json:"r0_ratio"`
	R3Out       float64 `json:"r3_out"`
	R3In        float64 `json:"r3_in"`
	R3Net       float64 `json:"r3_net"`
	R3Ratio     float64 `json:"r3_ratio"`
}

type StockMoneyTrendRow struct {
	OpenDate  string  `json:"opendate"`
	Trade     float64 `json:"trade"`
	NetAmount float64 `json:"netamount"`
	R0Net     float64 `json:"r0_net"`
}

type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeOut uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
	GlobalStockIndexesReadable(crawlTimeout uint) string
	GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any
	GetMoneyRankSina(sort string) []map[string]any
	GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any
}
