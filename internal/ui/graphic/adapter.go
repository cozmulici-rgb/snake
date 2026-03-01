package graphic

import (
	"fmt"
	"time"

	"snake/internal/app/session"
)

type HUD struct {
	Line1  string
	Line2  string
	Line3  string
	Msg    string
	Detail string
}

type StartOverlay struct {
	Title    string
	Subtitle string
	Hint     string
}

type MenuPreset struct {
	Name        string
	Description string
}

type MainMenu struct {
	Title       string
	Subtitle    string
	PresetLines []string
	StatsLine1  string
	StatsLine2  string
	HelpLine    string
}

func BuildHUD(modeName string, snap session.SessionSnapshot) HUD {
	speed := 0.0
	if snap.TickInterval > 0 {
		speed = 1 / snap.TickInterval.Seconds()
	}

	h := HUD{
		Line1: fmt.Sprintf("Mode:%s  Score:%d  Length:%d  Level:%d", modeName, snap.Score, len(snap.Snake), snap.Level),
		Line2: fmt.Sprintf("Food:%d  NextLvl:%d  Obst:%d  Time:%s  Speed:%.1f/s", snap.FoodEaten, snap.FoodsToNextLevel, len(snap.Obstacles), FormatDuration(snap.Elapsed), speed),
		Line3: fmt.Sprintf("BestScore:%d  BestLen:%d  BestTime:%s  Runs:%d", snap.BestScore, snap.BestLength, FormatDuration(snap.BestDuration), snap.RunsPlayed),
		Msg:   "WASD/Arrows move | P pause | F11 fullscreen | Q/Esc quit",
	}

	if !snap.Started {
		h.Msg = "Press any direction key to start"
	}
	if snap.Paused {
		h.Msg = "Paused | P resume | Q/Esc quit"
	}
	if snap.IsOver {
		if snap.IsWon {
			h.Msg = "You win! R restart | M menu | Q/Esc quit"
		} else {
			h.Msg = "Game Over | R restart | M menu | Q/Esc quit"
		}
		if snap.HasLastRun {
			h.Detail = fmt.Sprintf("Run: Score %d (%s)  Len %d (%s)  Time %s (%s)",
				snap.LastRun.Score,
				FormatSignedInt(snap.LastRun.ScoreDeltaVsPrevBest),
				snap.LastRun.Length,
				FormatSignedInt(snap.LastRun.LengthDeltaVsPrevBest),
				FormatDuration(snap.LastRun.Duration),
				FormatSignedDuration(snap.LastRun.DurationDeltaVsPrevBest))
		}
	}

	return h
}

func BuildStartOverlay(modeName string, snap session.SessionSnapshot) (StartOverlay, bool) {
	if snap.Started || snap.IsOver {
		return StartOverlay{}, false
	}

	if modeName == "" {
		modeName = "Balanced"
	}

	return StartOverlay{
		Title:    "SNAKE",
		Subtitle: fmt.Sprintf("Mode: %s", modeName),
		Hint:     "Press WASD/Arrows to start",
	}, true
}

func BuildMainMenu(selected int, presets []MenuPreset, profile session.Profile) MainMenu {
	menu := MainMenu{
		Title:    "SNAKE",
		Subtitle: "Choose a mode and start playing",
		HelpLine: "Up/Down or 1-3 select | Enter/Space start | Q/Esc quit",
	}

	menu.PresetLines = make([]string, 0, len(presets))
	for i, p := range presets {
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		menu.PresetLines = append(menu.PresetLines, fmt.Sprintf("%s%d. %s - %s", prefix, i+1, p.Name, p.Description))
	}

	menu.StatsLine1 = fmt.Sprintf("BestScore:%d  BestLen:%d  BestTime:%s  Runs:%d",
		profile.BestScore,
		profile.BestLength,
		FormatDuration(time.Duration(profile.BestDurationMillis)*time.Millisecond),
		profile.RunsPlayed)
	menu.StatsLine2 = fmt.Sprintf("TotalFood:%d  TotalPlay:%s",
		profile.TotalFoodEaten,
		FormatDuration(time.Duration(profile.TotalPlayTimeMillis)*time.Millisecond))

	return menu
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func FormatSignedInt(v int) string {
	if v > 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}

func FormatSignedDuration(d time.Duration) string {
	if d >= 0 {
		return "+" + FormatDuration(d)
	}
	return "-" + FormatDuration(-d)
}
