package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type PayloadRef struct {
	File   string `json:"file,omitempty"`
	Inline string `json:"inline,omitempty"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

type PayloadStore struct {
	root        string
	inlineLimit int
	maxTotal    int64
}

func NewPayloadStore(root string, inlineLimit int, maxTotal int64) *PayloadStore {
	return &PayloadStore{
		root:        root,
		inlineLimit: inlineLimit,
		maxTotal:    maxTotal,
	}
}

func (s *PayloadStore) Save(kind string, body []byte) (PayloadRef, error) {
	sum := sha256.Sum256(body)
	ref := PayloadRef{
		SHA256: hex.EncodeToString(sum[:]),
		Size:   len(body),
	}
	if len(body) <= s.inlineLimit {
		ref.Inline = string(body)
		return ref, nil
	}
	if s.maxTotal > 0 && int64(len(body)) > s.maxTotal {
		return PayloadRef{}, fmt.Errorf("payload exceeds maxTotal limit: payload=%d limit=%d", len(body), s.maxTotal)
	}

	// 大 payload 落盘，避免主日志文件因请求/响应体过大而膨胀且难以检索。
	dir := filepath.Join(s.root, "payloads", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return PayloadRef{}, fmt.Errorf("ensure payload dir %s: %w", dir, err)
	}
	file := filepath.Join(dir, fmt.Sprintf("%s-%d.log", kind, time.Now().UnixNano()))
	if err := os.WriteFile(file, body, 0o644); err != nil {
		return PayloadRef{}, fmt.Errorf("write payload spill file %s: %w", file, err)
	}
	ref.File = file
	return ref, nil
}

func (r *Runtime) NewTrace(source string) TraceContext {
	return TraceContext{
		TraceID:      uuid.NewString(),
		SpanID:       uuid.NewString(),
		AppSessionID: r.sessionID,
		Source:       source,
	}
}

func (r *Runtime) GoWithRecover(module, event string, fn func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				// 在 goroutine 边界兜底 recover，统一将 panic 打到独立 sink 便于审计与告警。
				r.ForSink(SinkPanic, module).Error(
					event,
					"goroutine panic recovered",
					String("error_class", "panic"),
					Any("panic_value", recovered),
					String("stack", string(debug.Stack())),
				)
			}
		}()
		fn()
	}()
}

func (r *Runtime) AttachPayloadStore(store *PayloadStore) {
	r.payloads = store
}
