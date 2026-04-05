package db

import (
	"strings"
	"testing"

	runtimelogger "go-stock/backend/logger"
	gormlogger "gorm.io/gorm/logger"
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

func TestNewGormConfig_UsesStructuredLogger(t *testing.T) {
	cfg := newGormConfig(runtimelogger.Default())

	gormLog, ok := cfg.Logger.(*runtimelogger.GormLogger)
	if !ok {
		t.Fatalf("expected custom structured gorm logger, got %T", cfg.Logger)
	}
	if gormLog.LogMode(gormlogger.Warn) == nil {
		t.Fatalf("expected gorm logger LogMode to return a logger")
	}
}
