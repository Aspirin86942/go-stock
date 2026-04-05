package logger

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPMiddleware_WritesRequestResponseAndPayloadRef(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	runtime.AttachPayloadStore(NewPayloadStore(t.TempDir(), 32, 5<<20))

	requestBody := strings.Repeat("q", 128)
	var seenBody string

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body in downstream handler: %v", err)
		}
		seenBody = string(body)
		_, _ = w.Write([]byte(strings.Repeat("x", 128)))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/chat/summary-stream", strings.NewReader(requestBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seenBody != requestBody {
		t.Fatalf("expected downstream handler to read original body, got %q", seenBody)
	}

	line := buf.String()
	for _, want := range []string{`"event":"http.request.completed"`, `"path":"/api/chat/summary-stream"`, `payload_file`} {
		if !strings.Contains(line, want) {
			t.Fatalf("expected middleware log to contain %s, got %s", want, line)
		}
	}
}

func TestHTTPMiddleware_PreservesFlusherForStreamingHandlers(t *testing.T) {
	runtime := newRuntimeForTest(io.Discard)

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatalf("expected middleware response writer to preserve http.Flusher")
		}
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected wrapped handler status 202, got %d", rec.Code)
	}
}
