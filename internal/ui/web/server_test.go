package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"snake/internal/app/session"
)

type fakeSessionService struct {
	profile           session.Profile
	snapshot          session.SessionSnapshot
	startErr          error
	restartErr        error
	loadProfileErr    error
	lastRun           session.RunSummaryView
	lastRunOK         bool
	startCalls        []session.PresetConfig
	applyCalls        []session.DirectionInput
	tickCalls         int
	togglePauseCalls  int
	restartCalls      int
	applyResult       bool
	togglePauseResult bool
	onStart           func(cfg session.PresetConfig)
	onApplyDirection  func(dir session.DirectionInput)
	onTick            func()
	onTogglePause     func() bool
	onRestart         func() error
}

func (f *fakeSessionService) LoadProfile(context.Context) error { return f.loadProfileErr }

func (f *fakeSessionService) Profile() session.Profile { return f.profile }

func (f *fakeSessionService) Start(_ context.Context, cfg session.PresetConfig) error {
	f.startCalls = append(f.startCalls, cfg)
	if f.onStart != nil {
		f.onStart(cfg)
	}
	return f.startErr
}

func (f *fakeSessionService) ApplyDirection(dir session.DirectionInput) bool {
	f.applyCalls = append(f.applyCalls, dir)
	if f.onApplyDirection != nil {
		f.onApplyDirection(dir)
	}
	return f.applyResult
}

func (f *fakeSessionService) Tick() {
	f.tickCalls++
	if f.onTick != nil {
		f.onTick()
	}
}

func (f *fakeSessionService) TogglePause() bool {
	f.togglePauseCalls++
	if f.onTogglePause != nil {
		return f.onTogglePause()
	}
	return f.togglePauseResult
}

func (f *fakeSessionService) Restart() error {
	f.restartCalls++
	if f.onRestart != nil {
		return f.onRestart()
	}
	return f.restartErr
}

func (f *fakeSessionService) Quit() error { return nil }

func (f *fakeSessionService) Snapshot() session.SessionSnapshot { return f.snapshot }

func (f *fakeSessionService) LastRunSummary() (session.RunSummaryView, bool) {
	return f.lastRun, f.lastRunOK
}

func newTestServer(svc *fakeSessionService) *server {
	return &server{
		svc:            svc,
		currentPreset:  0,
		currentTickDur: presets[0].BaseTick,
	}
}

func TestRunRequiresService(t *testing.T) {
	err := Run(nil)
	if err == nil || err.Error() != "session service is required" {
		t.Fatalf("expected nil-service error, got=%v", err)
	}
}

func TestParseDirectionAliasesAndInvalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    session.DirectionInput
		wantErr bool
	}{
		{name: "up alias", input: "W", want: session.DirectionUp},
		{name: "down alias", input: "down", want: session.DirectionDown},
		{name: "left alias", input: " a ", want: session.DirectionLeft},
		{name: "right alias", input: "RIGHT", want: session.DirectionRight},
		{name: "invalid", input: "north", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDirection(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected direction: got=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/start", strings.NewReader(`{"preset":1,"extra":true}`))
	var out startRequest

	err := decodeJSON(req, &out)
	if err == nil {
		t.Fatalf("expected decodeJSON to reject unknown fields")
	}
}

func TestMapSnapshotJSONUsesLowercasePointsAndEmptyArrays(t *testing.T) {
	snap := session.SessionSnapshot{
		Width:         40,
		Height:        40,
		Snake:         nil,
		Obstacles:     nil,
		Food:          session.Point{X: 4, Y: 7},
		Elapsed:       3 * time.Second,
		TickInterval:  140 * time.Millisecond,
		HasLastRun:    true,
		LastRun:       session.RunSummaryView{Duration: 2 * time.Second},
		BestDuration:  5 * time.Second,
		TotalPlayTime: 9 * time.Second,
	}

	data, err := json.Marshal(mapSnapshot(snap))
	if err != nil {
		t.Fatalf("marshal mapSnapshot: %v", err)
	}
	body := string(data)

	for _, want := range []string{
		`"snake":[]`,
		`"obstacles":[]`,
		`"food":{"x":4,"y":7}`,
		`"elapsed_millis":3000`,
		`"tick_interval_millis":140`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected JSON to contain %s, got %s", want, body)
		}
	}
}

func TestHandleStateRejectsWrongMethod(t *testing.T) {
	srv := newTestServer(&fakeSessionService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/state", nil)

	srv.handleState(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleStartRejectsInvalidPreset(t *testing.T) {
	srv := newTestServer(&fakeSessionService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/start", strings.NewReader(`{"preset":99}`))

	srv.handleStart(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleStartAppliesPresetConfigAndResetsTiming(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:        boardWidth,
			Height:       boardHeight,
			TickInterval: presets[1].BaseTick,
		},
	}
	srv := newTestServer(svc)
	srv.lastTickAt = time.Unix(123, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/start", strings.NewReader(`{"preset":1}`))

	srv.handleStart(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(svc.startCalls) != 1 {
		t.Fatalf("expected one start call, got=%d", len(svc.startCalls))
	}
	gotCfg := svc.startCalls[0]
	if gotCfg.Width != boardWidth || gotCfg.Height != boardHeight {
		t.Fatalf("unexpected board size: %+v", gotCfg)
	}
	if gotCfg.BaseTick != presets[1].BaseTick || gotCfg.MinTick != presets[1].MinTick || gotCfg.LevelStep != presets[1].LevelStep {
		t.Fatalf("unexpected preset timing: %+v", gotCfg)
	}
	if srv.currentPreset != 1 {
		t.Fatalf("expected preset index 1, got=%d", srv.currentPreset)
	}
	if !srv.lastTickAt.IsZero() {
		t.Fatalf("expected tick anchor reset after start")
	}
	if srv.currentTickDur != presets[1].BaseTick {
		t.Fatalf("unexpected current tick duration: got=%v want=%v", srv.currentTickDur, presets[1].BaseTick)
	}
}

func TestHandleInputRejectsInvalidDirection(t *testing.T) {
	srv := newTestServer(&fakeSessionService{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/input", strings.NewReader(`{"direction":"north"}`))

	srv.handleInput(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleInputSetsTickTimingOnSuccessfulStart(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:        boardWidth,
			Height:       boardHeight,
			Started:      false,
			TickInterval: 95 * time.Millisecond,
		},
		applyResult: true,
	}
	svc.onApplyDirection = func(dir session.DirectionInput) {
		svc.snapshot.Started = true
		svc.snapshot.Direction = dir
	}

	srv := newTestServer(svc)
	srv.lastTickAt = time.Time{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/input", strings.NewReader(`{"direction":"right"}`))

	srv.handleInput(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(svc.applyCalls) != 1 || svc.applyCalls[0] != session.DirectionRight {
		t.Fatalf("unexpected apply calls: %+v", svc.applyCalls)
	}
	if srv.lastTickAt.IsZero() {
		t.Fatalf("expected tick anchor to be set after successful start input")
	}
	if srv.currentTickDur != 95*time.Millisecond {
		t.Fatalf("unexpected tick duration: got=%v want=%v", srv.currentTickDur, 95*time.Millisecond)
	}
}

func TestHandlePauseResumeResetsTimer(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:   boardWidth,
			Height:  boardHeight,
			Started: true,
		},
		togglePauseResult: false,
	}
	srv := newTestServer(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/pause", nil)

	srv.handlePause(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if svc.togglePauseCalls != 1 {
		t.Fatalf("expected one pause toggle, got=%d", svc.togglePauseCalls)
	}
	if srv.lastTickAt.IsZero() {
		t.Fatalf("expected resume to reset tick anchor")
	}
}

func TestHandleRestartRequiresActiveSession(t *testing.T) {
	svc := &fakeSessionService{restartErr: errors.New("session is not started")}
	srv := newTestServer(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/restart", nil)

	srv.handleRestart(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleRestartResetsTimingOnSuccess(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:        boardWidth,
			Height:       boardHeight,
			TickInterval: 120 * time.Millisecond,
		},
	}
	srv := newTestServer(svc)
	srv.lastTickAt = time.Unix(456, 0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/restart", nil)

	srv.handleRestart(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if svc.restartCalls != 1 {
		t.Fatalf("expected one restart call, got=%d", svc.restartCalls)
	}
	if !srv.lastTickAt.IsZero() {
		t.Fatalf("expected tick anchor reset after restart")
	}
	if srv.currentTickDur != 120*time.Millisecond {
		t.Fatalf("unexpected tick duration: got=%v want=%v", srv.currentTickDur, 120*time.Millisecond)
	}
}

func TestTickLockedSkipsWhenGameIsNotRunnable(t *testing.T) {
	tests := []struct {
		name     string
		snapshot session.SessionSnapshot
	}{
		{name: "no board", snapshot: session.SessionSnapshot{}},
		{name: "not started", snapshot: session.SessionSnapshot{Width: boardWidth, Height: boardHeight, Started: false}},
		{name: "paused", snapshot: session.SessionSnapshot{Width: boardWidth, Height: boardHeight, Started: true, Paused: true}},
		{name: "over", snapshot: session.SessionSnapshot{Width: boardWidth, Height: boardHeight, Started: true, IsOver: true}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeSessionService{snapshot: tc.snapshot}
			srv := newTestServer(svc)

			srv.tickLocked(time.Unix(100, 0))

			if svc.tickCalls != 0 {
				t.Fatalf("expected no ticks, got=%d", svc.tickCalls)
			}
		})
	}
}

func TestTickLockedDoesNotTickImmediatelyWithoutAnchor(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:        boardWidth,
			Height:       boardHeight,
			Started:      true,
			TickInterval: 100 * time.Millisecond,
		},
	}
	srv := newTestServer(svc)
	srv.currentTickDur = 100 * time.Millisecond
	now := time.Unix(100, 0)

	srv.tickLocked(now)

	if svc.tickCalls != 0 {
		t.Fatalf("expected no tick on initial anchor set, got=%d", svc.tickCalls)
	}
	if !srv.lastTickAt.Equal(now) {
		t.Fatalf("expected tick anchor to be initialized to now")
	}
}

func TestTickLockedProcessesCatchUpTicks(t *testing.T) {
	svc := &fakeSessionService{
		snapshot: session.SessionSnapshot{
			Width:        boardWidth,
			Height:       boardHeight,
			Started:      true,
			TickInterval: 100 * time.Millisecond,
		},
	}
	srv := newTestServer(svc)
	srv.currentTickDur = 0
	now := time.Unix(100, 0)
	srv.lastTickAt = now.Add(-250 * time.Millisecond)

	srv.tickLocked(now)

	if svc.tickCalls != 2 {
		t.Fatalf("expected two catch-up ticks, got=%d", svc.tickCalls)
	}
	wantAnchor := now.Add(-50 * time.Millisecond)
	if !srv.lastTickAt.Equal(wantAnchor) {
		t.Fatalf("unexpected anchor after catch-up: got=%v want=%v", srv.lastTickAt, wantAnchor)
	}
	if srv.currentTickDur != 100*time.Millisecond {
		t.Fatalf("expected interval to come from snapshot, got=%v", srv.currentTickDur)
	}
}
