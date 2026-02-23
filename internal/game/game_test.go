package game

import (
	"testing"
	"time"
)

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
	if g.Level() != 1 {
		t.Fatalf("unexpected initial level after movement: got=%d want=1", g.Level())
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

func TestInvalidDirectionIsRejected(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})

	if g.SetDirection(DirNone) {
		t.Fatalf("none direction should be rejected")
	}
	if g.SetDirection(Direction{X: 1, Y: 1}) {
		t.Fatalf("diagonal direction should be rejected")
	}
	if g.Started() {
		t.Fatalf("game should not start on invalid direction")
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

func TestObstacleCollisionEndsGame(t *testing.T) {
	g := New(Config{Width: 5, Height: 5, ObstaclesStep: 1}, &sequenceRNG{})
	g.snake = []Point{{X: 1, Y: 1}}
	g.dir = DirRight
	g.started = true
	g.over = false
	g.food = Point{X: 4, Y: 4}
	g.obstacles = []Point{{X: 2, Y: 1}}

	g.Tick()

	if !g.IsOver() {
		t.Fatalf("expected obstacle collision to end game")
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
	if g.FoodEaten() != 1 {
		t.Fatalf("unexpected food eaten count: got=%d want=1", g.FoodEaten())
	}
	if g.FoodsToNextLevel() != DefaultFoodsPerLevel-1 {
		t.Fatalf("unexpected foods-to-next-level: got=%d want=%d", g.FoodsToNextLevel(), DefaultFoodsPerLevel-1)
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

func TestPlaceFoodAvoidsObstacles(t *testing.T) {
	g := New(Config{Width: 3, Height: 2, ObstaclesStep: 1}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 0, Y: 0}}
	g.obstacles = []Point{
		{X: 1, Y: 0},
		{X: 2, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	if !g.placeFood() {
		t.Fatalf("expected food placement to succeed")
	}
	if g.Food() != (Point{X: 2, Y: 1}) {
		t.Fatalf("unexpected food position with obstacles: got=%v want={2 1}", g.Food())
	}
}

func TestWinningMoveWhenBoardFills(t *testing.T) {
	g := New(Config{Width: 2, Height: 1}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 0, Y: 0}}
	g.dir = DirRight
	g.started = true
	g.over = false
	g.won = false
	g.score = 0
	g.food = Point{X: 1, Y: 0}

	g.Tick()

	if !g.IsOver() {
		t.Fatalf("expected game over after filling board")
	}
	if !g.IsWon() {
		t.Fatalf("expected win state after filling board")
	}
	if g.Score() != 1 {
		t.Fatalf("unexpected score after winning move: got=%d want=1", g.Score())
	}
	if len(g.Snake()) != 2 {
		t.Fatalf("unexpected snake length after winning move: got=%d want=2", len(g.Snake()))
	}
	if g.RunsPlayed() != 1 {
		t.Fatalf("unexpected run count after win: got=%d want=1", g.RunsPlayed())
	}
}

func TestLevelIncreasesEveryFoodsPerLevel(t *testing.T) {
	g := New(Config{Width: 5, Height: 5, FoodsPerLevel: 5, ObstaclesStep: 1}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 1, Y: 1}}
	g.dir = DirRight
	g.started = true
	g.over = false
	g.foodEaten = 4
	g.level = 1
	g.food = Point{X: 2, Y: 1}

	g.Tick()

	if g.Level() != 2 {
		t.Fatalf("expected level up after 5th food: got=%d want=2", g.Level())
	}
	if g.FoodsToNextLevel() != 5 {
		t.Fatalf("unexpected foods-to-next-level after level up: got=%d want=5", g.FoodsToNextLevel())
	}
	if g.ObstacleCount() != 1 {
		t.Fatalf("unexpected obstacle count after level up: got=%d want=1", g.ObstacleCount())
	}
	if contains(g.Obstacles(), g.Snake()[0]) {
		t.Fatalf("obstacle spawned on snake")
	}
}

func TestTickIntervalScalesWithLevelAndCaps(t *testing.T) {
	g := New(Config{Width: 5, Height: 5}, &sequenceRNG{})
	g.level = 1
	got := g.TickInterval(140*time.Millisecond, 70*time.Millisecond, 8*time.Millisecond)
	if got != 140*time.Millisecond {
		t.Fatalf("unexpected base tick interval: got=%v want=%v", got, 140*time.Millisecond)
	}

	g.level = 6
	got = g.TickInterval(140*time.Millisecond, 70*time.Millisecond, 8*time.Millisecond)
	if got != 100*time.Millisecond {
		t.Fatalf("unexpected scaled tick interval: got=%v want=%v", got, 100*time.Millisecond)
	}

	g.level = 20
	got = g.TickInterval(140*time.Millisecond, 70*time.Millisecond, 8*time.Millisecond)
	if got != 70*time.Millisecond {
		t.Fatalf("unexpected capped tick interval: got=%v want=%v", got, 70*time.Millisecond)
	}
}

func TestRunFinalizationUpdatesBestStats(t *testing.T) {
	g := New(Config{Width: 4, Height: 4}, &sequenceRNG{})
	g.snake = []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}
	g.dir = DirLeft
	g.started = true
	g.startedAt = time.Now().Add(-2 * time.Second)
	g.score = 2
	g.foodEaten = 2
	g.level = 1
	g.over = false
	g.food = Point{X: 3, Y: 3}

	g.Tick()

	if !g.IsOver() {
		t.Fatalf("expected run to end")
	}
	if g.RunsPlayed() != 1 {
		t.Fatalf("unexpected runs played: got=%d want=1", g.RunsPlayed())
	}
	if g.BestScore() != 2 {
		t.Fatalf("unexpected best score: got=%d want=2", g.BestScore())
	}
	if g.BestLength() != 3 {
		t.Fatalf("unexpected best length: got=%d want=3", g.BestLength())
	}
	if g.BestDuration() <= 0 {
		t.Fatalf("expected positive best duration")
	}
	if g.TotalFoodEaten() != 2 {
		t.Fatalf("unexpected total food eaten: got=%d want=2", g.TotalFoodEaten())
	}
	if g.TotalPlayTime() <= 0 {
		t.Fatalf("expected positive total play time")
	}
	summary, ok := g.LastRunSummary()
	if !ok {
		t.Fatalf("expected last run summary")
	}
	if summary.Score != 2 || summary.Length != 3 || summary.FoodEaten != 2 {
		t.Fatalf("unexpected run summary payload: %+v", summary)
	}
	if !summary.NewBestScore || !summary.NewBestLength || !summary.NewBestDuration {
		t.Fatalf("expected new best flags on first finalized run: %+v", summary)
	}
	if summary.ScoreDeltaVsPrevBest != 2 {
		t.Fatalf("unexpected score delta: got=%d want=2", summary.ScoreDeltaVsPrevBest)
	}
	if summary.LengthDeltaVsPrevBest != 3 {
		t.Fatalf("unexpected length delta: got=%d want=3", summary.LengthDeltaVsPrevBest)
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

func TestFinalizeNowRecordsRunWithoutGameOver(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	g.SetDirection(DirRight)
	g.startedAt = time.Now().Add(-time.Second)
	g.score = 3
	g.foodEaten = 3
	g.level = 1

	g.FinalizeNow()

	if g.RunsPlayed() != 1 {
		t.Fatalf("expected runs played to increment: got=%d want=1", g.RunsPlayed())
	}
	if g.BestScore() != 3 {
		t.Fatalf("unexpected best score after finalize: got=%d want=3", g.BestScore())
	}
}

func TestProfileExportImportRoundTrip(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	g.SetDirection(DirRight)
	g.startedAt = time.Now().Add(-2 * time.Second)
	g.score = 4
	g.foodEaten = 4
	g.level = 2
	g.FinalizeNow()

	p := g.Profile()

	g2 := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	g2.ApplyProfile(p)

	if g2.BestScore() != g.BestScore() {
		t.Fatalf("best score mismatch after profile import: got=%d want=%d", g2.BestScore(), g.BestScore())
	}
	if g2.BestLength() != g.BestLength() {
		t.Fatalf("best length mismatch after profile import: got=%d want=%d", g2.BestLength(), g.BestLength())
	}
	if g2.BestDuration() != g.BestDuration() {
		t.Fatalf("best duration mismatch after profile import: got=%v want=%v", g2.BestDuration(), g.BestDuration())
	}
	if g2.RunsPlayed() != g.RunsPlayed() {
		t.Fatalf("runs mismatch after profile import: got=%d want=%d", g2.RunsPlayed(), g.RunsPlayed())
	}
	if g2.TotalFoodEaten() != g.TotalFoodEaten() {
		t.Fatalf("total food mismatch after profile import: got=%d want=%d", g2.TotalFoodEaten(), g.TotalFoodEaten())
	}
	if g2.TotalPlayTime() != g.TotalPlayTime() {
		t.Fatalf("total play time mismatch after profile import: got=%v want=%v", g2.TotalPlayTime(), g.TotalPlayTime())
	}
}
