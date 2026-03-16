@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0tools\build.ps1"
if errorlevel 1 exit /b %errorlevel%
