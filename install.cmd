@echo off
setlocal EnableExtensions

set "INSTALL_SCRIPT=%TEMP%\pablo-install.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -Command "[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -OutFile '%INSTALL_SCRIPT%' -UseBasicParsing"
if errorlevel 1 (
    echo error: failed to download install.ps1
    exit /b 1
)

powershell -NoProfile -ExecutionPolicy Bypass -File "%INSTALL_SCRIPT%"
set "EXIT_CODE=%ERRORLEVEL%"
del /q "%INSTALL_SCRIPT%" >nul 2>&1
exit /b %EXIT_CODE%
