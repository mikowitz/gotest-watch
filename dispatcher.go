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
	readyChan chan bool,
) {
	testRunning := false

	// Signal that we're ready to process commands and show initial prompt
	readyChan <- true
	fmt.Print("> ")

	for {
		// Check context first
		select {
		case <-ctx.Done():
			// Wait for test to finish if running
			if testRunning {
				<-testCompleteChan
			}
			fmt.Println("Shutting down...")
			return
		default:
		}

		select {
		case <-fileChangeChan:
			if !testRunning {
				testRunning = true
				readyChan <- false
				go runTests(ctx, config, testCompleteChan, readyChan, nil, nil)
			}

		case cmd := <-commandChan:
			// Execute command handler
			if err := handleCommand(cmd.Command, config, cmd.Args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			// Spawn test runner if command requires it
			if !testRunning && cmd.Command == ForceRunCmd {
				testRunning = true
				readyChan <- false
				go runTests(ctx, config, testCompleteChan, readyChan, nil, nil)
			}

		case <-helpChan:
			// Handle help - does NOT spawn test runner
			if err := handleHelp(config, nil); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

		case <-testCompleteChan:
			testRunning = false
			readyChan <- true
			fmt.Println()
			fmt.Print("> ")

		case <-ctx.Done():
			// Wait for test to finish if running
			if testRunning {
				<-testCompleteChan
			}
			fmt.Println("Shutting down...")
			return
		}
	}
}
