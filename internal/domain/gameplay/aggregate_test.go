package gameplay

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

	before := g.Snapshot(time.Now()).Snake[0]
	g.Tick(time.Now())
	after := g.Snapshot(time.Now()).Snake[0]

	if before != after {
		t.Fatalf("snake moved before start: before=%v after=%v", before, after)
	}
	if g.Snapshot(time.Now()).Started {
		t.Fatalf("game should not be started before first direction")
	}
}

func TestStartDirectionStartsGameAndMoves(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()

	headBefore := g.Snapshot(now).Snake[0]
	changed := g.Start(DirRight, now)
	if !changed {
		t.Fatalf("expected start direction to start the game")
	}

	g.Tick(now.Add(100 * time.Millisecond))
	headAfter := g.Snapshot(now).Snake[0]
	want := Point{X: headBefore.X + 1, Y: headBefore.Y}
	if headAfter != want {
		t.Fatalf("unexpected head position after tick: got=%v want=%v", headAfter, want)
	}
	if g.Snapshot(now).Level != 1 {
		t.Fatalf("unexpected initial level after movement: got=%d want=1", g.Snapshot(now).Level)
	}
}

func TestOppositeDirectionIsRejected(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()

	g.Start(DirRight, now)
	if g.SetDirection(DirLeft) {
		t.Fatalf("reverse direction should be rejected")
	}
	if g.Snapshot(now).Direction != DirRight {
		t.Fatalf("direction changed unexpectedly: got=%v want=%v", g.Snapshot(now).Direction, DirRight)
	}
}

func TestInvalidDirectionIsRejected(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()

	if g.Start(DirNone, now) {
		t.Fatalf("none direction should be rejected")
	}
	if g.Start(Direction{X: 1, Y: 1}, now) {
		t.Fatalf("diagonal direction should be rejected")
	}
	if g.Snapshot(now).Started {
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

	g.Tick(time.Now())

	if !g.Snapshot(time.Now()).IsOver {
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

	g.Tick(time.Now())

	if !g.Snapshot(time.Now()).IsOver {
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

	g.Tick(time.Now())

	if !g.Snapshot(time.Now()).IsOver {
		t.Fatalf("expected obstacle collision to end game")
	}
}

func TestPauseStopsTickMovement(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()
	g.Start(DirRight, now)
	g.TogglePause()
	before := g.Snapshot(now).Snake[0]
	g.Tick(now.Add(100 * time.Millisecond))
	after := g.Snapshot(now).Snake[0]
	if before != after {
		t.Fatalf("snake moved while paused: before=%v after=%v", before, after)
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

	g.Tick(time.Now())
	snap := g.Snapshot(time.Now())
	if snap.Score != 1 {
		t.Fatalf("unexpected score after eating food: got=%d want=1", snap.Score)
	}
	if gotLen := len(snap.Snake); gotLen != 2 {
		t.Fatalf("unexpected snake length after eating food: got=%d want=2", gotLen)
	}
	if contains(snap.Snake, snap.Food) {
		t.Fatalf("food spawned on snake body after eating")
	}
	if snap.FoodEaten != 1 {
		t.Fatalf("unexpected food eaten count: got=%d want=1", snap.FoodEaten)
	}
	if snap.FoodsToNextLevel != DefaultFoodsPerLevel-1 {
		t.Fatalf("unexpected foods-to-next-level: got=%d want=%d", snap.FoodsToNextLevel, DefaultFoodsPerLevel-1)
	}
}

func TestPlaceFoodAvoidsSnakeAndObstacles(t *testing.T) {
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
	if g.Snapshot(time.Now()).Food != (Point{X: 2, Y: 1}) {
		t.Fatalf("unexpected food position with obstacles: got=%v want={2 1}", g.Snapshot(time.Now()).Food)
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

	g.Tick(time.Now())
	snap := g.Snapshot(time.Now())
	if !snap.IsOver {
		t.Fatalf("expected game over after filling board")
	}
	if !snap.IsWon {
		t.Fatalf("expected win state after filling board")
	}
	if snap.Score != 1 {
		t.Fatalf("unexpected score after winning move: got=%d want=1", snap.Score)
	}
	if len(snap.Snake) != 2 {
		t.Fatalf("unexpected snake length after winning move: got=%d want=2", len(snap.Snake))
	}
	if snap.RunsPlayed != 1 {
		t.Fatalf("unexpected run count after win: got=%d want=1", snap.RunsPlayed)
	}
}

func TestLevelIncreasesAndObstaclesRegenerate(t *testing.T) {
	g := New(Config{Width: 5, Height: 5, FoodsPerLevel: 5, ObstaclesStep: 1}, &sequenceRNG{values: []int{0}})
	g.snake = []Point{{X: 1, Y: 1}}
	g.dir = DirRight
	g.started = true
	g.over = false
	g.foodEaten = 4
	g.level = 1
	g.food = Point{X: 2, Y: 1}

	g.Tick(time.Now())
	snap := g.Snapshot(time.Now())
	if snap.Level != 2 {
		t.Fatalf("expected level up after 5th food: got=%d want=2", snap.Level)
	}
	if snap.FoodsToNextLevel != 5 {
		t.Fatalf("unexpected foods-to-next-level after level up: got=%d want=5", snap.FoodsToNextLevel)
	}
	if len(snap.Obstacles) != 1 {
		t.Fatalf("unexpected obstacle count after level up: got=%d want=1", len(snap.Obstacles))
	}
	if contains(snap.Obstacles, snap.Snake[0]) {
		t.Fatalf("obstacle spawned on snake")
	}
}

func TestRunFinalizationUpdatesBestStats(t *testing.T) {
	g := New(Config{Width: 4, Height: 4}, &sequenceRNG{})
	now := time.Now()
	g.snake = []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}
	g.dir = DirLeft
	g.started = true
	g.startedAt = now.Add(-2 * time.Second)
	g.score = 2
	g.foodEaten = 2
	g.level = 1
	g.over = false
	g.food = Point{X: 3, Y: 3}

	g.Tick(now)
	snap := g.Snapshot(now)
	if !snap.IsOver {
		t.Fatalf("expected run to end")
	}
	if snap.RunsPlayed != 1 {
		t.Fatalf("unexpected runs played: got=%d want=1", snap.RunsPlayed)
	}
	if snap.BestScore != 2 {
		t.Fatalf("unexpected best score: got=%d want=2", snap.BestScore)
	}
	if snap.BestLength != 3 {
		t.Fatalf("unexpected best length: got=%d want=3", snap.BestLength)
	}
	if snap.BestDuration <= 0 {
		t.Fatalf("expected positive best duration")
	}
	if snap.TotalFoodEaten != 2 {
		t.Fatalf("unexpected total food eaten: got=%d want=2", snap.TotalFoodEaten)
	}
	if snap.TotalPlayTime <= 0 {
		t.Fatalf("expected positive total play time")
	}
	if !snap.HasLastRun {
		t.Fatalf("expected last run summary")
	}
	if snap.LastRun.Score != 2 || snap.LastRun.Length != 3 || snap.LastRun.FoodEaten != 2 {
		t.Fatalf("unexpected run summary payload: %+v", snap.LastRun)
	}
}

func TestResetRestoresInitialState(t *testing.T) {
	g := New(Config{Width: 8, Height: 6}, &sequenceRNG{values: []int{0}})
	now := time.Now()
	g.snake = []Point{{X: 1, Y: 1}, {X: 1, Y: 2}}
	g.dir = DirRight
	g.started = true
	g.over = true
	g.won = true
	g.score = 12

	g.Reset(now)
	snap := g.Snapshot(now)
	if snap.IsOver {
		t.Fatalf("expected reset game to be active")
	}
	if snap.IsWon {
		t.Fatalf("expected reset game not to be won")
	}
	if snap.Started {
		t.Fatalf("expected reset game not started")
	}
	if snap.Score != 0 {
		t.Fatalf("unexpected reset score: got=%d want=0", snap.Score)
	}
	if snap.Direction != DirNone {
		t.Fatalf("unexpected reset direction: got=%v want=%v", snap.Direction, DirNone)
	}
	if len(snap.Snake) != 1 {
		t.Fatalf("unexpected reset snake length: got=%d want=1", len(snap.Snake))
	}
	wantHead := Point{X: 4, Y: 3}
	if snap.Snake[0] != wantHead {
		t.Fatalf("unexpected reset head position: got=%v want=%v", snap.Snake[0], wantHead)
	}
}

func TestFinalizeRecordsRunWithoutGameOver(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()
	g.Start(DirRight, now)
	g.startedAt = now.Add(-time.Second)
	g.score = 3
	g.foodEaten = 3
	g.level = 1

	g.Finalize(now)
	snap := g.Snapshot(now)
	if snap.RunsPlayed != 1 {
		t.Fatalf("expected runs played to increment: got=%d want=1", snap.RunsPlayed)
	}
	if snap.BestScore != 3 {
		t.Fatalf("unexpected best score after finalize: got=%d want=3", snap.BestScore)
	}
}

func TestProfileRoundTrip(t *testing.T) {
	g := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	now := time.Now()
	g.Start(DirRight, now)
	g.startedAt = now.Add(-2 * time.Second)
	g.score = 4
	g.foodEaten = 4
	g.level = 2
	g.Finalize(now)

	p := g.Profile()
	g2 := New(Config{Width: 6, Height: 6}, &sequenceRNG{})
	g2.ApplyProfile(p)

	s1 := g.Snapshot(now)
	s2 := g2.Snapshot(now)
	if s2.BestScore != s1.BestScore {
		t.Fatalf("best score mismatch after profile import: got=%d want=%d", s2.BestScore, s1.BestScore)
	}
	if s2.BestLength != s1.BestLength {
		t.Fatalf("best length mismatch after profile import: got=%d want=%d", s2.BestLength, s1.BestLength)
	}
	if s2.BestDuration.Milliseconds() != s1.BestDuration.Milliseconds() {
		t.Fatalf("best duration mismatch after profile import: got=%v want=%v", s2.BestDuration, s1.BestDuration)
	}
	if s2.RunsPlayed != s1.RunsPlayed {
		t.Fatalf("runs mismatch after profile import: got=%d want=%d", s2.RunsPlayed, s1.RunsPlayed)
	}
	if s2.TotalFoodEaten != s1.TotalFoodEaten {
		t.Fatalf("total food mismatch after profile import: got=%d want=%d", s2.TotalFoodEaten, s1.TotalFoodEaten)
	}
	if s2.TotalPlayTime.Milliseconds() != s1.TotalPlayTime.Milliseconds() {
		t.Fatalf("total play time mismatch after profile import: got=%v want=%v", s2.TotalPlayTime, s1.TotalPlayTime)
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
