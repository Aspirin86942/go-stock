package logger

import "context"

type traceContextKey struct{}

func WithTraceContext(ctx context.Context, trace TraceContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceContextKey{}, trace)
}

func TraceContextFromContext(ctx context.Context) (TraceContext, bool) {
	if ctx == nil {
		return TraceContext{}, false
	}
	val := ctx.Value(traceContextKey{})
	if val == nil {
		return TraceContext{}, false
	}
	trace, ok := val.(TraceContext)
	return trace, ok
}

func (r *Runtime) TraceOrNew(ctx context.Context, source string) TraceContext {
	if trace, ok := TraceContextFromContext(ctx); ok && trace.TraceID != "" {
		return trace
	}
	return r.NewTrace(source)
}

func (r *Runtime) EnsureTraceContext(ctx context.Context, source string) (context.Context, TraceContext) {
	if ctx == nil {
		ctx = context.Background()
	}
	existing, hasExisting := TraceContextFromContext(ctx)
	trace := r.TraceOrNew(ctx, source)
	if hasExisting && trace == existing {
		return ctx, trace
	}
	return WithTraceContext(ctx, trace), trace
}
