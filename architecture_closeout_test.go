package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppBridgeNoLongerOwnsBusinessDataApis(t *testing.T) {
	appGo := mustReadFile(t, "app.go")
	appCommon := mustReadFile(t, "app_common.go")
	raw := appGo + "\n" + appCommon

	disallowed := []string{
		"data.NewMarketNewsApi()",
		"data.NewSearchStockApi(",
		"data.NewStockDataApi()",
		"data.NewAiRecommendStocksService()",
		"data.NewStockChangesApi()",
		"data.NewStockChangeHistoryService()",
		"data.NewFundApi()",
		"data.NewStockGroupApi(",
		"db.Dao",
	}

	for _, token := range disallowed {
		if strings.Contains(raw, token) {
			t.Fatalf("forbidden legacy bridge token still present: %s", token)
		}
	}
}

func TestFrontendDirectWailsImportsAreLimitedToServices(t *testing.T) {
	root := filepath.Join("frontend", "src")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".vue") && !strings.HasSuffix(path, ".mjs") && !strings.HasSuffix(path, ".js") {
			return nil
		}
		if strings.Contains(path, filepath.Join("frontend", "src", "services")) {
			return nil
		}
		raw := mustReadFile(t, path)
		if strings.Contains(raw, "wailsjs/go/main/App") {
			t.Fatalf("direct App binding import outside services: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRouterUsesPagesForAllRouteComponents(t *testing.T) {
	raw := mustReadFile(t, filepath.Join("frontend", "src", "router", "router.js"))
	if strings.Contains(raw, "../components/") {
		t.Fatalf("router still imports route components directly from components")
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
