package main

import (
	"bufio"
	"context"
	"io"
	"strings"
)

func parseCommand(input string) (Command, []string) {
	input = strings.TrimSpace(input)
	inputs := strings.Fields(input)
	if len(inputs) == 0 {
		return Command(""), nil
	}
	if len(inputs) == 1 {
		return Command(inputs[0]), nil
	}
	return Command(inputs[0]), inputs[1:]
}

// readStdin reads commands from stdin and sends them to the appropriate channels.
// The readyChan parameter controls whether stdin processing is active or paused.
//
// IMPORTANT: readyChan MUST be buffered (capacity >= 1) to prevent deadlocks.
// The non-blocking select means readStdin isn't always listening on readyChan.
// If readyChan is unbuffered, senders will block indefinitely when readStdin isn't receiving.
func readStdin(ctx context.Context, r io.Reader, cmdChan chan CommandMessage, helpChan chan HelpMessage, readyChan chan bool) {
	scanner := bufio.NewScanner(r)
	ready := false

	for {
		// Wait until ready
		for !ready {
			select {
			case ready = <-readyChan:
			case <-ctx.Done():
				return
			}
		}

		// Check for ready state change (non-blocking) before scanning
		select {
		case ready = <-readyChan:
		case <-ctx.Done():
			return
		default:
		}

		// If not ready, loop back to wait
		if !ready {
			continue
		}

		// Scan next line
		if !scanner.Scan() {
			break
		}

		// Check if context was cancelled while we were scanning
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		cmd, args := parseCommand(line)

		if cmd == Command("") {
			continue
		}

		if cmd == HelpCmd {
			select {
			case helpChan <- HelpMessage{}:
			case <-ctx.Done():
				return
			}
		} else {
			select {
			case cmdChan <- CommandMessage{Command: cmd, Args: args}:
			case <-ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// TODO: log error
	}
}
