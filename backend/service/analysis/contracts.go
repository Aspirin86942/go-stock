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
	HistoryJSON string `json:"historyJSON"`
}

// StreamChunk is a normalized event consumed by the callers.
type StreamChunk struct {
	ChatID           string           `json:"chatId"`
	Question         string           `json:"question"`
	Content          string           `json:"content"`
	ExtraContent     string           `json:"extraContent"`
	Model            string           `json:"model"`
	Time             string           `json:"time"`
	ReasoningContent string           `json:"reasoning_content"`
	ToolCalls        []map[string]any `json:"tool_calls"`
}

// ResultArtifact describes a stable share/export view for the latest analysis result.
type ResultArtifact struct {
	StockCode        string `json:"stockCode"`
	StockName        string `json:"stockName"`
	Content          string `json:"content"`
	AnalysisDate     string `json:"analysisDate"`
	MarkdownFilename string `json:"markdownFilename"`
}
