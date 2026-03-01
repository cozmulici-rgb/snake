package console

import (
	"strings"
	"testing"
	"time"

	"snake/internal/app/session"

	"github.com/eiannone/keyboard"
)

func TestMapInput(t *testing.T) {
	ev, ok := MapInput('w', 0)
	if !ok || ev.Kind != EventDir || ev.Direction != session.DirectionUp {
		t.Fatalf("unexpected mapping for w: ev=%+v ok=%v", ev, ok)
	}
	ev, ok = MapInput(0, keyboard.KeyEsc)
	if !ok || ev.Kind != EventQuit {
		t.Fatalf("unexpected mapping for esc: ev=%+v ok=%v", ev, ok)
	}
}

func TestRenderIncludesStatsAndControls(t *testing.T) {
	rendered := Render(session.SessionSnapshot{
		Width:            3,
		Height:           2,
		Snake:            []session.Point{{X: 1, Y: 1}},
		Food:             session.Point{X: 2, Y: 0},
		Score:            1,
		Level:            1,
		FoodsToNextLevel: 4,
		TickInterval:     100 * time.Millisecond,
	}, false)

	if !strings.Contains(rendered, "Snake (Console)") {
		t.Fatalf("expected title in render output")
	}
	if !strings.Contains(rendered, "Controls: WASD or Arrow Keys") {
		t.Fatalf("expected controls line in render output")
	}
}
