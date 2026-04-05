package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"go-stock/backend/util"
	"go-stock/internal/testenv"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

// @Author spark
// @Date 2025/2/24 9:35
// @Desc
// -----------------------------------------------------------------------------------
func TestIsHKTradingTime(t *testing.T) {
	f := IsHKTradingTime(time.Now())
	t.Log(f)
}

func TestIsUSTradingTime(t *testing.T) {

	date := time.Now()
	hour, minute, _ := date.Clock()
	t.Logf("当前时间: %02d:%02d", hour, minute)

	t.Logf("美股交易时段=%v", IsUSTradingTime(time.Now()))
}

func TestManual_CheckStockBaseInfo(t *testing.T) {
	testenv.RequireManualTest(t)
	db.Init("./data/stock.db")
	NewApp().CheckStockBaseInfo(context.Background())
}

func TestManual_UpdateStockInfoUSFromJSON(t *testing.T) {
	testenv.RequireManualTest(t)
	db.Init("./data/stock.db")

	jsonStr := "{\n\t\t\"id\" : 3334,\n\t\t\"created_at\" : \"2025-02-28 16:49:31.8342514+08:00\",\n\t\t\"updated_at\" : \"2025-02-28 16:49:31.8342514+08:00\",\n\t\t\"deleted_at\" : null,\n\t\t\"code\" : \"PUK.US\",\n\t\t\"name\" : \"英国保诚集团\",\n\t\t\"full_name\" : \"\",\n\t\t\"e_name\" : \"\",\n\t\t\"exchange\" : \"NASDAQ\",\n\t\t\"type\" : \"stock\",\n\t\t\"is_del\" : 0,\n\t\t\"bk_name\" : null,\n\t\t\"bk_code\" : null\n\t}"

	v := &models.StockInfoUS{}
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		t.Fatalf("unmarshal stock info us: %v", err)
	}
	t.Logf("stock info us: %+v", v)
	if err := db.Dao.Model(v).Updates(v).Error; err != nil {
		t.Fatalf("update stock info us: %v", err)
	}
}

func TestUpdateCheck(t *testing.T) {
	requireIntegrationTest(t)
	releaseVersion := &models.GitHubReleaseVersion{}
	_, err := resty.New().R().
		SetResult(releaseVersion).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		Get("https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest")
	//  https://api.github.com/repos/OWNER/REPO/releases/latest
	if err != nil {
		t.Fatalf("get github release version: %v", err)
	}
	if strings.TrimSpace(releaseVersion.TagName) == "" {
		t.Fatalf("expected latest release tag name, got %#v", releaseVersion)
	}
	t.Logf("releaseVersion=%+v", releaseVersion)
}

func TestReleaseSmoke_GetScreenResolution(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	x, y, w, h, err := getScreenResolution()
	if err != nil {
		t.Fatalf("get screen resolution: %v", err)
	}
	t.Logf("screen resolution x=%d y=%d w=%d h=%d", x, y, w, h)

}

func TestManual_CheckUpdate(t *testing.T) {
	testenv.RequireManualTest(t)
	db.Init("./data/stock.db")
	NewApp().CheckUpdate(1)
}

func TestManual_GetAiRecommendStocksList(t *testing.T) {
	testenv.RequireManualTest(t)
	db.Init("./data/stock.db")

	str := "{\"startDate\": \"2026-03-20 00:00:00\", \"endDate\": \"2026-03-27 23:59:59\", \"page\": 1, \"pageSize\": 5000}"
	query := &models.AiRecommendStocksQuery{}
	if err := json.Unmarshal([]byte(str), query); err != nil {
		t.Fatalf("unmarshal ai recommend query: %v", err)
	}

	pageData, err := data.NewAiRecommendStocksService().GetAiRecommendStocksList(query)
	if err != nil {
		t.Fatalf("get ai recommend stocks list: %v", err)
	}
	if pageData == nil || len(pageData.List) == 0 {
		t.Fatalf("expected ai recommend stocks list to be non-empty")
	}
	var dataExport []models.AiRecommendStocksMdExport
	for _, v := range pageData.List {
		dataExport = append(dataExport, v.ToMdExportStruct())
	}
	content := util.MarkdownTableWithTitle("近期AI分析/推荐股票明细列表", dataExport)
	if strings.TrimSpace(content) == "" {
		t.Fatalf("expected ai recommend markdown to be non-empty")
	}
	t.Logf("ai recommend rows=%d", len(pageData.List))
}

func TestManual_SummaryStockNews(t *testing.T) {
	testenv.RequireManualTest(t)
	db.Init("./data/stock.db")
	question := "分析今日的市场行情走势是否和券商的观点一致"
	app := NewApp()
	msgs := data.NewDeepSeekOpenAi(app.ctx, 0).NewSummaryStockNewsStreamWithTools(question, nil, app.AiTools, true, nil)

	content := &strings.Builder{}
	for msg := range msgs {
		t.Logf("msg=%+v", msg)
		segment, ok := msg["content"].(string)
		if !ok {
			t.Fatalf("expected content string in stream message, got %#v", msg["content"])
		}
		content.WriteString(segment)
	}
	if content.Len() == 0 {
		t.Fatalf("expected summary content to be non-empty")
	}
	t.Logf("summary content len=%d", content.Len())
}

func TestCalculateNextRunTime(t *testing.T) {
	t.Log(NewApp().CalculateNextRunTime("0 0 0 * * ?"))
}

func TestManual_FetchAiModels(t *testing.T) {
	testenv.RequireManualTest(t)
	app := NewApp()
	models := app.FetchAiModels("https://ark.cn-beijing.volces.com/api/v3", "")
	t.Log(models)
}

func TestGetEffectiveSponsorVip_ReturnsOpenAccess(t *testing.T) {
	db.Init(fmt.Sprintf("%s/stock.db", t.TempDir()))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	app := NewApp()
	result := app.GetEffectiveSponsorVip()

	active, ok := result["active"].(bool)
	if !ok {
		t.Fatalf("expected active to be bool, got %#v", result["active"])
	}
	if !active {
		t.Fatalf("expected active=true, got %#v", result["active"])
	}

	level, ok := result["vipLevel"].(int)
	if !ok {
		t.Fatalf("expected vipLevel to be int, got %#v", result["vipLevel"])
	}
	if level < 2 {
		t.Fatalf("expected vipLevel >= 2, got %d", level)
	}
}
