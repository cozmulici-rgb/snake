# Snake (Go, Windows)

Snake is a Windows game written in Go with two play modes:
- Graphic mode (default)
- Console mode

This branch also includes an experimental browser-based Three.js frontend served by Go.

## Requirements
- Windows
- Go 1.24+

## Quick Start

Run graphic mode (default):

```powershell
go run .
```

Run graphic mode in fullscreen:

```powershell
go run . --fullscreen
```

Run console mode:

```powershell
go run ./cmd/console
```

Run the experimental web mode:

```powershell
go run ./cmd/web
```

Then open `http://127.0.0.1:8080` in a browser.

Use the Windows launcher:

```powershell
.\run.bat
```

Use the helper script:

```powershell
.\scripts\dev.ps1 graphic
.\scripts\dev.ps1 graphic-fullscreen
.\scripts\dev.ps1 console
```

Notes for web mode:
- The browser UI currently loads `three` from a CDN import map.
- Game rules and ticking still run in Go; the browser is only the rendering/input layer.

## Compile Binaries

Build Windows executables:

```powershell
go build -o snake-graphic.exe .
go build -o snake-console.exe ./cmd/console
```

Build all packages (sanity compile):

```powershell
go build ./...
```

## Game Rules
- The snake does not move until you press the first direction key.
- Eat food to increase score and snake length.
- Every 5 foods, the level increases.
- Each level increases speed.
- Obstacles are added as levels increase.
- You lose if you hit a wall, your body, or an obstacle.
- You win if the board fills and no food can be placed.

## Controls

In game:
- `W` or `Up` move up
- `A` or `Left` move left
- `S` or `Down` move down
- `D` or `Right` move right
- `P` pause/resume
- `Q` or `Esc` quit
- `R` restart after game over

Graphic mode only:
- `F11` toggle fullscreen
- `M` return to menu on game-over screen

Graphic menu:
- `Up`/`Down` or `W`/`S` change preset
- `1`, `2`, `3` quick-select preset
- `Enter` or `Space` start game

## Difficulty Presets (Graphic Mode)
- `Balanced` default progression.
- `Relaxed` slower speed and softer obstacle growth.
- `Hardcore` faster speed and denser obstacles.

## HUD and Statistics

Live HUD (both modes) shows:
- Current score
- Snake length
- Level
- Food eaten
- Foods needed for next level
- Obstacle count
- Elapsed time
- Current speed (cells/second)

Session and profile stats:
- Best score
- Best length
- Best survival time
- Runs played
- Total food eaten
- Total play time

After game over, run summary includes deltas vs previous best values.

## Saved Profile
- Stats persist automatically between sessions.
- Profile path on Windows: `%APPDATA%\snake\profile.json`

## Developer Architecture
- Layered architecture follows `cmd -> ui -> app -> domain`.
- `internal/domain/gameplay` contains pure gameplay rules and invariants.
- `internal/app/session` orchestrates use-cases and exposes immutable snapshots.
- `internal/infra/profile` and `internal/infra/system` provide storage/clock/rng adapters.
- `internal/ui/console` and `internal/ui/graphic` adapt platform input/rendering to the app layer.
- Architecture transition is complete; current boundaries are documented in `docs/architecture/target-architecture.md`.

## Documentation
- User/developer setup and run guide: this `README.md`.
- Internal architecture docs: `docs/README.md`.
- ADRs: `docs/adr/`.

## Troubleshooting
- `go` not found: install Go, reopen the terminal, or run via `.\scripts\dev.ps1` (it also checks `C:\Program Files\Go\bin\go.exe`).
- Console input errors: run in PowerShell/Windows Terminal with focus on the game terminal; avoid terminals that block raw keyboard input.
- Reset profile data: delete `%APPDATA%\snake\profile.json` and start the game again.

## Developer Commands

```powershell
.\scripts\dev.ps1 test
.\scripts\dev.ps1 arch
.\scripts\dev.ps1 build
.\scripts\dev.ps1 fmt
.\scripts\dev.ps1 cover
```
