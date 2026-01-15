# Pablo Development Shell Launcher (New Window)
$PABLO_ROOT = $PSScriptRoot
$BINARY_PATH = Join-Path $PABLO_ROOT "build"
if (-not (Test-Path (Join-Path $BINARY_PATH "pablo.exe"))) {
    $BINARY_PATH = Join-Path $PABLO_ROOT "src"
}

# Construct the command for the new window. 
# We use single quotes for the outer string and avoid backticks inside the Command variable.
$Command = "set-location '$PABLO_ROOT'; `$env:Path = '$BINARY_PATH;' + `$env:Path; Write-Host 'Pablo Dev Shell Active' -ForegroundColor Green; pablo --help"

Start-Process powershell -ArgumentList "-NoExit", "-Command", "$Command"
