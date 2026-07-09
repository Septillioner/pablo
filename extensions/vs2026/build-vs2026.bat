@echo off
setlocal EnableExtensions

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"
if errorlevel 1 (
    echo ERROR: Cannot cd to "%SCRIPT_DIR%"
    exit /b 1
)

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

if not exist "Pablo.sln" (
    echo ERROR: Pablo.sln not found in %CD%
    exit /b 1
)

echo Using MSBuild: %MSBUILD%
echo Building: %CD%\Pablo.sln
"%MSBUILD%" "Pablo.sln" /p:Configuration=Release /restore /t:Rebuild /v:m
if errorlevel 1 exit /b %ERRORLEVEL%

set "VSIX_PRIMARY=%CD%\Pablo\bin\Release\net472\Pablo.VisualStudio.vsix"
set "VSIX_FALLBACK=%CD%\Pablo\bin\Release\Pablo.VisualStudio.vsix"
set "VSIX_LEGACY=%CD%\Pablo\.vsix"
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
