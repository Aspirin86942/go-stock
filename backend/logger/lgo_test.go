package logger

import (
	"path/filepath"
	"testing"

	"go-stock/backend/apppath"
)

func TestBuildLogPaths_UsesSharedRuntimeRoot(t *testing.T) {
	root := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)
	paths := apppath.Paths{
		RootDir:      root,
		LogsDir:      filepath.Join(root, "logs"),
		WailsLogPath: filepath.Join(root, "logs", "wails.log"),
		InfoLogPath:  filepath.Join(root, "logs", "info.log"),
		ErrorLogPath: filepath.Join(root, "logs", "error.log"),
	}

	infoPath, errorPath := buildLogPaths(paths)

	if infoPath != filepath.Join(root, "logs", "info.log") {
		t.Fatalf("expected info log under runtime root, got %q", infoPath)
	}
	if errorPath != filepath.Join(root, "logs", "error.log") {
		t.Fatalf("expected error log under runtime root, got %q", errorPath)
	}
}
