package logger

import (
	"fmt"
	"go-stock/backend/apppath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var CoreLogger *zap.Logger
var SugaredLogger *zap.SugaredLogger

func init() {
	InitLogger()
}

func InitLogger() {
	paths, err := apppath.Ensure()
	if err != nil {
		panic(fmt.Sprintf("init app paths for logger: %v", err))
	}

	runtime := MustInit(DefaultConfig(paths))
	CoreLogger = runtime.getSinkLogger(SinkApp)
	initLegacyGlobals(runtime)
}

func initLegacyGlobals(runtime *Runtime) {
	CoreLogger = runtime.getSinkLogger(SinkApp)
	SugaredLogger = newLegacyShimLogger(runtime).Sugar()
}

func newLegacyShimLogger(runtime *Runtime) *zap.Logger {
	core := &legacyShimCore{
		appCore:   runtime.getSinkLogger(SinkApp).Core(),
		errorCore: runtime.getSinkLogger(SinkError).Core(),
	}
	return zap.New(core, zap.AddCaller())
}

type legacyShimCore struct {
	appCore   zapcore.Core
	errorCore zapcore.Core
}

func (c *legacyShimCore) Enabled(level zapcore.Level) bool {
	return c.routeCore(level).Enabled(level)
}

func (c *legacyShimCore) With(fields []zapcore.Field) zapcore.Core {
	return &legacyShimCore{
		appCore:   c.appCore.With(fields),
		errorCore: c.errorCore.With(fields),
	}
}

func (c *legacyShimCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	core := c.routeCore(entry.Level)
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *legacyShimCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return c.routeCore(entry.Level).Write(entry, fields)
}

func (c *legacyShimCore) Sync() error {
	appErr := c.appCore.Sync()
	errorErr := c.errorCore.Sync()
	if appErr != nil {
		return appErr
	}
	return errorErr
}

func (c *legacyShimCore) routeCore(level zapcore.Level) zapcore.Core {
	if level >= zapcore.ErrorLevel {
		return c.errorCore
	}
	return c.appCore
}
