package internal

import (
	"bufio"
	"context"
	"io"
	"log"
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
// It runs continuously in a goroutine, and the dispatcher decides whether to
// process or ignore commands based on whether tests are running.
func ReadStdin(
	ctx context.Context,
	r io.Reader,
	cmdChan chan CommandMessage,
	helpChan chan HelpMessage,
) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		// Check if context was cancelled
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
		log.Print(err)
	}
}
