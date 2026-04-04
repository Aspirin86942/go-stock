package db

import (
	"strings"
	"testing"
)

func TestDefaultSQLiteDSN_UsesProvidedStockDBPath(t *testing.T) {
	dsn := defaultSQLiteDSN(`C:\Users\Test\AppData\Local\go-stock\data\stock.db`)

	if !strings.Contains(dsn, `C:\Users\Test\AppData\Local\go-stock\data\stock.db`) {
		t.Fatalf("expected DSN to contain app-root stock db path, got %q", dsn)
	}
	if !strings.Contains(dsn, "_journal_mode=WAL") {
		t.Fatalf("expected DSN to keep WAL setting, got %q", dsn)
	}
}
