package testenv

import (
	"os"
	"testing"
)

const (
	ExternalEnv          = "GO_STOCK_RUN_EXTERNAL_TESTS"
	LegacyIntegrationEnv = "GO_STOCK_RUN_INTEGRATION_TESTS"
	ReleaseSmokeEnv      = "GO_STOCK_RUN_RELEASE_SMOKE"
	ManualEnv            = "GO_STOCK_RUN_MANUAL_TESTS"
)

func RequireExternalTest(t *testing.T) {
	t.Helper()

	enabled, reason := externalGateDecision(os.Getenv, testing.Short())
	if !enabled {
		t.Skip(reason)
	}
}

func RequireReleaseSmokeTest(t *testing.T) {
	t.Helper()

	enabled, reason := releaseSmokeGateDecision(os.Getenv, testing.Short())
	if !enabled {
		t.Skip(reason)
	}
}

func RequireManualTest(t *testing.T) {
	t.Helper()

	enabled, reason := manualGateDecision(os.Getenv, testing.Short())
	if !enabled {
		t.Skip(reason)
	}
}

func externalGateDecision(getenv func(string) string, shortMode bool) (bool, string) {
	if shortMode {
		return false, "skipping external test in short mode"
	}
	if getenv(ExternalEnv) == "1" || getenv(LegacyIntegrationEnv) == "1" {
		return true, ""
	}
	return false, "skipping external test; set GO_STOCK_RUN_EXTERNAL_TESTS=1 to enable"
}

func releaseSmokeGateDecision(getenv func(string) string, shortMode bool) (bool, string) {
	if shortMode {
		return false, "skipping release smoke test in short mode"
	}
	if getenv(ReleaseSmokeEnv) == "1" {
		return true, ""
	}
	return false, "skipping release smoke test; set GO_STOCK_RUN_RELEASE_SMOKE=1 to enable"
}

func manualGateDecision(getenv func(string) string, shortMode bool) (bool, string) {
	if shortMode {
		return false, "skipping manual test in short mode"
	}
	if getenv(ManualEnv) == "1" {
		return true, ""
	}
	return false, "skipping manual test; set GO_STOCK_RUN_MANUAL_TESTS=1 to enable"
}
