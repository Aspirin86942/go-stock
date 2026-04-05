package logger

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var guardExcludedDirs = map[string]struct{}{
	".git":         {},
	"frontend":     {},
	"build":        {},
	"node_modules": {},
}

func TestNoLegacyLoggerUsageInNonTestGoFiles(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	fset := token.NewFileSet()
	guardFiles, err := collectGuardTargetFiles(repoRoot)
	if err != nil {
		t.Fatalf("collect guard target files: %v", err)
	}

	var violations []string
	for _, relativePath := range guardFiles {
		absolutePath := filepath.Join(repoRoot, filepath.FromSlash(relativePath))
		fileNode, err := parser.ParseFile(fset, absolutePath, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", relativePath, err)
		}
		violations = append(violations, collectLegacyUsageViolations(fset, fileNode, relativePath)...)
	}

	if len(violations) > 0 {
		t.Fatalf("legacy logger usage is forbidden in non-test Go files:\n%s", strings.Join(violations, "\n"))
	}
}

func TestCollectLegacyUsageViolations_FlagsSugaredLoggerValueUsage(t *testing.T) {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, "sample.go", `package sample
import logger "go-stock/backend/logger"
func demo() any {
	_ = logger.SugaredLogger
	return logger.SugaredLogger
}
`, 0)
	if err != nil {
		t.Fatalf("parse sample file: %v", err)
	}

	violations := collectLegacyUsageViolations(fset, fileNode, "sample.go")
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations for SugaredLogger value usage, got %d: %#v", len(violations), violations)
	}
}

func TestCollectGuardTargetFiles_SkipsExcludedDirsAndTestFiles(t *testing.T) {
	repoRoot := t.TempDir()
	files := []string{
		filepath.Join(repoRoot, "backend", "data", "stock_data_api.go"),
		filepath.Join(repoRoot, "backend", "data", "stock_data_api_test.go"),
		filepath.Join(repoRoot, "frontend", "src", "fake.go"),
		filepath.Join(repoRoot, "build", "generated.go"),
		filepath.Join(repoRoot, "node_modules", "pkg", "index.go"),
		filepath.Join(repoRoot, ".git", "hooks", "pre-commit.go"),
		filepath.Join(repoRoot, "main.go"),
	}

	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", file, err)
		}
		if err := os.WriteFile(file, []byte("package sample\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}

	guardFiles, err := collectGuardTargetFiles(repoRoot)
	if err != nil {
		t.Fatalf("collect guard files: %v", err)
	}
	slices.Sort(guardFiles)

	expected := []string{
		"backend/data/stock_data_api.go",
		"main.go",
	}
	if !slices.Equal(guardFiles, expected) {
		t.Fatalf("unexpected guard files:\nwant=%v\ngot=%v", expected, guardFiles)
	}
}

func TestShouldScanGuardFile(t *testing.T) {
	testCases := []struct {
		path string
		want bool
	}{
		{path: "backend/data/openai_tools.go", want: true},
		{path: "backend/data/openai_tools_test.go", want: false},
		{path: "frontend/src/mock.go", want: false},
		{path: "build/mock.go", want: false},
		{path: ".git/hooks/pre-commit.go", want: false},
		{path: "node_modules/pkg/mock.go", want: false},
		{path: "backend/data/words.txt", want: false},
	}

	for _, tc := range testCases {
		if got := shouldScanGuardFile(filepath.FromSlash(tc.path)); got != tc.want {
			t.Fatalf("shouldScanGuardFile(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func collectGuardTargetFiles(repoRoot string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == repoRoot {
			return nil
		}

		relativePath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipGuardDir(relativePath) {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldScanGuardFile(relativePath) {
			return nil
		}
		files = append(files, filepath.ToSlash(relativePath))
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}

func shouldScanGuardFile(relativePath string) bool {
	normalized := filepath.ToSlash(filepath.Clean(relativePath))
	if shouldSkipGuardDir(normalized) {
		return false
	}
	if !strings.HasSuffix(normalized, ".go") {
		return false
	}
	return !strings.HasSuffix(normalized, "_test.go")
}

func shouldSkipGuardDir(relativePath string) bool {
	normalized := filepath.ToSlash(filepath.Clean(relativePath))
	for _, part := range strings.Split(normalized, "/") {
		if _, excluded := guardExcludedDirs[part]; excluded {
			return true
		}
	}
	return false
}

func collectLegacyUsageViolations(fset *token.FileSet, fileNode *ast.File, relativePath string) []string {
	loggerAliases := findImportAliases(fileNode, "go-stock/backend/logger")

	var violations []string
	ast.Inspect(fileNode, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.SelectorExpr:
			if isSugaredLoggerSelector(current, loggerAliases) {
				position := fset.Position(current.Pos())
				violations = append(violations, relativePath+":"+itoa(position.Line)+":"+itoa(position.Column)+": SugaredLogger")
			}
		case *ast.CallExpr:
			selector, ok := current.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}

			target := ident.Name + "." + selector.Sel.Name
			switch target {
			case "fmt.Printf", "log.Fatalf", "log.Printf", "log.Println", "log.Print":
				position := fset.Position(current.Pos())
				violations = append(violations, relativePath+":"+itoa(position.Line)+":"+itoa(position.Column)+": "+target)
			}
		}
		return true
	})

	return violations
}

func isSugaredLoggerSelector(expr ast.Expr, loggerAliases map[string]struct{}) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := selector.X.(*ast.Ident)
	if ok {
		_, matched := loggerAliases[pkgIdent.Name]
		return matched && selector.Sel.Name == "SugaredLogger"
	}

	rootSelector, ok := selector.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok = rootSelector.X.(*ast.Ident)
	if !ok {
		return false
	}

	_, matched := loggerAliases[pkgIdent.Name]
	return matched && rootSelector.Sel.Name == "SugaredLogger"
}

func findImportAliases(fileNode *ast.File, importPath string) map[string]struct{} {
	aliases := make(map[string]struct{})
	for _, imported := range fileNode.Imports {
		if strings.Trim(imported.Path.Value, `"`) != importPath {
			continue
		}
		if imported.Name != nil {
			aliases[imported.Name.Name] = struct{}{}
			continue
		}
		pathParts := strings.Split(importPath, "/")
		aliases[pathParts[len(pathParts)-1]] = struct{}{}
	}
	return aliases
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
