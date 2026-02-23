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
	boardWidth          = 20
	boardHeight         = 15
	defaultCellSize     = 32
	defaultPadding      = 16
	defaultHUDHeight    = 56
	minCellSize         = 12
	defaultWindowWidth  = boardWidth*defaultCellSize + defaultPadding*2
	defaultWindowHeight = boardHeight*defaultCellSize + defaultPadding*2 + defaultHUDHeight
	tickRate            = 140 * time.Millisecond
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

type sceneLayout struct {
	boardX        float64
	boardY        float64
	cell          float64
	boardPixelW   float64
	boardPixelH   float64
	gridLineWidth float64
	snakeInset    float64
	foodInset     float64
	hudY          int
	hudLine2Y     int
	hudLine3Y     int
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

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	lay := computeLayout(sw, sh)

	ebitenutil.DrawRect(screen, lay.boardX, lay.boardY, lay.boardPixelW, lay.boardPixelH, boardColor)

	for y := 0; y <= boardHeight; y++ {
		lineY := lay.boardY + float64(y)*lay.cell
		ebitenutil.DrawRect(screen, lay.boardX, lineY, lay.boardPixelW, lay.gridLineWidth, gridColor)
	}
	for x := 0; x <= boardWidth; x++ {
		lineX := lay.boardX + float64(x)*lay.cell
		ebitenutil.DrawRect(screen, lineX, lay.boardY, lay.gridLineWidth, lay.boardPixelH, gridColor)
	}

	food := a.state.Food()
	foodX := lay.boardX + float64(food.X)*lay.cell + lay.foodInset
	foodY := lay.boardY + float64(food.Y)*lay.cell + lay.foodInset
	foodSize := lay.cell - lay.foodInset*2
	ebitenutil.DrawRect(screen, foodX, foodY, foodSize, foodSize, foodColor)

	snake := a.state.Snake()
	for i, p := range snake {
		x := lay.boardX + float64(p.X)*lay.cell + lay.snakeInset
		y := lay.boardY + float64(p.Y)*lay.cell + lay.snakeInset
		size := lay.cell - lay.snakeInset*2
		c := snakeBodyColor
		if i == 0 {
			c = snakeHeadColor
		}
		ebitenutil.DrawRect(screen, x, y, size, size, c)
	}

	hud := fmt.Sprintf("SNAKE | Score: %d", a.state.Score())
	ebitenutil.DebugPrintAt(screen, hud, defaultPadding, lay.hudY)
	ebitenutil.DebugPrintAt(screen, "WASD/Arrows move | F11 fullscreen | Q/Esc quit", defaultPadding, lay.hudLine2Y)

	if !a.state.Started() {
		ebitenutil.DebugPrintAt(screen, "Press any direction key to start", defaultPadding, lay.hudLine3Y)
	}
	if a.state.IsOver() {
		ebitenutil.DebugPrintAt(screen, "Game Over | R restart | Q/Esc quit", defaultPadding, lay.hudLine3Y)
	}
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return defaultWindowWidth, defaultWindowHeight
	}
	return outsideWidth, outsideHeight
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

	ebiten.SetWindowSize(defaultWindowWidth, defaultWindowHeight)
	ebiten.SetWindowTitle("Snake - Graphic Mode")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(*fullscreen)

	if err := ebiten.RunGame(newApp()); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

func computeLayout(screenW, screenH int) sceneLayout {
	usableW := float64(screenW - defaultPadding*2)
	usableH := float64(screenH - defaultHUDHeight - defaultPadding*2)
	if usableW < 1 {
		usableW = 1
	}
	if usableH < 1 {
		usableH = 1
	}

	cell := minFloat(usableW/float64(boardWidth), usableH/float64(boardHeight))
	if cell < minCellSize {
		cell = minCellSize
	}

	boardW := cell * float64(boardWidth)
	boardH := cell * float64(boardHeight)
	boardX := (float64(screenW) - boardW) / 2
	boardY := float64(defaultPadding+defaultHUDHeight) + (usableH-boardH)/2
	minBoardY := float64(defaultPadding + defaultHUDHeight)
	if boardY < minBoardY {
		boardY = minBoardY
	}

	gridLine := maxFloat(1, cell*0.04)
	snakeInset := maxFloat(1, cell*0.08)
	foodInset := maxFloat(2, cell*0.22)

	return sceneLayout{
		boardX:        boardX,
		boardY:        boardY,
		cell:          cell,
		boardPixelW:   boardW,
		boardPixelH:   boardH,
		gridLineWidth: gridLine,
		snakeInset:    snakeInset,
		foodInset:     foodInset,
		hudY:          defaultPadding,
		hudLine2Y:     defaultPadding + 20,
		hudLine3Y:     defaultPadding + 36,
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
