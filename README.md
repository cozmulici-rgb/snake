# Snake (Go, Windows)

Snake game with both console and graphical modes.

## Requirements
- Windows
- Go 1.24+ installed and available on PATH

## Run

### Console mode
From this folder:

```powershell
go fmt ./...
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

## Controls
- `W` / `Up Arrow` up
- `A` / `Left Arrow` left
- `S` / `Down Arrow` down
- `D` / `Right Arrow` right
- `Q` or `Esc` quit
- `R` restart (game over screen)
- `F11` toggle fullscreen (graphic mode)
