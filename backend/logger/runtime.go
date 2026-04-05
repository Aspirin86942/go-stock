package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unicode"

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
	mu          sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.maxTotal > 0 {
		used, err := s.currentPayloadUsage()
		if err != nil {
			return PayloadRef{}, err
		}
		next := used + int64(len(body))
		if next > s.maxTotal {
			return PayloadRef{}, fmt.Errorf("payload exceeds maxTotal limit: used=%d payload=%d limit=%d", used, len(body), s.maxTotal)
		}
	}

	// 大 payload 落盘，避免主日志文件因请求/响应体过大而膨胀且难以检索。
	dir := filepath.Join(s.root, "payloads", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return PayloadRef{}, fmt.Errorf("ensure payload dir %s: %w", dir, err)
	}
	file := filepath.Join(dir, fmt.Sprintf("%s-%d.log", sanitizePayloadKind(kind), time.Now().UnixNano()))
	if err := os.WriteFile(file, body, 0o644); err != nil {
		return PayloadRef{}, fmt.Errorf("write payload spill file %s: %w", file, err)
	}
	ref.File = file
	return ref, nil
}

func (s *PayloadStore) currentPayloadUsage() (int64, error) {
	payloadRoot := filepath.Join(s.root, "payloads")
	if _, err := os.Stat(payloadRoot); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("stat payload root %s: %w", payloadRoot, err)
	}

	var total int64
	err := filepath.WalkDir(payloadRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk payload root %s: %w", payloadRoot, err)
	}
	return total, nil
}

func sanitizePayloadKind(kind string) string {
	trimmed := strings.TrimSpace(kind)
	if trimmed == "" {
		return "payload"
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, trimmed)
}

func (r *Runtime) NewTrace(source string) TraceContext {
	if r == nil {
		return TraceContext{
			TraceID: uuid.NewString(),
			SpanID:  uuid.NewString(),
			Source:  source,
		}
	}
	return TraceContext{
		TraceID:      uuid.NewString(),
		SpanID:       uuid.NewString(),
		AppSessionID: r.sessionID,
		Source:       source,
	}
}

func (r *Runtime) GoWithRecover(module, event string, fn func()) {
	trace := r.NewTrace("panic")
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				// 在 goroutine 边界兜底 recover，统一将 panic 打到独立 sink 便于审计与告警。
				r.ForSink(SinkPanic, module).WithTrace(trace).Error(
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
	if store == nil {
		return
	}
	r.payloads = store
}
