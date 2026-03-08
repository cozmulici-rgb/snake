package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

type fakeRandom struct{}

func (fakeRandom) Intn(n int) int { return 0 }

type fakeRepo struct {
	loaded Profile
	saved  []Profile
	err    error
}

func (r *fakeRepo) Load(context.Context) (Profile, error) {
	return r.loaded, r.err
}

func (r *fakeRepo) Save(_ context.Context, p Profile) error {
	r.saved = append(r.saved, p)
	return r.err
}

func TestStartLoadsProfileAndInitializesSnapshot(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{loaded: Profile{BestScore: 11, RunsPlayed: 3}}
	svc := NewService(clk, fakeRandom{}, repo)

	err := svc.Start(context.Background(), PresetConfig{Width: 8, Height: 6})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	snap := svc.Snapshot()
	if snap.Width != 8 || snap.Height != 6 {
		t.Fatalf("unexpected snapshot dimensions: %+v", snap)
	}
	if snap.BestScore != 11 {
		t.Fatalf("expected best score from loaded profile: got=%d want=11", snap.BestScore)
	}
	if snap.RunsPlayed != 3 {
		t.Fatalf("expected runs from loaded profile: got=%d want=3", snap.RunsPlayed)
	}
}

func TestStartContinuesWhenProfileLoadFails(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{err: errors.New("corrupt profile")}
	svc := NewService(clk, fakeRandom{}, repo)
	svc.profile = Profile{BestScore: 9, RunsPlayed: 2}

	if err := svc.Start(context.Background(), PresetConfig{Width: 8, Height: 6}); err != nil {
		t.Fatalf("start should not fail on profile load errors: %v", err)
	}

	snap := svc.Snapshot()
	if snap.Width != 8 || snap.Height != 6 {
		t.Fatalf("expected playable session despite load error: %+v", snap)
	}
	if snap.BestScore != 9 || snap.RunsPlayed != 2 {
		t.Fatalf("expected service to retain last-known profile on load error: %+v", snap)
	}
}

func TestPauseStopsTickMovement(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	svc := NewService(clk, fakeRandom{}, &fakeRepo{})
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !svc.ApplyDirection(DirectionRight) {
		t.Fatalf("failed to apply initial direction")
	}
	before := svc.Snapshot().Snake[0]
	svc.TogglePause()
	clk.now = clk.now.Add(200 * time.Millisecond)
	svc.Tick()
	after := svc.Snapshot().Snake[0]
	if before != after {
		t.Fatalf("snake moved while paused: before=%v after=%v", before, after)
	}
}

func TestQuitFinalizesAndSavesProfile(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{}
	svc := NewService(clk, fakeRandom{}, repo)
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	svc.ApplyDirection(DirectionRight)
	clk.now = clk.now.Add(1500 * time.Millisecond)
	if err := svc.Quit(); err != nil {
		t.Fatalf("quit failed: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected one profile save on quit, got=%d", len(repo.saved))
	}
	snap := svc.Snapshot()
	if snap.RunsPlayed != 1 {
		t.Fatalf("expected run finalization on quit: got runs=%d", snap.RunsPlayed)
	}
	if !snap.HasLastRun {
		t.Fatalf("expected last run summary after quit")
	}
}

func TestTickFinalizationSavesProfileOnGameOver(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{}
	svc := NewService(clk, fakeRandom{}, repo)
	if err := svc.Start(context.Background(), PresetConfig{Width: 2, Height: 1}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !svc.ApplyDirection(DirectionRight) {
		t.Fatalf("failed to start movement")
	}
	clk.now = clk.now.Add(200 * time.Millisecond)
	svc.Tick()

	snap := svc.Snapshot()
	if !snap.IsOver {
		t.Fatalf("expected game over after wall collision")
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected profile save after game over, got=%d", len(repo.saved))
	}
}

func TestProfileGetterReturnsLoadedProfile(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{loaded: Profile{BestScore: 8}}
	svc := NewService(clk, fakeRandom{}, repo)
	if err := svc.LoadProfile(context.Background()); err != nil {
		t.Fatalf("load profile failed: %v", err)
	}
	if svc.Profile().BestScore != 8 {
		t.Fatalf("unexpected profile after load: %+v", svc.Profile())
	}
}

func TestLoadProfileWithoutRepository(t *testing.T) {
	svc := NewService(&fakeClock{now: time.Unix(1, 0)}, fakeRandom{}, nil)
	if err := svc.LoadProfile(context.Background()); err != nil {
		t.Fatalf("expected nil error when repo is not configured: %v", err)
	}
}

func TestApplyDirectionRejectsNoneAndNoSession(t *testing.T) {
	svc := NewService(&fakeClock{now: time.Unix(1, 0)}, fakeRandom{}, &fakeRepo{})
	if svc.ApplyDirection(DirectionRight) {
		t.Fatalf("expected direction apply to fail before start")
	}
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6}); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if svc.ApplyDirection(DirectionNone) {
		t.Fatalf("expected none direction to be rejected")
	}
}

func TestRestartRequiresSession(t *testing.T) {
	svc := NewService(&fakeClock{now: time.Unix(1, 0)}, fakeRandom{}, &fakeRepo{})
	if err := svc.Restart(); err == nil {
		t.Fatalf("expected error when restarting without a session")
	}
}

func TestLastRunSummaryAfterQuit(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	svc := NewService(clk, fakeRandom{}, &fakeRepo{})
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6}); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	svc.ApplyDirection(DirectionRight)
	clk.now = clk.now.Add(time.Second)
	if err := svc.Quit(); err != nil {
		t.Fatalf("quit failed: %v", err)
	}
	summary, ok := svc.LastRunSummary()
	if !ok {
		t.Fatalf("expected run summary")
	}
	if summary.ScoreDeltaVsPrevBest < 0 {
		t.Fatalf("expected non-negative initial score delta: %+v", summary)
	}
}

func TestDeveloperModeBypassSetsSnapshotLevel(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	svc := NewService(clk, fakeRandom{}, &fakeRepo{})
	svc.SetDeveloperMode(true)
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6, FoodsPerLevel: 2, ObstaclesStep: 1}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if err := svc.BypassLevel(4); err != nil {
		t.Fatalf("bypass level failed: %v", err)
	}

	snap := svc.Snapshot()
	if snap.Level != 4 {
		t.Fatalf("unexpected snapshot level after bypass: got=%d want=4", snap.Level)
	}
	if len(snap.Obstacles) != 3 {
		t.Fatalf("unexpected obstacle count after bypass: got=%d want=3", len(snap.Obstacles))
	}
}

func TestDeveloperModeSkipsProfileSaveOnQuit(t *testing.T) {
	clk := &fakeClock{now: time.Unix(100, 0)}
	repo := &fakeRepo{}
	svc := NewService(clk, fakeRandom{}, repo)
	svc.SetDeveloperMode(true)
	if err := svc.Start(context.Background(), PresetConfig{Width: 6, Height: 6}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	svc.ApplyDirection(DirectionRight)
	clk.now = clk.now.Add(time.Second)
	if err := svc.Quit(); err != nil {
		t.Fatalf("quit failed: %v", err)
	}

	if len(repo.saved) != 0 {
		t.Fatalf("developer mode should not persist profile saves, got=%d", len(repo.saved))
	}
	if svc.Snapshot().RunsPlayed != 0 {
		t.Fatalf("developer mode should not count runs")
	}
}

func TestHelpers(t *testing.T) {
	cfg := withDefaults(PresetConfig{})
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.BaseTick <= 0 || cfg.MinTick <= 0 {
		t.Fatalf("defaults were not applied: %+v", cfg)
	}
	interval := calcTickInterval(PresetConfig{BaseTick: 100 * time.Millisecond, MinTick: 50 * time.Millisecond, LevelStep: 10 * time.Millisecond}, 10)
	if interval != 50*time.Millisecond {
		t.Fatalf("unexpected capped interval: got=%v", interval)
	}
	if toDomainDirection(DirectionLeft).X != -1 {
		t.Fatalf("expected left mapping")
	}
	if fromDomainDirection(toDomainDirection(DirectionUp)) != DirectionUp {
		t.Fatalf("expected round-trip direction conversion")
	}
}

func TestNilDependenciesFallbacks(t *testing.T) {
	svc := NewService(nil, nil, nil)
	if svc.clock.Now().IsZero() {
		t.Fatalf("expected real clock fallback")
	}
	if svc.random.Intn(10) < 0 {
		t.Fatalf("random fallback returned invalid value")
	}
}

func TestTogglePauseWithoutSession(t *testing.T) {
	svc := NewService(&fakeClock{now: time.Unix(1, 0)}, fakeRandom{}, &fakeRepo{})
	if svc.TogglePause() {
		t.Fatalf("expected false pause state without active session")
	}
}
