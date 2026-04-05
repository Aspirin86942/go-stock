package logger

import (
	"context"
	"testing"
)

func TestTraceContextRoundTripThroughContext(t *testing.T) {
	initial := TraceContext{
		TraceID: "trace-123",
		SpanID:  "span-1",
		Source:  "test",
	}

	ctx := WithTraceContext(context.Background(), initial)
	if got, ok := TraceContextFromContext(ctx); !ok {
		t.Fatalf("expected trace context to be stored")
	} else if got != initial {
		t.Fatalf("expected trace %+v, got %+v", initial, got)
	}

	nilCtx := WithTraceContext(nil, initial)
	if got, ok := TraceContextFromContext(nilCtx); !ok {
		t.Fatalf("expected trace context persisted even when ctx was nil")
	} else if got != initial {
		t.Fatalf("expected trace %+v, got %+v", initial, got)
	}

	if _, ok := TraceContextFromContext(context.Background()); ok {
		t.Fatalf("expected empty context to not report a trace")
	}
}

func TestRuntimeEnsureTraceContext_ReusesExistingTrace(t *testing.T) {
	runtime := &Runtime{}
	existing := TraceContext{
		TraceID: "existing-trace",
		SpanID:  "existing-span",
		Source:  "initial",
	}
	ctx := WithTraceContext(context.Background(), existing)

	resultCtx, trace := runtime.EnsureTraceContext(ctx, "recycled")
	if trace != existing {
		t.Fatalf("expected reuse of %v, got %v", existing, trace)
	}
	if resultCtx != ctx {
		t.Fatalf("expected ensure to return original context, got new instance")
	}

	if stored, ok := TraceContextFromContext(resultCtx); !ok {
		t.Fatalf("expected context to still expose trace")
	} else if stored != existing {
		t.Fatalf("expected context to store %v, got %v", existing, stored)
	}
}

func TestEnsureTraceContextCreatesTraceForNilContext(t *testing.T) {
	runtime := &Runtime{}
	ctx, trace := runtime.EnsureTraceContext(nil, "nil-source")
	if ctx == nil {
		t.Fatalf("expected ensure to return a context")
	}
	if trace.TraceID == "" {
		t.Fatalf("expected new trace with non empty id")
	}
	if stored, ok := TraceContextFromContext(ctx); !ok {
		t.Fatalf("expected returned context to expose trace")
	} else if stored != trace {
		t.Fatalf("expected context to store %v, got %v", trace, stored)
	}
}

func TestTraceOrNewCreatesTraceWhenMissing(t *testing.T) {
	runtime := &Runtime{}
	trace := runtime.TraceOrNew(context.Background(), "fallback")
	if trace.TraceID == "" {
		t.Fatalf("expected TraceOrNew to generate a trace id")
	}

	if trace.Source != "fallback" {
		t.Fatalf("expected source fallback, got %s", trace.Source)
	}
}

func TestEnsureTraceContextRegeneratesTraceWhenTraceIDEmpty(t *testing.T) {
	runtime := &Runtime{}
	partial := TraceContext{
		TraceID: "",
		SpanID:  "span-empty",
		Source:  "origin",
	}
	ctx := WithTraceContext(context.Background(), partial)

	resultCtx, trace := runtime.EnsureTraceContext(ctx, "regen")
	if trace.TraceID == "" {
		t.Fatalf("expected new trace id, got empty")
	}
	if trace.TraceID == partial.TraceID {
		t.Fatalf("expected new trace different from the partial one")
	}
	if stored, ok := TraceContextFromContext(resultCtx); !ok {
		t.Fatalf("expected regenerated trace stored")
	} else if stored != trace {
		t.Fatalf("expected context to store %v, got %v", trace, stored)
	}
}
