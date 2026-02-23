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
- `P` pause/resume
- `F11` toggle fullscreen (graphic mode)

## Sprint A Features
- Live stats bar in both modes: score, snake length, level, food eaten, foods to next level, elapsed time, and speed.
- Session best stats: best score, best length, best survival time, total runs, and total play time.
- Level progression: level increases every 5 food items, and movement speed increases each level (with a minimum speed cap).

## Sprint B Features
- Level obstacles: each new level adds obstacle blocks that must be avoided.
- Obstacle-aware spawning: food never appears on snake or obstacle tiles.
- Pause/resume support in both modes.
- Game-over run summary with deltas vs previous best stats.

## CI
- GitHub Actions runs `go vet ./...`, `go test ./...`, and `go build ./...` on pushes and pull requests.
