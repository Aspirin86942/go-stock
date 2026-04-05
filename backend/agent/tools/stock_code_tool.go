package tools

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"go-stock/backend/data"
	"go-stock/backend/logger"
)

// @Author spark
// @Date 2025/8/4 18:25
// @Desc
//-----------------------------------------------------------------------------------

func GetQueryStockCodeInfoTool() tool.InvokableTool {
	return &QueryStockCodeInfo{}
}

type QueryStockCodeInfo struct {
}

func (q QueryStockCodeInfo) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "QueryStockCodeInfo",
		Desc: "查询股票/指数信息(股票/指数名称,股票/指数代码,股票/指数拼音,股票/指数拼音首字母,股票/指数交易所等",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"searchWord": {
				Type:     "string",
				Desc:     "股票搜索关键词",
				Required: true,
			},
		}),
	}, nil
}

func (q QueryStockCodeInfo) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	trace := toolTrace("tool-query-stock-code")
	if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
		log.WithTrace(trace).Info(
			"tool.stock_code.called",
			"query stock code tool called",
			logger.String("arguments", argumentsInJSON),
		)
	}
	parms := map[string]any{}
	err := json.Unmarshal([]byte(argumentsInJSON), &parms)
	if err != nil {
		if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
			log.WithTrace(trace).Error(
				"tool.stock_code.arguments_invalid",
				"unmarshal stock code tool args failed",
				logger.Err(err),
			)
		}
		return "", err
	}
	searchWord, ok := parms["searchWord"].(string)
	if !ok {
		if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
			log.WithTrace(trace).Warn(
				"tool.stock_code.search_word_missing",
				"search word not found in tool args",
			)
		}
		return "未找到股票信息", nil
	}
	if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
		log.WithTrace(trace).Info(
			"tool.stock_code.searching",
			"searching stock code info",
			logger.String("search_word", searchWord),
		)
	}
	stockList := data.NewStockDataApi().GetStockList(searchWord)
	marshal, err := json.Marshal(stockList)
	if err != nil {
		if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
			log.WithTrace(trace).Error(
				"tool.stock_code.marshal_failed",
				"marshal stock code tool result failed",
				logger.Err(err),
			)
		}
		return "", err
	}
	if log := toolModuleLogger("agent.tool.stock_code"); log != nil {
		log.WithTrace(trace).Info(
			"tool.stock_code.completed",
			"stock code tool completed",
			logger.Int("result_length", len(marshal)),
		)
	}
	return string(marshal), nil
}
