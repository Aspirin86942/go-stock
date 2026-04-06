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
	AvgChangeRatio float64 `json:"avgChangeRatio"`
	InAmount       float64 `json:"inAmount"`
	OutAmount      float64 `json:"outAmount"`
	NetAmount      float64 `json:"netAmount"`
	RatioAmount    float64 `json:"ratioAmount"`
	TSName         string  `json:"tsName"`
	TSSymbol       string  `json:"tsSymbol"`
	TSChangeRatio  float64 `json:"tsChangeRatio"`
	TSTrade        float64 `json:"tsTrade"`
	TSRatioAmount  float64 `json:"tsRatioAmount"`
}

type MoneyRankRow struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Trade       float64 `json:"trade"`
	ChangeRatio float64 `json:"changeRatio"`
	Turnover    float64 `json:"turnover"`
	Amount      float64 `json:"amount"`
	OutAmount   float64 `json:"outAmount"`
	InAmount    float64 `json:"inAmount"`
	NetAmount   float64 `json:"netAmount"`
	RatioAmount float64 `json:"ratioAmount"`
	R0Out       float64 `json:"r0Out"`
	R0In        float64 `json:"r0In"`
	R0Net       float64 `json:"r0Net"`
	R0Ratio     float64 `json:"r0Ratio"`
	R3Out       float64 `json:"r3Out"`
	R3In        float64 `json:"r3In"`
	R3Net       float64 `json:"r3Net"`
	R3Ratio     float64 `json:"r3Ratio"`
}

type StockMoneyTrendRow struct {
	OpenDate  string  `json:"openDate"`
	Trade     float64 `json:"trade"`
	NetAmount float64 `json:"netAmount"`
	R0Net     float64 `json:"r0Net"`
}

type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeOut uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
}
