@echo off
:: Pablo Cross-Platform Publish Script Wrapper
:: This allows running the publish-self command directly from CMD.

powershell -ExecutionPolicy Bypass -File "%~dp0publish-self.ps1"
