package session

import (
	"context"
	"time"
)

type SessionService interface {
	Start(ctx context.Context, cfg PresetConfig) error
	ApplyDirection(dir DirectionInput) bool
	Tick()
	TogglePause() bool
	Restart() error
	Quit() error
	Snapshot() SessionSnapshot
	LastRunSummary() (RunSummaryView, bool)
}

type ProfileRepository interface {
	Load(ctx context.Context) (Profile, error)
	Save(ctx context.Context, profile Profile) error
}

type Clock interface {
	Now() time.Time
}

type Random interface {
	Intn(n int) int
}
