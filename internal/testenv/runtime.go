package testenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"

	"go-stock/backend/apppath"
	"go-stock/backend/logger"
	"go.uber.org/zap"
)

const ArtifactsDirEnv = "GO_STOCK_TEST_ARTIFACTS_DIR"

type RuntimeArtifacts struct {
	RootDir   string
	LogsDir   string
	TestRunID string
	TestSuite string
	TestCase  string
}

func NewLoggerRuntime(t *testing.T, suite string) (*logger.Runtime, RuntimeArtifacts) {
	t.Helper()

	previousGlobals := logger.CaptureGlobalState()

	testSuite := strings.TrimSpace(suite)
	if testSuite == "" {
		testSuite = "suite"
	}
	testCase := strings.TrimSpace(t.Name())
	if testCase == "" {
		testCase = "test"
	}

	testRunID := makeTestRunID()
	artifactsRoot := resolveArtifactsRoot(os.Getenv)
	artifactDir := filepath.Join(
		artifactsRoot,
		sanitizePathFragment(testRunID),
		sanitizePathFragment(testSuite),
		sanitizePathFragment(testCase),
	)
	logsDir := filepath.Join(artifactDir, "logs")

	runtime := logger.MustInit(logger.Config{
		Paths: apppath.Paths{
			RootDir: artifactDir,
			LogsDir: logsDir,
		},
		EnableStdout: false,
		Fields: []zap.Field{
			logger.String("execution_mode", "test"),
			logger.String("test_run_id", testRunID),
			logger.String("test_suite", testSuite),
			logger.String("test_case", testCase),
		},
	})
	t.Cleanup(func() {
		previousGlobals.Restore()
		if err := runtime.Close(); err != nil {
			t.Errorf("close test logger runtime: %v", err)
		}
	})

	return runtime, RuntimeArtifacts{
		RootDir:   artifactDir,
		LogsDir:   logsDir,
		TestRunID: testRunID,
		TestSuite: testSuite,
		TestCase:  testCase,
	}
}

func resolveArtifactsRoot(getenv func(string) string) string {
	root := strings.TrimSpace(getenv(ArtifactsDirEnv))
	if root == "" {
		return filepath.Join("artifacts", "testlogs")
	}
	return filepath.Clean(root)
}

func makeTestRunID() string {
	return fmt.Sprintf("%s-%d", time.Now().UTC().Format("20060102-150405.000000000"), os.Getpid())
}

func sanitizePathFragment(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range trimmed {
		allowed := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.'
		if allowed {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}

	result := strings.Trim(b.String(), "-_.")
	if result == "" {
		return "unknown"
	}
	return result
}
