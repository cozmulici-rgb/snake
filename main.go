package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	boardWidth  = 20
	boardHeight = 15
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
	in := bufio.NewReader(os.Stdin)

	for !g.over {
		clearScreen()
		render(g)
		fmt.Print("Move (W/A/S/D), Q to quit: ")

		text, _ := in.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		if len(text) > 0 {
			switch text[0] {
			case 'w':
				if g.dir.y != 1 {
					g.dir = point{x: 0, y: -1}
				}
			case 's':
				if g.dir.y != -1 {
					g.dir = point{x: 0, y: 1}
				}
			case 'a':
				if g.dir.x != 1 {
					g.dir = point{x: -1, y: 0}
				}
			case 'd':
				if g.dir.x != -1 {
					g.dir = point{x: 1, y: 0}
				}
			case 'q':
				fmt.Println("Goodbye!")
				return
			}
		}

		step(&g)
	}

	clearScreen()
	render(g)
	fmt.Println("Game Over!")
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
