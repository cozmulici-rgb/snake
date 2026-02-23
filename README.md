# Snake (Go, Windows)

Snake game with both console and graphical modes.

## Requirements
- Windows
- Go 1.24+ installed and available on PATH

## Run

### Console mode
From this folder:

```powershell
go test ./...
go run .
```

Or use the launcher:

```powershell
.\run.bat
```

### Graphic mode
Runs in a desktop window:

```powershell
go run ./cmd/graphic
```

Start directly in fullscreen:

```powershell
go run ./cmd/graphic --fullscreen
```

### Dev script (PowerShell)
Use helper tasks:

```powershell
.\scripts\dev.ps1 console
.\scripts\dev.ps1 graphic
.\scripts\dev.ps1 graphic-fullscreen
.\scripts\dev.ps1 test
.\scripts\dev.ps1 build
.\scripts\dev.ps1 fmt
```

## Controls
- `W` / `Up Arrow` up
- `A` / `Left Arrow` left
- `S` / `Down Arrow` down
- `D` / `Right Arrow` right
- `Q` or `Esc` quit
- `R` restart (game over screen)
- `F11` toggle fullscreen (graphic mode)

## CI
- GitHub Actions runs `go vet ./...`, `go test ./...`, and `go build ./...` on pushes and pull requests.
