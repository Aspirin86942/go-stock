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

type Source interface {
	GetTelegraphList(source string) *[]*models.Telegraph
	RefreshFeeds()
	GlobalStockIndexes(crawlTimeOut uint) map[string]any
	GetIndustryRank(sort string, cnt int) map[string]any
}
