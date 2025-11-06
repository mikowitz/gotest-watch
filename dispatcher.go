package main

import (
	"context"
	"fmt"
	"os"
)

func dispatcher(
	ctx context.Context,
	config *TestConfig,
	fileChangeChan chan FileChangeMessage,
	commandChan chan CommandMessage,
	helpChan chan HelpMessage,
	testCompleteChan chan TestCompleteMessage,
) {
	testRunning := false

	// Show initial prompt
	fmt.Print("> ")

	for {
		if testRunning {
			// While test is running, only listen for test completion and context cancellation
			// Ignore file changes and user commands
			select {
			case <-fileChangeChan:
				// Ignore file changes while test is running
			case <-commandChan:
				// Ignore user commands while test is running
			case <-helpChan:
				// Ignore help requests while test is running
			case <-testCompleteChan:
				testRunning = false
				// fmt.Println()
				fmt.Print("> ")
			case <-ctx.Done():
				// Wait for test to finish before shutting down
				<-testCompleteChan
				fmt.Println("Shutting down...")
				return
			}
		} else {
			// When idle, process all events
			select {
			case <-fileChangeChan:
				testRunning = true
				fmt.Println("\nFile change detected, running tests...")
				go runTests(ctx, config, testCompleteChan, nil, nil)

			case cmd := <-commandChan:
				// Execute command handler
				if err := handleCommand(cmd.Command, config, cmd.Args); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}

				// Spawn test runner if command requires it
				if cmd.Command == ForceRunCmd {
					testRunning = true
					go runTests(ctx, config, testCompleteChan, nil, nil)
				} else {
					// Show prompt after non-test commands
					fmt.Print("> ")
				}

			case <-helpChan:
				// Handle help - does NOT spawn test runner
				if err := handleHelp(config, nil); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				// Show prompt after help
				fmt.Print("> ")

			case <-ctx.Done():
				fmt.Println("Shutting down...")
				return
			}
		}
	}
}
