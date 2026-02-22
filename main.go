package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/eiannone/keyboard"
)

const (
	boardWidth  = 20
	boardHeight = 15
	tickRate    = 140 * time.Millisecond
)

type point struct {
	x int
	y int
}

type game struct {
	snake []point
	dir   point
	food  point
	score int
	over  bool
}

func newGame() game {
	rand.Seed(time.Now().UnixNano())
	g := game{
		snake: []point{{x: boardWidth / 2, y: boardHeight / 2}},
		dir:   point{x: 1, y: 0},
	}
	g.food = randomFood(g.snake)
	return g
}

func main() {
	g := newGame()

	if err := keyboard.Open(); err != nil {
		fmt.Printf("failed to read keyboard input: %v\n", err)
		return
	}
	defer func() {
		_ = keyboard.Close()
	}()

	dirCh := make(chan point, 1)
	quitCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go readInput(dirCh, quitCh, errCh)

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	clearScreen()
	render(g)

	for !g.over {
		select {
		case dir := <-dirCh:
			if !isOpposite(g.dir, dir) {
				g.dir = dir
			}
		case <-ticker.C:
			step(&g)
			clearScreen()
			render(g)
		case <-quitCh:
			clearScreen()
			render(g)
			fmt.Println("Goodbye!")
			return
		case err := <-errCh:
			fmt.Printf("input error: %v\n", err)
			return
		}
	}

	clearScreen()
	render(g)
	fmt.Println("Game Over!")
}

func readInput(dirCh chan point, quitCh chan<- struct{}, errCh chan<- error) {
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
			next point
			ok   bool
		)

		switch key {
		case keyboard.KeyArrowUp:
			next, ok = point{x: 0, y: -1}, true
		case keyboard.KeyArrowDown:
			next, ok = point{x: 0, y: 1}, true
		case keyboard.KeyArrowLeft:
			next, ok = point{x: -1, y: 0}, true
		case keyboard.KeyArrowRight:
			next, ok = point{x: 1, y: 0}, true
		case keyboard.KeyEsc:
			select {
			case quitCh <- struct{}{}:
			default:
			}
			return
		}

		switch char {
		case 'w', 'W':
			next, ok = point{x: 0, y: -1}, true
		case 's', 'S':
			next, ok = point{x: 0, y: 1}, true
		case 'a', 'A':
			next, ok = point{x: -1, y: 0}, true
		case 'd', 'D':
			next, ok = point{x: 1, y: 0}, true
		case 'q', 'Q':
			select {
			case quitCh <- struct{}{}:
			default:
			}
			return
		}

		if ok {
			select {
			case dirCh <- next:
			default:
				<-dirCh
				dirCh <- next
			}
		}
	}
}

func step(g *game) {
	head := g.snake[0]
	next := point{x: head.x + g.dir.x, y: head.y + g.dir.y}

	if next.x < 0 || next.y < 0 || next.x >= boardWidth || next.y >= boardHeight {
		g.over = true
		return
	}
	if contains(g.snake, next) {
		g.over = true
		return
	}

	g.snake = append([]point{next}, g.snake...)
	if next == g.food {
		g.score++
		g.food = randomFood(g.snake)
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}
}

func render(g game) {
	fmt.Printf("Simple Snake (Windows console) | Score: %d\n\n", g.score)
	for y := 0; y < boardHeight+2; y++ {
		for x := 0; x < boardWidth+2; x++ {
			switch {
			case x == 0 || y == 0 || x == boardWidth+1 || y == boardHeight+1:
				fmt.Print("#")
			default:
				p := point{x: x - 1, y: y - 1}
				switch {
				case p == g.food:
					fmt.Print("*")
				case p == g.snake[0]:
					fmt.Print("@")
				case contains(g.snake[1:], p):
					fmt.Print("o")
				default:
					fmt.Print(" ")
				}
			}
		}
		fmt.Println()
	}
	fmt.Println("\nControls: WASD or Arrow Keys to move, Q/Esc to quit.")
}

func isOpposite(a, b point) bool {
	return a.x == -b.x && a.y == -b.y
}

func contains(parts []point, p point) bool {
	for _, s := range parts {
		if s == p {
			return true
		}
	}
	return false
}

func randomFood(snake []point) point {
	for {
		p := point{x: rand.Intn(boardWidth), y: rand.Intn(boardHeight)}
		if !contains(snake, p) {
			return p
		}
	}
}

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
