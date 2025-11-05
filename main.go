package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

//nolint:funlen
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
	readyChan := make(chan bool, 1)

	// Start file watcher in background
	root, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	go watchFiles(ctx, root, fileChangeChan)

	// Start stdin reader in background
	go readStdin(ctx, os.Stdin, cmdChan, helpChan, readyChan)

	// Create test config for command handlers
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Signal that we're ready to process commands and show initial prompt
	readyChan <- true
	fmt.Print("> ")

	// Message handling loop
	for {
		select {
		case cmd := <-cmdChan:
			// Handle force run command specially (needs channels)
			if cmd.Command == ForceRunCmd {
				fmt.Println("==> Running tests...")
				go runTests(ctx, config, testCompleteChan, readyChan)
			} else {
				// Execute other commands through registry
				if err := handleCommand(cmd.Command, config, cmd.Args); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				// Show prompt after command completes
				fmt.Print("> ")
			}

		case <-helpChan:
			// Handle help command
			if err := handleHelp(config, nil); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			// Show prompt after help
			fmt.Print("> ")

		case <-fileChangeChan:
			// File change detected - run tests automatically
			fmt.Println("\n==> File change detected, running tests...")
			go runTests(ctx, config, testCompleteChan, readyChan)

		case <-testCompleteChan:
			// Tests completed
			fmt.Print("==> Tests completed\n> ")
		}
	}
}
