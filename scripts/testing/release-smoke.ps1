$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/release'
  $env:GO_STOCK_RUN_RELEASE_SMOKE = '1'
  go test . -run 'TestReleaseSmoke_' -count=1
  go test ./ai-assistant-web -run 'TestReleaseSmoke_' -count=1
  node --test frontend/src/utils/frontendLogger.test.mjs
  npm --prefix frontend run build
} finally {
  Pop-Location
}
