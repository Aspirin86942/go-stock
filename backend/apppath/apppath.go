package apppath

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const appName = "go-stock"

type Paths struct {
	RootDir      string
	DataDir      string
	LogsDir      string
	WebviewDir   string
	StockDBPath  string
	UserDictPath string
	WailsLogPath string
	InfoLogPath  string
	ErrorLogPath string
}

var (
	once      sync.Once
	cached    Paths
	cachedErr error
)

func Ensure() (Paths, error) {
	once.Do(func() {
		rootDir, err := resolveRootDir(runtime.GOOS, os.Getenv, os.UserConfigDir)
		if err != nil {
			cachedErr = err
			return
		}

		paths := buildPaths(rootDir)
		if err := ensureDirectories(paths); err != nil {
			cachedErr = err
			return
		}
		if err := migrateLegacyRuntime(paths); err != nil {
			cachedErr = err
			return
		}

		cached = paths
	})

	return cached, cachedErr
}

func buildPaths(root string) Paths {
	root = filepath.Clean(root)
	dataDir := filepath.Join(root, "data")
	logsDir := filepath.Join(root, "logs")
	return Paths{
		RootDir:      root,
		DataDir:      dataDir,
		LogsDir:      logsDir,
		WebviewDir:   filepath.Join(root, "webview"),
		StockDBPath:  filepath.Join(dataDir, "stock.db"),
		UserDictPath: filepath.Join(dataDir, "dict", "user.txt"),
		WailsLogPath: filepath.Join(logsDir, "wails.log"),
		InfoLogPath:  filepath.Join(logsDir, "info.log"),
		ErrorLogPath: filepath.Join(logsDir, "error.log"),
	}
}

func resolveRootDir(goos string, getenv func(string) string, userConfigDir func() (string, error)) (string, error) {
	if goos == "windows" {
		localAppData := strings.TrimSpace(getenv("LOCALAPPDATA"))
		if localAppData != "" {
			return filepath.Clean(filepath.Join(localAppData, appName)), nil
		}
	}

	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve app root dir: %w", err)
	}
	if strings.TrimSpace(configDir) == "" {
		return "", fmt.Errorf("resolve app root dir: empty config dir")
	}

	return filepath.Clean(filepath.Join(configDir, appName)), nil
}

func ensureDirectories(paths Paths) error {
	dirs := []string{
		paths.RootDir,
		paths.DataDir,
		filepath.Dir(paths.UserDictPath),
		paths.LogsDir,
		paths.WebviewDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ensure directory %s: %w", dir, err)
		}
	}

	return nil
}

func migrateLegacyRuntime(paths Paths) error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	legacyDir := filepath.Dir(executablePath)
	return migrateLegacyRuntimeFiles(legacyDir, paths)
}

func migrateLegacyRuntimeFiles(legacyDir string, paths Paths) error {
	legacyDir = filepath.Clean(legacyDir)
	if legacyDir == "." || legacyDir == "" {
		return nil
	}
	if samePath(legacyDir, paths.RootDir) {
		return nil
	}

	legacyFiles := map[string]string{
		filepath.Join(legacyDir, "data", "stock.db"):         paths.StockDBPath,
		filepath.Join(legacyDir, "data", "stock.db-wal"):     paths.StockDBPath + "-wal",
		filepath.Join(legacyDir, "data", "stock.db-shm"):     paths.StockDBPath + "-shm",
		filepath.Join(legacyDir, "data", "dict", "user.txt"): paths.UserDictPath,
		filepath.Join(legacyDir, "logs", "info.log"):         paths.InfoLogPath,
		filepath.Join(legacyDir, "logs", "error.log"):        paths.ErrorLogPath,
		filepath.Join(legacyDir, "logs", "wails.log"):        paths.WailsLogPath,
	}

	for sourcePath, targetPath := range legacyFiles {
		if err := moveFileIfNeeded(sourcePath, targetPath); err != nil {
			return err
		}
	}

	cleanupLegacyRuntimeDirs(legacyDir)
	return nil
}

func moveFileIfNeeded(sourcePath, targetPath string) error {
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat legacy file %s: %w", sourcePath, err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat target file %s: %w", targetPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("ensure target dir for %s: %w", targetPath, err)
	}

	if err := os.Rename(sourcePath, targetPath); err == nil {
		return nil
	} else if !isRenameFallbackError(err) {
		return fmt.Errorf("move legacy file %s -> %s: %w", sourcePath, targetPath, err)
	}

	if err := copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("copy legacy file %s -> %s: %w", sourcePath, targetPath, err)
	}
	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("remove migrated legacy file %s: %w", sourcePath, err)
	}

	return nil
}

func copyFile(sourcePath, targetPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	return targetFile.Close()
}

func isRenameFallbackError(err error) bool {
	return errors.Is(err, os.ErrInvalid) || strings.Contains(strings.ToLower(err.Error()), "cross-device")
}

func cleanupLegacyRuntimeDirs(legacyDir string) {
	dirs := []string{
		filepath.Join(legacyDir, "data", "dict"),
		filepath.Join(legacyDir, "data"),
		filepath.Join(legacyDir, "logs"),
	}

	for _, dir := range dirs {
		_ = os.Remove(dir)
	}
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}
