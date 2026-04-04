# Runtime Path Relocation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move go-stock runtime files out of the release directory so packaged Windows builds keep only the `exe` in the install location while databases, logs, and WebView data live under `%LOCALAPPDATA%\\go-stock`.

**Architecture:** Add one shared runtime-path package that owns application root resolution and directory creation, then make `db`, `logger`, `main`, `ai-assistant-web`, and the custom dictionary loader consume that package instead of hard-coded relative paths. Keep behavior compatible by preserving the same file names under a new absolute root.

**Tech Stack:** Go, Wails, PowerShell, Go `testing`

---

### Task 1: Add a shared runtime-path package with deterministic tests

**Files:**
- Create: `D:\codex_work\go-stock\backend\apppath\apppath.go`
- Create: `D:\codex_work\go-stock\backend\apppath\apppath_test.go`

- [ ] **Step 1: Write failing tests for Windows path resolution**

```go
func TestResolveRootDir_WindowsUsesLocalAppData(t *testing.T) {
	root, err := resolveRootDir("windows", func(key string) string {
		if key == "LOCALAPPDATA" {
			return `C:\Users\Test\AppData\Local`
		}
		return ""
	}, func() (string, error) {
		return "", errors.New("should not be called")
	})

	if err != nil {
		t.Fatalf("resolveRootDir returned error: %v", err)
	}
	want := filepath.Clean(`C:\Users\Test\AppData\Local\go-stock`)
	if root != want {
		t.Fatalf("expected %q, got %q", want, root)
	}
}
```

- [ ] **Step 2: Run the new package tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/apppath -count=1
```

Expected:

- compile failure because the package and resolver do not exist yet

- [ ] **Step 3: Implement the minimal shared path helpers**

```go
type Paths struct {
	RootDir         string
	DataDir         string
	LogsDir         string
	WebviewDir      string
	StockDBPath     string
	WailsLogPath    string
	InfoLogPath     string
	ErrorLogPath    string
	UserDictPath    string
}
```

- [ ] **Step 4: Re-run the path package tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/apppath -count=1
```

Expected:

- command exits `0`

### Task 2: Switch default database and logging paths to the shared runtime root

**Files:**
- Modify: `D:\codex_work\go-stock\backend\db\db.go`
- Modify: `D:\codex_work\go-stock\backend\logger\lgo.go`
- Modify: `D:\codex_work\go-stock\main.go`
- Modify: `D:\codex_work\go-stock\ai-assistant-web\server.go`
- Modify: `D:\codex_work\go-stock\backend\data\stock_sentiment_analysis.go`

- [ ] **Step 1: Add failing tests for the new default DB path**

```go
func TestDefaultSQLiteDSN_UsesAppRoot(t *testing.T) {
	dsn := defaultSQLiteDSNForTest(`C:\Users\Test\AppData\Local\go-stock\data\stock.db`)
	if !strings.Contains(dsn, `C:\Users\Test\AppData\Local\go-stock\data\stock.db`) {
		t.Fatalf("expected app-root database path in DSN, got %q", dsn)
	}
}
```

- [ ] **Step 2: Run targeted tests and confirm they fail**

Run:

```powershell
conda run -n test go test ./backend/db ./backend/logger -count=1
```

Expected:

- compile or assertion failure because the helpers still use relative paths

- [ ] **Step 3: Implement the shared-path wiring**

```go
paths, err := apppath.Ensure()
if err != nil {
	return err
}
sqlitePath := defaultSQLiteDSN(paths.StockDBPath)
logger.NewFileLogger(paths.WailsLogPath)
```

- [ ] **Step 4: Re-run the targeted tests and confirm they pass**

Run:

```powershell
conda run -n test go test ./backend/apppath ./backend/db ./backend/logger -count=1
```

Expected:

- command exits `0`

### Task 3: Verify the packaged app no longer writes runtime files into the release directory

**Files:**
- Verify only: `D:\codex_work\go-stock\build\bin\go-stock.exe`
- Verify only: `D:\codex_work\go-stock\build\bin`

- [ ] **Step 1: Run focused Go tests as final unit evidence**

Run:

```powershell
conda run -n test go test ./backend/apppath ./backend/db ./backend/logger -count=1
```

Expected:

- command exits `0`

- [ ] **Step 2: Run a desktop build verification**

Run:

```powershell
conda run -n test go test ./... -count=1
& 'C:\Program Files\Go\bin\go.exe' build .
```

Expected:

- targeted Go tests relevant to modified packages pass
- desktop build exits `0`

- [ ] **Step 3: Launch the app and inspect the release directory**

Run:

```powershell
Start-Process 'D:\codex_work\go-stock\build\bin\go-stock.exe'
```

Manual verification checklist:

- `D:\codex_work\go-stock\build\bin` does not receive new `data` or `logs` runtime directories from this launch
- `%LOCALAPPDATA%\go-stock` contains the new runtime files and directories

- [ ] **Step 4: Review only the intended slice**

```powershell
git diff -- backend/apppath backend/db/db.go backend/logger/lgo.go backend/data/stock_sentiment_analysis.go main.go ai-assistant-web/server.go docs/superpowers/specs/2026-04-04-runtime-path-relocation-design.md docs/superpowers/plans/2026-04-04-runtime-path-relocation.md
```
