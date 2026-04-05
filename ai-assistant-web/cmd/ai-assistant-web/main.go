package main

import (
	assistantweb "go-stock/ai-assistant-web"
	"go-stock/backend/logger"
	"os"
)

func main() {
	if err := assistantweb.Start(); err != nil {
		runtimeLogger := logger.Default()
		if runtimeLogger != nil {
			runtimeLogger.ForSink(logger.SinkApp, "ai-assistant-web").
				WithTrace(runtimeLogger.NewTrace("bootstrap")).
				Error(
					"startup.failed",
					"ai-assistant-web start failed",
					logger.Err(err),
				)
		}
		os.Exit(1)
	}
}
