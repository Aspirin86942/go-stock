package logger

import (
	"fmt"
	"go-stock/backend/apppath"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func DefaultConfig(paths apppath.Paths) Config {
	return Config{
		Paths:        paths,
		EnableStdout: true,
	}
}

func MustInit(cfg Config) *Runtime {
	if cfg.Paths.LogsDir == "" {
		paths, err := apppath.Ensure()
		if err != nil {
			panic(fmt.Sprintf("ensure app paths for logger runtime: %v", err))
		}
		cfg.Paths = paths
	}

	runtime := &Runtime{
		sessionID: uuid.NewString(),
		sinks:     make(map[Sink]*zap.Logger),
	}
	if err := runtime.bootstrapSinks(cfg); err != nil {
		panic(fmt.Sprintf("bootstrap logger sinks: %v", err))
	}
	runtime.AttachPayloadStore(newDefaultPayloadStore(cfg.Paths))

	defaultRuntime.Store(runtime)
	return runtime
}

func Default() *Runtime {
	return defaultRuntime.Load()
}
