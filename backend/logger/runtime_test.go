package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestGoWithRecover_WritesPanicEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)

	runtime.GoWithRecover("panic", "logger-test", func() {
		panic("boom")
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line := buf.String()
		if strings.Contains(line, `"event":"logger-test"`) && strings.Contains(line, `"error_class":"panic"`) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected panic event in panic sink, got %s", buf.String())
}
