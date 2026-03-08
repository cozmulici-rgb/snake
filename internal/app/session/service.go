package session

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"snake/internal/domain/gameplay"
)

type Service struct {
	clock         Clock
	random        Random
	repo          ProfileRepository
	game          *gameplay.Game
	preset        PresetConfig
	profile       Profile
	savedRuns     int
	developerMode bool
}

var _ SessionService = (*Service)(nil)

func NewService(clock Clock, random Random, repo ProfileRepository) *Service {
	if clock == nil {
		clock = realClock{}
	}
	if random == nil {
		random = &mathRandom{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
	}
	return &Service{
		clock:  clock,
		random: random,
		repo:   repo,
	}
}

func (s *Service) LoadProfile(ctx context.Context) error {
	if s.repo == nil {
		s.profile = Profile{}
		return nil
	}
	p, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}
	s.profile = p
	return nil
}

func (s *Service) Profile() Profile {
	return s.profile
}

func (s *Service) Start(ctx context.Context, cfg PresetConfig) error {
	cfg = withDefaults(cfg)
	if s.repo != nil {
		// Profile IO issues should not block starting a playable session.
		_ = s.LoadProfile(ctx)
	}

	g := gameplay.New(gameplay.Config{
		Width:         cfg.Width,
		Height:        cfg.Height,
		FoodsPerLevel: cfg.FoodsPerLevel,
		ObstaclesStep: cfg.ObstaclesStep,
	}, s.random)
	g.ApplyProfile(profileToDomain(s.profile))
	g.SetDeveloperMode(s.developerMode)

	s.game = g
	s.preset = cfg
	s.savedRuns = s.game.Snapshot(s.clock.Now()).RunsPlayed
	return nil
}

func (s *Service) ApplyDirection(dir DirectionInput) bool {
	if s.game == nil {
		return false
	}
	d := toDomainDirection(dir)
	if d == gameplay.DirNone {
		return false
	}
	now := s.clock.Now()
	snap := s.game.Snapshot(now)
	if !snap.Started {
		return s.game.Start(d, now)
	}
	return s.game.SetDirection(d)
}

func (s *Service) SetDeveloperMode(enabled bool) {
	s.developerMode = enabled
	if s.game != nil {
		s.game.SetDeveloperMode(enabled)
	}
}

func (s *Service) BypassLevel(level int) error {
	if s.game == nil {
		return errors.New("session is not started")
	}
	return s.game.BypassToLevel(level)
}

func (s *Service) Tick() {
	if s.game == nil {
		return
	}
	s.game.Tick(s.clock.Now())
	_ = s.persistIfNeeded(context.Background())
}

func (s *Service) TogglePause() bool {
	if s.game == nil {
		return false
	}
	return s.game.TogglePause()
}

func (s *Service) Restart() error {
	if s.game == nil {
		return errors.New("session is not started")
	}
	s.game.Reset(s.clock.Now())
	return s.persistIfNeeded(context.Background())
}

func (s *Service) Quit() error {
	if s.game == nil {
		return nil
	}
	s.game.Finalize(s.clock.Now())
	return s.persistIfNeeded(context.Background())
}

func (s *Service) Snapshot() SessionSnapshot {
	if s.game == nil {
		return SessionSnapshot{}
	}
	now := s.clock.Now()
	snap := s.game.Snapshot(now)

	out := SessionSnapshot{
		Width:            snap.Width,
		Height:           snap.Height,
		Snake:            mapPointsFromDomain(snap.Snake),
		Obstacles:        mapPointsFromDomain(snap.Obstacles),
		Food:             pointFromDomain(snap.Food),
		Direction:        fromDomainDirection(snap.Direction),
		Score:            snap.Score,
		FoodEaten:        snap.FoodEaten,
		Level:            snap.Level,
		FoodsToNextLevel: snap.FoodsToNextLevel,
		Started:          snap.Started,
		IsOver:           snap.IsOver,
		IsWon:            snap.IsWon,
		Paused:           snap.Paused,
		Elapsed:          snap.Elapsed,
		BestScore:        snap.BestScore,
		BestLength:       snap.BestLength,
		BestDuration:     snap.BestDuration,
		RunsPlayed:       snap.RunsPlayed,
		TotalFoodEaten:   snap.TotalFoodEaten,
		TotalPlayTime:    snap.TotalPlayTime,
		LastRun:          runSummaryFromDomain(snap.LastRun),
		HasLastRun:       snap.HasLastRun,
	}
	out.TickInterval = calcTickInterval(s.preset, out.Level)
	return out
}

func (s *Service) LastRunSummary() (RunSummaryView, bool) {
	snap := s.Snapshot()
	return snap.LastRun, snap.HasLastRun
}

func (s *Service) persistIfNeeded(ctx context.Context) error {
	if s.repo == nil || s.game == nil || s.developerMode {
		return nil
	}
	current := s.game.Snapshot(s.clock.Now())
	if current.RunsPlayed == s.savedRuns {
		return nil
	}
	p := profileFromDomain(s.game.Profile())
	if err := s.repo.Save(ctx, p); err != nil {
		return err
	}
	s.savedRuns = current.RunsPlayed
	s.profile = p
	return nil
}

func withDefaults(cfg PresetConfig) PresetConfig {
	if cfg.Width <= 0 {
		cfg.Width = 20
	}
	if cfg.Height <= 0 {
		cfg.Height = 15
	}
	if cfg.FoodsPerLevel <= 0 {
		cfg.FoodsPerLevel = gameplay.DefaultFoodsPerLevel
	}
	if cfg.ObstaclesStep <= 0 {
		cfg.ObstaclesStep = gameplay.DefaultObstaclesStep
	}
	if cfg.BaseTick <= 0 {
		cfg.BaseTick = 140 * time.Millisecond
	}
	if cfg.MinTick <= 0 {
		cfg.MinTick = 70 * time.Millisecond
	}
	if cfg.LevelStep < 0 {
		cfg.LevelStep = 0
	}
	if cfg.LevelStep == 0 {
		cfg.LevelStep = 8 * time.Millisecond
	}
	return cfg
}

func calcTickInterval(cfg PresetConfig, level int) time.Duration {
	cfg = withDefaults(cfg)
	lvl := level - 1
	if lvl < 0 {
		lvl = 0
	}
	interval := cfg.BaseTick - time.Duration(lvl)*cfg.LevelStep
	if interval < cfg.MinTick {
		return cfg.MinTick
	}
	return interval
}

func toDomainDirection(dir DirectionInput) gameplay.Direction {
	switch dir {
	case DirectionUp:
		return gameplay.DirUp
	case DirectionDown:
		return gameplay.DirDown
	case DirectionLeft:
		return gameplay.DirLeft
	case DirectionRight:
		return gameplay.DirRight
	default:
		return gameplay.DirNone
	}
}

func fromDomainDirection(dir gameplay.Direction) DirectionInput {
	switch dir {
	case gameplay.DirUp:
		return DirectionUp
	case gameplay.DirDown:
		return DirectionDown
	case gameplay.DirLeft:
		return DirectionLeft
	case gameplay.DirRight:
		return DirectionRight
	default:
		return DirectionNone
	}
}

func mapPointsFromDomain(in []gameplay.Point) []Point {
	out := make([]Point, len(in))
	for i, p := range in {
		out[i] = pointFromDomain(p)
	}
	return out
}

func pointFromDomain(p gameplay.Point) Point {
	return Point{X: p.X, Y: p.Y}
}

func profileFromDomain(p gameplay.Profile) Profile {
	return Profile{
		BestScore:           p.BestScore,
		BestLength:          p.BestLength,
		BestDurationMillis:  p.BestDurationMillis,
		RunsPlayed:          p.RunsPlayed,
		TotalFoodEaten:      p.TotalFoodEaten,
		TotalPlayTimeMillis: p.TotalPlayTimeMillis,
	}
}

func profileToDomain(p Profile) gameplay.Profile {
	return gameplay.Profile{
		BestScore:           p.BestScore,
		BestLength:          p.BestLength,
		BestDurationMillis:  p.BestDurationMillis,
		RunsPlayed:          p.RunsPlayed,
		TotalFoodEaten:      p.TotalFoodEaten,
		TotalPlayTimeMillis: p.TotalPlayTimeMillis,
	}
}

func runSummaryFromDomain(s gameplay.RunSummary) RunSummaryView {
	return RunSummaryView{
		Score:                   s.Score,
		Length:                  s.Length,
		FoodEaten:               s.FoodEaten,
		Level:                   s.Level,
		Duration:                s.Duration,
		Won:                     s.Won,
		ScoreDeltaVsPrevBest:    s.ScoreDeltaVsPrevBest,
		LengthDeltaVsPrevBest:   s.LengthDeltaVsPrevBest,
		DurationDeltaVsPrevBest: s.DurationDeltaVsPrevBest,
		NewBestScore:            s.NewBestScore,
		NewBestLength:           s.NewBestLength,
		NewBestDuration:         s.NewBestDuration,
	}
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type mathRandom struct {
	rnd *rand.Rand
}

func (r *mathRandom) Intn(n int) int {
	return r.rnd.Intn(n)
}
