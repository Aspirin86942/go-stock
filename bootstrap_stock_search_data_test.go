package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/models"
)

func TestSeedBundledStockSearchDataPopulatesEmptyRuntimeTables(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "stock.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := migrateStockSearchTables(); err != nil {
		t.Fatalf("migrate stock search tables: %v", err)
	}

	err = seedBundledStockSearchData(
		[]byte(`{"data":{"fields":["ts_code","symbol","name","area","industry","cnspell","market","list_date","act_name","act_ent_type","fullname","exchange","list_status","curr_type","enname","delist_date","is_hs"],"items":[["600519.SH","600519","贵州茅台","贵州","酿酒","gzmt","主板","20010827","","","贵州茅台股份有限公司","SSE","L","CNY","","","N"]],"has_more":false,"count":1}}`),
		[]byte(`[{"code":"hk00700","name":"腾讯控股","fullName":"腾讯控股有限公司","eName":"Tencent"}]`),
		[]byte(`[{"code":"gb_aapl","name":"苹果","fullName":"Apple Inc.","eName":"Apple","exchange":"NASDAQ","type":"stock"}]`),
	)
	if err != nil {
		t.Fatalf("seed bundled stock search data: %v", err)
	}

	assertTableCount(t, &data.StockBasic{}, 1)
	assertTableCount(t, &models.StockInfoHK{}, 1)
	assertTableCount(t, &models.StockInfoUS{}, 1)
}

func TestSeedBundledStockSearchDataSkipsTablesThatAlreadyHaveData(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "stock.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := migrateStockSearchTables(); err != nil {
		t.Fatalf("migrate stock search tables: %v", err)
	}

	if err := db.Dao.Create(&data.StockBasic{TsCode: "000001.SZ", Name: "平安银行"}).Error; err != nil {
		t.Fatalf("insert existing a-share stock: %v", err)
	}
	if err := db.Dao.Create(&models.StockInfoHK{Code: "hk00001", Name: "长和"}).Error; err != nil {
		t.Fatalf("insert existing hk stock: %v", err)
	}
	if err := db.Dao.Create(&models.StockInfoUS{Code: "gb_tsla", Name: "特斯拉"}).Error; err != nil {
		t.Fatalf("insert existing us stock: %v", err)
	}

	err = seedBundledStockSearchData(
		[]byte(`{"data":{"fields":["ts_code","symbol","name","area","industry","cnspell","market","list_date","act_name","act_ent_type","fullname","exchange","list_status","curr_type","enname","delist_date","is_hs"],"items":[["600519.SH","600519","贵州茅台","贵州","酿酒","gzmt","主板","20010827","","","贵州茅台股份有限公司","SSE","L","CNY","","","N"]],"has_more":false,"count":1}}`),
		[]byte(`[{"code":"hk00700","name":"腾讯控股","fullName":"腾讯控股有限公司","eName":"Tencent"}]`),
		[]byte(`[{"code":"gb_aapl","name":"苹果","fullName":"Apple Inc.","eName":"Apple","exchange":"NASDAQ","type":"stock"}]`),
	)
	if err != nil {
		t.Fatalf("seed bundled stock search data: %v", err)
	}

	assertTableCount(t, &data.StockBasic{}, 1)
	assertTableCount(t, &models.StockInfoHK{}, 1)
	assertTableCount(t, &models.StockInfoUS{}, 1)
}

func TestSeedBundledStockSearchDataBatchesLargeHKAndUSPayloads(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "stock.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := migrateStockSearchTables(); err != nil {
		t.Fatalf("migrate stock search tables: %v", err)
	}

	hkStocks := make([]models.StockInfoHK, 0, 1500)
	usStocks := make([]models.StockInfoUS, 0, 1500)
	for i := 0; i < 1500; i++ {
		hkStocks = append(hkStocks, models.StockInfoHK{
			Code: fmt.Sprintf("hk%05d", i),
			Name: fmt.Sprintf("港股%d", i),
		})
		usStocks = append(usStocks, models.StockInfoUS{
			Code:     fmt.Sprintf("gb_test_%d", i),
			Name:     fmt.Sprintf("美股%d", i),
			Exchange: "NASDAQ",
			Type:     "stock",
		})
	}

	hkPayload, err := json.Marshal(hkStocks)
	if err != nil {
		t.Fatalf("marshal hk payload: %v", err)
	}
	usPayload, err := json.Marshal(usStocks)
	if err != nil {
		t.Fatalf("marshal us payload: %v", err)
	}

	err = seedBundledStockSearchData(
		[]byte(`{"data":{"fields":["ts_code","symbol","name","area","industry","cnspell","market","list_date","act_name","act_ent_type","fullname","exchange","list_status","curr_type","enname","delist_date","is_hs"],"items":[],"has_more":false,"count":0}}`),
		hkPayload,
		usPayload,
	)
	if err != nil {
		t.Fatalf("seed bundled stock search data: %v", err)
	}

	assertTableCount(t, &models.StockInfoHK{}, 1500)
	assertTableCount(t, &models.StockInfoUS{}, 1500)
}

func TestSeedBundledStockSearchDataHandlesBundledHKAndUSPayloads(t *testing.T) {
	db.Init(filepath.Join(t.TempDir(), "stock.db"))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := migrateStockSearchTables(); err != nil {
		t.Fatalf("migrate stock search tables: %v", err)
	}

	err = seedBundledStockSearchData(
		[]byte(`{"data":{"fields":["ts_code"],"items":[],"has_more":false,"count":0}}`),
		stocksBinHK,
		stocksBinUS,
	)
	if err != nil {
		t.Fatalf("seed bundled stock search data with bundled payloads: %v", err)
	}

	var hkCount int64
	if err := db.Dao.Model(&models.StockInfoHK{}).Count(&hkCount).Error; err != nil {
		t.Fatalf("count hk stocks: %v", err)
	}
	if hkCount == 0 {
		t.Fatalf("expected bundled hk stock data to be imported")
	}

	var usCount int64
	if err := db.Dao.Model(&models.StockInfoUS{}).Count(&usCount).Error; err != nil {
		t.Fatalf("count us stocks: %v", err)
	}
	if usCount == 0 {
		t.Fatalf("expected bundled us stock data to be imported")
	}
}

func assertTableCount(t *testing.T, model any, expected int64) {
	t.Helper()

	var count int64
	if err := db.Dao.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if count != expected {
		t.Fatalf("unexpected %T count: got %d want %d", model, count, expected)
	}
}
