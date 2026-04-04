package data

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func drainToolTestMessages(ch chan map[string]any) []map[string]any {
	var msgs []map[string]any
	for {
		select {
		case msg := <-ch:
			msgs = append(msgs, msg)
		default:
			return msgs
		}
	}
}

func TestAskAiWithTools_ReportsUnsupportedFunctionCalling(t *testing.T) {
	var (
		mu          sync.Mutex
		requestBody [][]byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		mu.Lock()
		requestBody = append(requestBody, body)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"message":"Function call is not supported for this model."}}`))
	}))
	defer server.Close()

	ch := make(chan map[string]any, 8)
	AskAiWithTools(
		&OpenAi{
			BaseUrl:     server.URL,
			ApiKey:      "test-key",
			Model:       "test-model",
			MaxTokens:   256,
			Temperature: 0.1,
			TimeOut:     5,
		},
		nil,
		[]map[string]interface{}{
			{"role": "user", "content": "请分析这只股票"},
		},
		ch,
		"请分析这只股票",
		[]Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "FakeTool",
					Description: "test tool",
				},
			},
		},
		false,
	)

	mu.Lock()
	requestCount := len(requestBody)
	mu.Unlock()
	if requestCount != 1 {
		t.Fatalf("expected exactly 1 request when tool calling is unsupported, got %d", requestCount)
	}

	msgs := drainToolTestMessages(ch)
	if len(msgs) == 0 {
		t.Fatalf("expected at least one emitted message")
	}

	content, _ := msgs[0]["content"].(string)
	if !strings.Contains(content, "当前模型不支持工具调用") {
		t.Fatalf("expected explicit unsupported-tools message, got %q", content)
	}
}

func TestAskAiWithTools_AccumulatesStreamedToolArgumentsAndForcesInitialToolChoice(t *testing.T) {
	const fakeToolName = "FakeToolForOpenAITest"

	previousHandler, hadPreviousHandler := toolHandlers[fakeToolName]
	registerToolHandler(fakeToolName, func(o *OpenAi, funcArguments string, ctx *ToolContext) error {
		appendToolMessages(
			ctx.Messages,
			ctx.CurrentAIContent.String(),
			ctx.ReasoningContentText.String(),
			ctx.CurrentCallID,
			ctx.FuncName,
			funcArguments,
			"fake tool output",
		)
		return nil
	})
	defer func() {
		if hadPreviousHandler {
			toolHandlers[fakeToolName] = previousHandler
			return
		}
		delete(toolHandlers, fakeToolName)
	}()

	var (
		mu            sync.Mutex
		requestBodies [][]byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		mu.Lock()
		requestBodies = append(requestBodies, body)
		reqIndex := len(requestBodies)
		mu.Unlock()

		w.Header().Set("Content-Type", "text/event-stream")
		switch reqIndex {
		case 1:
			_, _ = w.Write([]byte(strings.Join([]string{
				`data: {"id":"resp-1","model":"test-model","choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"` + fakeToolName + `","arguments":"{\"stock"}}]},"finish_reason":""}]}`,
				`data: {"id":"resp-1","model":"test-model","choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"Code\":\"600519.SH\"}"}}]},"finish_reason":""}]}`,
				`data: {"id":"resp-1","model":"test-model","choices":[{"delta":{},"finish_reason":"tool_calls"}]}`,
				`data: [DONE]`,
				"",
			}, "\n\n")))
		case 2:
			_, _ = w.Write([]byte(strings.Join([]string{
				`data: {"id":"resp-2","model":"test-model","choices":[{"delta":{"content":"已完成"},"finish_reason":"stop"}]}`,
				`data: [DONE]`,
				"",
			}, "\n\n")))
		default:
			t.Fatalf("unexpected request count: %d", reqIndex)
		}
	}))
	defer server.Close()

	ch := make(chan map[string]any, 16)
	AskAiWithTools(
		&OpenAi{
			BaseUrl:     server.URL,
			ApiKey:      "test-key",
			Model:       "test-model",
			MaxTokens:   256,
			Temperature: 0.1,
			TimeOut:     5,
		},
		nil,
		[]map[string]interface{}{
			{"role": "user", "content": "请分析这只股票"},
		},
		ch,
		"请分析这只股票",
		[]Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        fakeToolName,
					Description: "test tool",
				},
			},
		},
		false,
	)

	mu.Lock()
	if len(requestBodies) != 2 {
		mu.Unlock()
		t.Fatalf("expected exactly 2 requests, got %d", len(requestBodies))
	}
	toolRequestBody := append([]byte(nil), requestBodies[0]...)
	finalRequestBody := append([]byte(nil), requestBodies[1]...)
	mu.Unlock()

	var firstReq map[string]any
	if err := json.Unmarshal(toolRequestBody, &firstReq); err != nil {
		t.Fatalf("unmarshal first request: %v", err)
	}
	if got, ok := firstReq["tool_choice"].(string); !ok || got != "required" {
		t.Fatalf("expected first request tool_choice=required, got %#v", firstReq["tool_choice"])
	}

	var secondReq map[string]any
	if err := json.Unmarshal(finalRequestBody, &secondReq); err != nil {
		t.Fatalf("unmarshal second request: %v", err)
	}
	if _, exists := secondReq["tool_choice"]; exists {
		t.Fatalf("expected second request to stop forcing tool_choice after one tool result, got %#v", secondReq["tool_choice"])
	}

	messages, ok := secondReq["messages"].([]any)
	if !ok {
		t.Fatalf("expected second request messages array, got %#v", secondReq["messages"])
	}

	var (
		sawToolCall bool
		sawToolMsg  bool
	)
	for _, item := range messages {
		message, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if role, _ := message["role"].(string); role == "assistant" {
			if toolCalls, ok := message["tool_calls"].([]any); ok && len(toolCalls) == 1 {
				call, _ := toolCalls[0].(map[string]any)
				function, _ := call["function"].(map[string]any)
				if name, _ := function["name"].(string); name == fakeToolName {
					if args, _ := function["arguments"].(string); args == "{\"stockCode\":\"600519.SH\"}" {
						sawToolCall = true
					}
				}
			}
		}

		if role, _ := message["role"].(string); role == "tool" {
			if content, _ := message["content"].(string); content == "fake tool output" {
				if callID, _ := message["tool_call_id"].(string); callID == "call_1" {
					sawToolMsg = true
				}
			}
		}
	}

	if !sawToolCall {
		t.Fatalf("expected second request to include full accumulated tool arguments")
	}
	if !sawToolMsg {
		t.Fatalf("expected second request to include tool result message")
	}
}
