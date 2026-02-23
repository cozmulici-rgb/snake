package main

import (
	"fmt"
	"strings"
	"time"

	"snake/internal/game"

	"github.com/eiannone/keyboard"
)

const (
	boardWidth  = 20
	boardHeight = 15
	baseTick    = 140 * time.Millisecond
	minTick     = 70 * time.Millisecond
	levelStep   = 8 * time.Millisecond
)

type inputEventKind int

const (
	eventDir inputEventKind = iota
	eventQuit
	eventRestart
	eventPause
)

type inputEvent struct {
	kind inputEventKind
	dir  game.Direction
}

func main() {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("failed to read keyboard input: %v\n", err)
		return
	}
	defer func() {
		_ = keyboard.Close()
	}()

	eventCh := make(chan inputEvent, 16)
	errCh := make(chan error, 1)
	go readInput(eventCh, errCh)

	initScreen()
	defer restoreScreen()

	cfg := game.Config{Width: boardWidth, Height: boardHeight}
	state := game.New(cfg, nil)

	for {
		currentTick := state.TickInterval(baseTick, minTick, levelStep)
		ticker := time.NewTicker(currentTick)
		paused := false
		drainInput(eventCh)
		render(state, false, paused, currentTick)

		for !state.IsOver() {
			select {
			case ev := <-eventCh:
				switch ev.kind {
				case eventDir:
					state.SetDirection(ev.dir)
				case eventQuit:
					ticker.Stop()
					render(state, false, paused, currentTick)
					fmt.Println("Goodbye!")
					return
				case eventPause:
					paused = !paused
					render(state, false, paused, currentTick)
				}
			case <-ticker.C:
				if !paused {
					state.Tick()
					nextTick := state.TickInterval(baseTick, minTick, levelStep)
					if nextTick != currentTick {
						ticker.Reset(nextTick)
						currentTick = nextTick
					}
				}
				render(state, false, paused, currentTick)
			case err := <-errCh:
				ticker.Stop()
				fmt.Printf("input error: %v\n", err)
				return
			}
		}

		ticker.Stop()
		render(state, true, false, currentTick)
		if !waitForEndMenu(eventCh, errCh) {
			fmt.Println("Goodbye!")
			return
		}
		state.Reset()
	}
}

func readInput(eventCh chan<- inputEvent, errCh chan<- error) {
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
			return
		}

		var (
			dir game.Direction
			ok  bool
		)

		switch key {
		case keyboard.KeyArrowUp:
			dir, ok = game.DirUp, true
		case keyboard.KeyArrowDown:
			dir, ok = game.DirDown, true
		case keyboard.KeyArrowLeft:
			dir, ok = game.DirLeft, true
		case keyboard.KeyArrowRight:
			dir, ok = game.DirRight, true
		case keyboard.KeyEsc:
			eventCh <- inputEvent{kind: eventQuit}
			continue
		}

		switch char {
		case 'w', 'W':
			dir, ok = game.DirUp, true
		case 's', 'S':
			dir, ok = game.DirDown, true
		case 'a', 'A':
			dir, ok = game.DirLeft, true
		case 'd', 'D':
			dir, ok = game.DirRight, true
		case 'r', 'R':
			eventCh <- inputEvent{kind: eventRestart}
			continue
		case 'p', 'P':
			eventCh <- inputEvent{kind: eventPause}
			continue
		case 'q', 'Q':
			eventCh <- inputEvent{kind: eventQuit}
			continue
		}

		if ok {
			eventCh <- inputEvent{kind: eventDir, dir: dir}
		}
	}
}

func render(state *game.State, showEndMenu bool, paused bool, tickInterval time.Duration) {
	var b strings.Builder
	width := state.Width()
	height := state.Height()
	snake := state.Snake()
	obstacles := state.Obstacles()
	food := state.Food()
	started := state.Started()
	speed := 0.0
	if tickInterval > 0 {
		speed = 1 / tickInterval.Seconds()
	}

	b.Grow((width + 4) * (height + 14))
	b.WriteString("\x1b[H")
	b.WriteString("Snake (Console)\n")
	b.WriteString(fmt.Sprintf("Score:%d  Length:%d  Level:%d  Food:%d  NextLvl:%d  Obst:%d  Speed:%.1f/s  Time:%s\n",
		state.Score(), state.SnakeLength(), state.Level(), state.FoodEaten(), state.FoodsToNextLevel(), state.ObstacleCount(), speed, formatDuration(state.Elapsed())))
	b.WriteString(fmt.Sprintf("BestScore:%d  BestLength:%d  BestTime:%s  Runs:%d  TotalTime:%s\n\n",
		state.BestScore(), state.BestLength(), formatDuration(state.BestDuration()), state.RunsPlayed(), formatDuration(state.TotalPlayTime())))

	for y := 0; y < height+2; y++ {
		for x := 0; x < width+2; x++ {
			switch {
			case x == 0 || y == 0 || x == width+1 || y == height+1:
				b.WriteByte('#')
			default:
				p := game.Point{X: x - 1, Y: y - 1}
				switch {
				case p == food:
					b.WriteByte('*')
				case contains(obstacles, p):
					b.WriteByte('x')
				case len(snake) > 0 && p == snake[0]:
					b.WriteByte('@')
				case contains(snake[1:], p):
					b.WriteByte('o')
				default:
					b.WriteByte(' ')
				}
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString("\nControls: WASD or Arrow Keys to move, P pause, Q/Esc quit.\n")
	if !started {
		b.WriteString("Press any direction key to start.\n")
	}
	if paused {
		b.WriteString("Paused. Press P to resume.\n")
	}
	if showEndMenu {
		if state.IsWon() {
			b.WriteString("You win! Press R to restart, Q/Esc to quit.\n")
		} else {
			b.WriteString("Game Over! Press R to restart, Q/Esc to quit.\n")
		}
		if summary, ok := state.LastRunSummary(); ok {
			b.WriteString(fmt.Sprintf("Run: Score %d (%s) | Length %d (%s) | Time %s (%s)\n",
				summary.Score,
				formatSignedInt(summary.ScoreDeltaVsPrevBest),
				summary.Length,
				formatSignedInt(summary.LengthDeltaVsPrevBest),
				formatDuration(summary.Duration),
				formatSignedDuration(summary.DurationDeltaVsPrevBest)))
		}
	}
	fmt.Print(b.String())
}

func waitForEndMenu(eventCh <-chan inputEvent, errCh <-chan error) bool {
	for {
		select {
		case ev := <-eventCh:
			switch ev.kind {
			case eventRestart:
				return true
			case eventQuit:
				return false
			}
		case err := <-errCh:
			fmt.Printf("input error: %v\n", err)
			return false
		}
	}
}

func drainInput(eventCh <-chan inputEvent) {
	for {
		select {
		case <-eventCh:
		default:
			return
		}
	}
}

func contains(parts []game.Point, p game.Point) bool {
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

func initScreen() {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
}

func restoreScreen() {
	fmt.Print("\x1b[?25h")
}
