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

func TestBuildStartOverlay_ActiveBeforeStart(t *testing.T) {
	overlay, ok := BuildStartOverlay("Balanced", session.SessionSnapshot{
		Started: false,
		IsOver:  false,
	})
	if !ok {
		t.Fatalf("expected overlay to be active before game start")
	}
	if !strings.Contains(overlay.Title, "SNAKE") {
		t.Fatalf("unexpected overlay title: %q", overlay.Title)
	}
	if !strings.Contains(overlay.Hint, "WASD/Arrows") {
		t.Fatalf("unexpected overlay hint: %q", overlay.Hint)
	}
}

func TestBuildStartOverlay_InactiveAfterStart(t *testing.T) {
	_, ok := BuildStartOverlay("Balanced", session.SessionSnapshot{Started: true})
	if ok {
		t.Fatalf("expected overlay to be inactive after start")
	}
}

func TestBuildStartOverlay_InactiveWhenOver(t *testing.T) {
	_, ok := BuildStartOverlay("Balanced", session.SessionSnapshot{IsOver: true})
	if ok {
		t.Fatalf("expected overlay to be inactive when game is over")
	}
}

func TestBuildMainMenu_IncludesGameNameAndSelection(t *testing.T) {
	menu := BuildMainMenu(1, []MenuPreset{
		{Name: "Balanced", Description: "Default progression"},
		{Name: "Hardcore", Description: "Faster and denser"},
	}, session.Profile{})

	if !strings.Contains(menu.Title, "SNAKE") {
		t.Fatalf("expected menu title to include game name, got=%q", menu.Title)
	}
	if len(menu.PresetLines) != 2 {
		t.Fatalf("unexpected number of preset lines: %d", len(menu.PresetLines))
	}
	if !strings.HasPrefix(menu.PresetLines[1], "> ") {
		t.Fatalf("expected selected preset marker on second item, got=%q", menu.PresetLines[1])
	}
}

func TestBuildMainMenu_IncludesProfileStats(t *testing.T) {
	menu := BuildMainMenu(0, []MenuPreset{{Name: "Balanced", Description: "Default"}}, session.Profile{
		BestScore:           20,
		RunsPlayed:          3,
		TotalFoodEaten:      42,
		BestDurationMillis:  90 * 1000,
		TotalPlayTimeMillis: 240 * 1000,
	})

	if !strings.Contains(menu.StatsLine1, "BestScore:20") {
		t.Fatalf("expected best score in stats line, got=%q", menu.StatsLine1)
	}
	if !strings.Contains(menu.StatsLine1, "Runs:3") {
		t.Fatalf("expected runs in stats line, got=%q", menu.StatsLine1)
	}
	if !strings.Contains(menu.StatsLine2, "TotalFood:42") {
		t.Fatalf("expected total food in stats line, got=%q", menu.StatsLine2)
	}
	if menu.HelpLine == "" {
		t.Fatalf("expected non-empty menu help line")
	}
}
