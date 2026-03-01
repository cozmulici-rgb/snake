package console

import (
	"fmt"
	"strings"
	"time"

	"snake/internal/app/session"

	"github.com/eiannone/keyboard"
)

type EventKind int

const (
	EventDir EventKind = iota
	EventQuit
	EventRestart
	EventPause
)

type Event struct {
	Kind      EventKind
	Direction session.DirectionInput
}

func MapInput(char rune, key keyboard.Key) (Event, bool) {
	switch key {
	case keyboard.KeyArrowUp:
		return Event{Kind: EventDir, Direction: session.DirectionUp}, true
	case keyboard.KeyArrowDown:
		return Event{Kind: EventDir, Direction: session.DirectionDown}, true
	case keyboard.KeyArrowLeft:
		return Event{Kind: EventDir, Direction: session.DirectionLeft}, true
	case keyboard.KeyArrowRight:
		return Event{Kind: EventDir, Direction: session.DirectionRight}, true
	case keyboard.KeyEsc:
		return Event{Kind: EventQuit}, true
	}

	switch char {
	case 'w', 'W':
		return Event{Kind: EventDir, Direction: session.DirectionUp}, true
	case 's', 'S':
		return Event{Kind: EventDir, Direction: session.DirectionDown}, true
	case 'a', 'A':
		return Event{Kind: EventDir, Direction: session.DirectionLeft}, true
	case 'd', 'D':
		return Event{Kind: EventDir, Direction: session.DirectionRight}, true
	case 'r', 'R':
		return Event{Kind: EventRestart}, true
	case 'p', 'P':
		return Event{Kind: EventPause}, true
	case 'q', 'Q':
		return Event{Kind: EventQuit}, true
	default:
		return Event{}, false
	}
}

func Render(snapshot session.SessionSnapshot, showEndMenu bool) string {
	var b strings.Builder
	width := snapshot.Width
	height := snapshot.Height
	speed := 0.0
	if snapshot.TickInterval > 0 {
		speed = 1 / snapshot.TickInterval.Seconds()
	}

	b.Grow((width + 4) * (height + 14))
	b.WriteString("\x1b[H")
	b.WriteString("Snake (Console)\n")
	b.WriteString(fmt.Sprintf("Score:%d  Length:%d  Level:%d  Food:%d  NextLvl:%d  Obst:%d  Speed:%.1f/s  Time:%s\n",
		snapshot.Score, len(snapshot.Snake), snapshot.Level, snapshot.FoodEaten, snapshot.FoodsToNextLevel, len(snapshot.Obstacles), speed, formatDuration(snapshot.Elapsed)))
	b.WriteString(fmt.Sprintf("BestScore:%d  BestLength:%d  BestTime:%s  Runs:%d  TotalTime:%s\n\n",
		snapshot.BestScore, snapshot.BestLength, formatDuration(snapshot.BestDuration), snapshot.RunsPlayed, formatDuration(snapshot.TotalPlayTime)))

	for y := 0; y < height+2; y++ {
		for x := 0; x < width+2; x++ {
			switch {
			case x == 0 || y == 0 || x == width+1 || y == height+1:
				b.WriteByte('#')
			default:
				p := session.Point{X: x - 1, Y: y - 1}
				switch {
				case p == snapshot.Food:
					b.WriteByte('*')
				case contains(snapshot.Obstacles, p):
					b.WriteByte('x')
				case len(snapshot.Snake) > 0 && p == snapshot.Snake[0]:
					b.WriteByte('@')
				case contains(snapshot.Snake[1:], p):
					b.WriteByte('o')
				default:
					b.WriteByte(' ')
				}
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString("\nControls: WASD or Arrow Keys to move, P pause, Q/Esc quit.\n")
	if !snapshot.Started {
		b.WriteString("Press any direction key to start.\n")
	}
	if snapshot.Paused {
		b.WriteString("Paused. Press P to resume.\n")
	}
	if showEndMenu {
		if snapshot.IsWon {
			b.WriteString("You win! Press R to restart, Q/Esc to quit.\n")
		} else {
			b.WriteString("Game Over! Press R to restart, Q/Esc to quit.\n")
		}
		if snapshot.HasLastRun {
			b.WriteString(fmt.Sprintf("Run: Score %d (%s) | Length %d (%s) | Time %s (%s)\n",
				snapshot.LastRun.Score,
				formatSignedInt(snapshot.LastRun.ScoreDeltaVsPrevBest),
				snapshot.LastRun.Length,
				formatSignedInt(snapshot.LastRun.LengthDeltaVsPrevBest),
				formatDuration(snapshot.LastRun.Duration),
				formatSignedDuration(snapshot.LastRun.DurationDeltaVsPrevBest)))
		}
	}
	return b.String()
}

func contains(parts []session.Point, p session.Point) bool {
	for _, s := range parts {
		if s == p {
			return true
		}
	}
	return false
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatSignedInt(v int) string {
	if v > 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}

func formatSignedDuration(d time.Duration) string {
	if d >= 0 {
		return "+" + formatDuration(d)
	}
	return "-" + formatDuration(-d)
}
