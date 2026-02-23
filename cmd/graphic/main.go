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
	defaultHUDHeight    = 92
	minCellSize         = 12
	defaultWindowWidth  = boardWidth*defaultCellSize + defaultPadding*2
	defaultWindowHeight = boardHeight*defaultCellSize + defaultPadding*2 + defaultHUDHeight
	baseTick            = 140 * time.Millisecond
	minTick             = 70 * time.Millisecond
	levelStep           = 8 * time.Millisecond
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
	state        *game.State
	lastTickAt   time.Time
	tickInterval time.Duration
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
	hudLine4Y     int
}

func newApp() *app {
	state := game.New(game.Config{Width: boardWidth, Height: boardHeight}, nil)
	return &app{
		state:        state,
		lastTickAt:   time.Now(),
		tickInterval: state.TickInterval(baseTick, minTick, levelStep),
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
			a.tickInterval = a.state.TickInterval(baseTick, minTick, levelStep)
			a.lastTickAt = time.Now()
		}
		return nil
	}

	if dir, ok := readDirection(); ok {
		a.state.SetDirection(dir)
	}

	now := time.Now()
	for now.Sub(a.lastTickAt) >= a.tickInterval {
		a.state.Tick()
		a.lastTickAt = a.lastTickAt.Add(a.tickInterval)
		a.tickInterval = a.state.TickInterval(baseTick, minTick, levelStep)
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

	speed := 0.0
	if a.tickInterval > 0 {
		speed = 1 / a.tickInterval.Seconds()
	}

	line1 := fmt.Sprintf("Score:%d  Length:%d  Level:%d", a.state.Score(), a.state.SnakeLength(), a.state.Level())
	line2 := fmt.Sprintf("Food:%d  NextLvl:%d  Time:%s  Speed:%.1f/s", a.state.FoodEaten(), a.state.FoodsToNextLevel(), formatDuration(a.state.Elapsed()), speed)
	line3 := fmt.Sprintf("BestScore:%d  BestLen:%d  BestTime:%s  Runs:%d", a.state.BestScore(), a.state.BestLength(), formatDuration(a.state.BestDuration()), a.state.RunsPlayed())

	ebitenutil.DebugPrintAt(screen, line1, defaultPadding, lay.hudY)
	ebitenutil.DebugPrintAt(screen, line2, defaultPadding, lay.hudLine2Y)
	ebitenutil.DebugPrintAt(screen, line3, defaultPadding, lay.hudLine3Y)

	msg := "WASD/Arrows move | F11 fullscreen | Q/Esc quit"
	if !a.state.Started() {
		msg = "Press any direction key to start"
	}
	if a.state.IsOver() {
		if a.state.IsWon() {
			msg = "You win! R restart | Q/Esc quit"
		} else {
			msg = "Game Over | R restart | Q/Esc quit"
		}
	}
	ebitenutil.DebugPrintAt(screen, msg, defaultPadding, lay.hudLine4Y)
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
		hudLine4Y:     defaultPadding + 54,
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

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
