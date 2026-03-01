package session

import "time"

type DirectionInput int

const (
	DirectionNone DirectionInput = iota
	DirectionUp
	DirectionDown
	DirectionLeft
	DirectionRight
)

type Point struct {
	X int
	Y int
}

type PresetConfig struct {
	Name          string
	Width         int
	Height        int
	FoodsPerLevel int
	ObstaclesStep int
	BaseTick      time.Duration
	MinTick       time.Duration
	LevelStep     time.Duration
}

type Profile struct {
	BestScore           int   `json:"best_score"`
	BestLength          int   `json:"best_length"`
	BestDurationMillis  int64 `json:"best_duration_millis"`
	RunsPlayed          int   `json:"runs_played"`
	TotalFoodEaten      int   `json:"total_food_eaten"`
	TotalPlayTimeMillis int64 `json:"total_play_time_millis"`
}

type RunSummaryView struct {
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

type SessionSnapshot struct {
	Width            int
	Height           int
	Snake            []Point
	Obstacles        []Point
	Food             Point
	Direction        DirectionInput
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
	LastRun          RunSummaryView
	HasLastRun       bool
	TickInterval     time.Duration
}
