$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
Push-Location $repoRoot
try {
  $env:GO_STOCK_TEST_ARTIFACTS_DIR = 'artifacts/testlogs/default'
  $coverageRoot = Join-Path $repoRoot 'artifacts/coverage/default'
  $goCoverageProfile = Join-Path $coverageRoot 'go-default.coverprofile'
  $goCoverageSummary = Join-Path $coverageRoot 'go-default-summary.txt'
  $goCoverageHtml = Join-Path $coverageRoot 'go-default.html'
  $frontendCoverageDir = Join-Path $coverageRoot 'node-v8'
  $frontendCoverageIndex = Join-Path $coverageRoot 'frontend-v8-files.txt'

  if (Test-Path $coverageRoot) {
    Remove-Item -Recurse -Force $coverageRoot
  }
  New-Item -ItemType Directory -Force -Path $frontendCoverageDir | Out-Null

  go test ./... -count=1 -covermode=atomic -coverprofile $goCoverageProfile

  $coverageSummary = go tool cover "-func=$goCoverageProfile"
  $coverageSummary | Set-Content -Path $goCoverageSummary -Encoding utf8
  $coverageSummary
  go tool cover "-html=$goCoverageProfile" "-o=$goCoverageHtml"

  $env:NODE_V8_COVERAGE = $frontendCoverageDir
  node --test frontend/src/utils/frontendLogger.test.mjs frontend/src/utils/stockCode.test.mjs frontend/src/utils/aiConfig.test.mjs frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
  $frontendCoverageFiles = Get-ChildItem -Path $frontendCoverageDir -Filter '*.json' -File
  if ($frontendCoverageFiles.Count -eq 0) {
    throw 'frontend default tests produced no V8 coverage files'
  }
  $frontendCoverageFiles.FullName | Set-Content -Path $frontendCoverageIndex -Encoding utf8
} finally {
  Remove-Item Env:NODE_V8_COVERAGE -ErrorAction SilentlyContinue
  Pop-Location
}
