package main

import (
	"context"
	"log"

	"snake/internal/app/session"
	infprofile "snake/internal/infra/profile"
	"snake/internal/infra/system"
	"snake/internal/ui/web"
)

func main() {
	svc := session.NewService(system.RealClock{}, system.NewMathRandom(0), infprofile.NewFileRepository(infprofile.DefaultPath()))
	if err := svc.LoadProfile(context.Background()); err != nil {
		log.Printf("warning: could not load profile: %v", err)
	}

	if err := web.Run(svc); err != nil {
		log.Fatal(err)
	}
}
