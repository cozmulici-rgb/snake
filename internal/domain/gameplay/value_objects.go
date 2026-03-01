package gameplay

import "time"

const (
	DefaultFoodsPerLevel = 5
	DefaultObstaclesStep = 2
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
	ObstaclesStep int
}

type Profile struct {
	BestScore           int   `json:"best_score"`
	BestLength          int   `json:"best_length"`
	BestDurationMillis  int64 `json:"best_duration_millis"`
	RunsPlayed          int   `json:"runs_played"`
	TotalFoodEaten      int   `json:"total_food_eaten"`
	TotalPlayTimeMillis int64 `json:"total_play_time_millis"`
}

type RunSummary struct {
	Score                   int
	Length                  int
	FoodEaten               int
	Level                   int
	Duration                time.Duration
	Won                     bool
	ScoreDeltaVsPrevBest    int
	LengthDeltaVsPrevBest   int
	DurationDeltaVsPrevBest time.Duration
	NewBestScore            bool
	NewBestLength           bool
	NewBestDuration         bool
}

type Snapshot struct {
	Width            int
	Height           int
	Snake            []Point
	Obstacles        []Point
	Food             Point
	Direction        Direction
	Score            int
	FoodEaten        int
	Level            int
	FoodsToNextLevel int
	Started          bool
	IsOver           bool
	IsWon            bool
	Paused           bool
	Elapsed          time.Duration
	BestScore        int
	BestLength       int
	BestDuration     time.Duration
	RunsPlayed       int
	TotalFoodEaten   int
	TotalPlayTime    time.Duration
	LastRun          RunSummary
	HasLastRun       bool
}

type RNG interface {
	Intn(n int) int
}

type Aggregate interface {
	Start(dir Direction, now time.Time) bool
	SetDirection(dir Direction) bool
	Tick(now time.Time)
	TogglePause() bool
	Reset(now time.Time)
	Finalize(now time.Time)
	Snapshot(now time.Time) Snapshot
}
