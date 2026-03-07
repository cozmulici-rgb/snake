package main

import (
	"log"

	"snake/internal/ui/web"
)

func main() {
	if err := web.Run(); err != nil {
		log.Fatal(err)
	}
}
