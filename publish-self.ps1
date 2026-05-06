# Pablo Cross-Platform Publish Script (PowerShell)
# This script runs the Pablo deployment pipeline for the local system.

# Check for Administrator privileges
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
$isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "Requiring Administrator privileges. Requesting elevation..." -ForegroundColor Yellow
    $scriptPath = $MyInvocation.MyCommand.Definition
    Start-Process powershell -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`"" -Verb RunAs
    exit
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location "$scriptDir/src"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Pablo Self-Publishing (Windows)"         -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

go run main.go -f ../pablo.yaml run -e windows-local

Pause