package db

import (
	"go-stock/backend/apppath"
	runtimelogger "go-stock/backend/logger"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Dao *gorm.DB

func defaultSQLiteDSN(stockDBPath string) string {
	return stockDBPath + "?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-524288"
}

func Init(sqlitePath string) {
	var openDb *gorm.DB
	var err error
	if sqlitePath == "" {
		paths, pathErr := apppath.Ensure()
		if pathErr != nil {
			log.Fatalf("init app paths error: %s", pathErr.Error())
		}
		sqlitePath = defaultSQLiteDSN(paths.StockDBPath)
	}
	openDb, err = gorm.Open(sqlite.Open(sqlitePath), newGormConfig(runtimelogger.Default()))

	if err != nil {
		log.Fatalf("db connection error is %s", err.Error())
	}

	// 兜底：确保 busy_timeout / WAL / synchronous 生效（不同驱动/DSN 参数支持可能存在差异）
	_ = openDb.Exec("PRAGMA busy_timeout=10000").Error
	_ = openDb.Exec("PRAGMA journal_mode=WAL").Error
	_ = openDb.Exec("PRAGMA synchronous=NORMAL").Error

	dbCon, err := openDb.DB()
	if err != nil {
		log.Fatalf("openDb.DB error is  %s", err.Error())
	}
	// SQLite 写入是串行锁模型：连接开太多会放大锁竞争导致 SQLITE_BUSY
	dbCon.SetMaxIdleConns(1)
	dbCon.SetMaxOpenConns(5)
	dbCon.SetConnMaxLifetime(time.Hour)
	Dao = openDb
	AutoMigrate()
}

func newGormConfig(runtime *runtimelogger.Runtime) *gorm.Config {
	return &gorm.Config{
		Logger:                                   runtimelogger.NewGormLogger(runtime, 3*time.Second),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
	}
}
