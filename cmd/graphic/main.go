package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"snake/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	boardWidth    = 20
	boardHeight   = 15
	cellSize      = 32
	padding       = 16
	hudHeight     = 56
	tickRate      = 140 * time.Millisecond
	boardPixelW   = boardWidth * cellSize
	boardPixelH   = boardHeight * cellSize
	windowWidth   = boardPixelW + padding*2
	windowHeight  = boardPixelH + padding*2 + hudHeight
	gridLineWidth = 1
	snakeInset    = 2
	foodInset     = 6
)

var (
	bgColor        = color.RGBA{16, 18, 23, 255}
	boardColor     = color.RGBA{28, 34, 44, 255}
	gridColor      = color.RGBA{37, 45, 58, 255}
	snakeBodyColor = color.RGBA{64, 200, 120, 255}
	snakeHeadColor = color.RGBA{94, 235, 145, 255}
	foodColor      = color.RGBA{245, 95, 78, 255}
)

type app struct {
	state      *game.State
	lastTickAt time.Time
}

func newApp() *app {
	state := game.New(game.Config{Width: boardWidth, Height: boardHeight}, nil)
	return &app{
		state:      state,
		lastTickAt: time.Now(),
	}
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if a.state.IsOver() {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			a.state.Reset()
			a.lastTickAt = time.Now()
		}
		return nil
	}

	if dir, ok := readDirection(); ok {
		a.state.SetDirection(dir)
	}

	now := time.Now()
	for now.Sub(a.lastTickAt) >= tickRate {
		a.state.Tick()
		a.lastTickAt = a.lastTickAt.Add(tickRate)
		if a.state.IsOver() {
			break
		}
	}

	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	boardX := padding
	boardY := padding + hudHeight

	ebitenutil.DrawRect(screen, float64(boardX), float64(boardY), float64(boardPixelW), float64(boardPixelH), boardColor)

	for y := 0; y <= boardHeight; y++ {
		lineY := boardY + y*cellSize
		ebitenutil.DrawRect(screen, float64(boardX), float64(lineY), float64(boardPixelW), gridLineWidth, gridColor)
	}
	for x := 0; x <= boardWidth; x++ {
		lineX := boardX + x*cellSize
		ebitenutil.DrawRect(screen, float64(lineX), float64(boardY), gridLineWidth, float64(boardPixelH), gridColor)
	}

	food := a.state.Food()
	foodX := boardX + food.X*cellSize + foodInset
	foodY := boardY + food.Y*cellSize + foodInset
	foodSize := cellSize - foodInset*2
	ebitenutil.DrawRect(screen, float64(foodX), float64(foodY), float64(foodSize), float64(foodSize), foodColor)

	snake := a.state.Snake()
	for i, p := range snake {
		x := boardX + p.X*cellSize + snakeInset
		y := boardY + p.Y*cellSize + snakeInset
		size := cellSize - snakeInset*2
		c := snakeBodyColor
		if i == 0 {
			c = snakeHeadColor
		}
		ebitenutil.DrawRect(screen, float64(x), float64(y), float64(size), float64(size), c)
	}

	hud := fmt.Sprintf("SNAKE | Score: %d", a.state.Score())
	ebitenutil.DebugPrintAt(screen, hud, padding, padding)
	ebitenutil.DebugPrintAt(screen, "WASD/Arrows move | F11 fullscreen | Q/Esc quit", padding, padding+20)

	if !a.state.Started() {
		ebitenutil.DebugPrintAt(screen, "Press any direction key to start", padding, padding+36)
	}
	if a.state.IsOver() {
		ebitenutil.DebugPrintAt(screen, "Game Over | R restart | Q/Esc quit", padding, padding+36)
	}
}

func (a *app) Layout(_, _ int) (int, int) {
	return windowWidth, windowHeight
}

func readDirection() (game.Direction, bool) {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyW), inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		return game.DirUp, true
	case inpututil.IsKeyJustPressed(ebiten.KeyS), inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		return game.DirDown, true
	case inpututil.IsKeyJustPressed(ebiten.KeyA), inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		return game.DirLeft, true
	case inpututil.IsKeyJustPressed(ebiten.KeyD), inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		return game.DirRight, true
	default:
		return game.DirNone, false
	}
}

func main() {
	fullscreen := flag.Bool("fullscreen", false, "start in fullscreen mode")
	flag.Parse()

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Snake - Graphic Mode")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(*fullscreen)

	if err := ebiten.RunGame(newApp()); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
