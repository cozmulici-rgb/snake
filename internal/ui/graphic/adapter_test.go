package graphic

import (
	"strings"
	"testing"
	"time"

	"snake/internal/app/session"
)

func TestBuildHUDLines(t *testing.T) {
	h := BuildHUD("Balanced", session.SessionSnapshot{
		Score:            3,
		Snake:            []session.Point{{X: 1, Y: 1}, {X: 1, Y: 2}},
		Level:            2,
		FoodEaten:        5,
		FoodsToNextLevel: 5,
		Obstacles:        []session.Point{{X: 0, Y: 0}},
		Elapsed:          2 * time.Minute,
		TickInterval:     100 * time.Millisecond,
	})

	if !strings.Contains(h.Line1, "Mode:Balanced") {
		t.Fatalf("unexpected line1: %s", h.Line1)
	}
	if !strings.Contains(h.Line2, "Speed:10.0/s") {
		t.Fatalf("unexpected line2: %s", h.Line2)
	}
}

func TestBuildHUDGameOverMessage(t *testing.T) {
	h := BuildHUD("Balanced", session.SessionSnapshot{IsOver: true, IsWon: false})
	if !strings.Contains(h.Msg, "Game Over") {
		t.Fatalf("expected game over message, got=%q", h.Msg)
	}
}
