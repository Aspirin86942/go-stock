package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-stock/backend/db"
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
