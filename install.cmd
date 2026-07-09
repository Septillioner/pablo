@echo off
powershell -NoProfile -ExecutionPolicy Bypass -Command "iex ((Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -UseBasicParsing).Content)"
