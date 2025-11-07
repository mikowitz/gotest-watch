package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func setupSignalHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			sig := <-sigChan
			fmt.Printf("\n\nReceived signal: %v\n", sig)
			fmt.Println("Shutting down gracefully...")
			cancel()
			os.Exit(0)
		}
	}()

	return ctx, cancel
}
