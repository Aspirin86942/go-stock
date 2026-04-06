package notification

import (
	"strings"

	"go-stock/backend/data"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
)

type Cache interface {
	TTL(key []byte) (uint32, error)
	Set(key []byte, value []byte, expireSeconds int) error
}

type Adapter interface {
	SendDingTalk(message string) string
	SendLocal(title, content string) bool
	LoadStockInfo(stockCode string) *data.StockInfo
}

type Delivery struct {
	DingResult   string `json:"dingResult"`
	EventTitle   string `json:"eventTitle"`
	EventContent string `json:"eventContent"`
}

type Service struct {
	cache   Cache
	adapter Adapter
}

func NewService(cache Cache, adapter Adapter) *Service {
	if cache == nil {
		panic("notification: cache dependency is required")
	}
	if adapter == nil {
		panic("notification: adapter dependency is required")
	}
	return &Service{
		cache:   cache,
		adapter: adapter,
	}
}

func (s *Service) SendDingTalk(message, stockCode string) string {
	ttl, _ := s.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return ""
	}
	if err := s.cache.Set([]byte(stockCode), []byte("1"), 60*5); err != nil {
		return ""
	}
	return s.adapter.SendDingTalk(message)
}

func (s *Service) SendTyped(message, stockCode string, msgType int) Delivery {
	ttl, _ := s.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return Delivery{}
	}
	if err := s.cache.Set([]byte(stockCode), []byte("1"), ttlForMessageType(msgType)); err != nil {
		return Delivery{}
	}

	content := formatNotificationContent(s.adapter.LoadStockInfo(stockCode))
	if strings.TrimSpace(content) != "" {
		s.adapter.SendLocal(messageTypeName(msgType), content)
	}

	return Delivery{
		DingResult:   s.adapter.SendDingTalk(message),
		EventTitle:   messageTypeName(msgType),
		EventContent: content,
	}
}

func ttlForMessageType(msgType int) int {
	switch msgType {
	case 1, 4, 5:
		return 60 * 5
	case 2, 3:
		return 60 * 30
	default:
		return 60 * 5
	}
}

func messageTypeName(msgType int) string {
	switch msgType {
	case 1:
		return "涨跌报警"
	case 2:
		return "股价报警"
	case 3:
		return "成本价报警"
	case 4:
		return "止盈报警"
	case 5:
		return "止损报警"
	default:
		return "未知类型"
	}
}

func formatNotificationContent(stockInfo *data.StockInfo) string {
	if stockInfo == nil {
		return ""
	}
	price, err := convertor.ToFloat(stockInfo.Price)
	if err != nil {
		price = 0
	}
	preClose, err := convertor.ToFloat(stockInfo.PreClose)
	if err != nil {
		preClose = 0
	}
	changePercent := float64(0)
	if preClose > 0 {
		changePercent = mathutil.RoundToFloat(((price-preClose)/preClose)*100, 2)
	}
	return "[" + stockInfo.Name + "] " + stockInfo.Price + " " + convertor.ToString(changePercent) + "% " + stockInfo.Date + " " + stockInfo.Time
}
