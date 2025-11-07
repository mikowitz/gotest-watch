package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	initRegistry()

	fmt.Println("gotest-watch started")

	// Create a cancellable context for graceful shutdown
	ctx, _ := setupSignalHandler()

	// Create test config for command handlers
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Store config in context
	ctx = withConfig(ctx, config)

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

	// Start dispatcher (blocks until context is cancelled)
	dispatcher(ctx, fileChangeChan, cmdChan, helpChan, testCompleteChan)
}
