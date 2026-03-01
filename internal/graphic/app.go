//go:build windows

package graphic

import (
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"snake/internal/app/session"
	infprofile "snake/internal/infra/profile"
	"snake/internal/infra/system"
	graphicui "snake/internal/ui/graphic"

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
)

var (
	bgColor        = color.RGBA{16, 18, 23, 255}
	boardColor     = color.RGBA{28, 34, 44, 255}
	gridColor      = color.RGBA{37, 45, 58, 255}
	obstacleColor  = color.RGBA{148, 158, 172, 255}
	snakeBodyColor = color.RGBA{64, 200, 120, 255}
	snakeHeadColor = color.RGBA{94, 235, 145, 255}
	foodColor      = color.RGBA{245, 95, 78, 255}
)

type sceneState int

const (
	sceneMenu sceneState = iota
	sceneGame
)

type preset struct {
	name          string
	description   string
	baseTick      time.Duration
	minTick       time.Duration
	levelStep     time.Duration
	foodsPerLevel int
	obstaclesStep int
}

var presets = []preset{
	{
		name:          "Balanced",
		description:   "Default progression and obstacle density",
		baseTick:      140 * time.Millisecond,
		minTick:       70 * time.Millisecond,
		levelStep:     8 * time.Millisecond,
		foodsPerLevel: 5,
		obstaclesStep: 2,
	},
	{
		name:          "Relaxed",
		description:   "Slower speed, gentler obstacle growth",
		baseTick:      180 * time.Millisecond,
		minTick:       95 * time.Millisecond,
		levelStep:     6 * time.Millisecond,
		foodsPerLevel: 6,
		obstaclesStep: 1,
	},
	{
		name:          "Hardcore",
		description:   "Faster speed, denser obstacles",
		baseTick:      120 * time.Millisecond,
		minTick:       55 * time.Millisecond,
		levelStep:     10 * time.Millisecond,
		foodsPerLevel: 4,
		obstaclesStep: 3,
	},
}

type app struct {
	scene          sceneState
	selectedPreset int
	currentPreset  preset
	service        *session.Service
	profileData    session.Profile
	lastTickAt     time.Time
	tickInterval   time.Duration
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
	hudLine5Y     int
}

func newApp() *app {
	svc := session.NewService(system.RealClock{}, system.NewMathRandom(0), infprofile.NewFileRepository(infprofile.DefaultPath()))
	if err := svc.LoadProfile(context.Background()); err != nil {
		log.Printf("warning: could not load profile: %v", err)
	}

	return &app{
		scene:          sceneMenu,
		selectedPreset: 0,
		service:        svc,
		profileData:    svc.Profile(),
	}
}

func (a *app) startGameFromSelectedPreset() {
	cfgPreset := presets[a.selectedPreset]
	err := a.service.Start(context.Background(), session.PresetConfig{
		Name:          cfgPreset.name,
		Width:         boardWidth,
		Height:        boardHeight,
		FoodsPerLevel: cfgPreset.foodsPerLevel,
		ObstaclesStep: cfgPreset.obstaclesStep,
		BaseTick:      cfgPreset.baseTick,
		MinTick:       cfgPreset.minTick,
		LevelStep:     cfgPreset.levelStep,
	})
	if err != nil {
		log.Printf("warning: could not start session: %v", err)
		return
	}

	a.currentPreset = cfgPreset
	a.tickInterval = a.service.Snapshot().TickInterval
	a.lastTickAt = time.Now()
	a.scene = sceneGame
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	switch a.scene {
	case sceneMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			a.selectedPreset = (a.selectedPreset - 1 + len(presets)) % len(presets)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			a.selectedPreset = (a.selectedPreset + 1) % len(presets)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			a.selectedPreset = 0
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) && len(presets) > 1 {
			a.selectedPreset = 1
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) && len(presets) > 2 {
			a.selectedPreset = 2
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			a.startGameFromSelectedPreset()
		}
		return nil
	case sceneGame:
		snap := a.service.Snapshot()
		if snap.Width == 0 || snap.Height == 0 {
			a.scene = sceneMenu
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) && !snap.IsOver {
			a.service.TogglePause()
			snap = a.service.Snapshot()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			_ = a.service.Quit()
			a.profileData = a.service.Profile()
			return ebiten.Termination
		}

		if snap.IsOver {
			a.profileData = a.service.Profile()
			if inpututil.IsKeyJustPressed(ebiten.KeyR) {
				if err := a.service.Restart(); err != nil {
					log.Printf("warning: could not restart session: %v", err)
					return nil
				}
				a.tickInterval = a.service.Snapshot().TickInterval
				a.lastTickAt = time.Now()
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyM) {
				a.scene = sceneMenu
			}
			return nil
		}

		if dir, ok := readDirection(); ok {
			a.service.ApplyDirection(dir)
			snap = a.service.Snapshot()
		}
		if snap.Paused {
			return nil
		}
		if a.tickInterval <= 0 {
			a.tickInterval = snap.TickInterval
		}

		now := time.Now()
		for now.Sub(a.lastTickAt) >= a.tickInterval {
			a.service.Tick()
			a.lastTickAt = a.lastTickAt.Add(a.tickInterval)
			snap = a.service.Snapshot()
			a.tickInterval = snap.TickInterval
			if snap.IsOver {
				a.profileData = a.service.Profile()
				break
			}
		}
		return nil
	default:
		return nil
	}
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	if a.scene == sceneMenu {
		a.drawMenu(screen)
		return
	}

	snap := a.service.Snapshot()
	if snap.Width == 0 || snap.Height == 0 {
		a.drawMenu(screen)
		return
	}

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

	foodX := lay.boardX + float64(snap.Food.X)*lay.cell + lay.foodInset
	foodY := lay.boardY + float64(snap.Food.Y)*lay.cell + lay.foodInset
	foodSize := lay.cell - lay.foodInset*2
	ebitenutil.DrawRect(screen, foodX, foodY, foodSize, foodSize, foodColor)

	for _, p := range snap.Obstacles {
		x := lay.boardX + float64(p.X)*lay.cell + lay.snakeInset
		y := lay.boardY + float64(p.Y)*lay.cell + lay.snakeInset
		size := lay.cell - lay.snakeInset*2
		ebitenutil.DrawRect(screen, x, y, size, size, obstacleColor)
	}

	for i, p := range snap.Snake {
		x := lay.boardX + float64(p.X)*lay.cell + lay.snakeInset
		y := lay.boardY + float64(p.Y)*lay.cell + lay.snakeInset
		size := lay.cell - lay.snakeInset*2
		c := snakeBodyColor
		if i == 0 {
			c = snakeHeadColor
		}
		ebitenutil.DrawRect(screen, x, y, size, size, c)
	}

	hud := graphicui.BuildHUD(a.currentPreset.name, snap)
	ebitenutil.DebugPrintAt(screen, hud.Line1, defaultPadding, lay.hudY)
	ebitenutil.DebugPrintAt(screen, hud.Line2, defaultPadding, lay.hudLine2Y)
	ebitenutil.DebugPrintAt(screen, hud.Line3, defaultPadding, lay.hudLine3Y)
	ebitenutil.DebugPrintAt(screen, hud.Msg, defaultPadding, lay.hudLine4Y)
	if hud.Detail != "" {
		ebitenutil.DebugPrintAt(screen, hud.Detail, defaultPadding, lay.hudLine5Y)
	}
}

func (a *app) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "SNAKE - MODE SELECT", defaultPadding, defaultPadding)
	ebitenutil.DebugPrintAt(screen, "Use Up/Down or 1-3, Enter to start, Q/Esc quit", defaultPadding, defaultPadding+20)

	baseY := defaultPadding + 48
	for i, p := range presets {
		prefix := "  "
		if i == a.selectedPreset {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%d. %s - %s", prefix, i+1, p.name, p.description)
		ebitenutil.DebugPrintAt(screen, line, defaultPadding, baseY+i*18)
	}

	bestLine := fmt.Sprintf("BestScore:%d  BestLen:%d  BestTime:%s  Runs:%d",
		a.profileData.BestScore,
		a.profileData.BestLength,
		graphicui.FormatDuration(time.Duration(a.profileData.BestDurationMillis)*time.Millisecond),
		a.profileData.RunsPlayed)
	ebitenutil.DebugPrintAt(screen, bestLine, defaultPadding, baseY+len(presets)*20+20)

	totalLine := fmt.Sprintf("TotalFood:%d  TotalPlay:%s", a.profileData.TotalFoodEaten, graphicui.FormatDuration(time.Duration(a.profileData.TotalPlayTimeMillis)*time.Millisecond))
	ebitenutil.DebugPrintAt(screen, totalLine, defaultPadding, baseY+len(presets)*20+40)
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return defaultWindowWidth, defaultWindowHeight
	}
	return outsideWidth, outsideHeight
}

func readDirection() (session.DirectionInput, bool) {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyW), inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		return session.DirectionUp, true
	case inpututil.IsKeyJustPressed(ebiten.KeyS), inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		return session.DirectionDown, true
	case inpututil.IsKeyJustPressed(ebiten.KeyA), inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		return session.DirectionLeft, true
	case inpututil.IsKeyJustPressed(ebiten.KeyD), inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		return session.DirectionRight, true
	default:
		return session.DirectionNone, false
	}
}

func Run() {
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
		hudLine5Y:     defaultPadding + 70,
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
