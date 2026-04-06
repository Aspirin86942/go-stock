package fund

import "go-stock/backend/data"

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) GetFundList(key string) []data.FundBasic {
	return data.NewFundApi().GetFundList(key)
}

func (a *Adapter) GetFollowedFund() []data.FollowedFund {
	return data.NewFundApi().GetFollowedFund()
}

func (a *Adapter) FollowFund(fundCode string) string {
	return data.NewFundApi().FollowFund(fundCode)
}

func (a *Adapter) UnFollowFund(fundCode string) string {
	return data.NewFundApi().UnFollowFund(fundCode)
}

func (a *Adapter) CrawlFundBasic(fundCode string) (*data.FundBasic, error) {
	return data.NewFundApi().CrawlFundBasic(fundCode)
}

func (a *Adapter) CrawlFundNetEstimatedUnit(code string) {
	data.NewFundApi().CrawlFundNetEstimatedUnit(code)
}

func (a *Adapter) CrawlFundNetUnitValue(code string) {
	data.NewFundApi().CrawlFundNetUnitValue(code)
}

func (a *Adapter) AllFund() {
	data.NewFundApi().AllFund()
}
