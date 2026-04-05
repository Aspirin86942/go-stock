package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	runtime.AttachPayloadStore(NewPayloadStore(t.TempDir(), 32, 5<<20))

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatalf("expected middleware response writer to preserve http.Flusher")
		}
		w.WriteHeader(http.StatusAccepted)
		for i := 0; i < 4; i++ {
			_, _ = w.Write([]byte(strings.Repeat("s", 24)))
			w.(http.Flusher).Flush()
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected wrapped handler status 202, got %d", rec.Code)
	}

	line := buf.String()
	if !strings.Contains(line, "response_payload_file") {
		t.Fatalf("expected streaming response to spill payload to file, got %s", line)
	}
}

func TestHTTPMiddleware_DoesNotExposeFlusherWhenUnderlyingWriterDoesNotSupportIt(t *testing.T) {
	runtime := newRuntimeForTest(io.Discard)

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); ok {
			t.Fatalf("expected middleware not to expose http.Flusher when base writer lacks it")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := newBasicResponseWriter()
	handler.ServeHTTP(rec, req)

	if rec.status != http.StatusNoContent {
		t.Fatalf("expected wrapped handler status 204, got %d", rec.status)
	}
}

func TestResponseRecorder_SpillsStreamedPayloadWithoutKeepingWholeBuffer(t *testing.T) {
	store := NewPayloadStore(t.TempDir(), 32, 5<<20)
	base := httptest.NewRecorder()
	recorder := newResponseRecorder(base, store, "http-response")
	writer := wrapResponseWriter(recorder, base)

	_, err := writer.Write([]byte(strings.Repeat("a", 16)))
	if err != nil {
		t.Fatalf("write first chunk: %v", err)
	}
	if got := recorder.capture.BufferedLen(); got != 16 {
		t.Fatalf("expected first chunk to stay inline, got %d bytes", got)
	}

	writer.(http.Flusher).Flush()
	if !recorder.capture.HasSpilled() {
		t.Fatalf("expected flush to start streaming spill")
	}
	if got := recorder.capture.BufferedLen(); got != 0 {
		t.Fatalf("expected inline buffer to drain after spill starts, got %d", got)
	}

	_, err = writer.Write([]byte(strings.Repeat("b", 256)))
	if err != nil {
		t.Fatalf("write streamed chunk: %v", err)
	}
	if got := recorder.capture.BufferedLen(); got != 0 {
		t.Fatalf("expected recorder not to retain streamed payload in memory, got %d", got)
	}

	ref, err := recorder.PayloadRef()
	if err != nil {
		t.Fatalf("finalize payload ref: %v", err)
	}
	if ref.File == "" {
		t.Fatalf("expected streamed payload to spill to file, got %#v", ref)
	}

	body, err := os.ReadFile(ref.File)
	if err != nil {
		t.Fatalf("read spilled response payload: %v", err)
	}
	if len(body) != 272 {
		t.Fatalf("expected spilled payload length 272, got %d", len(body))
	}
}

func TestHTTPMiddleware_StreamedPayloadHonorsMaxTotal(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)
	runtime.AttachPayloadStore(NewPayloadStore(t.TempDir(), 16, 32))

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("expected middleware to preserve flusher")
		}

		_, _ = w.Write([]byte(strings.Repeat("a", 8)))
		flusher.Flush()
		_, _ = w.Write([]byte(strings.Repeat("b", 40)))
		flusher.Flush()
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/chat/summary-stream", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	line := buf.String()
	if !strings.Contains(line, "response_payload_error") || !strings.Contains(line, "maxTotal") {
		t.Fatalf("expected maxTotal overflow to be logged, got %s", line)
	}
	if strings.Contains(line, "response_payload_file") {
		t.Fatalf("expected overflowed streamed payload not to expose spill file, got %s", line)
	}
}

func TestHTTPMiddleware_ReusesRequestContextTrace(t *testing.T) {
	buf := &bytes.Buffer{}
	runtime := newRuntimeForTest(buf)

	handler := runtime.HTTPMiddleware("ai-assistant-web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace, ok := TraceContextFromContext(r.Context())
		if !ok || strings.TrimSpace(trace.TraceID) == "" {
			t.Fatalf("expected request context to carry trace, got %#v", trace)
		}

		runtime.ForSink(SinkHTTP, "ai-assistant-web").WithTrace(trace).Info(
			"http.request.handler_trace",
			"handler observed request trace",
		)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/chat/summary-stream", nil)
	req.Body = errReader{}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entries := parseJSONLogEntries(t, buf.String())
	traceByEvent := map[string]string{}
	for _, entry := range entries {
		event, _ := entry["event"].(string)
		traceID, _ := entry["trace_id"].(string)
		if event == "http.request.body.read_failed" || event == "http.request.handler_trace" || event == "http.request.completed" {
			traceByEvent[event] = traceID
		}
	}

	for _, event := range []string{"http.request.body.read_failed", "http.request.handler_trace", "http.request.completed"} {
		traceID := strings.TrimSpace(traceByEvent[event])
		if traceID == "" {
			t.Fatalf("expected event %s to have non-empty trace_id, got %#v", event, traceByEvent)
		}
	}
	if traceByEvent["http.request.body.read_failed"] != traceByEvent["http.request.completed"] {
		t.Fatalf("expected read_failed and completed to reuse trace_id, got %#v", traceByEvent)
	}
	if traceByEvent["http.request.handler_trace"] != traceByEvent["http.request.completed"] {
		t.Fatalf("expected handler context trace to match completed trace, got %#v", traceByEvent)
	}
}

func parseJSONLogEntries(t *testing.T, content string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(content), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(trimmed), &entry); err != nil {
			t.Fatalf("parse json log entry %q: %v", trimmed, err)
		}
		entries = append(entries, entry)
	}
	return entries
}

type basicResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBasicResponseWriter() *basicResponseWriter {
	return &basicResponseWriter{header: make(http.Header)}
}

func (w *basicResponseWriter) Header() http.Header {
	return w.header
}

func (w *basicResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(body)
}

func (w *basicResponseWriter) WriteHeader(status int) {
	w.status = status
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func (errReader) Close() error {
	return nil
}
