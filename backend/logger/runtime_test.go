package logger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestPayloadStore_SpillsLargeBodiesToPayloadDir(t *testing.T) {
	store := NewPayloadStore(filepath.Join(t.TempDir(), "logs"), 32, 5<<20)

	ref, err := store.Save("http-request", bytes.Repeat([]byte("x"), 128))
	if err != nil {
		t.Fatalf("save payload: %v", err)
	}
	if ref.File == "" {
		t.Fatalf("expected payload spill file, got inline payload %#v", ref)
	}
	if _, err := os.Stat(ref.File); err != nil {
		t.Fatalf("expected spill file to exist: %v", err)
	}
}

func TestPayloadStore_InlineSmallPayloadWithDigestAndSize(t *testing.T) {
	body := []byte("ok")
	store := NewPayloadStore(filepath.Join(t.TempDir(), "logs"), 32, 5<<20)

	ref, err := store.Save("http-request", body)
	if err != nil {
		t.Fatalf("save payload: %v", err)
	}
	if ref.File != "" {
		t.Fatalf("expected inline payload only, got file=%q", ref.File)
	}
	if ref.Inline != "ok" {
		t.Fatalf("expected inline payload to be preserved, got %q", ref.Inline)
	}
	if ref.Size != len(body) {
		t.Fatalf("expected size=%d, got %d", len(body), ref.Size)
	}
	wantHash := sha256.Sum256(body)
	if ref.SHA256 != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("expected sha256=%s, got %s", hex.EncodeToString(wantHash[:]), ref.SHA256)
	}
}

func TestPayloadStore_ReturnsErrorWhenTotalWouldExceedMaxTotal(t *testing.T) {
	store := NewPayloadStore(filepath.Join(t.TempDir(), "logs"), 16, 200)

	if _, err := store.Save("http-request", bytes.Repeat([]byte("a"), 128)); err != nil {
		t.Fatalf("save first payload: %v", err)
	}
	_, err := store.Save("http-request", bytes.Repeat([]byte("b"), 100))
	if err == nil {
		t.Fatalf("expected maxTotal overflow error, got nil")
	}
	if !strings.Contains(err.Error(), "maxTotal") {
		t.Fatalf("expected maxTotal error, got %v", err)
	}
}

func TestGoWithRecover_WritesPanicEvent(t *testing.T) {
	buf := &lockedBuffer{}
	runtime := newRuntimeForTest(buf)
	runtime.sessionID = "session-test"

	runtime.GoWithRecover("panic", "logger-test", func() {
		panic("boom")
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line := buf.String()
		if strings.Contains(line, `"event":"logger-test"`) &&
			strings.Contains(line, `"error_class":"panic"`) &&
			strings.Contains(line, `"app_session_id":"session-test"`) &&
			strings.Contains(line, `"panic_value":"boom"`) &&
			strings.Contains(line, `"stack":"`) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected panic event in panic sink, got %s", buf.String())
}
