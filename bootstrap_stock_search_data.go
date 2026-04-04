package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"

	"github.com/duke-git/lancet/v2/slice"
)

const bundledStockSeedBatchSize = 50

// 正式版首次启动可能会落到 build/bin/data/stock.db，新库为空时需要用随包基础数据回填搜索候选。
func ensureBundledStockSearchData() error {
	if err := migrateStockSearchTables(); err != nil {
		return err
	}
	return seedBundledStockSearchData(stocksBin, stocksBinHK, stocksBinUS)
}

func migrateStockSearchTables() error {
	return db.Dao.AutoMigrate(&data.StockBasic{}, &models.StockInfoHK{}, &models.StockInfoUS{})
}

func seedBundledStockSearchData(aSharePayload, hkPayload, usPayload []byte) error {
	if err := seedBundledAStockData(aSharePayload); err != nil {
		return err
	}
	if err := seedBundledHKStockData(hkPayload); err != nil {
		return err
	}
	if err := seedBundledUSStockData(usPayload); err != nil {
		return err
	}
	return nil
}

func seedBundledAStockData(payload []byte) error {
	var total int64
	if err := db.Dao.Model(&data.StockBasic{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 || len(payload) == 0 {
		return nil
	}

	fields := "ts_code,symbol,name,area,industry,cnspell,market,list_date,act_name,act_ent_type,fullname,exchange,list_status,curr_type,enname,delist_date,is_hs"
	res := &data.TushareStockBasicResponse{}
	if err := json.Unmarshal(payload, res); err != nil {
		return err
	}

	inserted := 0
	for _, item := range res.Data.Items {
		stock := &data.StockBasic{}
		stockData := map[string]any{}
		for _, field := range strings.Split(fields, ",") {
			idx := slice.IndexOf(res.Data.Fields, field)
			if idx == -1 || idx >= len(item) {
				continue
			}
			stockData[field] = item[idx]
		}
		jsonData, _ := json.Marshal(stockData)
		if err := json.Unmarshal(jsonData, stock); err != nil {
			continue
		}
		if stock.TsCode == "" {
			continue
		}
		stock.ID = 0
		if err := db.Dao.Create(stock).Error; err != nil {
			return fmt.Errorf("seed bundled A-share stock %s: %w", stock.TsCode, err)
		}
		inserted++
	}

	logger.SugaredLogger.Infof("seed bundled A-share stock data: %d", inserted)
	return nil
}

func seedBundledHKStockData(payload []byte) error {
	var total int64
	if err := db.Dao.Model(&models.StockInfoHK{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 || len(payload) == 0 {
		return nil
	}

	var stocks []models.StockInfoHK
	if err := json.Unmarshal(payload, &stocks); err != nil {
		return err
	}
	if len(stocks) == 0 {
		return nil
	}
	if err := db.Dao.CreateInBatches(&stocks, bundledStockSeedBatchSize).Error; err != nil {
		return fmt.Errorf("seed bundled HK stock data: %w", err)
	}

	logger.SugaredLogger.Infof("seed bundled HK stock data: %d", len(stocks))
	return nil
}

func seedBundledUSStockData(payload []byte) error {
	var total int64
	if err := db.Dao.Model(&models.StockInfoUS{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 || len(payload) == 0 {
		return nil
	}

	var stocks []models.StockInfoUS
	if err := json.Unmarshal(payload, &stocks); err != nil {
		return err
	}
	if len(stocks) == 0 {
		return nil
	}
	if err := db.Dao.CreateInBatches(&stocks, bundledStockSeedBatchSize).Error; err != nil {
		return fmt.Errorf("seed bundled US stock data: %w", err)
	}

	logger.SugaredLogger.Infof("seed bundled US stock data: %d", len(stocks))
	return nil
}
