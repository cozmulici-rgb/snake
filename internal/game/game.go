package game

import (
	"math/rand"
	"time"
)

const (
	DefaultFoodsPerLevel = 5
)

type Point struct {
	X int
	Y int
}

type Direction Point

var (
	DirNone  = Direction{X: 0, Y: 0}
	DirUp    = Direction{X: 0, Y: -1}
	DirDown  = Direction{X: 0, Y: 1}
	DirLeft  = Direction{X: -1, Y: 0}
	DirRight = Direction{X: 1, Y: 0}
)

type Config struct {
	Width         int
	Height        int
	FoodsPerLevel int
}

type RNG interface {
	Intn(n int) int
}

type State struct {
	width         int
	height        int
	foodsPerLevel int
	snake         []Point
	dir           Direction
	food          Point
	score         int
	foodEaten     int
	level         int
	started       bool
	over          bool
	won           bool
	startedAt     time.Time
	endedAt       time.Time
	runFinalized  bool
	bestScore     int
	bestLength    int
	bestDuration  time.Duration
	runsPlayed    int
	totalFood     int
	totalPlayTime time.Duration
	rng           RNG
}

func New(cfg Config, rng RNG) *State {
	if cfg.Width <= 0 {
		cfg.Width = 20
	}
	if cfg.Height <= 0 {
		cfg.Height = 15
	}
	if cfg.FoodsPerLevel <= 0 {
		cfg.FoodsPerLevel = DefaultFoodsPerLevel
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	s := &State{
		width:         cfg.Width,
		height:        cfg.Height,
		foodsPerLevel: cfg.FoodsPerLevel,
		rng:           rng,
	}
	s.Reset()
	return s
}

func (s *State) Reset() {
	s.finalizeRun(time.Now())

	s.snake = []Point{{X: s.width / 2, Y: s.height / 2}}
	s.dir = DirNone
	s.score = 0
	s.foodEaten = 0
	s.level = 1
	s.started = false
	s.over = false
	s.won = false
	s.startedAt = time.Time{}
	s.endedAt = time.Time{}
	s.runFinalized = false
	if !s.placeFood() {
		s.over = true
		s.won = true
	}
}

func (s *State) SetDirection(dir Direction) bool {
	if !isCardinal(dir) || s.over {
		return false
	}
	if !s.started {
		s.dir = dir
		s.started = true
		s.startedAt = time.Now()
		return true
	}
	if isOpposite(s.dir, dir) {
		return false
	}
	s.dir = dir
	return true
}

func (s *State) Tick() {
	if s.over || !s.started {
		return
	}

	now := time.Now()
	head := s.snake[0]
	next := Point{X: head.X + dirX(s.dir), Y: head.Y + dirY(s.dir)}

	if next.X < 0 || next.Y < 0 || next.X >= s.width || next.Y >= s.height {
		s.over = true
		s.endedAt = now
		s.finalizeRun(now)
		return
	}
	if contains(s.snake, next) {
		s.over = true
		s.endedAt = now
		s.finalizeRun(now)
		return
	}

	s.snake = append([]Point{next}, s.snake...)
	if next == s.food {
		s.score++
		s.foodEaten++
		s.level = 1 + s.foodEaten/s.foodsPerLevel
		if !s.placeFood() {
			s.over = true
			s.won = true
			s.endedAt = now
			s.finalizeRun(now)
		}
		return
	}

	s.snake = s.snake[:len(s.snake)-1]
}

func (s *State) Width() int {
	return s.width
}

func (s *State) Height() int {
	return s.height
}

func (s *State) Snake() []Point {
	out := make([]Point, len(s.snake))
	copy(out, s.snake)
	return out
}

func (s *State) SnakeLength() int {
	return len(s.snake)
}

func (s *State) Food() Point {
	return s.food
}

func (s *State) Score() int {
	return s.score
}

func (s *State) FoodEaten() int {
	return s.foodEaten
}

func (s *State) Level() int {
	return s.level
}

func (s *State) FoodsToNextLevel() int {
	if s.foodsPerLevel <= 0 {
		return 0
	}
	rem := s.foodsPerLevel - (s.foodEaten % s.foodsPerLevel)
	if rem <= 0 {
		return s.foodsPerLevel
	}
	return rem
}

func (s *State) Started() bool {
	return s.started
}

func (s *State) IsOver() bool {
	return s.over
}

func (s *State) IsWon() bool {
	return s.won
}

func (s *State) Direction() Direction {
	return s.dir
}

func (s *State) Elapsed() time.Duration {
	if !s.started || s.startedAt.IsZero() {
		return 0
	}
	if s.over && !s.endedAt.IsZero() {
		d := s.endedAt.Sub(s.startedAt)
		if d < 0 {
			return 0
		}
		return d
	}
	d := time.Since(s.startedAt)
	if d < 0 {
		return 0
	}
	return d
}

func (s *State) BestScore() int {
	return s.bestScore
}

func (s *State) BestLength() int {
	return s.bestLength
}

func (s *State) BestDuration() time.Duration {
	return s.bestDuration
}

func (s *State) RunsPlayed() int {
	return s.runsPlayed
}

func (s *State) TotalFoodEaten() int {
	return s.totalFood
}

func (s *State) TotalPlayTime() time.Duration {
	return s.totalPlayTime
}

func (s *State) TickInterval(base, min, levelStep time.Duration) time.Duration {
	if base <= 0 {
		return time.Millisecond
	}
	if levelStep < 0 {
		levelStep = 0
	}
	if min <= 0 {
		min = time.Millisecond
	}

	lvl := s.level - 1
	if lvl < 0 {
		lvl = 0
	}
	interval := base - time.Duration(lvl)*levelStep
	if interval < min {
		return min
	}
	return interval
}

func (s *State) placeFood() bool {
	total := s.width * s.height
	if len(s.snake) >= total {
		return false
	}

	occupied := make([]bool, total)
	for _, p := range s.snake {
		occupied[p.Y*s.width+p.X] = true
	}

	available := make([]Point, 0, total-len(s.snake))
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			idx := y*s.width + x
			if !occupied[idx] {
				available = append(available, Point{X: x, Y: y})
			}
		}
	}

	s.food = available[s.rng.Intn(len(available))]
	return true
}

func (s *State) finalizeRun(now time.Time) {
	if s.runFinalized || !s.started {
		return
	}
	if s.endedAt.IsZero() {
		s.endedAt = now
	}
	duration := s.endedAt.Sub(s.startedAt)
	if duration < 0 {
		duration = 0
	}

	s.runsPlayed++
	s.totalFood += s.foodEaten
	s.totalPlayTime += duration

	if s.score > s.bestScore {
		s.bestScore = s.score
	}
	if len(s.snake) > s.bestLength {
		s.bestLength = len(s.snake)
	}
	if duration > s.bestDuration {
		s.bestDuration = duration
	}

	s.runFinalized = true
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
