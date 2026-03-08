package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"snake/internal/app/session"
)

const (
	boardWidth   = 40
	boardHeight  = 40
	defaultAddr  = "127.0.0.1:8080"
	loopInterval = 16 * time.Millisecond
)

type preset struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	BaseTick      time.Duration `json:"-"`
	MinTick       time.Duration `json:"-"`
	LevelStep     time.Duration `json:"-"`
	FoodsPerLevel int           `json:"foods_per_level"`
	ObstaclesStep int           `json:"obstacles_step"`
}

var presets = []preset{
	{
		Name:          "Balanced",
		Description:   "Default progression and obstacle density",
		BaseTick:      140 * time.Millisecond,
		MinTick:       70 * time.Millisecond,
		LevelStep:     8 * time.Millisecond,
		FoodsPerLevel: 5,
		ObstaclesStep: 2,
	},
	{
		Name:          "Relaxed",
		Description:   "Slower speed, gentler obstacle growth",
		BaseTick:      180 * time.Millisecond,
		MinTick:       95 * time.Millisecond,
		LevelStep:     6 * time.Millisecond,
		FoodsPerLevel: 6,
		ObstaclesStep: 1,
	},
	{
		Name:          "Hardcore",
		Description:   "Faster speed, denser obstacles",
		BaseTick:      120 * time.Millisecond,
		MinTick:       55 * time.Millisecond,
		LevelStep:     10 * time.Millisecond,
		FoodsPerLevel: 4,
		ObstaclesStep: 3,
	},
}

//go:embed static/*
var staticFiles embed.FS

type server struct {
	mu             sync.Mutex
	svc            session.SessionService
	currentPreset  int
	developerMode  bool
	lastTickAt     time.Time
	currentTickDur time.Duration
}

type stateResponse struct {
	Presets       []preset          `json:"presets"`
	CurrentPreset int               `json:"current_preset"`
	DeveloperMode bool              `json:"developer_mode"`
	Snapshot      snapshotResponse  `json:"snapshot"`
	Profile       profileResponse   `json:"profile"`
	Controls      map[string]string `json:"controls"`
	ServerTime    string            `json:"server_time"`
}

type snapshotResponse struct {
	Width            int                `json:"width"`
	Height           int                `json:"height"`
	Snake            []pointResponse    `json:"snake"`
	Obstacles        []pointResponse    `json:"obstacles"`
	Food             pointResponse      `json:"food"`
	Direction        string             `json:"direction"`
	Score            int                `json:"score"`
	FoodEaten        int                `json:"food_eaten"`
	Level            int                `json:"level"`
	FoodsToNextLevel int                `json:"foods_to_next_level"`
	Started          bool               `json:"started"`
	IsOver           bool               `json:"is_over"`
	IsWon            bool               `json:"is_won"`
	Paused           bool               `json:"paused"`
	ElapsedMillis    int64              `json:"elapsed_millis"`
	BestScore        int                `json:"best_score"`
	BestLength       int                `json:"best_length"`
	BestDurationMs   int64              `json:"best_duration_millis"`
	RunsPlayed       int                `json:"runs_played"`
	TotalFoodEaten   int                `json:"total_food_eaten"`
	TotalPlayTimeMs  int64              `json:"total_play_time_millis"`
	HasLastRun       bool               `json:"has_last_run"`
	LastRun          runSummaryResponse `json:"last_run"`
	TickIntervalMs   int64              `json:"tick_interval_millis"`
}

type pointResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type runSummaryResponse struct {
	Score                   int   `json:"score"`
	Length                  int   `json:"length"`
	FoodEaten               int   `json:"food_eaten"`
	Level                   int   `json:"level"`
	DurationMillis          int64 `json:"duration_millis"`
	Won                     bool  `json:"won"`
	ScoreDeltaVsPrevBest    int   `json:"score_delta_vs_prev_best"`
	LengthDeltaVsPrevBest   int   `json:"length_delta_vs_prev_best"`
	DurationDeltaVsPrevBest int64 `json:"duration_delta_vs_prev_best_millis"`
	NewBestScore            bool  `json:"new_best_score"`
	NewBestLength           bool  `json:"new_best_length"`
	NewBestDuration         bool  `json:"new_best_duration"`
}

type profileResponse struct {
	BestScore           int   `json:"best_score"`
	BestLength          int   `json:"best_length"`
	BestDurationMillis  int64 `json:"best_duration_millis"`
	RunsPlayed          int   `json:"runs_played"`
	TotalFoodEaten      int   `json:"total_food_eaten"`
	TotalPlayTimeMillis int64 `json:"total_play_time_millis"`
}

type startRequest struct {
	Preset int `json:"preset"`
}

type inputRequest struct {
	Direction string `json:"direction"`
}

type developerModeRequest struct {
	Enabled bool `json:"enabled"`
}

type developerLevelRequest struct {
	Level int `json:"level"`
}

func Run(svc session.SessionService) error {
	if svc == nil {
		return errors.New("session service is required")
	}

	srv := &server{
		svc:           svc,
		currentPreset: 0,
	}
	srv.currentTickDur = presets[0].BaseTick

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go srv.loop(ctx)

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticRoot)))
	mux.HandleFunc("/api/state", srv.handleState)
	mux.HandleFunc("/api/start", srv.handleStart)
	mux.HandleFunc("/api/input", srv.handleInput)
	mux.HandleFunc("/api/pause", srv.handlePause)
	mux.HandleFunc("/api/restart", srv.handleRestart)
	mux.HandleFunc("/api/developer-mode", srv.handleDeveloperMode)
	mux.HandleFunc("/api/developer-level", srv.handleDeveloperLevel)

	log.Printf("web UI listening at http://%s", defaultAddr)
	return http.ListenAndServe(defaultAddr, mux)
}

func (s *server) loop(ctx context.Context) {
	ticker := time.NewTicker(loopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.Lock()
			s.tickLocked(now)
			s.mu.Unlock()
		}
	}
}

func (s *server) tickLocked(now time.Time) {
	snap := s.svc.Snapshot()
	if snap.Width == 0 || !snap.Started || snap.Paused || snap.IsOver {
		return
	}
	if s.currentTickDur <= 0 {
		s.currentTickDur = snap.TickInterval
	}
	if s.lastTickAt.IsZero() {
		s.lastTickAt = now
	}
	for now.Sub(s.lastTickAt) >= s.currentTickDur {
		s.svc.Tick()
		s.lastTickAt = s.lastTickAt.Add(s.currentTickDur)
		snap = s.svc.Snapshot()
		s.currentTickDur = snap.TickInterval
		if snap.IsOver || snap.Paused {
			return
		}
	}
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON(w, http.StatusOK, s.stateLocked())
}

func (s *server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req startRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Preset < 0 || req.Preset >= len(presets) {
		writeError(w, http.StatusBadRequest, "invalid preset")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := presets[req.Preset]
	err := s.svc.Start(r.Context(), session.PresetConfig{
		Name:          p.Name,
		Width:         boardWidth,
		Height:        boardHeight,
		FoodsPerLevel: p.FoodsPerLevel,
		ObstaclesStep: p.ObstaclesStep,
		BaseTick:      p.BaseTick,
		MinTick:       p.MinTick,
		LevelStep:     p.LevelStep,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.currentPreset = req.Preset
	s.lastTickAt = time.Time{}
	s.currentTickDur = p.BaseTick
	writeJSON(w, http.StatusOK, s.stateLocked())
}

func (s *server) handleInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req inputRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	dir, err := parseDirection(req.Direction)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	applied := s.svc.ApplyDirection(dir)
	snap := s.svc.Snapshot()
	if applied && snap.Started {
		s.lastTickAt = time.Now()
		s.currentTickDur = snap.TickInterval
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"applied": applied,
		"state":   s.stateLocked(),
	})
}

func (s *server) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	paused := s.svc.TogglePause()
	if !paused {
		snap := s.svc.Snapshot()
		if snap.Started && !snap.IsOver {
			s.lastTickAt = time.Now()
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"paused": paused,
		"state":  s.stateLocked(),
	})
}

func (s *server) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.svc.Restart(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.lastTickAt = time.Time{}
	s.currentTickDur = s.svc.Snapshot().TickInterval
	writeJSON(w, http.StatusOK, s.stateLocked())
}

func (s *server) handleDeveloperMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req developerModeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.developerMode = req.Enabled
	s.svc.SetDeveloperMode(req.Enabled)
	writeJSON(w, http.StatusOK, s.stateLocked())
}

func (s *server) handleDeveloperLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req developerLevelRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.developerMode {
		writeError(w, http.StatusForbidden, "developer mode is not enabled")
		return
	}
	if err := s.svc.BypassLevel(req.Level); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	snap := s.svc.Snapshot()
	s.currentTickDur = snap.TickInterval
	if snap.Started && !snap.Paused && !snap.IsOver {
		s.lastTickAt = time.Now()
	} else {
		s.lastTickAt = time.Time{}
	}

	writeJSON(w, http.StatusOK, s.stateLocked())
}

func (s *server) stateLocked() stateResponse {
	snap := s.svc.Snapshot()
	profile := s.svc.Profile()

	return stateResponse{
		Presets:       presets,
		CurrentPreset: s.currentPreset,
		DeveloperMode: s.developerMode,
		Snapshot:      mapSnapshot(snap),
		Profile: profileResponse{
			BestScore:           profile.BestScore,
			BestLength:          profile.BestLength,
			BestDurationMillis:  profile.BestDurationMillis,
			RunsPlayed:          profile.RunsPlayed,
			TotalFoodEaten:      profile.TotalFoodEaten,
			TotalPlayTimeMillis: profile.TotalPlayTimeMillis,
		},
		Controls: map[string]string{
			"move":    "WASD or Arrow keys",
			"pause":   "P",
			"restart": "R",
			"start":   "Enter",
		},
		ServerTime: time.Now().Format(time.RFC3339),
	}
}

func mapSnapshot(snap session.SessionSnapshot) snapshotResponse {
	return snapshotResponse{
		Width:            snap.Width,
		Height:           snap.Height,
		Snake:            mapPoints(snap.Snake),
		Obstacles:        mapPoints(snap.Obstacles),
		Food:             mapPoint(snap.Food),
		Direction:        directionName(snap.Direction),
		Score:            snap.Score,
		FoodEaten:        snap.FoodEaten,
		Level:            snap.Level,
		FoodsToNextLevel: snap.FoodsToNextLevel,
		Started:          snap.Started,
		IsOver:           snap.IsOver,
		IsWon:            snap.IsWon,
		Paused:           snap.Paused,
		ElapsedMillis:    snap.Elapsed.Milliseconds(),
		BestScore:        snap.BestScore,
		BestLength:       snap.BestLength,
		BestDurationMs:   snap.BestDuration.Milliseconds(),
		RunsPlayed:       snap.RunsPlayed,
		TotalFoodEaten:   snap.TotalFoodEaten,
		TotalPlayTimeMs:  snap.TotalPlayTime.Milliseconds(),
		HasLastRun:       snap.HasLastRun,
		LastRun: runSummaryResponse{
			Score:                   snap.LastRun.Score,
			Length:                  snap.LastRun.Length,
			FoodEaten:               snap.LastRun.FoodEaten,
			Level:                   snap.LastRun.Level,
			DurationMillis:          snap.LastRun.Duration.Milliseconds(),
			Won:                     snap.LastRun.Won,
			ScoreDeltaVsPrevBest:    snap.LastRun.ScoreDeltaVsPrevBest,
			LengthDeltaVsPrevBest:   snap.LastRun.LengthDeltaVsPrevBest,
			DurationDeltaVsPrevBest: snap.LastRun.DurationDeltaVsPrevBest.Milliseconds(),
			NewBestScore:            snap.LastRun.NewBestScore,
			NewBestLength:           snap.LastRun.NewBestLength,
			NewBestDuration:         snap.LastRun.NewBestDuration,
		},
		TickIntervalMs: snap.TickInterval.Milliseconds(),
	}
}

func mapPoints(in []session.Point) []pointResponse {
	if len(in) == 0 {
		return []pointResponse{}
	}
	out := make([]pointResponse, len(in))
	for i, p := range in {
		out[i] = mapPoint(p)
	}
	return out
}

func mapPoint(in session.Point) pointResponse {
	return pointResponse{X: in.X, Y: in.Y}
}

func parseDirection(in string) (session.DirectionInput, error) {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "up", "w":
		return session.DirectionUp, nil
	case "down", "s":
		return session.DirectionDown, nil
	case "left", "a":
		return session.DirectionLeft, nil
	case "right", "d":
		return session.DirectionRight, nil
	default:
		return session.DirectionNone, fmt.Errorf("invalid direction %q", in)
	}
}

func directionName(dir session.DirectionInput) string {
	switch dir {
	case session.DirectionUp:
		return "up"
	case session.DirectionDown:
		return "down"
	case session.DirectionLeft:
		return "left"
	case session.DirectionRight:
		return "right"
	default:
		return "none"
	}
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	if r.Body == nil {
		return errors.New("missing request body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error":      message,
		"statusCode": strconv.Itoa(status),
	})
}
