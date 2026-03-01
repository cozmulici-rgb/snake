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
go run ./cmd/console
```

Or use the launcher:

```powershell
.\run.bat
```

### Graphic mode
Default mode (`go run .`) and desktop window mode:

```powershell
go run .
```

Start directly in fullscreen:

```powershell
go run . --fullscreen
```

Alternative explicit graphic entrypoint:

```powershell
go run ./cmd/graphic
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
.\scripts\dev.ps1 cover
```

## Controls
- `W` / `Up Arrow` up
- `A` / `Left Arrow` left
- `S` / `Down Arrow` down
- `D` / `Right Arrow` right
- `Q` or `Esc` quit
- `R` restart (game over screen)
- `M` back to menu (graphic mode game-over screen)
- `P` pause/resume
- `F11` toggle fullscreen (graphic mode)

### Graphic Menu Controls
- `Up` / `Down` (or `W` / `S`) select mode preset
- `1`, `2`, `3` quick preset select
- `Enter` or `Space` start game

## Sprint A Features
- Live stats bar in both modes: score, snake length, level, food eaten, foods to next level, elapsed time, and speed.
- Session best stats: best score, best length, best survival time, total runs, and total play time.
- Level progression: level increases every 5 food items, and movement speed increases each level (with a minimum speed cap).

## Sprint B Features
- Level obstacles: each new level adds obstacle blocks that must be avoided.
- Obstacle-aware spawning: food never appears on snake or obstacle tiles.
- Pause/resume support in both modes.
- Game-over run summary with deltas vs previous best stats.

## Sprint C Features
- Persistent profile stats across app restarts (best score/length/time, runs played, total food, total play time).
- Profile is stored at `%APPDATA%\\snake\\profile.json` on Windows.

## CI
- GitHub Actions runs `go vet ./...`, `go test ./...`, and `go build ./...` on pushes and pull requests.
- Coverage gate: `internal/game` must stay at or above 80% statement coverage.

## Release
- Tag a version like `v1.0.0` and push the tag.
- Release workflow builds Windows binaries for console and graphic modes and uploads:
  - `snake-windows-amd64.zip`
  - `snake-windows-amd64.zip.sha256`
