package market

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

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

type StockResearchReportEntry struct {
	InfoCode     string `json:"infoCode"`
	StockCode    string `json:"stockCode"`
	StockName    string `json:"stockName"`
	Market       string `json:"market"`
	IndvInduName string `json:"indvInduName"`
	Title        string `json:"title"`
	EmRatingName string `json:"emRatingName"`
	RatingChange int    `json:"ratingChange"`
	SRatingName  string `json:"sRatingName"`
	Researcher   string `json:"researcher"`
	OrgSName     string `json:"orgSName"`
	PublishDate  string `json:"publishDate"`
}

type IndustryResearchReportEntry struct {
	InfoCode     string `json:"infoCode"`
	IndustryName string `json:"industryName"`
	Title        string `json:"title"`
	EmRatingName string `json:"emRatingName"`
	RatingChange int    `json:"ratingChange"`
	SRatingName  string `json:"sRatingName"`
	Researcher   string `json:"researcher"`
	OrgSName     string `json:"orgSName"`
	PublishDate  string `json:"publishDate"`
}

type StockNoticeEntry struct {
	ArtCode     string              `json:"art_code"`
	Title       string              `json:"title"`
	NoticeDate  string              `json:"notice_date"`
	DisplayTime string              `json:"display_time"`
	Codes       []StockNoticeCode   `json:"codes"`
	Columns     []StockNoticeColumn `json:"columns"`
}

type StockNoticeCode struct {
	StockCode  string `json:"stock_code"`
	ShortName  string `json:"short_name"`
	MarketCode string `json:"market_code"`
}

type StockNoticeColumn struct {
	ColumnName string `json:"column_name"`
}

type EMDictCodeEntry struct {
	BKCode      string `json:"bkCode"`
	BKName      string `json:"bkName"`
	FirstLetter string `json:"firstLetter"`
}

type HotTopicEntry struct {
	Nickname    string          `json:"nickname"`
	Desc        string          `json:"desc"`
	SquareImg   string          `json:"squareImg"`
	StockList   []HotTopicStock `json:"stock_list"`
	ClickNumber int             `json:"clickNumber"`
	PostNumber  int             `json:"postNumber"`
	HTID        string          `json:"htid"`
}

type HotTopicStock struct {
	Name string `json:"name"`
}

type InvestCalendarDay struct {
	Date string               `json:"date"`
	List []InvestCalendarItem `json:"list"`
}

type InvestCalendarItem struct {
	ArticleID string `json:"article_id"`
	Title     string `json:"title"`
	LikeCount int    `json:"like_count"`
}

type ClsCalendarDay struct {
	CalendarDay string            `json:"calendar_day"`
	Week        string            `json:"week"`
	Items       []ClsCalendarItem `json:"items"`
}

type ClsCalendarItem struct {
	ID       string               `json:"id"`
	Title    string               `json:"title"`
	Event    *ClsCalendarEvent    `json:"event"`
	Economic *ClsCalendarEconomic `json:"economic"`
}

type ClsCalendarEvent struct {
	Star int `json:"star"`
}

type ClsCalendarEconomic struct {
	Star      int    `json:"star"`
	Actual    string `json:"actual"`
	Consensus string `json:"consensus"`
	Front     string `json:"front"`
}

type SearchStockResponse struct {
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Message string          `json:"message"`
	Data    SearchStockData `json:"data"`
}

type SearchStockData struct {
	TraceInfo SearchStockTraceInfo `json:"traceInfo"`
	Result    SearchStockResultSet `json:"result"`
}

type SearchStockTraceInfo struct {
	ShowText string `json:"showText"`
}

type SearchStockResultSet struct {
	Columns  []SearchStockColumn `json:"columns"`
	DataList []map[string]any    `json:"dataList"`
}

type SearchStockColumn struct {
	Title      string              `json:"title"`
	Key        string              `json:"key"`
	Unit       string              `json:"unit"`
	HiddenNeed bool                `json:"hiddenNeed"`
	DateMsg    string              `json:"dateMsg"`
	Children   []SearchStockColumn `json:"children"`
}

type MinutePriceLine struct {
	StockCode string            `json:"stockCode"`
	StockName string            `json:"stockName"`
	Date      string            `json:"date"`
	PriceData []data.MinuteData `json:"priceData"`
}

type EastMoneyKLinePageResult struct {
	OK              bool             `json:"ok"`
	Data            []data.KLineData `json:"data"`
	Message         string           `json:"message"`
	ErrorCode       string           `json:"errorCode"`
	UsedCookieRetry bool             `json:"usedCookieRetry"`
}

type RealtimePrice struct {
	StockCode string `json:"stockCode"`
	StockName string `json:"stockName"`
	Price     string `json:"price"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	PreClose  string `json:"preClose"`
	Date      string `json:"date"`
	Time      string `json:"time"`
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
