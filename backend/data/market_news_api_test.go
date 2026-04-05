package data

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/models"
	"go-stock/backend/util"

	"github.com/coocood/freecache"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// @Author spark
// @Date 2025/4/23 17:58
// @Desc
//-----------------------------------------------------------------------------------

func TestGetSinaNews(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	news := NewMarketNewsApi().GetSinaNews(30)
	requirePositiveLen(t, "Sina news", len(*news))
	logValue(t, "first_sina_news", (*news)[0])
	//NewMarketNewsApi().GetNewTelegraph(30)
}

func TestGlobalStockIndexes(t *testing.T) {
	requireManualTest(t)
	resp := NewMarketNewsApi().GlobalStockIndexes(30)

	bs, _ := json.Marshal(resp)
	t.Logf("JSON:\n%s", string(bs))

	md := NewMarketNewsApi().GlobalStockIndexesToReadable(resp)

	t.Logf("resp: \n%s", md)
}

func TestGetIndustryRank(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetIndustryRank("0", 10)
	for s, a := range res["data"].([]any) {
		t.Logf("key: %+v, value: %+v", s, a)
	}
}
func TestGetIndustryMoneyRankSina(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetIndustryMoneyRankSina("0", "netamount")
	for i, re := range res {
		t.Logf("key: %+v, value: %+v", i, re)

	}
}
func TestGetMoneyRankSina(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetMoneyRankSina("r3_net")
	for i, re := range res {
		t.Logf("key: %+v, value: %+v", i, re)
	}
}

func TestGetStockMoneyTrendByDay(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetStockMoneyTrendByDay("sh600438", 360)
	for i, re := range res {
		t.Logf("key: %+v, value: %+v", i, re)
	}
}

func TestLongTiger(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")

	NewMarketNewsApi().LongTiger("2025-06-08")
}

func TestStockResearchReport(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().StockResearchReport("002046", 7)
	for _, a := range resp {
		t.Logf("value: %+v", a)
		data := a.(map[string]any)
		t.Logf("value: %s  infoCode:%s", data["title"], data["infoCode"])
		NewMarketNewsApi().GetIndustryReportInfo(data["infoCode"].(string))
	}
}

func TestIndustryResearchReport(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().IndustryResearchReport("", 7)
	for _, a := range resp {
		t.Logf("value: %+v", a)
		data := a.(map[string]any)
		t.Logf("value: %s  infoCode:%s", data["title"], data["infoCode"])
		t.Logf("url: https://pdf.dfcfw.com/pdf/H3_%s_1.pdf", data["infoCode"])
		//NewMarketNewsApi().GetIndustryReportInfo(data["infoCode"].(string))
	}
}

func TestStockNotice(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().StockNotice("600584,600900")
	requirePositiveLen(t, "stock notice rows", len(resp))
	// 转为可表格化的结构：东方财富接口返回的 list 中每项为 map，字段名多为驼峰
	type row struct {
		Title      string `md:"公告标题"`
		NoticeDate string `md:"公告日期"`
		ColumnName string `md:"公告类型"`
	}
	var rows []row
	for _, a := range resp {
		m, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if m["columns"].([]any) != nil && len(m["columns"].([]any)) > 0 {
			columns := m["columns"].([]any)[0].(map[string]any)
			rows = append(rows, row{
				Title:      convertor.ToString(m["title"]),
				NoticeDate: convertor.ToString(m["notice_date"]),
				ColumnName: convertor.ToString(columns["column_name"]),
			})
		} else {
			rows = append(rows, row{
				Title:      convertor.ToString(m["title"]),
				NoticeDate: convertor.ToString(m["notice_date"]),
			})
		}

	}
	md := util.MarkdownTableWithTitle("上市公司公告", rows)
	requireNotBlank(t, "stock notice markdown", md)
	t.Logf("stock notice markdown:\n%s", md)
}

func TestEMDictCode(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	resp := NewMarketNewsApi().EMDictCode("016", freecache.NewCache(100))
	for _, a := range resp {
		t.Logf("value: %+v", a)
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		return
	}
	dict := &[]models.BKDict{}
	json.Unmarshal(bytes, dict)
	t.Logf("value: %s", string(bytes))
	md := util.MarkdownTableWithTitle("行业/板块代码", dict)
	t.Log(md)

}

func TestTradingViewNews(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	NewMarketNewsApi().TradingViewNews()
}

func TestXUEQIUHotStock(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().XUEQIUHotStock(50, "10")
	for _, a := range *res {
		t.Logf("value: %+v", a)
	}

	md := util.MarkdownTableWithTitle("当前热门股票排名", res)
	t.Log(md)
}

func TestHotEvent(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().HotEvent(50)
	for _, a := range *res {
		t.Logf("value: %+v", a)
	}

}

func TestHotTopic(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().HotTopic(10)
	for _, a := range res {
		t.Logf("value: %+v", a)
	}

}

func TestInvestCalendar(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().InvestCalendar("2025-06")
	for _, a := range res {
		bytes, err := json.Marshal(a)
		if err != nil {
			continue
		}
		date := gjson.Get(string(bytes), "date")
		list := gjson.Get(string(bytes), "list")

		t.Logf("value: %+v,list: %+v", date.String(), list)
	}
}

func TestClsCalendar(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewMarketNewsApi().ClsCalendar()
	md := strings.Builder{}
	for _, a := range res {
		bytes, err := json.Marshal(a)
		if err != nil {
			continue
		}
		//t.Logf("value: %+v", string(bytes))
		date := gjson.Get(string(bytes), "calendar_day")
		md.WriteString("\n### 事件/会议日期：" + date.String())
		list := gjson.Get(string(bytes), "items")
		//t.Logf("value: %+v,list: %+v", date.String(), list)
		list.ForEach(func(key, value gjson.Result) bool {
			t.Logf("key: %+v,value: %+v", key.String(), gjson.Get(value.String(), "title"))
			md.WriteString("\n- " + gjson.Get(value.String(), "title").String())
			return true
		})
	}
	t.Logf("md:\n %s", md.String())
}

func TestGetGDP(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetGDP()
	md := util.MarkdownTableWithTitle("国内生产总值(GDP)", res.GDPResult.Data)
	t.Log(md)
}
func TestGetCPI(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetCPI()
	md := util.MarkdownTableWithTitle("居民消费价格指数(CPI)", res.CPIResult.Data)
	t.Log(md)
}

// PPI
func TestGetPPI(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetPPI()
	md := util.MarkdownTableWithTitle("工业品出厂价格指数(PPI)", res.PPIResult.Data)
	t.Log(md)
}

// PMI
func TestGetPMI(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetPMI()
	md := util.MarkdownTableWithTitle("采购经理人指数(PMI)", res.PMIResult.Data)
	t.Log(md)
}
func TestGetIndustryReportInfo(t *testing.T) {
	requireManualTest(t)
	NewMarketNewsApi().GetIndustryReportInfo("AP202507151709216483")
}

func TestReutersNew(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	NewMarketNewsApi().ReutersNew()
}

func TestInteractiveAnswer(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	datas := NewMarketNewsApi().InteractiveAnswer(1, 100, "立讯精密")
	t.Logf("PageSize:%d", datas.PageSize)
	md := util.MarkdownTableWithTitle("投资互动", datas.Results)
	t.Log(md)

}
func TestGetNewsList2(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	news := NewMarketNewsApi().GetNewsList2("财联社电报", random.RandInt(100, 500))
	messageText := strings.Builder{}
	for _, telegraph := range *news {
		messageText.WriteString("## " + telegraph.Time + ":" + "\n")
		messageText.WriteString("### " + telegraph.Content + "\n")
	}
	t.Logf("value: %s", messageText.String())
}

func TestTelegraphList(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	InitAnalyzeSentiment()
	NewMarketNewsApi().TelegraphList(30)
}

func TestProxy(t *testing.T) {
	requireManualTest(t)
	response, err := resty.New().
		SetProxy("http://go-stock:778d4ff2-73f3-4d56-b3c3-d9a730a06ae3@stock.sparkmemory.top:8888").
		R().
		SetHeader("Host", "news-mediator.tradingview.com").
		SetHeader("Origin", "https://cn.tradingview.com").
		SetHeader("Referer", "https://cn.tradingview.com/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0").
		//Get("https://api.ipify.org")
		Get("https://news-mediator.tradingview.com/news-flow/v2/news?filter=lang%3Azh-Hans&client=screener&streaming=false&user_prostatus=non_pro")
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	t.Logf("value: %s", response.String())

}

func TestNtfy(t *testing.T) {
	requireManualTest(t)

	//attach := "http://go-stock.sparkmemory.top/%E5%88%86%E6%9E%90%E6%8A%A5%E5%91%8A/%E8%B5%84%E9%87%91%E6%B5%81%E5%90%91/2025-12/AI%EF%BC%9A%E5%B8%82%E5%9C%BA%E5%88%86%E6%9E%90%E6%8A%A5%E5%91%8A-[2025.12.11_12.02.01].html"
	//post, err := resty.New().SetBaseURL("https://go-stock.sparkmemory.top:16667").R().
	//	SetHeader("Filename", "AI：市场分析报告-[2025.12.11_12.02.01].html").
	//	SetHeader("Icon", "https://go-stock.sparkmemory.top/appicon.png").
	//	SetHeader("Attach", attach).
	//	SetBody("AI：市场分析报告-[2025.12.11_12.02.01]").Post("/go-stock")
	//if err != nil {
	//	t.Fatal(err)
	//	return
	//}
	//t.Logf("value: %s", post.String())
	t.Logf("value: %s", filepath.Base("https://go-stock.sparkmemory.top/%E5%88%86%E6%9E%90%E6%8A%A5%E5%91%8A/2025/12/11/%E5%B8%82%E5%9C%BA%E8%B5%84%E8%AE%AF[%E5%B8%82%E5%9C%BA%E8%B5%84%E8%AE%AF]-(2025-12-11)AI%E5%88%86%E6%9E%90%E7%BB%93%E6%9E%9C_20251211131509.html"))
	t.Logf("value: %s", strutil.After("/data/go-stock-site/docs/分析报告/2025/12/09/市场资讯[市场资讯]-(2025-12-09)AI分析结果.md", "/data/go-stock-site/docs/"))
}

func TestGetSecuritiesCompanyOpinion(t *testing.T) {
	requireManualTest(t)
	res := NewMarketNewsApi().GetSecuritiesCompanyOpinion("2026-03-01", "2026-03-03")
	md := strings.Builder{}
	for _, d := range res.Data {
		md.WriteString(d.OpinionData + "\n")
	}
	t.Logf("%s", md.String())

}
func TestGetNewsListData(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	list, total := NewMarketNewsApi().GetNewsListData("", time.Now().Add(-time.Hour*24*2), 1, 2)
	md := strings.Builder{}
	md.WriteString("### " + "最近新闻资讯" + "（共 " + strconv.FormatInt(total, 10) + " 条）\r\n")
	for _, d := range *list {
		md.WriteString(d.DataTime.Format(time.DateTime) + " " + d.Content + "\r\n")
	}
	//t.Logf("%s", md.String())
}
