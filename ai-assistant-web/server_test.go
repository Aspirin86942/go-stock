package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-stock/backend/db"
	"go-stock/internal/testenv"
)

func TestRequireVip2_AllowsRequestsWithoutSponsorCode(t *testing.T) {
	db.Init(fmt.Sprintf("%s/stock.db", t.TempDir()))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	recorder := httptest.NewRecorder()

	if ok := requireVip2(recorder); !ok {
		t.Fatalf("expected requireVip2 to allow requests without sponsor code")
	}
}

func TestVipStatus_ReturnsOpenAccess(t *testing.T) {
	db.Init(fmt.Sprintf("%s/stock.db", t.TempDir()))
	sqlDB, err := db.Dao.DB()
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/vip-status", nil)
	recorder := httptest.NewRecorder()

	(&app{}).vipStatus(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal vip status response: %v", err)
	}

	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %#v", payload["ok"])
	}

	if active, _ := payload["active"].(bool); !active {
		t.Fatalf("expected active=true, got %#v", payload["active"])
	}
}

func TestNewHandler_HealthRouteStillWorksThroughLoggingMiddleware(t *testing.T) {
	assertHealthRouteStillWorksThroughLoggingMiddleware(t, "ai-assistant-web")
}

func TestReleaseSmoke_HealthRouteStillWorksThroughLoggingMiddleware(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	assertHealthRouteStillWorksThroughLoggingMiddleware(t, "release-smoke-ai-assistant-web")
}

func assertHealthRouteStillWorksThroughLoggingMiddleware(t *testing.T, suite string) {
	t.Helper()

	runtime, artifacts := testenv.NewLoggerRuntime(t, suite)

	handler, err := newHandler(runtime)
	if err != nil {
		t.Fatalf("build server handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS header to be preserved, got %q", got)
	}

	httpLog, err := os.ReadFile(filepath.Join(artifacts.LogsDir, "http.log"))
	if err != nil {
		t.Fatalf("read http log: %v", err)
	}
	if !strings.Contains(string(httpLog), `"path":"/api/health"`) {
		t.Fatalf("expected health request to be logged through middleware, got %s", string(httpLog))
	}
	if !strings.Contains(string(httpLog), `"execution_mode":"test"`) {
		t.Fatalf("expected health request log to carry test metadata, got %s", string(httpLog))
	}
	if !strings.Contains(string(httpLog), `"status_code":200`) {
		t.Fatalf("expected health request log to record 200 status, got %s", string(httpLog))
	}
}
