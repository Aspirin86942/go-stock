package data

import (
	"encoding/json"
	"go-stock/internal/testenv"
	"strings"
	"testing"
)

func logValue(t *testing.T, label string, value any) {
	t.Helper()

	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", label, err)
	}
	t.Logf("%s:\n%s", label, string(payload))
}

func requirePositiveLen(t *testing.T, label string, size int) {
	t.Helper()

	if size <= 0 {
		t.Fatalf("expected %s to contain data", label)
	}
}

func requireNotBlank(t *testing.T, label, value string) {
	t.Helper()

	if strings.TrimSpace(value) == "" {
		t.Fatalf("expected %s to be non-blank", label)
	}
}

func requireManualTest(t *testing.T) {
	t.Helper()

	testenv.RequireManualTest(t)
}
