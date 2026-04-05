package logger

import (
	"fmt"
	"go-stock/backend/apppath"

	"go.uber.org/zap"
)

var CoreLogger *zap.Logger
var SugaredLogger *zap.SugaredLogger
var Logger *zap.Logger

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
	Logger = CoreLogger
	SugaredLogger = Logger.Sugar()
}
