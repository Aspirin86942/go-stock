package db

import (
	"fmt"
	"go-stock/backend/apppath"
	runtimelogger "go-stock/backend/logger"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Dao *gorm.DB

func defaultSQLiteDSN(stockDBPath string) string {
	return stockDBPath + "?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-524288"
}

func Init(sqlitePath string) {
	runtime := runtimelogger.Default()
	var openDb *gorm.DB
	var err error
	if sqlitePath == "" {
		paths, pathErr := apppath.Ensure()
		if pathErr != nil {
			panicInit(runtime, "db.init.paths_failed", "init app paths failed", pathErr)
		}
		sqlitePath = defaultSQLiteDSN(paths.StockDBPath)
	}
	openDb, err = gorm.Open(sqlite.Open(sqlitePath), newGormConfig(runtime))

	if err != nil {
		panicInit(runtime, "db.init.open_failed", "open sqlite connection failed", err)
	}

	// 兜底：确保 busy_timeout / WAL / synchronous 生效（不同驱动/DSN 参数支持可能存在差异）
	logInitWarning(runtime, "db.init.pragma_busy_timeout_failed", "set sqlite busy_timeout failed", openDb.Exec("PRAGMA busy_timeout=10000").Error)
	logInitWarning(runtime, "db.init.pragma_journal_mode_failed", "set sqlite journal_mode failed", openDb.Exec("PRAGMA journal_mode=WAL").Error)
	logInitWarning(runtime, "db.init.pragma_synchronous_failed", "set sqlite synchronous failed", openDb.Exec("PRAGMA synchronous=NORMAL").Error)

	dbCon, err := openDb.DB()
	if err != nil {
		panicInit(runtime, "db.init.sql_db_failed", "open sql.DB handle failed", err)
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

func panicInit(runtime *runtimelogger.Runtime, event string, message string, err error) {
	if err == nil {
		return
	}

	if runtime == nil {
		runtime = runtimelogger.Default()
	}

	wrappedErr := fmt.Errorf("%s: %w", message, err)
	if runtime != nil {
		runtime.ForSink(runtimelogger.SinkDB, "db").WithTrace(runtime.NewTrace("db")).Error(
			event,
			message,
			runtimelogger.String("error_message", err.Error()),
			runtimelogger.Err(err),
		)
	}
	panic(wrappedErr)
}

func logInitWarning(runtime *runtimelogger.Runtime, event string, message string, err error) {
	if err == nil {
		return
	}

	if runtime == nil {
		runtime = runtimelogger.Default()
	}
	if runtime == nil {
		return
	}

	runtime.ForSink(runtimelogger.SinkDB, "db").WithTrace(runtime.NewTrace("db")).Warn(
		event,
		message,
		runtimelogger.String("error_message", err.Error()),
		runtimelogger.Err(err),
	)
}
