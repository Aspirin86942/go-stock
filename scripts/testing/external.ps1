$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/external'
  $env:GO_STOCK_RUN_EXTERNAL_TESTS = '1'
  go test ./... -count=1
  node --test frontend/src/utils/frontendLogger.test.mjs frontend/src/utils/stockCode.test.mjs frontend/src/utils/aiConfig.test.mjs frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
} finally {
  Pop-Location
}
