package logger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type responseRecorder struct {
	base        http.ResponseWriter
	status      int
	wroteHeader bool
	capture     *responseCapture
	captureErr  error
}

type flushableResponseRecorder struct {
	*responseRecorder
	flusher http.Flusher
}

type responseCapture struct {
	store   *PayloadStore
	kind    string
	inline  bytes.Buffer
	hasher  hash.Hash
	size    int
	file    *os.File
	path    string
	spilled bool
}

func newResponseRecorder(base http.ResponseWriter, store *PayloadStore, kind string) *responseRecorder {
	return &responseRecorder{
		base:    base,
		status:  http.StatusOK,
		capture: newResponseCapture(store, kind),
	}
}

func wrapResponseWriter(recorder *responseRecorder, base http.ResponseWriter) http.ResponseWriter {
	flusher, ok := base.(http.Flusher)
	if !ok {
		return recorder
	}
	return &flushableResponseRecorder{
		responseRecorder: recorder,
		flusher:          flusher,
	}
}

func newResponseCapture(store *PayloadStore, kind string) *responseCapture {
	return &responseCapture{
		store:  store,
		kind:   kind,
		hasher: sha256.New(),
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.base.Header()
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.wroteHeader = true
	r.base.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	written, err := r.base.Write(body)
	if written > 0 {
		r.record(body[:written])
	}
	return written, err
}

func (r *responseRecorder) record(body []byte) {
	if r.captureErr != nil || r.capture == nil {
		return
	}
	if err := r.capture.Write(body); err != nil {
		r.captureErr = err
	}
}

func (r *responseRecorder) markStreaming() {
	if r.captureErr != nil || r.capture == nil {
		return
	}
	if err := r.capture.ForceSpill(); err != nil {
		r.captureErr = err
	}
}

func (r *responseRecorder) PayloadRef() (PayloadRef, error) {
	if r.capture == nil {
		return PayloadRef{}, r.captureErr
	}
	ref, err := r.capture.PayloadRef()
	if r.captureErr != nil {
		if err != nil {
			return ref, fmt.Errorf("%v; finalize payload ref: %w", r.captureErr, err)
		}
		return ref, r.captureErr
	}
	return ref, err
}

func (r *flushableResponseRecorder) Flush() {
	r.markStreaming()
	r.flusher.Flush()
}

func (c *responseCapture) Write(body []byte) error {
	if len(body) == 0 {
		return nil
	}

	if _, err := c.hasher.Write(body); err != nil {
		return fmt.Errorf("hash response payload: %w", err)
	}

	nextSize := c.size + len(body)
	if !c.spilled && nextSize > c.inlineLimit() {
		if err := c.startSpill(); err != nil {
			return err
		}
	}

	if c.spilled {
		if _, err := c.file.Write(body); err != nil {
			return fmt.Errorf("append spilled payload: %w", err)
		}
	} else {
		if _, err := c.inline.Write(body); err != nil {
			return fmt.Errorf("buffer inline payload: %w", err)
		}
	}

	c.size = nextSize
	return nil
}

func (c *responseCapture) ForceSpill() error {
	if c.spilled {
		return nil
	}
	return c.startSpill()
}

func (c *responseCapture) PayloadRef() (PayloadRef, error) {
	ref := PayloadRef{
		SHA256: hex.EncodeToString(c.hasher.Sum(nil)),
		Size:   c.size,
	}

	if c.spilled {
		if c.file != nil {
			if err := c.file.Close(); err != nil {
				return ref, fmt.Errorf("close spilled payload file %s: %w", c.path, err)
			}
			c.file = nil
		}
		ref.File = c.path
		return ref, nil
	}

	ref.Inline = c.inline.String()
	return ref, nil
}

func (c *responseCapture) BufferedLen() int {
	return c.inline.Len()
}

func (c *responseCapture) HasSpilled() bool {
	return c.spilled
}

func (c *responseCapture) inlineLimit() int {
	if c.store != nil && c.store.inlineLimit > 0 {
		return c.store.inlineLimit
	}
	return defaultPayloadInlineLimit
}

func (c *responseCapture) startSpill() error {
	if c.spilled {
		return nil
	}

	file, path, err := c.createSpillFile()
	if err != nil {
		return err
	}
	if _, err := file.Write(c.inline.Bytes()); err != nil {
		_ = file.Close()
		return fmt.Errorf("spill inline payload to %s: %w", path, err)
	}

	c.inline.Reset()
	c.file = file
	c.path = path
	c.spilled = true
	return nil
}

func (c *responseCapture) createSpillFile() (*os.File, string, error) {
	if c.store == nil {
		file, err := os.CreateTemp("", sanitizePayloadKind(c.kind)+"-*.log")
		if err != nil {
			return nil, "", fmt.Errorf("create temp spill file: %w", err)
		}
		return file, file.Name(), nil
	}

	dir := filepath.Join(c.store.root, "payloads", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", fmt.Errorf("ensure payload dir %s: %w", dir, err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%s-%d.log", sanitizePayloadKind(c.kind), time.Now().UnixNano()))
	file, err := os.Create(path)
	if err != nil {
		return nil, "", fmt.Errorf("create spill file %s: %w", path, err)
	}
	return file, path, nil
}

func (rt *Runtime) HTTPMiddleware(module string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		recorder := newResponseRecorder(w, rt.payloads, "http-response")
		writer := wrapResponseWriter(recorder, w)

		requestBody, readErr := readRequestBody(req)
		if readErr != nil {
			rt.ForSink(SinkHTTP, module).WithTrace(rt.NewTrace("http")).Warn(
				"http.request.body.read_failed",
				"failed to read http request body",
				String("method", req.Method),
				String("path", req.URL.Path),
				String("error_message", readErr.Error()),
			)
		}

		next.ServeHTTP(writer, req)

		responseRef, responseErr := recorder.PayloadRef()
		fields := []zap.Field{
			String("method", req.Method),
			String("path", req.URL.Path),
			Int("status_code", recorder.status),
			Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		fields = append(fields, payloadFields("request_payload", rt.savePayload("http-request", requestBody))...)
		fields = append(fields, payloadFields("response_payload", responseRef)...)
		if responseErr != nil {
			fields = append(fields, String("response_payload_error", responseErr.Error()))
		}

		rt.ForSink(SinkHTTP, module).WithTrace(rt.NewTrace("http")).Info(
			"http.request.completed",
			"handled http request",
			fields...,
		)
	})
}

func readRequestBody(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func (rt *Runtime) savePayload(kind string, body []byte) PayloadRef {
	if rt == nil || rt.payloads == nil {
		return PayloadRef{Size: len(body)}
	}
	ref, err := rt.payloads.Save(kind, body)
	if err != nil {
		return PayloadRef{
			Inline: err.Error(),
			Size:   len(body),
		}
	}
	return ref
}

func payloadFields(prefix string, ref PayloadRef) []zap.Field {
	fields := []zap.Field{
		Int(fmt.Sprintf("%s_size", prefix), ref.Size),
	}
	if ref.SHA256 != "" {
		fields = append(fields, String(fmt.Sprintf("%s_sha256", prefix), ref.SHA256))
	}
	if ref.File != "" {
		fields = append(fields, String(fmt.Sprintf("%s_file", prefix), ref.File))
	}
	if ref.Inline != "" {
		fields = append(fields, String(fmt.Sprintf("%s_inline", prefix), ref.Inline))
	}
	return fields
}
