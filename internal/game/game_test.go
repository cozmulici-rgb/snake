package game

import "testing"

type sequenceRNG struct {
	values []int
	idx    int
}

func (r *sequenceRNG) Intn(n int) int {
	if len(r.values) == 0 {
		return 0
	}
	v := r.values[r.idx%len(r.values)]
	r.idx++
	if v < 0 {
		v = -v
	}
	return v % n
}

func TestTickRequiresStartDirection(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})

	before := g.Snake()[0]
	g.Tick()
	after := g.Snake()[0]

	if before != after {
		t.Fatalf("snake moved before start: before=%v after=%v", before, after)
	}
	if g.Started() {
		t.Fatalf("game should not be started before first direction")
	}
}

func TestSetDirectionStartsGameAndMoves(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})

	headBefore := g.Snake()[0]
	changed := g.SetDirection(DirRight)
	if !changed {
		t.Fatalf("expected direction change to start the game")
	}

	g.Tick()
	headAfter := g.Snake()[0]
	want := Point{X: headBefore.X + 1, Y: headBefore.Y}
	if headAfter != want {
		t.Fatalf("unexpected head position after tick: got=%v want=%v", headAfter, want)
	}
}

func TestOppositeDirectionIsRejected(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})

	g.SetDirection(DirRight)
	if g.SetDirection(DirLeft) {
		t.Fatalf("reverse direction should be rejected")
	}
	if g.Direction() != DirRight {
		t.Fatalf("direction changed unexpectedly: got=%v want=%v", g.Direction(), DirRight)
	}
}

func TestWallCollisionEndsGame(t *testing.T) {
	g := New(Config{Width: 4, Height: 4}, &sequenceRNG{})
	g.snake = []Point{{X: 0, Y: 0}}
	g.dir = DirLeft
	g.started = true
	g.over = false
	g.food = Point{X: 3, Y: 3}

	g.Tick()

	if !g.IsOver() {
		t.Fatalf("expected wall collision to end game")
	}
}

func TestSelfCollisionEndsGame(t *testing.T) {
	g := New(Config{Width: 5, Height: 5}, &sequenceRNG{})
	g.snake = []Point{
		{X: 2, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 2},
		{X: 2, Y: 2},
	}
	g.dir = DirDown
	g.started = true
	g.over = false
	g.food = Point{X: 0, Y: 0}

	g.Tick()

	if !g.IsOver() {
		t.Fatalf("expected self collision to end game")
	}
}

func TestEatingFoodIncreasesScoreAndLength(t *testing.T) {
	g := New(Config{Width: 5, Height: 5}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 1, Y: 1}}
	g.dir = DirRight
	g.started = true
	g.over = false
	g.score = 0
	g.food = Point{X: 2, Y: 1}

	g.Tick()

	if g.Score() != 1 {
		t.Fatalf("unexpected score after eating food: got=%d want=1", g.Score())
	}
	if gotLen := len(g.Snake()); gotLen != 2 {
		t.Fatalf("unexpected snake length after eating food: got=%d want=2", gotLen)
	}
	if contains(g.Snake(), g.Food()) {
		t.Fatalf("food spawned on snake body after eating")
	}
}

func TestPlaceFoodAvoidsSnake(t *testing.T) {
	g := New(Config{Width: 2, Height: 2}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
	}

	if !g.placeFood() {
		t.Fatalf("expected food placement to succeed")
	}
	if g.Food() != (Point{X: 1, Y: 1}) {
		t.Fatalf("unexpected food position: got=%v want={1 1}", g.Food())
	}
}

func TestResetRestoresInitialState(t *testing.T) {
	g := New(Config{Width: 8, Height: 6}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 1, Y: 1}, {X: 1, Y: 2}}
	g.dir = DirRight
	g.started = true
	g.over = true
	g.won = true
	g.score = 12

	g.Reset()

	if g.IsOver() {
		t.Fatalf("expected reset game to be active")
	}
	if g.IsWon() {
		t.Fatalf("expected reset game not to be won")
	}
	if g.Started() {
		t.Fatalf("expected reset game not started")
	}
	if g.Score() != 0 {
		t.Fatalf("unexpected reset score: got=%d want=0", g.Score())
	}
	if g.Direction() != DirNone {
		t.Fatalf("unexpected reset direction: got=%v want=%v", g.Direction(), DirNone)
	}

	snake := g.Snake()
	if len(snake) != 1 {
		t.Fatalf("unexpected reset snake length: got=%d want=1", len(snake))
	}
	wantHead := Point{X: 4, Y: 3}
	if snake[0] != wantHead {
		t.Fatalf("unexpected reset head position: got=%v want=%v", snake[0], wantHead)
	}
	if contains(snake, g.Food()) {
		t.Fatalf("food spawned on snake after reset")
	}
}
