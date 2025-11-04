package main

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseCommand tests the parseCommand helper function with various inputs
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedCommand string
		expectedArgs    []string
	}{
		{
			name:            "command only",
			input:           "help",
			expectedCommand: "help",
			expectedArgs:    nil,
		},
		{
			name:            "command with single arg",
			input:           "r TestFoo",
			expectedCommand: "r",
			expectedArgs:    []string{"TestFoo"},
		},
		{
			name:            "command with multiple args",
			input:           "p /path/to/dir arg2 arg3",
			expectedCommand: "p",
			expectedArgs:    []string{"/path/to/dir", "arg2", "arg3"},
		},
		{
			name:            "empty string",
			input:           "",
			expectedCommand: "",
			expectedArgs:    nil,
		},
		{
			name:            "whitespace only spaces",
			input:           "   ",
			expectedCommand: "",
			expectedArgs:    nil,
		},
		{
			name:            "whitespace only tabs",
			input:           "\t\t",
			expectedCommand: "",
			expectedArgs:    nil,
		},
		{
			name:            "whitespace only mixed",
			input:           "  \t  \n  ",
			expectedCommand: "",
			expectedArgs:    nil,
		},
		{
			name:            "leading whitespace",
			input:           "  help",
			expectedCommand: "help",
			expectedArgs:    nil,
		},
		{
			name:            "trailing whitespace",
			input:           "help  ",
			expectedCommand: "help",
			expectedArgs:    nil,
		},
		{
			name:            "leading and trailing whitespace",
			input:           "  v  ",
			expectedCommand: "v",
			expectedArgs:    nil,
		},
		{
			name:            "multiple spaces between args",
			input:           "r   TestFoo",
			expectedCommand: "r",
			expectedArgs:    []string{"TestFoo"},
		},
		{
			name:            "tabs between args",
			input:           "r\tTestFoo\tTestBar",
			expectedCommand: "r",
			expectedArgs:    []string{"TestFoo", "TestBar"},
		},
		{
			name:            "mixed whitespace between args",
			input:           "p  \t /some/path  \t arg2",
			expectedCommand: "p",
			expectedArgs:    []string{"/some/path", "arg2"},
		},
		{
			name:            "command with path containing spaces would be split",
			input:           "p /path with spaces",
			expectedCommand: "p",
			expectedArgs:    []string{"/path", "with", "spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args := parseCommand(tt.input)

			assert.Equal(t, Command(tt.expectedCommand), command, "command should match")
			assert.Equal(t, tt.expectedArgs, args, "args should match")
		})
	}
}

// TestParseCommand_EdgeCases tests additional edge cases
func TestParseCommand_EdgeCases(t *testing.T) {
	t.Run("newline characters are treated as whitespace", func(t *testing.T) {
		command, args := parseCommand("help\n")
		assert.Equal(t, Command("help"), command)
		assert.Nil(t, args)
	})

	t.Run("carriage return characters are treated as whitespace", func(t *testing.T) {
		command, args := parseCommand("help\r\n")
		assert.Equal(t, Command("help"), command)
		assert.Nil(t, args)
	})

	t.Run("multiple consecutive whitespace types normalized", func(t *testing.T) {
		command, args := parseCommand("  \t\n  v  \t\n  ")
		assert.Equal(t, Command("v"), command)
		assert.Nil(t, args)
	})
}

// TestParseCommand_RealWorldExamples tests realistic command inputs
func TestParseCommand_RealWorldExamples(t *testing.T) {
	t.Run("verbose toggle", func(t *testing.T) {
		command, args := parseCommand("v")
		assert.Equal(t, Command("v"), command)
		assert.Nil(t, args)
	})

	t.Run("set run pattern", func(t *testing.T) {
		command, args := parseCommand("r TestMyFunction")
		assert.Equal(t, Command("r"), command)
		assert.Equal(t, []string{"TestMyFunction"}, args)
	})

	t.Run("clear run pattern", func(t *testing.T) {
		command, args := parseCommand("r")
		assert.Equal(t, Command("r"), command)
		assert.Nil(t, args)
	})

	t.Run("set test path", func(t *testing.T) {
		command, args := parseCommand("p ./internal/server")
		assert.Equal(t, Command("p"), command)
		assert.Equal(t, []string{"./internal/server"}, args)
	})

	t.Run("clear screen", func(t *testing.T) {
		command, args := parseCommand("cls")
		assert.Equal(t, Command("cls"), command)
		assert.Nil(t, args)
	})

	t.Run("force run", func(t *testing.T) {
		command, args := parseCommand("f")
		assert.Equal(t, Command("f"), command)
		assert.Nil(t, args)
	})

	t.Run("show help", func(t *testing.T) {
		command, args := parseCommand("h")
		assert.Equal(t, Command("h"), command)
		assert.Nil(t, args)
	})

	t.Run("clear all", func(t *testing.T) {
		command, args := parseCommand("clear")
		assert.Equal(t, Command("clear"), command)
		assert.Nil(t, args)
	})
}

// ============================================================================
// readStdin Tests
// ============================================================================

// TestReadStdin_SendsHelpMessage tests that help command sends HelpMessage
func TestReadStdin_SendsHelpMessage(t *testing.T) {
	// Create mock stdin
	input := "h\n"
	mockStdin := strings.NewReader(input)

	// Create channels
	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 10)

	// Start readStdin with mock stdin
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

	// Signal ready
	readyChan <- true

	// Wait for message
	select {
	case msg := <-helpChan:
		assert.NotNil(t, msg, "should receive HelpMessage")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for HelpMessage")
	}

	// Should not receive CommandMessage
	select {
	case <-commandChan:
		t.Fatal("should not receive CommandMessage for help command")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

// TestReadStdin_SendsCommandMessage tests that regular commands send CommandMessage
func TestReadStdin_SendsCommandMessage(t *testing.T) {
	// Create mock stdin
	input := "v\n"
	mockStdin := strings.NewReader(input)

	// Create channels
	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 10)

	// Start readStdin with mock stdin
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

	// Signal ready
	readyChan <- true

	// Wait for message
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("v"), msg.Command, "should receive verbose command")
		assert.Nil(t, msg.Args, "should have no args")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for CommandMessage")
	}

	// Should not receive HelpMessage
	select {
	case <-helpChan:
		t.Fatal("should not receive HelpMessage for regular command")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

// TestReadStdin_CommandWithArgs tests command parsing with arguments
func TestReadStdin_CommandWithArgs(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCmd  Command
		expectedArgs []string
	}{
		{
			name:         "set run pattern",
			input:        "r TestFoo\n",
			expectedCmd:  Command("r"),
			expectedArgs: []string{"TestFoo"},
		},
		{
			name:         "set test path",
			input:        "p ./internal\n",
			expectedCmd:  Command("p"),
			expectedArgs: []string{"./internal"},
		},
		{
			name:         "command with multiple args",
			input:        "p /path arg2 arg3\n",
			expectedCmd:  Command("p"),
			expectedArgs: []string{"/path", "arg2", "arg3"},
		},
		{
			name:         "command without args",
			input:        "clear\n",
			expectedCmd:  Command("clear"),
			expectedArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStdin := strings.NewReader(tt.input)

			commandChan := make(chan CommandMessage, 10)
			helpChan := make(chan HelpMessage, 10)
			readyChan := make(chan bool, 10)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

			readyChan <- true

			select {
			case msg := <-commandChan:
				assert.Equal(t, tt.expectedCmd, msg.Command, "command should match")
				assert.Equal(t, tt.expectedArgs, msg.Args, "args should match")
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout waiting for CommandMessage")
			}
		})
	}
}

// TestReadStdin_IgnoresEmptyLines tests that empty lines are ignored
func TestReadStdin_IgnoresEmptyLines(t *testing.T) {
	// Create mock stdin with empty lines
	input := "\n\n  \n\t\nv\n"
	mockStdin := strings.NewReader(input)

	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

	readyChan <- true

	// Should only receive one message (the "v" command)
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("v"), msg.Command, "should receive only the non-empty command")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for CommandMessage")
	}

	// Should not receive any more messages
	select {
	case <-commandChan:
		t.Fatal("should not receive additional messages for empty lines")
	case <-time.After(50 * time.Millisecond):
		// Expected - no more messages
	}
}

// TestReadStdin_ReadyChannelBlocking tests that readyChan controls processing
func TestReadStdin_ReadyChannelBlocking(t *testing.T) {
	// Create mock stdin with multiple commands
	input := "v\nclear\n"
	mockStdin := strings.NewReader(input)

	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 1) // Buffered to prevent deadlock

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

	// Don't send ready signal yet - should not receive any messages
	select {
	case <-commandChan:
		t.Fatal("should not receive message when not ready")
	case <-time.After(50 * time.Millisecond):
		// Expected - blocked
	}

	// Now send ready signal
	readyChan <- true

	// Should receive first message
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("v"), msg.Command, "should receive first command")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for first CommandMessage")
	}

	// Second command should be processed automatically (still ready)
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("clear"), msg.Command, "should receive second command")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for second CommandMessage")
	}
}

// TestReadStdin_ReadyChannelBlocksAndUnblocks tests blocking and unblocking
func TestReadStdin_ReadyChannelBlocksAndUnblocks(t *testing.T) {
	// Use a pipe to control when input becomes available
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 1) // Buffered to prevent deadlock

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, pipeReader, commandChan, helpChan, readyChan)

	// Send true to start processing
	readyChan <- true

	// Write first command
	pipeWriter.Write([]byte("v\n"))

	// Send false immediately to ensure it's in the buffer before readStdin loops back
	readyChan <- false

	// Receive first command
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("v"), msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for first command")
	}

	// Give readStdin time to process the state change and enter blocking state
	time.Sleep(20 * time.Millisecond)

	// Write second command while blocked (in goroutine because pipe writes block until read)
	go pipeWriter.Write([]byte("clear\n"))

	// Should not receive second command while blocked
	select {
	case <-commandChan:
		t.Fatal("should not receive command while blocked")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message received
	}

	// Send true to resume processing
	readyChan <- true

	// Should now receive the second command
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("clear"), msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for second command after unblocking")
	}

	// Write third command
	pipeWriter.Write([]byte("f\n"))

	// Should receive third command (still unblocked)
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("f"), msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for third command")
	}
}

// TestReadStdin_MultipleCommands tests processing multiple commands
func TestReadStdin_MultipleCommands(t *testing.T) {
	input := "v\nr TestFoo\np .\nclear\nh\n"
	mockStdin := strings.NewReader(input)

	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readStdin(ctx, mockStdin, commandChan, helpChan, readyChan)

	readyChan <- true

	// Should receive 4 CommandMessages
	expectedCommands := []struct {
		cmd  Command
		args []string
	}{
		{Command("v"), nil},
		{Command("r"), []string{"TestFoo"}},
		{Command("p"), []string{"."}},
		{Command("clear"), nil},
	}

	for i, expected := range expectedCommands {
		select {
		case msg := <-commandChan:
			assert.Equal(t, expected.cmd, msg.Command, "command %d should match", i)
			assert.Equal(t, expected.args, msg.Args, "args %d should match", i)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for command %d", i)
		}
	}

	// Should receive 1 HelpMessage
	select {
	case msg := <-helpChan:
		assert.NotNil(t, msg, "should receive help message")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for help message")
	}
}

// TestReadStdin_ContextCancellation tests that context cancellation stops reading
func TestReadStdin_ContextCancellation(t *testing.T) {
	// Create infinite input stream
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 10)
	readyChan := make(chan bool, 10)

	ctx, cancel := context.WithCancel(context.Background())

	go readStdin(ctx, pipeReader, commandChan, helpChan, readyChan)

	readyChan <- true

	// Write a command
	pipeWriter.Write([]byte("v\n"))

	// Should receive it
	select {
	case msg := <-commandChan:
		assert.Equal(t, Command("v"), msg.Command)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for command")
	}

	// Cancel context
	cancel()

	// Give goroutine time to stop
	time.Sleep(50 * time.Millisecond)

	// Write another command
	pipeWriter.Write([]byte("clear\n"))

	// Should not receive it (goroutine should be stopped)
	select {
	case <-commandChan:
		t.Fatal("should not receive command after context cancellation")
	case <-time.After(100 * time.Millisecond):
		// Expected - goroutine stopped
	}
}
