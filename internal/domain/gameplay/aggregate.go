package gameplay

import (
	"errors"
	"math/rand"
	"time"
)

type Game struct {
	width         int
	height        int
	foodsPerLevel int
	obstaclesStep int
	snake         []Point
	obstacles     []Point
	dir           Direction
	food          Point
	score         int
	foodEaten     int
	level         int
	started       bool
	over          bool
	won           bool
	paused        bool
	startedAt     time.Time
	endedAt       time.Time
	runFinalized  bool
	bestScore     int
	bestLength    int
	bestDuration  time.Duration
	runsPlayed    int
	totalFood     int
	totalPlayTime time.Duration
	lastRun       RunSummary
	hasLastRun    bool
	levelOffset   int
	developerMode bool
	rng           RNG
}

func New(cfg Config, rng RNG) *Game {
	if cfg.Width <= 0 {
		cfg.Width = 20
	}
	if cfg.Height <= 0 {
		cfg.Height = 15
	}
	if cfg.FoodsPerLevel <= 0 {
		cfg.FoodsPerLevel = DefaultFoodsPerLevel
	}
	if cfg.ObstaclesStep < 0 {
		cfg.ObstaclesStep = 0
	}
	if cfg.ObstaclesStep == 0 {
		cfg.ObstaclesStep = DefaultObstaclesStep
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}

	g := &Game{
		width:         cfg.Width,
		height:        cfg.Height,
		foodsPerLevel: cfg.FoodsPerLevel,
		obstaclesStep: cfg.ObstaclesStep,
		rng:           rng,
	}
	g.Reset(time.Time{})
	return g
}

func (g *Game) Start(dir Direction, now time.Time) bool {
	if !isCardinal(dir) || g.over {
		return false
	}
	if g.started {
		return g.SetDirection(dir)
	}
	g.dir = dir
	g.started = true
	g.startedAt = now
	return true
}

func (g *Game) SetDirection(dir Direction) bool {
	if !isCardinal(dir) || g.over {
		return false
	}
	if !g.started {
		return false
	}
	if isOpposite(g.dir, dir) {
		return false
	}
	g.dir = dir
	return true
}

func (g *Game) Tick(now time.Time) {
	if g.over || !g.started || g.paused {
		return
	}

	head := g.snake[0]
	next := Point{X: head.X + dirX(g.dir), Y: head.Y + dirY(g.dir)}

	if next.X < 0 || next.Y < 0 || next.X >= g.width || next.Y >= g.height {
		g.over = true
		g.endedAt = now
		g.finalizeRun(now)
		return
	}
	if contains(g.snake, next) {
		g.over = true
		g.endedAt = now
		g.finalizeRun(now)
		return
	}
	if contains(g.obstacles, next) {
		g.over = true
		g.endedAt = now
		g.finalizeRun(now)
		return
	}

	g.snake = append([]Point{next}, g.snake...)
	if next == g.food {
		prevLevel := g.level
		g.score++
		g.foodEaten++
		g.level = g.effectiveLevel()
		if g.level != prevLevel {
			g.regenerateObstacles()
		}
		if !g.placeFood() {
			g.over = true
			g.won = true
			g.endedAt = now
			g.finalizeRun(now)
		}
		return
	}

	g.snake = g.snake[:len(g.snake)-1]
}

func (g *Game) TogglePause() bool {
	if g.over {
		return g.paused
	}
	g.paused = !g.paused
	return g.paused
}

func (g *Game) Reset(now time.Time) {
	g.finalizeRun(now)

	g.snake = []Point{{X: g.width / 2, Y: g.height / 2}}
	g.obstacles = nil
	g.dir = DirNone
	g.score = 0
	g.foodEaten = 0
	g.level = 1
	g.levelOffset = 0
	g.started = false
	g.over = false
	g.won = false
	g.paused = false
	g.startedAt = time.Time{}
	g.endedAt = time.Time{}
	g.runFinalized = false
	if !g.placeFood() {
		g.over = true
		g.won = true
	}
}

func (g *Game) Finalize(now time.Time) {
	if !g.started || g.runFinalized {
		return
	}
	if g.endedAt.IsZero() {
		g.endedAt = now
	}
	g.finalizeRun(now)
}

func (g *Game) SetDeveloperMode(enabled bool) {
	g.developerMode = enabled
}

func (g *Game) BypassToLevel(level int) error {
	if level < 1 {
		return errors.New("level must be at least 1")
	}
	if g.over {
		return errors.New("cannot bypass level after game over")
	}

	g.levelOffset = level - g.baseLevel()
	g.level = g.effectiveLevel()
	g.regenerateObstacles()
	if contains(g.obstacles, g.food) && !g.placeFood() {
		g.over = true
		g.won = true
	}

	return nil
}

func (g *Game) Snapshot(now time.Time) Snapshot {
	snake := make([]Point, len(g.snake))
	copy(snake, g.snake)
	obstacles := make([]Point, len(g.obstacles))
	copy(obstacles, g.obstacles)

	return Snapshot{
		Width:            g.width,
		Height:           g.height,
		Snake:            snake,
		Obstacles:        obstacles,
		Food:             g.food,
		Direction:        g.dir,
		Score:            g.score,
		FoodEaten:        g.foodEaten,
		Level:            g.level,
		FoodsToNextLevel: g.foodsToNextLevel(),
		Started:          g.started,
		IsOver:           g.over,
		IsWon:            g.won,
		Paused:           g.paused,
		Elapsed:          g.elapsedAt(now),
		BestScore:        g.bestScore,
		BestLength:       g.bestLength,
		BestDuration:     g.bestDuration,
		RunsPlayed:       g.runsPlayed,
		TotalFoodEaten:   g.totalFood,
		TotalPlayTime:    g.totalPlayTime,
		LastRun:          g.lastRun,
		HasLastRun:       g.hasLastRun,
	}
}

func (g *Game) TickInterval(base, min, levelStep time.Duration) time.Duration {
	if base <= 0 {
		return time.Millisecond
	}
	if levelStep < 0 {
		levelStep = 0
	}
	if min <= 0 {
		min = time.Millisecond
	}

	lvl := g.level - 1
	if lvl < 0 {
		lvl = 0
	}
	interval := base - time.Duration(lvl)*levelStep
	if interval < min {
		return min
	}
	return interval
}

func (g *Game) Profile() Profile {
	return Profile{
		BestScore:           g.bestScore,
		BestLength:          g.bestLength,
		BestDurationMillis:  g.bestDuration.Milliseconds(),
		RunsPlayed:          g.runsPlayed,
		TotalFoodEaten:      g.totalFood,
		TotalPlayTimeMillis: g.totalPlayTime.Milliseconds(),
	}
}

func (g *Game) ApplyProfile(p Profile) {
	if p.BestScore > 0 {
		g.bestScore = p.BestScore
	}
	if p.BestLength > 0 {
		g.bestLength = p.BestLength
	}
	if p.BestDurationMillis > 0 {
		g.bestDuration = time.Duration(p.BestDurationMillis) * time.Millisecond
	}
	if p.RunsPlayed > 0 {
		g.runsPlayed = p.RunsPlayed
	}
	if p.TotalFoodEaten > 0 {
		g.totalFood = p.TotalFoodEaten
	}
	if p.TotalPlayTimeMillis > 0 {
		g.totalPlayTime = time.Duration(p.TotalPlayTimeMillis) * time.Millisecond
	}
}

func (g *Game) foodsToNextLevel() int {
	if g.foodsPerLevel <= 0 {
		return 0
	}
	rem := g.foodsPerLevel - (g.foodEaten % g.foodsPerLevel)
	if rem <= 0 {
		return g.foodsPerLevel
	}
	return rem
}

func (g *Game) elapsedAt(now time.Time) time.Duration {
	if !g.started || g.startedAt.IsZero() {
		return 0
	}
	if g.over && !g.endedAt.IsZero() {
		d := g.endedAt.Sub(g.startedAt)
		if d < 0 {
			return 0
		}
		return d
	}
	d := now.Sub(g.startedAt)
	if d < 0 {
		return 0
	}
	return d
}

func (g *Game) placeFood() bool {
	total := g.width * g.height
	if len(g.snake)+len(g.obstacles) >= total {
		return false
	}

	occupied := make([]bool, total)
	for _, p := range g.snake {
		occupied[p.Y*g.width+p.X] = true
	}
	for _, p := range g.obstacles {
		occupied[p.Y*g.width+p.X] = true
	}

	available := make([]Point, 0, total-len(g.snake)-len(g.obstacles))
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			idx := y*g.width + x
			if !occupied[idx] {
				available = append(available, Point{X: x, Y: y})
			}
		}
	}

	g.food = available[g.rng.Intn(len(available))]
	return true
}

func (g *Game) regenerateObstacles() {
	target := (g.level - 1) * g.obstaclesStep
	if target < 0 {
		target = 0
	}
	total := g.width * g.height
	maxObstacles := total - len(g.snake) - 1 // keep at least one cell for food
	if maxObstacles < 0 {
		maxObstacles = 0
	}
	if target > maxObstacles {
		target = maxObstacles
	}

	candidates := make([]Point, 0, total-len(g.snake))
	occupied := make([]bool, total)
	for _, p := range g.snake {
		occupied[p.Y*g.width+p.X] = true
	}

	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			idx := y*g.width + x
			if !occupied[idx] {
				candidates = append(candidates, Point{X: x, Y: y})
			}
		}
	}

	g.obstacles = g.obstacles[:0]
	for i := 0; i < target && len(candidates) > 0; i++ {
		j := g.rng.Intn(len(candidates))
		g.obstacles = append(g.obstacles, candidates[j])
		candidates[j] = candidates[len(candidates)-1]
		candidates = candidates[:len(candidates)-1]
	}
}

func (g *Game) finalizeRun(now time.Time) {
	if g.runFinalized || !g.started {
		return
	}
	if g.endedAt.IsZero() {
		g.endedAt = now
	}
	duration := g.endedAt.Sub(g.startedAt)
	if duration < 0 {
		duration = 0
	}

	prevBestScore := g.bestScore
	prevBestLength := g.bestLength
	prevBestDuration := g.bestDuration

	if !g.developerMode {
		g.runsPlayed++
		g.totalFood += g.foodEaten
		g.totalPlayTime += duration

		if g.score > g.bestScore {
			g.bestScore = g.score
		}
		if len(g.snake) > g.bestLength {
			g.bestLength = len(g.snake)
		}
		if duration > g.bestDuration {
			g.bestDuration = duration
		}
	}

	g.lastRun = RunSummary{
		Score:                   g.score,
		Length:                  len(g.snake),
		FoodEaten:               g.foodEaten,
		Level:                   g.level,
		Duration:                duration,
		Won:                     g.won,
		ScoreDeltaVsPrevBest:    g.score - prevBestScore,
		LengthDeltaVsPrevBest:   len(g.snake) - prevBestLength,
		DurationDeltaVsPrevBest: duration - prevBestDuration,
		NewBestScore:            !g.developerMode && g.score > prevBestScore,
		NewBestLength:           !g.developerMode && len(g.snake) > prevBestLength,
		NewBestDuration:         !g.developerMode && duration > prevBestDuration,
	}
	g.hasLastRun = true
	g.runFinalized = true
}

func (g *Game) baseLevel() int {
	if g.foodsPerLevel <= 0 {
		return 1
	}
	return 1 + g.foodEaten/g.foodsPerLevel
}

func (g *Game) effectiveLevel() int {
	level := g.baseLevel() + g.levelOffset
	if level < 1 {
		return 1
	}
	return level
}

func isCardinal(dir Direction) bool {
	x := dirX(dir)
	y := dirY(dir)
	return (x == 0 && (y == -1 || y == 1)) || (y == 0 && (x == -1 || x == 1))
}

func isOpposite(a, b Direction) bool {
	return dirX(a) == -dirX(b) && dirY(a) == -dirY(b)
}

func dirX(dir Direction) int {
	return Point(dir).X
}

func dirY(dir Direction) int {
	return Point(dir).Y
}

func contains(parts []Point, p Point) bool {
	for _, s := range parts {
		if s == p {
			return true
		}
	}
	return false
}
