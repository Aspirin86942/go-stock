package logger

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var allowedLegacyTestFiles = map[string]struct{}{
	"backend/logger/core_test.go":         {},
	"backend/logger/legacy_usage_test.go": {},
}

func TestNoLegacySugaredLoggerUsageInGoTestsOutsideCompatibilityCoverage(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	fset := token.NewFileSet()
	testFiles, err := collectLegacyTestTargetFiles(repoRoot)
	if err != nil {
		t.Fatalf("collect legacy test target files: %v", err)
	}

	var violations []string
	for _, relativePath := range testFiles {
		fileNode, err := parser.ParseFile(fset, filepath.Join(repoRoot, filepath.FromSlash(relativePath)), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", relativePath, err)
		}
		violations = append(violations, collectLegacySugaredLoggerViolations(fset, fileNode, relativePath)...)
	}

	if len(violations) > 0 {
		t.Fatalf("legacy SugaredLogger usage is forbidden in Go tests outside compatibility coverage:\n%s", strings.Join(violations, "\n"))
	}
}

func collectLegacyTestTargetFiles(repoRoot string) ([]string, error) {
	var testFiles []string
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

		normalized := filepath.ToSlash(relativePath)
		if !strings.HasSuffix(normalized, "_test.go") {
			return nil
		}
		if _, allowed := allowedLegacyTestFiles[normalized]; allowed {
			return nil
		}
		testFiles = append(testFiles, normalized)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(testFiles)
	return testFiles, nil
}

func collectLegacySugaredLoggerViolations(fset *token.FileSet, fileNode *ast.File, relativePath string) []string {
	loggerAliases := findImportAliases(fileNode, "go-stock/backend/logger")

	var violations []string
	ast.Inspect(fileNode, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if !isSugaredLoggerSelector(selector, loggerAliases) {
			return true
		}

		position := fset.Position(selector.Pos())
		violations = append(violations, relativePath+":"+itoa(position.Line)+":"+itoa(position.Column)+": SugaredLogger")
		return true
	})

	return violations
}
