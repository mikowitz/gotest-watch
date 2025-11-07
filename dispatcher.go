package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

//nolint:funlen
func dispatcher(
	ctx context.Context,
	fileChangeChan chan FileChangeMessage,
	commandChan chan CommandMessage,
	helpChan chan HelpMessage,
	testCompleteChan chan TestCompleteMessage,
) {
	testRunning := false

	config := getConfig(ctx)

	// Show initial prompt
	fmt.Print("> ")

	for {
		if testRunning {
			// While test is running, only listen for test completion and context cancellation
			// Ignore file changes and user commands (but show feedback for commands)
			select {
			case <-fileChangeChan:
				// Ignore file changes while test is running
			case cmd := <-commandChan:
				// Show the full line that was typed, so user knows what was ignored
				fullCmd := string(cmd.Command)
				if len(cmd.Args) > 0 {
					fullCmd = fullCmd + " " + strings.Join(cmd.Args, " ")
				}
				fmt.Printf("\n(Tests running - ignored input: '%s')\n", fullCmd)
			case <-helpChan:
				// Show that help was requested but ignored
				fmt.Println("\n(Tests running - ignored input: 'h')")
			case <-testCompleteChan:
				testRunning = false

				// Drain any commands that accumulated during test run
				drainedCommands := 0
				drainedHelp := 0
			drainLoop:
				for {
					select {
					case cmd := <-commandChan:
						drainedCommands++
						fullCmd := string(cmd.Command)
						if len(cmd.Args) > 0 {
							fullCmd = fullCmd + " " + strings.Join(cmd.Args, " ")
						}
						fmt.Printf("(Ignored during test: '%s')\n", fullCmd)
					case <-helpChan:
						drainedHelp++
						fmt.Println("(Ignored during test: 'h')")
					default:
						break drainLoop
					}
				}

				if drainedCommands > 0 || drainedHelp > 0 {
					fmt.Println()
				}

				// Show prompt
				fmt.Print("> ")
			case <-ctx.Done():
				// Wait for test to finish before shutting down
				select {
				case <-testCompleteChan:
					fmt.Println("Shutting down...")
					return
				case <-time.After(5 * time.Second):
					fmt.Fprintln(os.Stderr, "Timeout waiting for test to complete, forcing shutdown...")
					return
				}
			}
		} else {
			// When idle, process all events
			select {
			case <-fileChangeChan:
				testRunning = true
				fmt.Println("\nFile change detected, running tests...")
				go runTests(ctx, testCompleteChan, nil, nil)

			case cmd := <-commandChan:
				// Execute command handler
				if err := handleCommand(cmd.Command, config, cmd.Args); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}

				// Spawn test runner if command requires it
				if cmd.Command == ForceRunCmd {
					testRunning = true
					go runTests(ctx, testCompleteChan, nil, nil)
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
