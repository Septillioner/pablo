# Pablo one-liner installer (Windows)
# Usage: $s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
# Usage (cmd): install.cmd
# Pin version: $env:PABLO_VERSION = "v1.4.0"; then run the command above

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$GITHUB_REPO = "septillioner/pablo"
$GITHUB_API_BASE = "https://api.github.com/repos/$GITHUB_REPO/releases"
$RELEASES_PAGE = "https://github.com/$GITHUB_REPO/releases"
$GITHUB_HEADERS = @{ "User-Agent" = "pablo-installer" }

$SYSTEM_INSTALL_DIR = Join-Path $env:ProgramFiles "Pablo"
$SYSTEM_INSTALL_PATH = Join-Path $SYSTEM_INSTALL_DIR "pablo.exe"
$USER_INSTALL_DIR = Join-Path $env:LOCALAPPDATA "Pablo"
$USER_INSTALL_PATH = Join-Path $USER_INSTALL_DIR "pablo.exe"

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message"
}

function Fail {
    param([string]$Message)
    throw $Message
}

function Get-PlatformAssetName {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { $assetArch = "amd64" }
        "ARM64" { $assetArch = "arm64" }
        default { Fail "unsupported architecture: $arch. See $RELEASES_PAGE" }
    }

    return @{
        AssetName = "pablo-windows-$assetArch.exe"
        Label = "windows/$assetArch"
    }
}

function Get-ReleaseTag {
    if ($env:PABLO_VERSION -and $env:PABLO_VERSION.Trim().Length -gt 0) {
        $tag = $env:PABLO_VERSION.Trim()
        if (-not $tag.StartsWith("v")) {
            $tag = "v$tag"
        }
        return $tag
    }

    $latest = Invoke-RestMethod -Uri "$GITHUB_API_BASE/latest" -Headers $GITHUB_HEADERS
    if (-not $latest -or -not $latest.tag_name) {
        Fail "could not read latest release from GitHub API. Set PABLO_VERSION or see $RELEASES_PAGE"
    }

    return $latest.tag_name
}

function Get-ReleaseAssets {
    param(
        [string]$ReleaseTag,
        [string]$AssetName
    )

    $release = Invoke-RestMethod -Uri "$GITHUB_API_BASE/tags/$ReleaseTag" -Headers $GITHUB_HEADERS
    if (-not $release -or -not $release.assets) {
        Fail "could not read release $ReleaseTag from GitHub API. See $RELEASES_PAGE"
    }

    $asset = $release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
    if (-not $asset -or -not $asset.browser_download_url) {
        Fail "release $ReleaseTag has no asset $AssetName. See $RELEASES_PAGE"
    }

    $checksumAsset = $release.assets | Where-Object { $_.name -eq "checksums.txt" } | Select-Object -First 1
    return @{
        DownloadUrl = $asset.browser_download_url
        ChecksumsUrl = if ($checksumAsset) { $checksumAsset.browser_download_url } else { $null }
    }
}

function Test-AssetChecksum {
    param(
        [string]$FilePath,
        [string]$AssetName,
        [string]$ChecksumsUrl
    )

    if (-not $ChecksumsUrl) {
        Write-Step "checksums.txt not found in release; skipping verification"
        return
    }

    $checksumFile = Join-Path ([System.IO.Path]::GetTempPath()) "pablo-checksums.txt"
    Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $checksumFile -UseBasicParsing

    $expectedHash = $null
    foreach ($line in Get-Content $checksumFile) {
        $parts = $line -split '\s+', 2
        if ($parts.Count -eq 2 -and $parts[1] -eq $AssetName -and $parts[0]) {
            $expectedHash = $parts[0].ToLower()
            break
        }
    }

    if (-not $expectedHash) {
        Fail "checksum for $AssetName not found in checksums.txt"
    }

    $actualHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
    if ($actualHash -ne $expectedHash) {
        Fail "checksum mismatch for $AssetName"
    }

    Write-Step "checksum verified"
}

function Install-ToDirectory {
    param(
        [string]$SourceFile,
        [string]$TargetDir,
        [string]$TargetPath
    )

    New-Item -ItemType Directory -Force -Path $TargetDir | Out-Null
    Copy-Item -Path $SourceFile -Destination $TargetPath -Force
}

function Test-IsAdministrator {
    $principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Add-DirectoryToPath {
    param(
        [string]$Directory,
        [System.EnvironmentVariableTarget]$Scope
    )

    $currentPath = [Environment]::GetEnvironmentVariable("Path", $Scope)
    $segments = @()
    if ($currentPath) {
        $segments = $currentPath -split ";" | Where-Object { $_ -ne "" }
    }

    if ($segments -contains $Directory) {
        return
    }

    $newPath = ($segments + $Directory) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, $Scope)
}

function Install-PabloBinary {
    param([string]$SourceFile)

    if (Test-IsAdministrator) {
        Install-ToDirectory -SourceFile $SourceFile -TargetDir $SYSTEM_INSTALL_DIR -TargetPath $SYSTEM_INSTALL_PATH
        Add-DirectoryToPath -Directory $SYSTEM_INSTALL_DIR -Scope Machine
        return @{
            Path = $SYSTEM_INSTALL_PATH
            Scope = "system"
        }
    }

    try {
        Install-ToDirectory -SourceFile $SourceFile -TargetDir $SYSTEM_INSTALL_DIR -TargetPath $SYSTEM_INSTALL_PATH
        Add-DirectoryToPath -Directory $SYSTEM_INSTALL_DIR -Scope Machine
        return @{
            Path = $SYSTEM_INSTALL_PATH
            Scope = "system"
        }
    }
    catch {
        Write-Step "system install unavailable; using user install"
        Install-ToDirectory -SourceFile $SourceFile -TargetDir $USER_INSTALL_DIR -TargetPath $USER_INSTALL_PATH
        Add-DirectoryToPath -Directory $USER_INSTALL_DIR -Scope User
        return @{
            Path = $USER_INSTALL_PATH
            Scope = "user"
        }
    }
}

function Verify-Installation {
    param(
        [string]$InstalledPath,
        [string]$Scope
    )

    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $env:Path = (@($machinePath, $userPath) | Where-Object { $_ }) -join ";"

    if (Get-Command pablo -ErrorAction SilentlyContinue) {
        pablo version
        return
    }

    if (Test-Path $InstalledPath) {
        & $InstalledPath version
        if ($Scope -eq "user") {
            Write-Step "pablo is installed at $InstalledPath; open a new terminal to use it globally"
        }
        return
    }

    Fail "installation finished but pablo could not be executed"
}

function Main {
    $platform = Get-PlatformAssetName
    $releaseTag = Get-ReleaseTag
    $assets = Get-ReleaseAssets -ReleaseTag $releaseTag -AssetName $platform.AssetName

    $tempFile = Join-Path ([System.IO.Path]::GetTempPath()) $platform.AssetName
    Write-Step "downloading $($platform.AssetName) from $releaseTag"
    Invoke-WebRequest -Uri $assets.DownloadUrl -OutFile $tempFile -UseBasicParsing

    Test-AssetChecksum -FilePath $tempFile -AssetName $platform.AssetName -ChecksumsUrl $assets.ChecksumsUrl

    Write-Step "installing Pablo for $($platform.Label) ($releaseTag)"
    $installResult = Install-PabloBinary -SourceFile $tempFile
    Write-Step "installed to $($installResult.Path) ($($installResult.Scope))"

    Verify-Installation -InstalledPath $installResult.Path -Scope $installResult.Scope
    Write-Step "done"
}

try {
    Main
}
catch {
    Write-Host "error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
