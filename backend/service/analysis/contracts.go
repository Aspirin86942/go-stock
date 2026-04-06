package analysis

// StockRequest describes the client request for streaming stock analysis chunks.
type StockRequest struct {
	StockName   string `json:"stockName"`
	StockCode   string `json:"stockCode"`
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
}

// MarketSummaryRequest describes the client request for streaming a market summary.
type MarketSummaryRequest struct {
	Question    string `json:"question"`
	AIConfigID  int    `json:"aiConfigId"`
	SysPromptID *int   `json:"sysPromptId"`
	EnableTools bool   `json:"enableTools"`
	Think       bool   `json:"think"`
	HistoryJSON string `json:"historyJson"`
}

// StreamChunk is a normalized event consumed by the callers.
type StreamChunk struct {
	ChatID       string `json:"chatId"`
	Question     string `json:"question"`
	Content      string `json:"content"`
	ExtraContent string `json:"extraContent"`
	Model        string `json:"model"`
	Time         string `json:"time"`
}
