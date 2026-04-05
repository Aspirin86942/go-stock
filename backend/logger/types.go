package logger

import (
	"go-stock/backend/apppath"
	"sync/atomic"

	"go.uber.org/zap"
)

type Sink string

const (
	SinkApp      Sink = "app"
	SinkError    Sink = "error"
	SinkHTTP     Sink = "http"
	SinkAI       Sink = "ai"
	SinkTask     Sink = "task"
	SinkDB       Sink = "db"
	SinkFrontend Sink = "frontend"
	SinkPanic    Sink = "panic"
)

type TraceContext struct {
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	AppSessionID string `json:"app_session_id"`
	Source       string `json:"source"`
}

type Runtime struct {
	sinks map[Sink]*zap.Logger
}

type Config struct {
	Paths        apppath.Paths
	EnableStdout bool
}

type ModuleLogger struct {
	runtime *Runtime
	sink    Sink
	module  string
	trace   TraceContext
}

var defaultRuntime atomic.Pointer[Runtime]
