package agent

import (
	"context"
	"testing"

	"go-stock/backend/logger"
)

func TestModuleTrace_ReusesContextTrace(t *testing.T) {
	existing := logger.TraceContext{
		TraceID:      "trace-agent-1",
		SpanID:       "span-agent-1",
		AppSessionID: "session-agent-1",
		Source:       "http",
	}
	ctx := logger.WithTraceContext(context.Background(), existing)

	trace := moduleTrace(ctx, "agent-test")
	if trace != existing {
		t.Fatalf("expected moduleTrace to reuse %+v, got %+v", existing, trace)
	}
}
