package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRemoveNonPrintable tests the RemoveAllBlankChar function.
func TestRemoveNonPrintable(t *testing.T) {
	if got := RemoveAllBlankChar("新 希 望"); got != "新希望" {
		t.Fatalf("RemoveAllBlankChar returned %q, want %q", got, "新希望")
	}
}

func TestConvertStockCodeToTushareCode(t *testing.T) {
	if got := ConvertStockCodeToTushareCode("sz000802"); got != "000802.SZ" {
		t.Fatalf("ConvertStockCodeToTushareCode returned %q", got)
	}
	if got := ConvertTushareCodeToStockCode("000802.SZ"); got != "sz000802" {
		t.Fatalf("ConvertTushareCodeToStockCode returned %q", got)
	}
}
func TestReplaceSensitiveWords(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	txt := "新 希 望习近平"
	got := ReplaceSensitiveWords(txt)
	if strings.Contains(got, "习近平") {
		t.Fatalf("expected sensitive words to be removed, got %q", got)
	}
	if RemoveAllBlankChar(got) != "新希望" {
		t.Fatalf("expected filtered text to keep non-sensitive content after whitespace normalization, got %q", got)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "words.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected ReplaceSensitiveWords test not to write words.txt, stat err=%v", err)
	}
}
