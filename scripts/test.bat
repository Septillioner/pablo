@echo off
setlocal EnableExtensions

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0test.ps1" %*
exit /b %ERRORLEVEL%
