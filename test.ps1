# Pablo test runner (PowerShell)
# Usage: .\test.ps1 [unit|integration|e2e|all]

param(
    [Parameter(Position = 0)]
    [ValidateSet("unit", "integration", "e2e", "all")]
    [string]$Mode = "all"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Definition

function Show-Usage {
    Write-Host @"
Usage: .\test.ps1 [unit|integration|e2e|all]

  unit         Run unit tests in src/
  integration  Run integration-tagged tests in src/
  e2e          Run Docker-based E2E tests in tests/e2e/
  all          Run unit, then integration, then e2e (default)
"@
}

function Test-DockerAvailable {
    if ($env:PABLO_E2E_SKIP_DOCKER -eq "1") {
        return
    }
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw "docker is required for e2e tests (set PABLO_E2E_SKIP_DOCKER=1 to skip Docker setup)"
    }
    & docker info *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "docker daemon is not running"
    }
}

function Invoke-UnitTests {
    Write-Host "==> unit tests (src/)"
    Push-Location (Join-Path $Root "src")
    try {
        go test ./...
        if ($LASTEXITCODE -ne 0) { throw "unit tests failed" }
    } finally {
        Pop-Location
    }
}

function Invoke-IntegrationTests {
    Write-Host "==> integration tests (src/, -tags=integration)"
    Push-Location (Join-Path $Root "src")
    try {
        go test -tags=integration ./...
        if ($LASTEXITCODE -ne 0) { throw "integration tests failed" }
    } finally {
        Pop-Location
    }
}

function Invoke-E2ETests {
    Write-Host "==> e2e tests (tests/e2e/)"
    Test-DockerAvailable
    Push-Location (Join-Path $Root "tests\e2e")
    try {
        go test -tags=integration -v -timeout 10m ./...
        if ($LASTEXITCODE -ne 0) { throw "e2e tests failed" }
    } finally {
        Pop-Location
    }
}

switch ($Mode) {
    "unit" { Invoke-UnitTests }
    "integration" { Invoke-IntegrationTests }
    "e2e" { Invoke-E2ETests }
    "all" {
        Invoke-UnitTests
        Invoke-IntegrationTests
        Invoke-E2ETests
    }
    default {
        Show-Usage
        exit 1
    }
}
