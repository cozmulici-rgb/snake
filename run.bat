@echo off
setlocal

where go >nul 2>nul
if errorlevel 1 (
  echo Go was not found on PATH.
  echo Install Go from https://go.dev/dl/ and reopen this terminal.
  exit /b 1
)

echo Running Snake...
go run .
endlocal