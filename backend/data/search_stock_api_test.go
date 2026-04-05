package data

import (
	"encoding/json"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"go-stock/backend/util"
	"math"
	"testing"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/random"
)

func TestSearchStock(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")

	e := convertor.ToString(math.Floor(float64(9*random.RandFloat(0, 1, 12) + 1)))
	for i := 0; i < 19; i++ {
		e += convertor.ToString(math.Floor(float64(9 * random.RandFloat(0, 1, 12))))
	}
	t.Logf("e:%s", e)

	//res := NewSearchStockApi("量比大于2，基本面优秀，2025年三季报已披露，主力连续3日净流入，非创业板非科创板非ST").SearchStock(20)
	//res := NewSearchStockApi("今日涨幅前5的概念板块").SearchBk(50)
	res := NewSearchStockApi("今日涨幅前15的ETF").SearchETF(50)

	t.Logf("res:%+v", res)
	data := res["data"].(map[string]any)
	result := data["result"].(map[string]any)
	dataList := result["dataList"].([]any)
	columns := result["columns"].([]any)
	headers := map[string]string{}
	for _, v := range columns {
		//t.Logf("v:%+v", v)
		d := v.(map[string]any)
		//t.Logf("key:%s title:%s dateMsg:%s unit:%s", d["key"], d["title"], d["dateMsg"], d["unit"])
		title := convertor.ToString(d["title"])
		if convertor.ToString(d["dateMsg"]) != "" {
			title = title + "[" + convertor.ToString(d["dateMsg"]) + "]"
		}
		if convertor.ToString(d["unit"]) != "" {
			title = title + "(" + convertor.ToString(d["unit"]) + ")"
		}
		headers[d["key"].(string)] = title
	}
	table := &[]map[string]any{}
	for _, v := range dataList {
		//t.Logf("v:%+v", v)
		d := v.(map[string]any)
		tmp := map[string]any{}
		for key, title := range headers {
			//t.Logf("%s:%s", title, convertor.ToString(d[key]))
			tmp[title] = convertor.ToString(d[key])
		}
		*table = append(*table, tmp)
		//t.Logf("--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------")
	}
	jsonData, _ := json.Marshal(*table)
	markdownTable, _ := JSONToMarkdownTable(jsonData)
	t.Logf("markdownTable=\n%s", markdownTable)
}

func TestGetStockFinancialInfo(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	res := NewStockDataApi().GetStockFinancialInfo("600519.SH")
	MD := util.MarkdownTableWithTitle("600519.SH股票财报信息", res.Result.Data)
	requireNotBlank(t, "stock financial markdown", MD)
	t.Logf("financial markdown:\n%s", MD)
}
func TestGetStockHolderNum(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewStockDataApi().GetStockHolderNum("600519.SH")
	MD := util.MarkdownTableWithTitle("股票股东人数信息", res.Result.Data)
	t.Logf("res:\n%s", MD)
}

func TestSearchStockApi_HotStrategy(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewSearchStockApi("").HotStrategy()
	bytes, err := json.Marshal(res)
	if err != nil {
		return
	}
	strategy := &models.HotStrategy{}
	json.Unmarshal(bytes, strategy)
	for _, data := range strategy.Data {
		data.Chg = mathutil.RoundToFloat(100*data.Chg, 2)
	}
	markdownTable := util.MarkdownTable(strategy.Data)
	t.Logf("res:%s", markdownTable)
	//dataList := res["data"].([]any)
	//for _, v := range dataList {
	//	d := v.(map[string]any)
	//	t.Logf("v:%+v", d)
	//}
}
func TestSearchStockApi_HotStrategyTable(t *testing.T) {
	requireManualTest(t)
	db.Init("../../data/stock.db")
	res := NewSearchStockApi("").StrategySquare()
	t.Logf("res:%+v", res)
}
