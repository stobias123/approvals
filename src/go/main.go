package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/stobias123/slack_button/managers"
	"github.com/stobias123/slack_button/repositories"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// -- Repositories
	ar := repositories.NewApprovalRepositoryInMemory()

	// Managers
	am := managers.NewApprovalManager(ar)
	sm := managers.NewSlackManager(am)

	// Server
	s := NewServer(sm, am)

	go func() {
		if err := s.Router.Run(":8080"); err != nil {
			fmt.Printf("failed to run server: %v", err)
			cancel()
		}
	}()
	go func() {
		sm.Run(ctx)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()

	// Wait a moment for the goroutines to finish
	time.Sleep(1 * time.Second)
}
