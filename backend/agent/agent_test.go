package agent

import (
	"context"
	"errors"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
	"github.com/duke-git/lancet/v2/fileutil"
)

// @Author spark
// @Date 2025/8/4 17:32
// @Desc
//-----------------------------------------------------------------------------------

func TestGetStockAiAgent(t *testing.T) {
	requireIntegrationTest(t)
	ctx := context.Background()
	db.Init("../../data/stock.db")
	config := data.GetSettingConfig()
	aiConfig := requireFirstAIConfig(t, config)
	aiAgent := GetStockAiAgent(&ctx, *aiConfig)

	opt := []agent.AgentOption{
		//agent.WithComposeOptions(compose.WithCallbacks(&tool_logger.LoggerCallback{})),
		//react.WithChatModelOptions(ark.WithCache(cacheOption)),
	}

	sr, err := aiAgent.Stream(ctx, []*schema.Message{
		{
			Role:    schema.System,
			Content: config.Settings.Prompt,
		},
		{
			Role:    schema.User,
			Content: "结合以上提供的宏观经济数据/市场指数行情/国内外市场资讯/电报/会议/事件/投资者关注的问题，\n结合宏观经济，事件驱动，政策支持，投资者关注的问题，分析当前市场情绪和热点 找出有潜力/优质的板块/行业/概念/标的/主题，\n多因子深度分析计算上涨或下跌的逻辑和概率，\n最后按风险和投资周期给出具体推荐标的操作建议",
		},
	}, opt...)
	if err != nil {
		t.Fatalf("stream error: %v", err)
	}

	defer sr.Close() // remember to close the stream

	md := strings.Builder{}
	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// finish
				break
			}
			t.Fatalf("recv stream chunk: %v", err)
		}
		t.Logf("stream recv: %#v", msg)
		if msg.ReasoningContent != "" {
			md.WriteString(msg.ReasoningContent)
		}
		if msg.Content != "" {
			md.WriteString(msg.Content)
		}
	}
	if md.Len() == 0 {
		t.Fatalf("expected streamed agent output to be non-empty")
	}
	outputPath := filepath.Join(t.TempDir(), "result.md")
	if err := fileutil.WriteStringToFile(outputPath, md.String(), false); err != nil {
		t.Fatalf("write streamed result to %s: %v", outputPath, err)
	}
}

func TestAgent(t *testing.T) {
	requireIntegrationTest(t)
	db.Init("../../data/stock.db")
	aiConfig := requireFirstAIConfig(t, data.GetSettingConfig())

	md := strings.Builder{}
	ch := NewStockAiAgentApi().Chat("分析一下立讯精密", int(aiConfig.ID), nil)
	for message := range ch {
		t.Logf("res=%s", message.String())
		md.WriteString(message.String())
	}
	if md.Len() == 0 {
		t.Fatalf("expected chat output to be non-empty")
	}
}

func requireFirstAIConfig(t *testing.T, config *data.SettingConfig) *data.AIConfig {
	t.Helper()

	if config == nil || len(config.AiConfigs) == 0 {
		t.Skip("skipping external agent test: no AI config available")
	}
	for _, item := range config.AiConfigs {
		if item != nil {
			return item
		}
	}
	t.Skip("skipping external agent test: AI config entries are nil")
	return nil
}
