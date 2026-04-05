package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"

	"github.com/cloudwego/eino/components/tool"
)

func TestGetAllDataTools(t *testing.T) {
	tools := GetAllDataTools()
	t.Logf("Total tools count: %d", len(tools))

	toolNames := make(map[string]int)
	for i, tool := range tools {
		info, err := tool.Info(nil)
		if err != nil {
			t.Errorf("Tool %d: failed to get info: %v", i, err)
			continue
		}
		t.Logf("Tool %d: %s - %s", i+1, info.Name, info.Desc)

		if count, exists := toolNames[info.Name]; exists {
			t.Errorf("Duplicate tool name found: %s (count: %d)", info.Name, count+1)
		}
		toolNames[info.Name]++
	}

	t.Log("\n=== Tool List ===")
	for name, count := range toolNames {
		if count > 1 {
			t.Errorf("Duplicate tool: %s (count: %d)", name, count)
		}
	}
}

func TestGetMarketData_ReusesContextTrace(t *testing.T) {
	originalLoader := marketDataLoader
	t.Cleanup(func() {
		marketDataLoader = originalLoader
	})

	marketDataLoader = func(ctx context.Context) (APIResponse, error) {
		return APIResponse{
			Code: 200,
			Data: APIData{
				IndexQuote: []APIIndexQuote{
					{
						SecuCode: "sh000001",
						SecuName: "上证指数",
						LastPx:   3000.12,
						Change:   0.01,
						ChangePx: 30.01,
						UpNum:    10,
						DownNum:  5,
						FlatNum:  2,
					},
				},
				UpDownDis: APIUpDownDis{
					UpNum:       12,
					DownNum:     3,
					AverageRise: 0.02,
					RiseNum:     100,
					FallNum:     80,
				},
			},
		}, nil
	}

	logRoot, err := os.MkdirTemp("", "go-stock-market-data-log-*")
	if err != nil {
		t.Fatalf("create temp log root: %v", err)
	}
	logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: logRoot,
			LogsDir: filepath.Join(logRoot, "logs"),
		},
		EnableStdout: false,
	})

	trace := logger.TraceContext{
		TraceID:      "trace-market-data-1",
		SpanID:       "span-market-data-1",
		AppSessionID: "session-market-data-1",
		Source:       "http",
	}
	ctx := logger.WithTraceContext(context.Background(), trace)

	var target any
	for _, item := range GetAllDataTools() {
		info, err := item.Info(ctx)
		if err != nil {
			t.Fatalf("get tool info: %v", err)
		}
		if info.Name == "GetMarketData" {
			target = item
			break
		}
	}
	if target == nil {
		t.Fatalf("GetMarketData tool not found")
	}

	invokable, ok := target.(tool.InvokableTool)
	if !ok {
		t.Fatalf("GetMarketData tool does not implement invokable interface")
	}

	result, err := invokable.InvokableRun(ctx, `{}`)
	if err != nil {
		t.Fatalf("run GetMarketData tool: %v", err)
	}
	if !strings.Contains(result, "# 市场行情数据") {
		t.Fatalf("expected market data content, got %q", result)
	}

	content, err := os.ReadFile(filepath.Join(logRoot, "logs", "ai.log"))
	if err != nil {
		t.Fatalf("read ai log: %v", err)
	}
	traceByEvent := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("unmarshal log line %q: %v", line, err)
		}
		event, _ := entry["event"].(string)
		traceID, _ := entry["trace_id"].(string)
		traceByEvent[event] = traceID
	}

	for _, event := range []string{"tool.data_wrapper.called", "tool.market_data.generated"} {
		if traceByEvent[event] != "trace-market-data-1" {
			t.Fatalf("expected %s to reuse trace-market-data-1, got %#v", event, traceByEvent)
		}
	}
}
