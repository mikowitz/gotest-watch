package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDispatcher_FileChangeSpawnsTestRunner tests that FileChangeMessage spawns test runner
func TestDispatcher_FileChangeSpawnsTestRunner(t *testing.T) {
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
	readyChan := make(chan bool, 1)

	// Start dispatcher in background
	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Send file change message
	fileChangeChan <- FileChangeMessage{}

	// Should receive false on readyChan (stdin paused)
	select {
	case ready := <-readyChan:
		assert.False(t, ready, "readyChan should receive false when test starts")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("readyChan should receive value when test starts")
	}

	// Simulate test completion
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true on readyChan (stdin resumed)
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "readyChan should receive true when test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("readyChan should receive value when test completes")
	}

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
	readyChan := make(chan bool, 10)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Start first test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	select {
	case <-readyChan:
		// Test started
	case <-time.After(100 * time.Millisecond):
		t.Fatal("first test should start")
	}

	// Send another file change while test is running
	fileChangeChan <- FileChangeMessage{}

	// Should NOT receive another readyChan message (test already running)
	select {
	case <-readyChan:
		t.Fatal("should not start second test while first is running")
	case <-time.After(100 * time.Millisecond):
		// Correct - no second test started
	}

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true when test completes
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "should receive true when test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should receive completion signal")
	}

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
	readyChan := make(chan bool, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Send verbose command
	commandChan <- CommandMessage{Command: VerboseCmd, Args: nil}

	// Give time for command to execute
	time.Sleep(50 * time.Millisecond)

	// Verbose should have been toggled
	assert.True(t, config.Verbose, "verbose command should have been executed")

	cancel()
}

// TestDispatcher_CommandMessageSpawnsTestRunner tests that CommandMessage spawns test runner
func TestDispatcher_CommandMessageSpawnsTestRunner(t *testing.T) {
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
	readyChan := make(chan bool, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Send force run command
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Should receive false on readyChan (stdin paused)
	select {
	case ready := <-readyChan:
		assert.False(t, ready, "readyChan should receive false when test starts")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("readyChan should receive value when test starts")
	}

	// Simulate test completion
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true on readyChan (stdin resumed)
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "readyChan should receive true when test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("readyChan should receive value when test completes")
	}

	cancel()
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
	readyChan := make(chan bool, 10)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Start first test
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Wait for test to start
	select {
	case <-readyChan:
		// Test started
	case <-time.After(100 * time.Millisecond):
		t.Fatal("first test should start")
	}

	// Send another command while test is running
	commandChan <- CommandMessage{Command: ForceRunCmd, Args: nil}

	// Should NOT receive another readyChan message (test already running)
	select {
	case <-readyChan:
		t.Fatal("should not start second test while first is running")
	case <-time.After(100 * time.Millisecond):
		// Correct - no second test started
	}

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true when test completes
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "should receive true when test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should receive completion signal")
	}

	cancel()
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
	readyChan := make(chan bool, 1)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Send help message
	helpChan <- HelpMessage{}

	// Should NOT receive anything on readyChan (no test started)
	select {
	case <-readyChan:
		t.Fatal("help command should not start test runner")
	case <-time.After(100 * time.Millisecond):
		// Correct - no test started
	}

	cancel()
}

// TestDispatcher_TestCompleteMessageUpdatesStateAndReEnablesStdin tests TestCompleteMessage updates state and re-enables stdin
func TestDispatcher_TestCompleteMessageUpdatesStateAndReEnablesStdin(t *testing.T) {
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
	readyChan := make(chan bool, 10)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Start a test
	fileChangeChan <- FileChangeMessage{}

	// Drain the false signal
	<-readyChan

	// Send completion
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true (stdin re-enabled)
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "stdin should be re-enabled after test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should receive completion signal")
	}

	// Now another file change should be able to start a new test
	fileChangeChan <- FileChangeMessage{}

	select {
	case ready := <-readyChan:
		assert.False(t, ready, "new test should be able to start after previous completion")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should be able to start new test")
	}

	cancel()
}

// TestDispatcher_ContextDoneWaitsForTestsToFinish tests ctx.Done() waits for tests to finish
func TestDispatcher_ContextDoneWaitsForTestsToFinish(t *testing.T) {
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
	readyChan := make(chan bool, 10)

	done := make(chan struct{})
	go func() {
		dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)
		close(done)
	}()

	// Drain initial ready signal
	<-readyChan

	// Start a test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	<-readyChan

	// Cancel context while test is running
	cancel()

	// Dispatcher should NOT exit yet
	select {
	case <-done:
		t.Fatal("dispatcher should wait for test to complete before exiting")
	case <-time.After(100 * time.Millisecond):
		// Correct - still waiting
	}

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Now dispatcher should exit
	select {
	case <-done:
		// Correct - dispatcher exited after test completed
	case <-time.After(200 * time.Millisecond):
		t.Fatal("dispatcher should exit after test completes on shutdown")
	}
}

// TestDispatcher_ReadyChannelReceivesCorrectValues tests ready channel receives correct values
func TestDispatcher_ReadyChannelReceivesCorrectValues(t *testing.T) {
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
	readyChan := make(chan bool, 10)

	go dispatcher(ctx, config, fileChangeChan, commandChan, helpChan, testCompleteChan, readyChan)

	// Drain initial ready signal
	<-readyChan

	// Start test
	fileChangeChan <- FileChangeMessage{}

	// Should receive false (test starting, stdin paused)
	select {
	case ready := <-readyChan:
		assert.False(t, ready, "should receive false when test starts")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should receive signal when test starts")
	}

	// Complete test
	testCompleteChan <- TestCompleteMessage{}

	// Should receive true (test completed, stdin resumed)
	select {
	case ready := <-readyChan:
		assert.True(t, ready, "should receive true when test completes")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should receive signal when test completes")
	}

	cancel()
}
