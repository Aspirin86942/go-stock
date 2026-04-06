package market

import (
	"encoding/json"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"slices"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
)

type Service struct {
	source Source
}

type legacyMarketSource interface {
	LongTiger(date string) *[]models.LongTigerRankData
	StockResearchReport(stockCode string, days int) []any
	StockNotice(stockCode string) []any
	IndustryResearchReport(industryCode string, days int) []any
	EMDictCode(code string) []any
	XueQiuHotStock(size int, marketType string) *[]models.HotItem
	HotEvent(size int) *[]models.HotEvent
	HotTopic(size int) []any
	InvestCalendar(yearMonth string) []any
	ClsCalendar() []any
	SearchStock(words string, pageSize int) map[string]any
	HotStrategy() map[string]any
	GetStockKLine(stockCode string, days int64) *[]data.KLineData
	GetStockCommonKLine(stockCode string, days int64) *[]data.KLineData
	GetStockMinutePriceData(stockCode string) (*[]data.MinuteData, string)
	GetStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) *[]data.KLineData
	GetStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) map[string]any
	GetStockRealtimePrice(stockCode string) *data.StockInfo
}

func NewService(source Source) *Service {
	return &Service{source: source}
}

func (s *Service) LoadFeeds() FeedSet {
	telegraph := s.loadFeed("财联社电报")
	sina := s.loadFeed("新浪财经")
	foreign := s.loadFeed("外媒")

	return FeedSet{
		Telegraph: telegraph.Items,
		Sina:      sina.Items,
		Foreign:   foreign.Items,
	}
}

func (s *Service) RefreshFeed(source string) Feed {
	s.source.RefreshFeeds()
	return s.loadFeed(source)
}

func (s *Service) RefreshAllFeeds() FeedSet {
	s.source.RefreshFeeds()
	return s.LoadFeeds()
}

func (s *Service) LoadGlobalIndexes(crawlTimeOut uint) IndexSet {
	raw := s.source.GlobalStockIndexes(crawlTimeOut)
	return IndexSet{
		Common:  mapGlobalIndexesByRegion(raw, "common"),
		America: mapGlobalIndexesByRegion(raw, "america"),
		Europe:  mapGlobalIndexesByRegion(raw, "europe"),
		Asia:    mapGlobalIndexesByRegion(raw, "asia"),
		Other:   mapGlobalIndexesByRegion(raw, "other"),
	}
}

func (s *Service) LoadIndustryRanks(sort string, cnt int) []IndustryRankEntry {
	raw := s.source.GetIndustryRank(sort, cnt)
	rows, ok := raw["data"].([]any)
	if !ok || len(rows) == 0 {
		return []IndustryRankEntry{}
	}

	result := make([]IndustryRankEntry, 0, len(rows))
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, IndustryRankEntry{
			BoardCode:             convertor.ToString(row["bd_code"]),
			BoardName:             convertor.ToString(row["bd_name"]),
			BoardChangePercent:    convertor.ToString(row["bd_zdf"]),
			BoardChangePercent5D:  convertor.ToString(row["bd_zdf5"]),
			BoardChangePercent20D: convertor.ToString(row["bd_zdf20"]),
			LeaderCode:            convertor.ToString(row["nzg_code"]),
			LeaderName:            convertor.ToString(row["nzg_name"]),
			LeaderChangePercent:   convertor.ToString(row["nzg_zdf"]),
			LeaderPrice:           convertor.ToString(row["nzg_zxj"]),
		})
	}
	return result
}

func (s *Service) LoadGlobalIndexesReadable(crawlTimeout uint) string {
	return strings.TrimSpace(s.source.GlobalStockIndexesReadable(crawlTimeout))
}

func (s *Service) LoadIndustryMoneyRanks(fenlei, sort string) []IndustryMoneyRankRow {
	raw := s.source.GetIndustryMoneyRankSina(fenlei, sort)
	if len(raw) == 0 {
		return []IndustryMoneyRankRow{}
	}

	result := make([]IndustryMoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, IndustryMoneyRankRow{
			Category:       convertor.ToString(row["category"]),
			Name:           convertor.ToString(row["name"]),
			AvgChangeRatio: toFloat(row["avg_changeratio"]),
			InAmount:       toFloat(row["inamount"]),
			OutAmount:      toFloat(row["outamount"]),
			NetAmount:      toFloat(row["netamount"]),
			RatioAmount:    toFloat(row["ratioamount"]),
			TSName:         convertor.ToString(row["ts_name"]),
			TSSymbol:       convertor.ToString(row["ts_symbol"]),
			TSChangeRatio:  toFloat(row["ts_changeratio"]),
			TSTrade:        toFloat(row["ts_trade"]),
			TSRatioAmount:  toFloat(row["ts_ratioamount"]),
		})
	}
	return result
}

func (s *Service) LoadMoneyRanks(sort string) []MoneyRankRow {
	raw := s.source.GetMoneyRankSina(sort)
	if len(raw) == 0 {
		return []MoneyRankRow{}
	}

	result := make([]MoneyRankRow, 0, len(raw))
	for _, row := range raw {
		result = append(result, MoneyRankRow{
			Symbol:      convertor.ToString(row["symbol"]),
			Name:        convertor.ToString(row["name"]),
			Trade:       toFloat(row["trade"]),
			ChangeRatio: toFloat(row["changeratio"]),
			Turnover:    toFloat(row["turnover"]),
			Amount:      toFloat(row["amount"]),
			OutAmount:   toFloat(row["outamount"]),
			InAmount:    toFloat(row["inamount"]),
			NetAmount:   toFloat(row["netamount"]),
			RatioAmount: toFloat(row["ratioamount"]),
			R0Out:       toFloat(row["r0_out"]),
			R0In:        toFloat(row["r0_in"]),
			R0Net:       toFloat(row["r0_net"]),
			R0Ratio:     toFloat(row["r0_ratio"]),
			R3Out:       toFloat(row["r3_out"]),
			R3In:        toFloat(row["r3_in"]),
			R3Net:       toFloat(row["r3_net"]),
			R3Ratio:     toFloat(row["r3_ratio"]),
		})
	}
	return result
}

func (s *Service) LoadStockMoneyTrend(stockCode string, days int) []StockMoneyTrendRow {
	raw := s.source.GetStockMoneyTrendByDay(stockCode, days)
	if len(raw) == 0 {
		return []StockMoneyTrendRow{}
	}

	result := make([]StockMoneyTrendRow, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		row := raw[i]
		result = append(result, StockMoneyTrendRow{
			OpenDate:  convertor.ToString(row["opendate"]),
			Trade:     toFloat(row["trade"]),
			NetAmount: toFloat(row["netamount"]),
			R0Net:     toFloat(row["r0_net"]),
		})
	}
	return result
}

func (s *Service) LoadLongTiger(date string) []models.LongTigerRankData {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []models.LongTigerRankData{}
	}
	items := source.LongTiger(date)
	if items == nil {
		return []models.LongTigerRankData{}
	}
	copied := append([]models.LongTigerRankData(nil), (*items)...)
	return copied
}

func (s *Service) LoadStockResearchReports(stockCode string) []StockResearchReportEntry {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []StockResearchReportEntry{}
	}
	return decodeSlice[StockResearchReportEntry](source.StockResearchReport(stockCode, 7))
}

func (s *Service) LoadStockNotices(stockCode string) []StockNoticeEntry {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []StockNoticeEntry{}
	}
	return decodeSlice[StockNoticeEntry](source.StockNotice(stockCode))
}

func (s *Service) LoadIndustryResearchReports(industryCode string) []IndustryResearchReportEntry {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []IndustryResearchReportEntry{}
	}
	return decodeSlice[IndustryResearchReportEntry](source.IndustryResearchReport(industryCode, 7))
}

func (s *Service) LoadEMDictCodes(code string) []EMDictCodeEntry {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []EMDictCodeEntry{}
	}
	return decodeSlice[EMDictCodeEntry](source.EMDictCode(code))
}

func (s *Service) LoadHotStocks(marketType string) []models.HotItem {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []models.HotItem{}
	}
	items := source.XueQiuHotStock(100, marketType)
	if items == nil {
		return []models.HotItem{}
	}
	return append([]models.HotItem(nil), (*items)...)
}

func (s *Service) LoadHotEvents(size int) []models.HotEvent {
	if size <= 0 {
		size = 10
	}
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []models.HotEvent{}
	}
	items := source.HotEvent(size)
	if items == nil {
		return []models.HotEvent{}
	}
	return append([]models.HotEvent(nil), (*items)...)
}

func (s *Service) LoadHotTopics(size int) []HotTopicEntry {
	if size <= 0 {
		size = 10
	}
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []HotTopicEntry{}
	}
	return decodeSlice[HotTopicEntry](source.HotTopic(size))
}

func (s *Service) LoadInvestCalendar(yearMonth string) []InvestCalendarDay {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []InvestCalendarDay{}
	}
	return decodeSlice[InvestCalendarDay](source.InvestCalendar(yearMonth))
}

func (s *Service) LoadClsCalendar() []ClsCalendarDay {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []ClsCalendarDay{}
	}
	return decodeSlice[ClsCalendarDay](source.ClsCalendar())
}

func (s *Service) SearchStocks(words string) SearchStockResponse {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return SearchStockResponse{Data: SearchStockData{Result: SearchStockResultSet{Columns: []SearchStockColumn{}, DataList: []map[string]any{}}}}
	}
	resp := decodeStruct[SearchStockResponse](source.SearchStock(words, 5000))
	resp.Data.Result.Columns = normalizeSearchColumns(resp.Data.Result.Columns)
	if resp.Data.Result.Columns == nil {
		resp.Data.Result.Columns = []SearchStockColumn{}
	}
	if resp.Data.Result.DataList == nil {
		resp.Data.Result.DataList = []map[string]any{}
	}
	return resp
}

func (s *Service) LoadHotStrategies() models.HotStrategy {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return models.HotStrategy{Data: []*models.HotStrategyData{}}
	}
	resp := decodeStruct[models.HotStrategy](source.HotStrategy())
	if resp.Data == nil {
		resp.Data = []*models.HotStrategyData{}
	}
	return resp
}

func (s *Service) LoadStockList(keyword string) []data.StockBasic {
	result := data.NewStockDataApi().GetStockList(keyword)
	if result == nil {
		return []data.StockBasic{}
	}
	return append([]data.StockBasic(nil), result...)
}

func (s *Service) SaveNtfyNews(news models.NtfyNews) (*models.Telegraph, bool) {
	source := normalizeNtfySource(news.Tags)
	if source == "" {
		return nil, false
	}

	dataTime := time.UnixMilli(int64(news.Time * 1000))
	telegraph := &models.Telegraph{
		Title:           news.Title,
		Content:         news.Message,
		DataTime:        &dataTime,
		IsRed:           slices.Contains(news.Tags, "rotating_light"),
		Time:            dataTime.Format("15:04:05"),
		Source:          source,
		SentimentResult: data.AnalyzeSentiment(news.Message).Description,
	}

	cnt := int64(0)
	query := db.Dao.Model(telegraph)
	if strings.TrimSpace(telegraph.Title) == "" {
		query = query.Where("content = ?", telegraph.Content)
	} else {
		query = query.Where("title = ?", telegraph.Title)
	}
	query.Count(&cnt)
	if cnt > 0 {
		return nil, false
	}

	db.Dao.Model(telegraph).Create(telegraph)
	for _, subject := range filterNtfySubjects(news.Tags) {
		tag := &models.Tags{
			Name: subject,
			Type: "subject",
		}
		db.Dao.Model(tag).Where("name = ? and type = ?", subject, "subject").FirstOrCreate(tag)
		db.Dao.Model(&models.TelegraphTags{}).Where("telegraph_id = ? and tag_id = ?", telegraph.ID, tag.ID).FirstOrCreate(&models.TelegraphTags{
			TelegraphId: telegraph.ID,
			TagId:       tag.ID,
		})
	}
	return telegraph, true
}

func (s *Service) LoadStockKLine(stockCode string, days int64) []data.KLineData {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []data.KLineData{}
	}
	return cloneKLineData(source.GetStockKLine(stockCode, days))
}

func (s *Service) LoadStockCommonKLine(stockCode string, days int64) []data.KLineData {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []data.KLineData{}
	}
	return cloneKLineData(source.GetStockCommonKLine(stockCode, days))
}

func (s *Service) LoadStockMinutePriceLine(stockCode, stockName string) MinutePriceLine {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return MinutePriceLine{
			StockCode: stockCode,
			StockName: stockName,
			PriceData: []data.MinuteData{},
		}
	}
	priceData, date := source.GetStockMinutePriceData(stockCode)
	line := MinutePriceLine{
		StockCode: stockCode,
		StockName: stockName,
		Date:      strings.TrimSpace(date),
		PriceData: []data.MinuteData{},
	}
	if priceData == nil {
		return line
	}
	line.PriceData = append([]data.MinuteData(nil), (*priceData)...)
	return line
}

func (s *Service) LoadStockEastMoneyKLine(stockCode, klt string, limit int) []data.KLineData {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []data.KLineData{}
	}
	return cloneKLineData(source.GetStockEastMoneyKLinePage(stockCode, normalizeKlt(klt), normalizeLimit(limit), ""))
}

func (s *Service) LoadStockEastMoneyKLineResult(stockCode, klt string, limit int) EastMoneyKLinePageResult {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return EastMoneyKLinePageResult{Data: []data.KLineData{}}
	}
	return normalizeEastMoneyResult(source.GetStockEastMoneyKLinePageResult(stockCode, normalizeKlt(klt), normalizeLimit(limit), ""))
}

func (s *Service) LoadStockEastMoneyKLinePage(stockCode, klt string, limit int, end string) []data.KLineData {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return []data.KLineData{}
	}
	return cloneKLineData(source.GetStockEastMoneyKLinePage(stockCode, normalizeKlt(klt), normalizeLimit(limit), strings.TrimSpace(end)))
}

func (s *Service) LoadStockEastMoneyKLinePageResult(stockCode, klt string, limit int, end string) EastMoneyKLinePageResult {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return EastMoneyKLinePageResult{Data: []data.KLineData{}}
	}
	return normalizeEastMoneyResult(source.GetStockEastMoneyKLinePageResult(stockCode, normalizeKlt(klt), normalizeLimit(limit), strings.TrimSpace(end)))
}

func (s *Service) LoadRealtimePrice(stockCode string) RealtimePrice {
	source, ok := s.source.(legacyMarketSource)
	if !ok {
		return RealtimePrice{StockCode: stockCode}
	}
	raw := source.GetStockRealtimePrice(stockCode)
	if raw == nil {
		return RealtimePrice{
			StockCode: stockCode,
		}
	}
	return RealtimePrice{
		StockCode: firstNonEmpty(raw.Code, stockCode),
		StockName: raw.Name,
		Price: firstNonEmpty(
			raw.Price,
			raw.A1P,
			raw.B1P,
			raw.PreClose,
		),
		Bid:      raw.Bid,
		Ask:      raw.Ask,
		Open:     raw.Open,
		High:     raw.High,
		Low:      raw.Low,
		PreClose: raw.PreClose,
		Date:     raw.Date,
		Time:     raw.Time,
	}
}

func (s *Service) loadFeed(source string) Feed {
	items := s.source.GetTelegraphList(source)
	if items == nil {
		return Feed{Source: source, Items: []*models.Telegraph{}}
	}
	copied := make([]*models.Telegraph, len(*items))
	copy(copied, *items)
	return Feed{Source: source, Items: copied}
}

func mapGlobalIndexesByRegion(raw map[string]any, region string) []GlobalIndexEntry {
	rows, ok := raw[region].([]any)
	if !ok || len(rows) == 0 {
		return []GlobalIndexEntry{}
	}

	result := make([]GlobalIndexEntry, 0, len(rows))
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, GlobalIndexEntry{
			Code:     convertor.ToString(row["code"]),
			Name:     convertor.ToString(row["name"]),
			Location: convertor.ToString(row["location"]),
			Qtcode:   convertor.ToString(row["qtcode"]),
			State:    convertor.ToString(row["state"]),
			Zdf:      convertor.ToString(row["zdf"]),
			Zxj:      convertor.ToString(row["zxj"]),
			Img:      convertor.ToString(row["img"]),
			Region:   region,
		})
	}
	return result
}

func toFloat(value any) float64 {
	v, err := convertor.ToFloat(value)
	if err != nil {
		return 0
	}
	return v
}

func decodeSlice[T any](raw any) []T {
	if raw == nil {
		return []T{}
	}
	result := []T{}
	payload, err := json.Marshal(raw)
	if err != nil {
		return []T{}
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return []T{}
	}
	if result == nil {
		return []T{}
	}
	return result
}

func decodeStruct[T any](raw any) T {
	var result T
	if raw == nil {
		return result
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		return result
	}
	_ = json.Unmarshal(payload, &result)
	return result
}

func cloneKLineData(items *[]data.KLineData) []data.KLineData {
	if items == nil {
		return []data.KLineData{}
	}
	return append([]data.KLineData(nil), (*items)...)
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 500
	}
	if limit > 5000 {
		return 5000
	}
	return limit
}

func normalizeKlt(klt string) string {
	klt = strings.TrimSpace(klt)
	if klt == "" {
		return "1"
	}
	return klt
}

func normalizeEastMoneyResult(raw map[string]any) EastMoneyKLinePageResult {
	result := decodeStruct[EastMoneyKLinePageResult](raw)
	if result.Data == nil {
		result.Data = []data.KLineData{}
	}
	return result
}

func normalizeSearchColumns(columns []SearchStockColumn) []SearchStockColumn {
	if columns == nil {
		return []SearchStockColumn{}
	}
	result := make([]SearchStockColumn, 0, len(columns))
	for _, column := range columns {
		column.Children = normalizeSearchColumns(column.Children)
		result = append(result, column)
	}
	return result
}

func normalizeNtfySource(tags []string) string {
	if containsAny(tags, []string{"外媒简讯", "外媒资讯", "外媒"}) {
		return "外媒"
	}
	if slices.Contains(tags, "财联社电报") {
		return "财联社电报"
	}
	if slices.Contains(tags, "新浪财经") {
		return "新浪财经"
	}
	return ""
}

func filterNtfySubjects(tags []string) []string {
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		switch tag {
		case "rotating_light", "loudspeaker":
			continue
		default:
			result = append(result, tag)
		}
	}
	return result
}

func containsAny(tags []string, targets []string) bool {
	for _, target := range targets {
		if slices.Contains(tags, target) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
