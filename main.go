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
	tickRate    = 140 * time.Millisecond
)

type inputEventKind int

const (
	eventDir inputEventKind = iota
	eventQuit
	eventRestart
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
	for {
		state := game.New(cfg, nil)
		ticker := time.NewTicker(tickRate)
		drainInput(eventCh)
		render(state, false)

		for !state.IsOver() {
			select {
			case ev := <-eventCh:
				switch ev.kind {
				case eventDir:
					state.SetDirection(ev.dir)
				case eventQuit:
					ticker.Stop()
					render(state, false)
					fmt.Println("Goodbye!")
					return
				}
			case <-ticker.C:
				state.Tick()
				render(state, false)
			case err := <-errCh:
				ticker.Stop()
				fmt.Printf("input error: %v\n", err)
				return
			}
		}

		ticker.Stop()
		render(state, true)
		if !waitForEndMenu(eventCh, errCh) {
			fmt.Println("Goodbye!")
			return
		}
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
		case 'q', 'Q':
			eventCh <- inputEvent{kind: eventQuit}
			continue
		}

		if ok {
			eventCh <- inputEvent{kind: eventDir, dir: dir}
		}
	}
}

func render(state *game.State, showEndMenu bool) {
	var b strings.Builder
	width := state.Width()
	height := state.Height()
	snake := state.Snake()
	food := state.Food()
	started := state.Started()

	b.Grow((width + 4) * (height + 8))
	b.WriteString("\x1b[H")
	b.WriteString(fmt.Sprintf("Simple Snake (Windows console) | Score: %d\n\n", state.Score()))
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
	b.WriteString("\nControls: WASD or Arrow Keys to move, Q/Esc to quit.\n")
	if !started {
		b.WriteString("Press any direction key to start.\n")
	}
	if showEndMenu {
		b.WriteString("Game Over! Press R to restart, Q/Esc to quit.\n")
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

func initScreen() {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
}

func restoreScreen() {
	fmt.Print("\x1b[?25h")
}
