package main

import (
	"context"
	"fmt"
	"time"

	"snake/internal/app/session"
	infprofile "snake/internal/infra/profile"
	"snake/internal/infra/system"
	consoleui "snake/internal/ui/console"

	"github.com/eiannone/keyboard"
)

const (
	boardWidth  = 20
	boardHeight = 15
	baseTick    = 140 * time.Millisecond
	minTick     = 70 * time.Millisecond
	levelStep   = 8 * time.Millisecond
)

func main() {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("failed to read keyboard input: %v\n", err)
		return
	}
	defer func() {
		_ = keyboard.Close()
	}()

	eventCh := make(chan consoleui.Event, 16)
	errCh := make(chan error, 1)
	go readInput(eventCh, errCh)

	initScreen()
	defer restoreScreen()

	svc := session.NewService(system.RealClock{}, system.NewMathRandom(0), infprofile.NewFileRepository(infprofile.DefaultPath()))
	if err := svc.Start(context.Background(), session.PresetConfig{
		Name:          "Balanced",
		Width:         boardWidth,
		Height:        boardHeight,
		FoodsPerLevel: 5,
		ObstaclesStep: 2,
		BaseTick:      baseTick,
		MinTick:       minTick,
		LevelStep:     levelStep,
	}); err != nil {
		fmt.Printf("failed to initialize game session: %v\n", err)
		return
	}

	for {
		snap := svc.Snapshot()
		currentTick := snap.TickInterval
		ticker := time.NewTicker(currentTick)
		drainInput(eventCh)
		render(svc.Snapshot(), false)

		for !svc.Snapshot().IsOver {
			select {
			case ev := <-eventCh:
				switch ev.Kind {
				case consoleui.EventDir:
					svc.ApplyDirection(ev.Direction)
				case consoleui.EventQuit:
					err := svc.Quit()
					ticker.Stop()
					if err != nil {
						fmt.Printf("warning: could not persist profile: %v\n", err)
					}
					render(svc.Snapshot(), false)
					fmt.Println("Goodbye!")
					return
				case consoleui.EventPause:
					svc.TogglePause()
					render(svc.Snapshot(), false)
				}
			case <-ticker.C:
				svc.Tick()
				snap := svc.Snapshot()
				nextTick := snap.TickInterval
				if nextTick != currentTick {
					ticker.Reset(nextTick)
					currentTick = nextTick
				}
				render(snap, false)
			case err := <-errCh:
				ticker.Stop()
				fmt.Printf("input error: %v\n", err)
				return
			}
		}

		ticker.Stop()
		render(svc.Snapshot(), true)
		if !waitForEndMenu(eventCh, errCh) {
			if err := svc.Quit(); err != nil {
				fmt.Printf("warning: could not persist profile: %v\n", err)
			}
			fmt.Println("Goodbye!")
			return
		}
		if err := svc.Restart(); err != nil {
			fmt.Printf("failed to restart game: %v\n", err)
			return
		}
	}
}

func readInput(eventCh chan<- consoleui.Event, errCh chan<- error) {
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
			return
		}
		if ev, ok := consoleui.MapInput(char, key); ok {
			eventCh <- ev
		}
	}
}

func render(snapshot session.SessionSnapshot, showEndMenu bool) {
	fmt.Print(consoleui.Render(snapshot, showEndMenu))
}

func waitForEndMenu(eventCh <-chan consoleui.Event, errCh <-chan error) bool {
	for {
		select {
		case ev := <-eventCh:
			switch ev.Kind {
			case consoleui.EventRestart:
				return true
			case consoleui.EventQuit:
				return false
			}
		case err := <-errCh:
			fmt.Printf("input error: %v\n", err)
			return false
		}
	}
}

func drainInput(eventCh <-chan consoleui.Event) {
	for {
		select {
		case <-eventCh:
		default:
			return
		}
	}
}

func initScreen() {
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
}

func restoreScreen() {
	fmt.Print("\x1b[?25h")
}
