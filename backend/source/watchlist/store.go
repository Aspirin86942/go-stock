package watchlist

import (
	"context"

	"go-stock/backend/data"
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
