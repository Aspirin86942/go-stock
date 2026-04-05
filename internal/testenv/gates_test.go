package testenv

import "testing"

func TestExternalGateDecision_UsesNewOrLegacyEnv(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		env         map[string]string
		shortMode   bool
		wantEnabled bool
		wantReason  string
	}{
		{
			name: "short mode with explicit external env still skips",
			env: map[string]string{
				ExternalEnv: "1",
			},
			shortMode:   true,
			wantEnabled: false,
			wantReason:  "skipping external test in short mode",
		},
		{
			name: "new env enables external tests",
			env: map[string]string{
				ExternalEnv: "1",
			},
			wantEnabled: true,
		},
		{
			name: "legacy integration env keeps compatibility",
			env: map[string]string{
				LegacyIntegrationEnv: "1",
			},
			wantEnabled: true,
		},
		{
			name:        "missing env disables external tests",
			env:         map[string]string{},
			wantEnabled: false,
			wantReason:  "skipping external test; set GO_STOCK_RUN_EXTERNAL_TESTS=1 to enable",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enabled, reason := externalGateDecision(getenvFromMap(tc.env), tc.shortMode)
			if enabled != tc.wantEnabled {
				t.Fatalf("enabled = %v, want %v", enabled, tc.wantEnabled)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}

func TestReleaseSmokeGateDecision_RequiresDedicatedEnv(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		env         map[string]string
		shortMode   bool
		wantEnabled bool
		wantReason  string
	}{
		{
			name: "short mode skips release smoke tests",
			env: map[string]string{
				ReleaseSmokeEnv: "1",
			},
			shortMode:   true,
			wantEnabled: false,
			wantReason:  "skipping release smoke test in short mode",
		},
		{
			name: "release smoke env enables tests",
			env: map[string]string{
				ReleaseSmokeEnv: "1",
			},
			wantEnabled: true,
		},
		{
			name:        "missing release smoke env disables tests",
			env:         map[string]string{},
			wantEnabled: false,
			wantReason:  "skipping release smoke test; set GO_STOCK_RUN_RELEASE_SMOKE=1 to enable",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enabled, reason := releaseSmokeGateDecision(getenvFromMap(tc.env), tc.shortMode)
			if enabled != tc.wantEnabled {
				t.Fatalf("enabled = %v, want %v", enabled, tc.wantEnabled)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}

func TestManualGateDecision_RequiresDedicatedEnv(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		env         map[string]string
		shortMode   bool
		wantEnabled bool
		wantReason  string
	}{
		{
			name: "short mode skips manual tests",
			env: map[string]string{
				ManualEnv: "1",
			},
			shortMode:   true,
			wantEnabled: false,
			wantReason:  "skipping manual test in short mode",
		},
		{
			name: "manual env enables tests",
			env: map[string]string{
				ManualEnv: "1",
			},
			wantEnabled: true,
		},
		{
			name:        "missing manual env disables tests",
			env:         map[string]string{},
			wantEnabled: false,
			wantReason:  "skipping manual test; set GO_STOCK_RUN_MANUAL_TESTS=1 to enable",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enabled, reason := manualGateDecision(getenvFromMap(tc.env), tc.shortMode)
			if enabled != tc.wantEnabled {
				t.Fatalf("enabled = %v, want %v", enabled, tc.wantEnabled)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}

func getenvFromMap(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}
