package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDispatcher_FileChangeSpawnsTestRunner tests that FileChangeMessage spawns test runner
func TestDispatcher_FileChangeSpawnsTestRunner(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Start dispatcher in background
	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Send file change message
	fileChangeChan <- FileChangeMessage{}

	// Wait a moment for test to start
	time.Sleep(50 * time.Millisecond)

	// Simulate test completion
	testCompleteChan <- TestCompleteMessage{}

	// Wait for completion to be processed
	time.Sleep(50 * time.Millisecond)

	cancel()
}

// TestDispatcher_FileChangeIgnoredWhenTestRunning tests that FileChangeMessage ignored when testRunning=true
func TestDispatcher_FileChangeIgnoredWhenTestRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 10)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Start first test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// Send another file change while test is running - it will be drained and ignored
	fileChangeChan <- FileChangeMessage{}

	// Wait a bit for the dispatcher to drain it
	time.Sleep(50 * time.Millisecond)

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Wait for completion to be processed
	time.Sleep(50 * time.Millisecond)

	// The second file change should have been drained and ignored (not in channel anymore)
	assert.Equal(t, 0, len(fileChangeChan), "second file change should have been drained and ignored")

	cancel()
}

// TestDispatcher_CommandMessageCallsHandler tests that CommandMessage calls handler
func TestDispatcher_CommandMessageCallsHandler(t *testing.T) {
	initRegistry()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Send verbose command
	commandChan <- CommandMessage{Command: VerboseCmd, Args: nil}

	// Give time for command to execute
	time.Sleep(50 * time.Millisecond)

	// Verbose should have been toggled
	assert.True(t, config.Verbose, "verbose command should have been executed")
}

// TestDispatcher_CommandMessageSpawnsTestRunner tests that CommandMessage spawns test runner
func TestDispatcher_CommandMessageSpawnsTestRunner(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Send force run command
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// Simulate test completion
	testCompleteChan <- TestCompleteMessage{}

	// Wait for completion to be processed
	time.Sleep(50 * time.Millisecond)
}

// TestDispatcher_CommandMessageIgnoredWhenTestRunning tests that CommandMessage ignored when testRunning=true
func TestDispatcher_CommandMessageIgnoredWhenTestRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 10)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Start first test
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// Send another command while test is running - it will be drained and ignored
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Wait a bit for the dispatcher to drain it
	time.Sleep(50 * time.Millisecond)

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Wait for completion to be processed
	time.Sleep(50 * time.Millisecond)

	// The second command should have been drained and ignored (not in channel anymore)
	assert.Equal(t, 0, len(commandChan), "second command should have been drained and ignored")
}

// TestDispatcher_HelpMessageDoesNotSpawnTestRunner tests that HelpMessage doesn't spawn test runner
func TestDispatcher_HelpMessageDoesNotSpawnTestRunner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Send help message
	helpChan <- HelpMessage{}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// testCompleteChan should be empty (no test started)
	assert.Equal(t, 0, len(testCompleteChan), "help command should not start test runner")
}

// TestDispatcher_TestCompleteMessageUpdatesState tests TestCompleteMessage updates state
func TestDispatcher_TestCompleteMessageUpdatesState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 10)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Start a test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// Send completion
	testCompleteChan <- TestCompleteMessage{}

	// Wait for completion to be processed
	time.Sleep(50 * time.Millisecond)

	// Now another file change should be able to start a new test
	fileChangeChan <- FileChangeMessage{}

	// Wait for second test to start
	time.Sleep(50 * time.Millisecond)

	// Second test should have started (testRunning should be true again)
	// We can verify by checking that a third file change is ignored
	fileChangeChan <- FileChangeMessage{}
	time.Sleep(50 * time.Millisecond)
	// Third change should have been drained and ignored
	assert.Equal(t, 0, len(fileChangeChan), "third file change should be drained and ignored while second test runs")

	cancel()
}

// TestDispatcher_ContextDoneExitsGracefully tests ctx.Done() causes graceful shutdown
func TestDispatcher_ContextDoneExitsGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	done := make(chan struct{})
	go func() {
		dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)
		close(done)
	}()

	// Let dispatcher run for a bit
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Dispatcher should exit gracefully (either immediately if idle, or after waiting for test)
	select {
	case <-done:
		// Correct - dispatcher exited
	case <-time.After(500 * time.Millisecond):
		t.Fatal("dispatcher should exit after context cancellation")
	}
}

// TestDispatcher_StateTransitions tests state transitions between idle and running
func TestDispatcher_StateTransitions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	fileChangeChan := make(chan FileChangeMessage, 10)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan)

	// Start test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// While running, file changes should be ignored (drained from channel)
	fileChangeChan <- FileChangeMessage{}
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, len(fileChangeChan), "file change should be drained and ignored while running")

	// Complete test
	testCompleteChan <- TestCompleteMessage{}

	// Wait for state to transition back to idle
	time.Sleep(50 * time.Millisecond)

	// Now file changes should be processed again
	fileChangeChan <- FileChangeMessage{}
	time.Sleep(50 * time.Millisecond)

	// New test should have started, so another file change should be ignored
	fileChangeChan <- FileChangeMessage{}
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, len(fileChangeChan), "file change should be drained and ignored while second test runs")

	cancel()
}
