package logger

import (
	"go-stock/backend/apppath"
	"io"
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
	sessionID string
	sinks     map[Sink]*zap.Logger
	payloads  *PayloadStore
	closers   []io.Closer
}

type Config struct {
	Paths        apppath.Paths
	EnableStdout bool
	Fields       []zap.Field
}

// GlobalState captures the mutable package-level logger globals so test helpers
// can temporarily replace them and restore the previous state during cleanup.
type GlobalState struct {
	Runtime *Runtime
	Core    *zap.Logger
	Sugared *zap.SugaredLogger
}

func CaptureGlobalState() GlobalState {
	return GlobalState{
		Runtime: Default(),
		Core:    CoreLogger,
		Sugared: SugaredLogger,
	}
}

func (s GlobalState) Restore() {
	defaultRuntime.Store(s.Runtime)
	CoreLogger = s.Core
	SugaredLogger = s.Sugared
}

type Logger struct {
	runtime *Runtime
	sink    Sink
	module  string
	trace   TraceContext
}

var defaultRuntime atomic.Pointer[Runtime]
