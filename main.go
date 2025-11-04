package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	initRegistry()

	fmt.Println("gotest-watch started")

	ctx := context.Background()
	cmdChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	fileChangeChan := make(chan FileChangeMessage, 10)
	readyChan := make(chan bool, 1)

	// Start file watcher in background
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	go watchFiles(ctx, root, fileChangeChan)

	// Start stdin reader in background
	go readStdin(ctx, os.Stdin, cmdChan, helpChan, readyChan)

	// Signal that we're ready to process commands
	readyChan <- true

	// Create test config for command handlers
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Message handling loop
	for {
		select {
		case cmd := <-cmdChan:
			// Execute command through registry
			if err := handleCommand(cmd.Command, config, cmd.Args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

		case <-helpChan:
			// Handle help command
			if err := handleHelp(config, nil); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

		case <-fileChangeChan:
			// File change detected
			fmt.Println("\n==> File change detected")
		}
	}
}
