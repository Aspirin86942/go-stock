package apppath

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootDir_WindowsUsesLocalAppData(t *testing.T) {
	root, err := resolveRootDir("windows", func(key string) string {
		if key == "LOCALAPPDATA" {
			return `C:\Users\Test\AppData\Local`
		}
		return ""
	}, func() (string, error) {
		return "", errors.New("user config dir should not be called on windows with LOCALAPPDATA")
	})
	if err != nil {
		t.Fatalf("resolveRootDir returned error: %v", err)
	}

	want := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)
	if root != want {
		t.Fatalf("expected root %q, got %q", want, root)
	}
}

func TestBuildPaths_UsesExpectedFilesUnderRoot(t *testing.T) {
	root := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)

	paths := buildPaths(root)

	if paths.RootDir != root {
		t.Fatalf("expected RootDir %q, got %q", root, paths.RootDir)
	}
	if paths.DataDir != filepath.Join(root, "data") {
		t.Fatalf("expected DataDir under root, got %q", paths.DataDir)
	}
	if paths.LogsDir != filepath.Join(root, "logs") {
		t.Fatalf("expected LogsDir under root, got %q", paths.LogsDir)
	}
	if paths.WebviewDir != filepath.Join(root, "webview") {
		t.Fatalf("expected WebviewDir under root, got %q", paths.WebviewDir)
	}
	if paths.StockDBPath != filepath.Join(root, "data", "stock.db") {
		t.Fatalf("expected StockDBPath under root, got %q", paths.StockDBPath)
	}
	if paths.UserDictPath != filepath.Join(root, "data", "dict", "user.txt") {
		t.Fatalf("expected UserDictPath under root, got %q", paths.UserDictPath)
	}
	if paths.WailsLogPath != filepath.Join(root, "logs", "wails.log") {
		t.Fatalf("expected WailsLogPath under root, got %q", paths.WailsLogPath)
	}
	if paths.InfoLogPath != filepath.Join(root, "logs", "info.log") {
		t.Fatalf("expected InfoLogPath under root, got %q", paths.InfoLogPath)
	}
	if paths.ErrorLogPath != filepath.Join(root, "logs", "error.log") {
		t.Fatalf("expected ErrorLogPath under root, got %q", paths.ErrorLogPath)
	}
}

func TestMigrateLegacyRuntimeFiles_MovesKnownFilesToAppRoot(t *testing.T) {
	legacyDir := t.TempDir()
	root := filepath.Join(t.TempDir(), "go-stock")
	paths := buildPaths(root)

	sourceFiles := map[string]string{
		filepath.Join(legacyDir, "data", "stock.db"):         "db",
		filepath.Join(legacyDir, "data", "stock.db-wal"):     "wal",
		filepath.Join(legacyDir, "data", "stock.db-shm"):     "shm",
		filepath.Join(legacyDir, "data", "dict", "user.txt"): "dict",
		filepath.Join(legacyDir, "logs", "info.log"):         "info",
		filepath.Join(legacyDir, "logs", "error.log"):        "error",
		filepath.Join(legacyDir, "logs", "wails.log"):        "wails",
	}

	for filePath, content := range sourceFiles {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatalf("mkdir source dir failed: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatalf("write source file failed: %v", err)
		}
	}

	if err := ensureDirectories(paths); err != nil {
		t.Fatalf("ensureDirectories failed: %v", err)
	}
	if err := migrateLegacyRuntimeFiles(legacyDir, paths); err != nil {
		t.Fatalf("migrateLegacyRuntimeFiles failed: %v", err)
	}

	destinations := map[string]string{
		paths.StockDBPath:          "db",
		paths.StockDBPath + "-wal": "wal",
		paths.StockDBPath + "-shm": "shm",
		paths.UserDictPath:         "dict",
		paths.InfoLogPath:          "info",
		paths.ErrorLogPath:         "error",
		paths.WailsLogPath:         "wails",
	}

	for filePath, want := range destinations {
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read migrated file %s failed: %v", filePath, err)
		}
		if string(data) != want {
			t.Fatalf("expected migrated content %q for %s, got %q", want, filePath, string(data))
		}
	}

	for filePath := range sourceFiles {
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Fatalf("expected legacy file %s to be removed, stat err=%v", filePath, err)
		}
	}
}
