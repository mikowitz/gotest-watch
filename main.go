package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initRegistry()

	fmt.Println("gotest-watch started")

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\n\nReceived signal: %v\n", sig)
		fmt.Println("Shutting down gracefully...")
		cancel() // Cancel context to stop all goroutines
		os.Exit(0)
	}()

	cmdChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	fileChangeChan := make(chan FileChangeMessage, 10)
	testCompleteChan := make(chan TestCompleteMessage, 10)

	// Start file watcher in background
	root, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	go watchFiles(ctx, root, fileChangeChan)

	// Start stdin reader in background
	go readStdin(ctx, os.Stdin, cmdChan, helpChan)

	// Create test config for command handlers
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Start dispatcher (blocks until context is cancelled)
	dispatcher(ctx, config, fileChangeChan, cmdChan, helpChan, testCompleteChan)
}
