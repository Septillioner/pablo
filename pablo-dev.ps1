# Pablo Development Shell Setup (Current Session)

$PABLO_ROOT = $PSScriptRoot
# Check build folder first, then src
$BINARY_PATH = Join-Path $PABLO_ROOT "build"
if (-not (Test-Path (Join-Path $BINARY_PATH "pablo.exe"))) {
    $BINARY_PATH = Join-Path $PABLO_ROOT "src"
}

# Update PATH for the current session
$env:Path = "$BINARY_PATH;" + $env:Path

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Pablo PowerShell Dev Environment" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Binary Path: $BINARY_PATH" -ForegroundColor Gray
Write-Host "Status:      Successfully added to PATH`n" -ForegroundColor Green

Write-Host "IMPORTANT: To use in THIS window, you must run it with a dot:" -ForegroundColor Yellow
Write-Host ". .\pablo-dev.ps1" -ForegroundColor White
Write-Host "`nTo open a NEW window with this environment, run:" -ForegroundColor Gray
Write-Host ".\pablo-shell.ps1" -ForegroundColor Yellow
Write-Host "`n"

