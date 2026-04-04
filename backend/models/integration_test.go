package models_test

import (
	"os"
	"testing"
)

func requireIntegrationTest(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("GO_STOCK_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("skipping integration test; set GO_STOCK_RUN_INTEGRATION_TESTS=1 to enable")
	}
}
