package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StartSignalShutdown tests complete lifecycle: start, signal, verify clean shutdown
func TestIntegration_StartSignalShutdown(t *testing.T) {
	t.Skip("Cannot test actual signal handling within same process - signal would terminate the test")
	// Create context with signal handler
	ctx, cancel := setupSignalHandler()
	defer cancel()

	// Create config
	config := NewTestConfig()

	// Store config in context
	ctxWithConfig := withConfig(ctx, config)

	// Create channels
	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Start dispatcher
	dispatcherDone := make(chan struct{})
	go func() {
		dispatcher(ctxWithConfig, config, fileChangeChan, commandChan, helpChan, testCompleteChan)
		close(dispatcherDone)
	}()

	// Wait for dispatcher to start
	time.Sleep(50 * time.Millisecond)

	// Send SIGINT signal
	err := syscall.Kill(os.Getpid(), syscall.SIGINT)
	require.NoError(t, err, "should be able to send SIGINT")

	// Dispatcher should exit cleanly
	select {
	case <-dispatcherDone:
		// Expected - clean shutdown
	case <-time.After(1 * time.Second):
		t.Fatal("dispatcher should exit after SIGINT signal")
	}

	// Context should be cancelled
	select {
	case <-ctxWithConfig.Done():
		// Expected
	default:
		t.Fatal("context should be cancelled after shutdown")
	}
}

// TestIntegration_SignalDuringTestRun tests signal handling when test is running
func TestIntegration_SignalDuringTestRun(t *testing.T) {
	t.Skip("Cannot test actual signal handling within same process - signal would terminate the test")
	// Create context with signal handler
	ctx, cancel := setupSignalHandler()
	defer cancel()

	// Create config
	config := NewTestConfig()
	ctxWithConfig := withConfig(ctx, config)

	// Create channels
	fileChangeChan := make(chan FileChangeMessage, 1)
	commandChan := make(chan CommandMessage, 1)
	helpChan := make(chan HelpMessage, 1)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Start dispatcher
	dispatcherDone := make(chan struct{})
	go func() {
		dispatcher(ctxWithConfig, config, fileChangeChan, commandChan, helpChan, testCompleteChan)
		close(dispatcherDone)
	}()

	// Start a test
	fileChangeChan <- FileChangeMessage{}

	// Wait for test to start
	time.Sleep(50 * time.Millisecond)

	// Send signal while test is running
	err := syscall.Kill(os.Getpid(), syscall.SIGTERM)
	require.NoError(t, err, "should be able to send SIGTERM")

	// Dispatcher should NOT exit yet
	select {
	case <-dispatcherDone:
		t.Fatal("dispatcher should wait for test completion")
	case <-time.After(100 * time.Millisecond):
		// Expected - still waiting
	}

	// Complete the test
	testCompleteChan <- TestCompleteMessage{}

	// Now dispatcher should exit
	select {
	case <-dispatcherDone:
		// Expected - clean shutdown
	case <-time.After(500 * time.Millisecond):
		t.Fatal("dispatcher should exit after test completion")
	}
}

// TestIntegration_ContextChainPreservation tests that context chain is preserved through operations
func TestIntegration_ContextChainPreservation(t *testing.T) {
	// Create base context with custom value
	type customKey struct{}
	baseCtx := context.WithValue(context.Background(), customKey{}, "base value")

	// Create signal handler context as child
	signalCtx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	// Add config to context
	config := NewTestConfig()
	fullCtx := withConfig(signalCtx, config)

	// Verify all values are accessible
	assert.Equal(t, "base value", fullCtx.Value(customKey{}), "should preserve base context value")

	retrievedConfig := getConfig(fullCtx)
	assert.NotNil(t, retrievedConfig, "should retrieve config")
	assert.Equal(t, config, retrievedConfig, "should retrieve correct config")

	// Verify context cancellation propagates
	cancel()
	select {
	case <-fullCtx.Done():
		// Expected - cancellation propagated
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context cancellation should propagate")
	}
}

// TestIntegration_ConfigAccessFromContext tests accessing config from context in different goroutines
func TestIntegration_ConfigAccessFromContext(t *testing.T) {
	config := NewTestConfig()
	config.SetVerbose(true)
	config.SetRunPattern("TestIntegration")

	ctx := withConfig(context.Background(), config)

	// Launch multiple goroutines that access config from context
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			retrievedConfig := getConfig(ctx)
			assert.NotNil(t, retrievedConfig)
			assert.True(t, retrievedConfig.GetVerbose())
			assert.Equal(t, "TestIntegration", retrievedConfig.GetRunPattern())
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Expected
		case <-time.After(1 * time.Second):
			t.Fatal("goroutine timed out")
		}
	}
}
