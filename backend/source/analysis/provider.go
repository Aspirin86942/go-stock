package analysis

import (
	"context"

	"go-stock/backend/data"
	analysisservice "go-stock/backend/service/analysis"
)

type Provider struct {
	tools []data.Tool
}

func NewProvider(tools []data.Tool) *Provider {
	return &Provider{tools: append([]data.Tool(nil), tools...)}
}

func (p *Provider) StockStream(ctx context.Context, request analysisservice.StockRequest) <-chan map[string]any {
	if request.EnableTools {
		return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewChatStream(
			request.StockName,
			request.StockCode,
			request.Question,
			request.SysPromptID,
			p.tools,
			request.Think,
		)
	}
	return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewChatStream(
		request.StockName,
		request.StockCode,
		request.Question,
		request.SysPromptID,
		[]data.Tool{},
		request.Think,
	)
}

func (p *Provider) MarketSummaryStream(ctx context.Context, request analysisservice.MarketSummaryRequest, history []map[string]interface{}) <-chan map[string]any {
	if request.EnableTools {
		return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewSummaryStockNewsStreamWithTools(
			request.Question,
			request.SysPromptID,
			p.tools,
			request.Think,
			history,
		)
	}
	return data.NewDeepSeekOpenAi(ctx, request.AIConfigID).NewSummaryStockNewsStream(
		request.Question,
		request.SysPromptID,
		request.Think,
		history,
	)
}
