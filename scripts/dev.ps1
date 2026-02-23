param(
    [ValidateSet("console", "graphic", "graphic-fullscreen", "test", "build", "fmt", "cover")]
    [string]$Task = "console"
)

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if ($goCmd) {
    $goExe = $goCmd.Source
} elseif (Test-Path "C:\Program Files\Go\bin\go.exe") {
    $goExe = "C:\Program Files\Go\bin\go.exe"
} else {
    Write-Error "Go executable not found. Install Go and ensure it is on PATH."
    exit 1
}

switch ($Task) {
    "console" {
        & $goExe run .
    }
    "graphic" {
        & $goExe run ./cmd/graphic
    }
    "graphic-fullscreen" {
        & $goExe run ./cmd/graphic --fullscreen
    }
    "test" {
        & $goExe test ./...
    }
    "build" {
        & $goExe build ./...
    }
    "fmt" {
        & $goExe fmt ./...
    }
    "cover" {
        & $goExe test ./internal/game "-coverprofile=coverage.out" "-covermode=atomic"
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        & $goExe tool cover "-func=coverage.out"
        Remove-Item coverage.out -ErrorAction SilentlyContinue
    }
}

exit $LASTEXITCODE