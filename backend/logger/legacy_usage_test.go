package logger

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var task6GuardFiles = map[string]struct{}{
	"ai-assistant-web/cmd/ai-assistant-web/main.go": {},
	"app.go":                         {},
	"app_common.go":                  {},
	"app_darwin.go":                  {},
	"app_linux.go":                   {},
	"app_windows.go":                 {},
	"backend/agent/agent.go":         {},
	"backend/agent/agent_api.go":     {},
	"backend/agent/chat_memory.go":   {},
	"backend/agent/cron_task_api.go": {},
	"backend/agent/tools/choice_stock_by_indicators_tool.go": {},
	"backend/agent/tools/data_tools_wrapper.go":              {},
	"backend/agent/tools/market_news_tool.go":                {},
	"backend/agent/tools/stock_code_tool.go":                 {},
	"bootstrap_stock_search_data.go":                         {},
	"main.go":                                                {},
}

func TestNoLegacyLoggerUsageInNonTestGoFiles(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	fset := token.NewFileSet()

	var violations []string
	for relativePath := range task6GuardFiles {
		absolutePath := filepath.Join(repoRoot, filepath.FromSlash(relativePath))
		fileNode, err := parser.ParseFile(fset, absolutePath, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", relativePath, err)
		}
		loggerAliases := findImportAliases(fileNode, "go-stock/backend/logger")

		ast.Inspect(fileNode, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isSugaredLoggerSelector(call.Fun, loggerAliases) {
				position := fset.Position(call.Pos())
				violations = append(violations, relativePath+":"+itoa(position.Line)+":"+itoa(position.Column)+": SugaredLogger")
				return true
			}

			selector, ok := call.Fun.(*ast.SelectorExpr)
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
				position := fset.Position(call.Pos())
				violations = append(violations, relativePath+":"+itoa(position.Line)+":"+itoa(position.Column)+": "+target)
			}
			return true
		})
	}

	if len(violations) > 0 {
		t.Fatalf("legacy logger usage is forbidden in Task 6 scope:\n%s", strings.Join(violations, "\n"))
	}
}

func isSugaredLoggerSelector(expr ast.Expr, loggerAliases map[string]struct{}) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	rootSelector, ok := selector.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := rootSelector.X.(*ast.Ident)
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
