@echo off
setlocal EnableExtensions

cd /d "%~dp0"

set "VSWHERE=%ProgramFiles(x86)%\Microsoft Visual Studio\Installer\vswhere.exe"
set "MSBUILD="
if exist "%VSWHERE%" (
    for /f "usebackq delims=" %%M in (`"%VSWHERE%" -latest -prerelease -requires Microsoft.Component.MSBuild -find MSBuild\**\Bin\MSBuild.exe`) do (
        set "MSBUILD=%%M"
        goto :found_msbuild
    )
    for /f "usebackq delims=" %%M in (`"%VSWHERE%" -latest -requires Microsoft.Component.MSBuild -find MSBuild\**\Bin\MSBuild.exe`) do (
        set "MSBUILD=%%M"
        goto :found_msbuild
    )
)
:found_msbuild
if not defined MSBUILD (
    echo ERROR: MSBuild not found. Install Visual Studio 2026 Insiders or 2022 with MSBuild component.
    exit /b 1
)

echo Using MSBuild: %MSBUILD%
"%MSBUILD%" "%~dp0Pablo.sln" /p:Configuration=Release /restore /t:Rebuild /v:m
if errorlevel 1 exit /b %ERRORLEVEL%

set "VSIX_PRIMARY=%~dp0Pablo\bin\Release\net472\Pablo.VisualStudio.vsix"
set "VSIX_FALLBACK=%~dp0Pablo\bin\Release\Pablo.VisualStudio.vsix"
set "VSIX_LEGACY=%~dp0Pablo\.vsix"
if exist "%VSIX_PRIMARY%" (
    echo VSIX: %VSIX_PRIMARY%
    exit /b 0
)
if exist "%VSIX_FALLBACK%" (
    echo VSIX: %VSIX_FALLBACK%
    exit /b 0
)
if exist "%VSIX_LEGACY%" (
    move /Y "%VSIX_LEGACY%" "%VSIX_PRIMARY%" >nul
    if exist "%VSIX_PRIMARY%" (
        echo VSIX: %VSIX_PRIMARY%
        exit /b 0
    )
)

echo ERROR: Pablo.VisualStudio.vsix was not produced.
exit /b 1
