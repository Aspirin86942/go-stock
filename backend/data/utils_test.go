package data

import (
	"go-stock/backend/logger"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRemoveNonPrintable tests the RemoveAllBlankChar function.
func TestRemoveNonPrintable(t *testing.T) {
	//tests := []struct {
	//	input    string
	//	expected string
	//}{
	//	{"新 希 望", "新希望"},
	//	{"", ""},
	//	{"Hello, World!", "Hello, World!"},
	//	{"\x00\x01\x02", ""},
	//	{"Hello\x00World", "HelloWorld"},
	//	{"\x1F\x20\x7E\x7F", " \x7E"},
	//}

	//for _, test := range tests {
	//	actual := RemoveAllBlankChar(test.input)
	//	if actual != test.expected {
	//		t.Errorf("RemoveAllBlankChar(%q) = %q; expected %q", test.input, actual, test.expected)
	//	}
	//}
	txt := "新 希 望"
	txt2 := RemoveAllBlankChar(txt)
	logger.SugaredLogger.Infof("RemoveAllBlankChar(%s)", txt2)
	logger.SugaredLogger.Infof("RemoveAllBlankChar(%s)", txt)

}

func TestConvertStockCodeToTushareCode(t *testing.T) {
	logger.SugaredLogger.Infof("ConvertStockCodeToTushareCode(%s)", ConvertStockCodeToTushareCode("sz000802"))
	logger.SugaredLogger.Infof("ConvertTushareCodeToStockCode(%s)", ConvertTushareCodeToStockCode("000802.SZ"))
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
