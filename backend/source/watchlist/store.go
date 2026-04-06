package watchlist

import (
	"context"

	"go-stock/backend/data"
	"go-stock/backend/db"
	marketservice "go-stock/backend/service/market"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) SaveStockAICron(ctx context.Context, cronText, stockCode string) {
	_ = ctx
	data.NewStockDataApi().SetStockAICron(cronText, stockCode)
}

func (s *Store) GetFollowedStock(ctx context.Context, stockCode string) data.FollowedStock {
	_ = ctx
	return data.NewStockDataApi().GetFollowedStockByStockCode(stockCode)
}

func (s *Store) ListFollowedStocks(ctx context.Context) []data.FollowedStock {
	_ = ctx
	result := data.NewStockDataApi().GetFollowList(0)
	if result == nil {
		return []data.FollowedStock{}
	}
	return append([]data.FollowedStock(nil), (*result)...)
}

func (s *Store) ListFollowedStocksByGroup(ctx context.Context, groupID int) []data.FollowedStock {
	_ = ctx
	result := data.NewStockDataApi().GetFollowList(groupID)
	if result == nil {
		return []data.FollowedStock{}
	}
	return append([]data.FollowedStock(nil), (*result)...)
}

func (s *Store) Follow(ctx context.Context, stockCode string) string {
	_ = ctx
	return data.NewStockDataApi().Follow(stockCode)
}

func (s *Store) Unfollow(ctx context.Context, stockCode string) string {
	_ = ctx
	return data.NewStockDataApi().UnFollow(stockCode)
}

func (s *Store) ListGroups(ctx context.Context) []data.Group {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).GetGroupList()
}

func (s *Store) AddGroup(ctx context.Context, group data.Group) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).AddGroup(group)
}

func (s *Store) UpdateGroupSort(ctx context.Context, id int, newSort int) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).UpdateGroupSort(id, newSort)
}

func (s *Store) InitializeGroupSort(ctx context.Context) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).InitializeGroupSort()
}

func (s *Store) ListGroupStocks(ctx context.Context, groupID int) []data.GroupStock {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).GetGroupStockByGroupId(groupID)
}

func (s *Store) AddGroupStock(ctx context.Context, groupID int, stockCode string) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).AddStockGroup(groupID, stockCode)
}

func (s *Store) RemoveGroupStock(ctx context.Context, stockCode, name string, groupID int) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).RemoveStockGroup(stockCode, name, groupID)
}

func (s *Store) RemoveGroup(ctx context.Context, groupID int) bool {
	_ = ctx
	return data.NewStockGroupApi(db.Dao).RemoveGroup(groupID)
}

func (s *Store) SetCostPriceAndVolume(ctx context.Context, stockCode string, price float64, volume int64) string {
	_ = ctx
	return data.NewStockDataApi().SetCostPriceAndVolume(price, volume, stockCode)
}

func (s *Store) SetTradingPrice(ctx context.Context, stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	_ = ctx
	return data.NewStockDataApi().SetTradingPrice(entryPrice, takeProfitPrice, stopLossPrice, costPrice, stockCode)
}

func (s *Store) SetAlarmChangePercent(ctx context.Context, stockCode string, val, alarmPrice float64) string {
	_ = ctx
	return data.NewStockDataApi().SetAlarmChangePercent(val, alarmPrice, stockCode)
}

func (s *Store) SetStockSort(ctx context.Context, stockCode string, sort int64) {
	_ = ctx
	data.NewStockDataApi().SetStockSort(sort, stockCode)
}

func (s *Store) GetRealtimePrices(ctx context.Context, stockCodes ...string) []marketservice.RealtimePrice {
	_ = ctx
	if len(stockCodes) == 0 {
		return []marketservice.RealtimePrice{}
	}

	quotes, err := data.NewStockDataApi().GetStockCodeRealTimeData(stockCodes...)
	if err != nil || quotes == nil {
		return []marketservice.RealtimePrice{}
	}

	result := make([]marketservice.RealtimePrice, 0, len(*quotes))
	for _, item := range *quotes {
		result = append(result, marketservice.RealtimePrice{
			StockCode: item.Code,
			StockName: item.Name,
			Price:     item.Price,
			Bid:       item.Bid,
			Ask:       item.Ask,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			PreClose:  item.PreClose,
			Date:      item.Date,
			Time:      item.Time,
		})
	}
	return result
}

func (s *Store) UpdateObservedPrice(ctx context.Context, stockCode string, price float64) {
	_ = ctx
	db.Dao.Model(&data.FollowedStock{}).Where("stock_code = ?", stockCode).Updates(map[string]any{
		"price": price,
	})
}
