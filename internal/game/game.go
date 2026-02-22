package game

import (
	"math/rand"
	"time"
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
	Width  int
	Height int
}

type RNG interface {
	Intn(n int) int
}

type State struct {
	width   int
	height  int
	snake   []Point
	dir     Direction
	food    Point
	score   int
	started bool
	over    bool
	won     bool
	rng     RNG
}

func New(cfg Config, rng RNG) *State {
	if cfg.Width <= 0 {
		cfg.Width = 20
	}
	if cfg.Height <= 0 {
		cfg.Height = 15
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	s := &State{
		width:  cfg.Width,
		height: cfg.Height,
		rng:    rng,
	}
	s.Reset()
	return s
}

func (s *State) Reset() {
	s.snake = []Point{{X: s.width / 2, Y: s.height / 2}}
	s.dir = DirNone
	s.score = 0
	s.started = false
	s.over = false
	s.won = false
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

	head := s.snake[0]
	next := Point{X: head.X + dirX(s.dir), Y: head.Y + dirY(s.dir)}

	if next.X < 0 || next.Y < 0 || next.X >= s.width || next.Y >= s.height {
		s.over = true
		return
	}
	if contains(s.snake, next) {
		s.over = true
		return
	}

	s.snake = append([]Point{next}, s.snake...)
	if next == s.food {
		s.score++
		if !s.placeFood() {
			s.over = true
			s.won = true
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

func (s *State) Food() Point {
	return s.food
}

func (s *State) Score() int {
	return s.score
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
