package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	_, _ = r.body.Write(body)
	return r.ResponseWriter.Write(body)
}

func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rt *Runtime) HTTPMiddleware(module string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

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

		next.ServeHTTP(recorder, req)

		fields := []zap.Field{
			String("method", req.Method),
			String("path", req.URL.Path),
			Int("status_code", recorder.status),
			Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		fields = append(fields, payloadFields("request_payload", rt.savePayload("http-request", requestBody))...)
		fields = append(fields, payloadFields("response_payload", rt.savePayload("http-response", recorder.body.Bytes()))...)

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
