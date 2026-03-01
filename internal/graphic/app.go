package graphic

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"snake/internal/game"
	"snake/internal/profile"

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
	state          *game.State
	lastTickAt     time.Time
	tickInterval   time.Duration
	paused         bool
	profilePath    string
	profileData    game.Profile
	savedRuns      int
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
	path := profile.DefaultPath()
	var p game.Profile
	if loaded, err := profile.Load(path); err == nil {
		p = loaded
	} else if !os.IsNotExist(err) {
		log.Printf("warning: could not load profile: %v", err)
	}

	return &app{
		scene:          sceneMenu,
		selectedPreset: 0,
		profilePath:    path,
		profileData:    p,
	}
}

func (a *app) startGameFromSelectedPreset() {
	cfgPreset := presets[a.selectedPreset]
	state := game.New(game.Config{
		Width:         boardWidth,
		Height:        boardHeight,
		FoodsPerLevel: cfgPreset.foodsPerLevel,
		ObstaclesStep: cfgPreset.obstaclesStep,
	}, nil)
	state.ApplyProfile(a.profileData)

	a.currentPreset = cfgPreset
	a.state = state
	a.tickInterval = state.TickInterval(cfgPreset.baseTick, cfgPreset.minTick, cfgPreset.levelStep)
	a.lastTickAt = time.Now()
	a.paused = false
	a.savedRuns = state.RunsPlayed()
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
		if a.state == nil {
			a.scene = sceneMenu
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyP) && !a.state.IsOver() {
			a.paused = !a.paused
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.state.FinalizeNow()
			a.persistProfile()
			return ebiten.Termination
		}

		if a.state.IsOver() {
			a.persistProfile()
			if inpututil.IsKeyJustPressed(ebiten.KeyR) {
				a.state.Reset()
				a.persistProfile()
				a.tickInterval = a.state.TickInterval(a.currentPreset.baseTick, a.currentPreset.minTick, a.currentPreset.levelStep)
				a.lastTickAt = time.Now()
				a.paused = false
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyM) {
				a.scene = sceneMenu
			}
			return nil
		}
		if a.paused {
			return nil
		}

		if dir, ok := readDirection(); ok {
			a.state.SetDirection(dir)
		}

		now := time.Now()
		for now.Sub(a.lastTickAt) >= a.tickInterval {
			a.state.Tick()
			a.persistProfile()
			a.lastTickAt = a.lastTickAt.Add(a.tickInterval)
			a.tickInterval = a.state.TickInterval(a.currentPreset.baseTick, a.currentPreset.minTick, a.currentPreset.levelStep)
			if a.state.IsOver() {
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

	if a.state == nil {
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

	food := a.state.Food()
	foodX := lay.boardX + float64(food.X)*lay.cell + lay.foodInset
	foodY := lay.boardY + float64(food.Y)*lay.cell + lay.foodInset
	foodSize := lay.cell - lay.foodInset*2
	ebitenutil.DrawRect(screen, foodX, foodY, foodSize, foodSize, foodColor)

	obstacles := a.state.Obstacles()
	for _, p := range obstacles {
		x := lay.boardX + float64(p.X)*lay.cell + lay.snakeInset
		y := lay.boardY + float64(p.Y)*lay.cell + lay.snakeInset
		size := lay.cell - lay.snakeInset*2
		ebitenutil.DrawRect(screen, x, y, size, size, obstacleColor)
	}

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

	line1 := fmt.Sprintf("Mode:%s  Score:%d  Length:%d  Level:%d", a.currentPreset.name, a.state.Score(), a.state.SnakeLength(), a.state.Level())
	line2 := fmt.Sprintf("Food:%d  NextLvl:%d  Obst:%d  Time:%s  Speed:%.1f/s", a.state.FoodEaten(), a.state.FoodsToNextLevel(), a.state.ObstacleCount(), formatDuration(a.state.Elapsed()), speed)
	line3 := fmt.Sprintf("BestScore:%d  BestLen:%d  BestTime:%s  Runs:%d", a.state.BestScore(), a.state.BestLength(), formatDuration(a.state.BestDuration()), a.state.RunsPlayed())

	ebitenutil.DebugPrintAt(screen, line1, defaultPadding, lay.hudY)
	ebitenutil.DebugPrintAt(screen, line2, defaultPadding, lay.hudLine2Y)
	ebitenutil.DebugPrintAt(screen, line3, defaultPadding, lay.hudLine3Y)

	msg := "WASD/Arrows move | P pause | F11 fullscreen | Q/Esc quit"
	detail := ""
	if !a.state.Started() {
		msg = "Press any direction key to start"
	}
	if a.paused {
		msg = "Paused | P resume | Q/Esc quit"
	}
	if a.state.IsOver() {
		if a.state.IsWon() {
			msg = "You win! R restart | M menu | Q/Esc quit"
		} else {
			msg = "Game Over | R restart | M menu | Q/Esc quit"
		}
		if summary, ok := a.state.LastRunSummary(); ok {
			detail = fmt.Sprintf("Run: Score %d (%s)  Len %d (%s)  Time %s (%s)",
				summary.Score,
				formatSignedInt(summary.ScoreDeltaVsPrevBest),
				summary.Length,
				formatSignedInt(summary.LengthDeltaVsPrevBest),
				formatDuration(summary.Duration),
				formatSignedDuration(summary.DurationDeltaVsPrevBest))
		}
	}
	ebitenutil.DebugPrintAt(screen, msg, defaultPadding, lay.hudLine4Y)
	if detail != "" {
		ebitenutil.DebugPrintAt(screen, detail, defaultPadding, lay.hudLine5Y)
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
		formatDuration(time.Duration(a.profileData.BestDurationMillis)*time.Millisecond),
		a.profileData.RunsPlayed)
	ebitenutil.DebugPrintAt(screen, bestLine, defaultPadding, baseY+len(presets)*20+20)

	totalLine := fmt.Sprintf("TotalFood:%d  TotalPlay:%s", a.profileData.TotalFoodEaten, formatDuration(time.Duration(a.profileData.TotalPlayTimeMillis)*time.Millisecond))
	ebitenutil.DebugPrintAt(screen, totalLine, defaultPadding, baseY+len(presets)*20+40)
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

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatSignedInt(v int) string {
	if v > 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}

func formatSignedDuration(d time.Duration) string {
	if d >= 0 {
		return "+" + formatDuration(d)
	}
	return "-" + formatDuration(-d)
}

func (a *app) persistProfile() {
	if a.state == nil || a.state.RunsPlayed() == a.savedRuns {
		return
	}
	if err := profile.Save(a.profilePath, a.state.Profile()); err != nil {
		log.Printf("warning: could not save profile: %v", err)
		return
	}
	a.savedRuns = a.state.RunsPlayed()
	a.profileData = a.state.Profile()
}
