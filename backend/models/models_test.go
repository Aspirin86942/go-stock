package models_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/duke-git/lancet/v2/strutil"

	"go-stock/backend/db"
	"go-stock/backend/models"
)

// @Author spark
// @Date 2025/2/22 16:09
// @Desc
// -----------------------------------------------------------------------------------
type StockInfoHKResp struct {
	Code       int              `json:"code"`
	Status     string           `json:"status"`
	StockInfos *[]StockInfoData `json:"data"`
}

type StockInfoData struct {
	C string `json:"c"`
	N string `json:"n"`
	T string `json:"t"`
	E string `json:"e"`
}

func TestStockInfoHK(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	if err := db.Dao.AutoMigrate(&models.StockInfoHK{}); err != nil {
		t.Fatalf("migrate stock_info_hk: %v", err)
	}
	bs, err := os.ReadFile("../../build/hk.json")
	if err != nil {
		t.Fatalf("read hk fixture: %v", err)
	}
	v := &StockInfoHKResp{}
	if err := json.Unmarshal(bs, v); err != nil {
		t.Fatalf("unmarshal hk fixture: %v", err)
	}
	if v.StockInfos == nil || len(*v.StockInfos) == 0 {
		t.Fatalf("expected hk stock infos to be non-empty")
	}
	hks := &[]models.StockInfoHK{}
	for i, data := range *v.StockInfos {
		t.Logf("第%d条数据: %+v", i, data)
		hk := &models.StockInfoHK{
			Code:  strutil.PadStart(data.C, 5, "0") + ".HK",
			EName: data.N,
		}
		*hks = append(*hks, *hk)
	}
	if err := db.Dao.Create(&hks).Error; err != nil {
		t.Fatalf("create hk rows: %v", err)
	}
}
