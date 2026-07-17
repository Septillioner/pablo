# Pablo test runner (PowerShell)
# Usage: .\scripts\test.ps1 [unit|integration|e2e|all]

param(
    [Parameter(Position = 0)]
    [ValidateSet("unit", "integration", "e2e", "all")]
    [string]$Mode = "all"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$Root = Split-Path -Parent $ScriptDir

$script:E2EScenarioMap = @{
    "TestSSH_StaticDeploy"       = "ssh-static"
    "TestSSH_RenameReplace"      = "ssh-rename-replace"
    "TestSSH_DockerRemoteDeploy" = "ssh-docker-remote"
}

$script:Summary = @{}

function Show-Usage {
    Write-Host @"
Usage: .\scripts\test.ps1 [unit|integration|e2e|all]

  unit         Run unit tests in src/
  integration  Run integration-tagged tests in src/
  e2e          Run Docker-based E2E tests in tests/e2e/
  all          Run unit, then integration, then e2e (default)
"@
}

function Write-SectionHeader {
    param([string]$Title)
    Write-Host ""
    Write-Host "======== $Title ========"
}

function Format-Elapsed {
    param($Seconds)
    if ($null -eq $Seconds -or $Seconds -eq 0) { return "" }
    return ("{0:N2}s" -f $Seconds)
}

function Invoke-ExternalOutput {
    param([scriptblock]$Command)

    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $raw = & $Command 2>&1
        $exitCode = $LASTEXITCODE
        $lines = @($raw | ForEach-Object {
            if ($_ -is [System.Management.Automation.ErrorRecord]) { $_.ToString() }
            else { "$_" }
        })
        return @{ Output = $lines; ExitCode = $exitCode }
    } finally {
        $ErrorActionPreference = $prev
    }
}

function Invoke-GoTestJson {
    param(
        [string[]]$GoArgs,
        [string]$WorkingDir
    )

    Push-Location $WorkingDir
    try {
        $result = Invoke-ExternalOutput { go test @GoArgs -json ./... }
    } finally {
        Pop-Location
    }

    $output = $result.Output
    $exitCode = $result.ExitCode

    $passed = 0
    $failed = 0

    foreach ($line in $output) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }
        try {
            $ev = $line | ConvertFrom-Json
        } catch {
            continue
        }

        if ($ev.Action -eq "output") { continue }

        $isPackageLevel = [string]::IsNullOrEmpty($ev.Test)
        if (-not $isPackageLevel) { continue }

        if ($ev.Action -eq "skip") {
            if ($ev.Output -match "no test files") { continue }
        }

        if ($ev.Action -eq "pass") {
            $elapsed = Format-Elapsed $ev.Elapsed
            $pkg = $ev.Package
            Write-Host ("  PASS  {0,-42} {1}" -f $pkg, $elapsed) -ForegroundColor Green
            $passed++
        } elseif ($ev.Action -eq "fail") {
            Write-Host ("  FAIL  {0}" -f $ev.Package) -ForegroundColor Red
            $failed++
        }
    }

    Write-Host ""
    Write-Host "  $passed packages passed, $failed failed"

    if ($exitCode -ne 0 -or $failed -gt 0) {
        return "FAIL"
    }
    return "PASS"
}

function Invoke-PackageTests {
    param(
        [string]$SectionTitle,
        [string]$SummaryKey,
        [string[]]$GoArgs
    )

    Write-SectionHeader $SectionTitle
    $result = Invoke-GoTestJson -GoArgs $GoArgs -WorkingDir (Join-Path $Root "src")
    $script:Summary[$SummaryKey] = $result
    if ($result -eq "FAIL") { throw "$SummaryKey tests failed" }
}

function Test-DockerAvailable {
    if ($env:PABLO_E2E_SKIP_DOCKER -eq "1") {
        return
    }
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw "docker is required for e2e tests (set PABLO_E2E_SKIP_DOCKER=1 to skip Docker setup)"
    }
    $result = Invoke-ExternalOutput { docker info }
    if ($result.ExitCode -ne 0) {
        throw "docker daemon is not running"
    }
}

function Get-E2EScenarioNames {
    $dir = Join-Path $Root "tests\e2e\scenarios"
    if (-not (Test-Path $dir)) { return @() }
    return @(Get-ChildItem $dir -Directory | ForEach-Object { $_.Name } | Sort-Object)
}

function Invoke-E2ETests {
    Write-SectionHeader "E2E"

    $scenarios = Get-E2EScenarioNames
    if ($scenarios.Count -gt 0) {
        Write-Host ("  Scenarios: {0}" -f ($scenarios -join ", "))
        Write-Host ""
    }

    Test-DockerAvailable

    Push-Location (Join-Path $Root "tests\e2e")
    try {
        $result = Invoke-ExternalOutput { go test -tags=integration -json -timeout 10m ./... }
    } finally {
        Pop-Location
    }

    $output = $result.Output
    $exitCode = $result.ExitCode

    $passed = 0
    $failed = 0

    foreach ($line in $output) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }
        try {
            $ev = $line | ConvertFrom-Json
        } catch {
            continue
        }

        if ([string]::IsNullOrEmpty($ev.Test)) { continue }
        if ($ev.Test -match "/") { continue }

        $scenario = $script:E2EScenarioMap[$ev.Test]
        if (-not $scenario) { $scenario = "unknown" }

        if ($ev.Action -eq "pass") {
            $elapsed = Format-Elapsed $ev.Elapsed
            Write-Host ("  PASS  {0,-32} ({1,-22}) {2}" -f $ev.Test, $scenario, $elapsed) -ForegroundColor Green
            $passed++
        } elseif ($ev.Action -eq "fail") {
            Write-Host ("  FAIL  {0,-32} ({1})" -f $ev.Test, $scenario) -ForegroundColor Red
            $failed++
        }
    }

    Write-Host ""
    Write-Host "  $passed scenarios passed, $failed failed"

    if ($exitCode -ne 0 -or $failed -gt 0) {
        $script:Summary["e2e"] = "FAIL"
        throw "e2e tests failed"
    }
    $script:Summary["e2e"] = "PASS"
}

function Write-Summary {
    Write-SectionHeader "SUMMARY"
    foreach ($key in @("unit", "integration", "e2e")) {
        if ($script:Summary.ContainsKey($key)) {
            $status = $script:Summary[$key]
            $color = if ($status -eq "PASS") { "Green" } else { "Red" }
            Write-Host ("  {0,-14} {1}" -f "${key}:", $status) -ForegroundColor $color
        }
    }
}

function Invoke-UnitTests {
    Invoke-PackageTests -SectionTitle "UNIT" -SummaryKey "unit" -GoArgs @()
}

function Invoke-IntegrationTests {
    Invoke-PackageTests -SectionTitle "INTEGRATION" -SummaryKey "integration" -GoArgs @("-tags=integration")
}

try {
    switch ($Mode) {
        "unit" { Invoke-UnitTests }
        "integration" { Invoke-IntegrationTests }
        "e2e" { Invoke-E2ETests }
        "all" {
            Invoke-UnitTests
            Invoke-IntegrationTests
            Invoke-E2ETests
            Write-Summary
        }
        default {
            Show-Usage
            exit 1
        }
    }
} catch {
    if ($Mode -eq "all" -and $script:Summary.Count -gt 0) {
        Write-Summary
    }
    Write-Host ""
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}
