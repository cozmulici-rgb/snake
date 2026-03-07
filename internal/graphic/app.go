//go:build windows

package graphic

import (
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"snake/internal/app/session"
	infprofile "snake/internal/infra/profile"
	"snake/internal/infra/system"
	graphicui "snake/internal/ui/graphic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	ebitentext "github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
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
	debugCharWidth      = 7
	menuCardHeight      = 34
	menuCardSpacing     = 12
	menuAccentWidth     = 5
	overlayPanelWidth   = 0.68
	overlayPanelHeight  = 0.38
	overlayMinPanelW    = 280
	overlayMinPanelH    = 120
	menuTitleFontPath   = "assets/fonts/DigitaltsStrawberry-DYMR1.ttf"
	menuBodyFontPath    = "assets/fonts/DigitaltsOrange-nRAPg.ttf"
	menuSmallFontPath   = "assets/fonts/DigitaltsLime-lgxmd.ttf"
)

var (
	bgColor         = color.RGBA{16, 18, 23, 255}
	boardColor      = color.RGBA{28, 34, 44, 255}
	gridColor       = color.RGBA{37, 45, 58, 255}
	obstacleColor   = color.RGBA{148, 158, 172, 255}
	snakeBodyColor  = color.RGBA{64, 200, 120, 255}
	snakeHeadColor  = color.RGBA{94, 235, 145, 255}
	foodColor       = color.RGBA{245, 95, 78, 255}
	menuCardColor   = color.RGBA{29, 35, 47, 255}
	menuSelected    = color.RGBA{42, 52, 70, 255}
	menuAccent      = color.RGBA{50, 120, 206, 255}
	menuScrim       = color.RGBA{12, 24, 46, 24}
	menuPanel       = color.RGBA{234, 242, 252, 240}
	menuPanelSoft   = color.RGBA{214, 226, 244, 240}
	menuCardBase    = color.RGBA{246, 250, 255, 248}
	menuCardHot     = color.RGBA{255, 243, 201, 248}
	menuCta         = color.RGBA{52, 123, 219, 240}
	menuCtaText     = color.RGBA{250, 252, 255, 255}
	menuTextMain    = color.RGBA{24, 38, 57, 255}
	menuTextMuted   = color.RGBA{52, 71, 96, 255}
	menuOutlineDark = color.RGBA{10, 21, 36, 210}
	overlayScrim    = color.RGBA{12, 14, 18, 180}
	overlayPanel    = color.RGBA{26, 31, 42, 235}
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
	service        session.SessionService
	profileData    session.Profile
	menuTitleFace  font.Face
	menuBodyFace   font.Face
	menuSmallFace  font.Face
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
		menuTitleFace:  loadMenuFontFace([]string{menuTitleFontPath, menuBodyFontPath, menuSmallFontPath}, 38),
		menuBodyFace:   loadMenuFontFace([]string{menuBodyFontPath, menuTitleFontPath, menuSmallFontPath}, 22),
		menuSmallFace:  loadMenuFontFace([]string{menuSmallFontPath, menuBodyFontPath, menuTitleFontPath}, 18),
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
	if overlay, ok := graphicui.BuildStartOverlay(a.currentPreset.name, snap); ok {
		a.drawStartOverlay(screen, lay, overlay)
	}
}

func (a *app) drawMenu(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	a.drawMenuBackground(screen)
	menu := graphicui.BuildMainMenu(a.selectedPreset, toMenuPresets(presets), a.profileData)

	panelW := minFloat(float64(sw-defaultPadding*2), 920)
	panelH := minFloat(float64(sh-defaultPadding*2), 640)
	panelX := (float64(sw) - panelW) / 2
	panelY := (float64(sh) - panelH) / 2

	ebitenutil.DrawRect(screen, panelX+8, panelY+10, panelW, panelH, color.RGBA{78, 96, 122, 96})
	ebitenutil.DrawRect(screen, panelX, panelY, panelW, panelH, menuPanel)
	ebitenutil.DrawRect(screen, panelX, panelY, panelW, 4, menuAccent)
	ebitenutil.DrawRect(screen, panelX+16, panelY+92, panelW-32, 1, color.RGBA{80, 110, 146, 120})

	title := menu.Title
	if title == "" {
		title = "SNAKE"
	}
	titleX := centeredTextInRect(title, int(panelX), int(panelW), a.menuTitleFace)
	drawMenuText(screen, a.menuTitleFace, title, titleX, int(panelY)+18, menuTextMain)

	subtitle := menu.Subtitle
	subX := centeredTextInRect(subtitle, int(panelX), int(panelW), a.menuBodyFace)
	drawMenuText(screen, a.menuBodyFace, subtitle, subX, int(panelY)+60, menuTextMuted)

	contentX := panelX + 18
	contentW := panelW - 36
	cardYBase := panelY + 110
	cardH := 62.0
	cardGap := 10.0
	for i, p := range presets {
		y := cardYBase + float64(i)*(cardH+cardGap)
		selected := i == a.selectedPreset
		cardColor := menuCardBase
		if selected {
			cardColor = menuCardHot
		}
		drawMenuCapsule(screen, contentX, y, contentW, cardH, menuPanelSoft, cardColor)
		if selected {
			ebitenutil.DrawRect(screen, contentX, y, menuAccentWidth, cardH, menuAccent)
			label := "ACTIVE"
			labelW := float64(textWidth(label, a.menuSmallFace) + 18)
			labelX := contentX + contentW - labelW - 14
			drawMenuCapsule(screen, labelX, y+12, labelW, 36, menuCta, color.RGBA{74, 143, 234, 245})
			drawMenuText(screen, a.menuSmallFace, label, int(labelX)+9, int(y)+20, menuCtaText)
		}
		name := fmt.Sprintf("%d. %s", i+1, p.name)
		drawMenuText(screen, a.menuBodyFace, name, int(contentX)+16, int(y)+10, menuTextMain)
		drawMenuText(screen, a.menuSmallFace, p.description, int(contentX)+16, int(y)+36, menuTextMuted)
	}

	ctaY := panelY + panelH - 154
	ctaH := 48.0
	ctaW := contentW
	drawMenuCapsule(screen, contentX, ctaY, ctaW, ctaH, menuCta, color.RGBA{74, 143, 234, 245})
	ctaText := fmt.Sprintf("Press Enter or Space to Start (%s)", presets[a.selectedPreset].name)
	ctaX := centeredTextInRect(ctaText, int(contentX), int(ctaW), a.menuBodyFace)
	drawMenuText(screen, a.menuBodyFace, ctaText, ctaX, int(ctaY)+10, menuCtaText)

	helpY := ctaY + ctaH + 10
	drawMenuCapsule(screen, contentX, helpY, ctaW, 36, menuPanelSoft, color.RGBA{230, 238, 250, 240})
	helpX := centeredTextInRect(menu.HelpLine, int(contentX), int(ctaW), a.menuSmallFace)
	drawMenuText(screen, a.menuSmallFace, menu.HelpLine, helpX, int(helpY)+8, menuTextMuted)

	statsY := helpY + 44
	stats1X := centeredTextInRect(menu.StatsLine1, int(contentX), int(ctaW), a.menuSmallFace)
	stats2X := centeredTextInRect(menu.StatsLine2, int(contentX), int(ctaW), a.menuSmallFace)
	drawMenuText(screen, a.menuSmallFace, menu.StatsLine1, stats1X, int(statsY), menuTextMuted)
	drawMenuText(screen, a.menuSmallFace, menu.StatsLine2, stats2X, int(statsY)+20, menuTextMuted)
}

func (a *app) drawMenuBackground(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if sw <= 0 || sh <= 0 {
		return
	}
	for i := 0; i < 16; i++ {
		h := float64(sh) / 16
		y := float64(i) * h
		ebitenutil.DrawRect(screen, 0, y, float64(sw), h+1, color.RGBA{
			R: 126 + uint8(i*4),
			G: 170 + uint8(i*3),
			B: 222 + uint8(i*2),
			A: 255,
		})
	}
	ebitenutil.DrawRect(screen, 0, float64(sh)*0.76, float64(sw), float64(sh)*0.24, color.RGBA{176, 211, 181, 226})
	ebitenutil.DrawRect(screen, 0, 0, float64(sw), float64(sh), menuScrim)
}

func (a *app) drawStartOverlay(screen *ebiten.Image, lay sceneLayout, overlay graphicui.StartOverlay) {
	ebitenutil.DrawRect(screen, lay.boardX, lay.boardY, lay.boardPixelW, lay.boardPixelH, overlayScrim)

	panelW := lay.boardPixelW * overlayPanelWidth
	panelH := lay.boardPixelH * overlayPanelHeight
	if panelW < overlayMinPanelW {
		panelW = overlayMinPanelW
	}
	if panelH < overlayMinPanelH {
		panelH = overlayMinPanelH
	}
	maxW := lay.boardPixelW - 24
	maxH := lay.boardPixelH - 24
	if panelW > maxW {
		panelW = maxW
	}
	if panelH > maxH {
		panelH = maxH
	}

	panelX := lay.boardX + (lay.boardPixelW-panelW)/2
	panelY := lay.boardY + (lay.boardPixelH-panelH)/2
	ebitenutil.DrawRect(screen, panelX, panelY, panelW, panelH, overlayPanel)
	ebitenutil.DrawRect(screen, panelX, panelY, panelW, 3, menuAccent)

	centerX := int(panelX + panelW/2)
	titleY := int(panelY) + 20
	subtitleY := titleY + 22
	hintY := subtitleY + 24
	ebitenutil.DebugPrintAt(screen, overlay.Title, centerX-(len(overlay.Title)*debugCharWidth)/2, titleY)
	ebitenutil.DebugPrintAt(screen, overlay.Subtitle, centerX-(len(overlay.Subtitle)*debugCharWidth)/2, subtitleY)
	ebitenutil.DebugPrintAt(screen, overlay.Hint, centerX-(len(overlay.Hint)*debugCharWidth)/2, hintY)
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

func centeredTextX(text string, screenW int, face font.Face) int {
	if face == nil {
		return (screenW - len(text)*debugCharWidth) / 2
	}
	return (screenW - ebitentext.BoundString(face, text).Dx()) / 2
}

func centeredTextInRect(text string, x, w int, face font.Face) int {
	return x + (w-textWidth(text, face))/2
}

func textWidth(text string, face font.Face) int {
	if face == nil {
		return len(text) * debugCharWidth
	}
	return ebitentext.BoundString(face, text).Dx()
}

func toMenuPresets(in []preset) []graphicui.MenuPreset {
	out := make([]graphicui.MenuPreset, len(in))
	for i, p := range in {
		out[i] = graphicui.MenuPreset{
			Name:        p.name,
			Description: p.description,
		}
	}
	return out
}

func loadMenuFontFace(paths []string, size float64) font.Face {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		tt, err := opentype.Parse(data)
		if err != nil {
			log.Printf("warning: failed to parse font %q: %v", p, err)
			continue
		}
		face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			log.Printf("warning: failed to create font face %q: %v", p, err)
			continue
		}
		return face
	}
	return nil
}

func drawMenuText(screen *ebiten.Image, face font.Face, text string, x, y int, clr color.Color) {
	if face == nil {
		ebitenutil.DebugPrintAt(screen, text, x, y)
		return
	}
	ascent := face.Metrics().Ascent.Ceil()
	ebitentext.Draw(screen, text, face, x, y+ascent, clr)
}

func drawMenuTextOutlined(screen *ebiten.Image, face font.Face, text string, x, y int, clr color.Color, outline color.Color) {
	drawMenuText(screen, face, text, x-1, y, outline)
	drawMenuText(screen, face, text, x+1, y, outline)
	drawMenuText(screen, face, text, x, y-1, outline)
	drawMenuText(screen, face, text, x, y+1, outline)
	drawMenuText(screen, face, text, x, y, clr)
}

func drawMenuCapsule(screen *ebiten.Image, x, y, w, h float64, outer color.Color, inner color.Color) {
	ebitenutil.DrawRect(screen, x, y, w, h, outer)
	ebitenutil.DrawRect(screen, x+2, y+2, w-4, h-4, inner)
}
