package notification

import (
	"go-stock/backend/data"
	"go-stock/backend/db"
)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) SendDingTalk(message string) string {
	return data.NewDingDingAPI().SendDingDingMessage(message)
}

func (a *Adapter) LoadStockInfo(stockCode string) *data.StockInfo {
	stockInfo := &data.StockInfo{}
	db.Dao.Model(stockInfo).Where("code = ?", stockCode).First(stockInfo)
	return stockInfo
}

func (a *Adapter) SendLocal(title, content string) bool {
	return sendLocalNotification("go-stock消息通知", title, content, "")
}
