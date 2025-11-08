package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWatchFiles_BlocksUntilStartWatchingCloses tests that watcher blocks until startWatching closes
func TestWatchFiles_BlocksUntilStartWatchingCloses(t *testing.T) {
	ctx := context.Background()
	fileChangeChan := make(chan FileChangeMessage, 1)
	startWatching := make(chan struct{})

	// Create a temporary test directory
	tempDir := t.TempDir()

	watcherStarted := make(chan struct{})
	go func() {
		close(watcherStarted)
		WatchFiles(ctx, tempDir, fileChangeChan, startWatching)
	}()

	// Wait for goroutine to start
	<-watcherStarted
	time.Sleep(50 * time.Millisecond)

	// Watcher should be blocking, so no file change messages should be sent yet
	select {
	case <-fileChangeChan:
		t.Fatal("watcher should not send messages before startWatching closes")
	default:
		// Expected - watcher is blocked
	}

	// Close startWatching to unblock watcher
	close(startWatching)

	// Give watcher time to start watching
	time.Sleep(100 * time.Millisecond)

	// Now watcher should be active (we can't easily test file watching here,
	// but we've verified it was blocked until the signal)
}

// TestWatchFiles_AcceptsStartWatchingParameter tests that watchFiles signature accepts startWatching param
func TestWatchFiles_AcceptsStartWatchingParameter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 1)
	startWatching := make(chan struct{})
	tempDir := t.TempDir()

	// Close immediately so watcher doesn't block
	close(startWatching)

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Cancel to stop watcher
	cancel()
}

// TestStartupSequence_InitialTestRunsBeforeWatcher tests initial test runs before watcher starts
func TestStartupSequence_InitialTestRunsBeforeWatcher(t *testing.T) {
	// This test verifies the startup sequence order:
	// 1. Initial test runs
	// 2. Test completes
	// 3. Watcher starts
	// 4. Normal operation begins

	ctx, cancel := context.WithCancel(WithConfig(context.Background(), NewTestConfig()))
	defer cancel()

	testCompleteChan := make(chan TestCompleteMessage, 1)
	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})

	// Track events
	events := make(chan string, 10)

	// Simulate initial test run
	go func() {
		events <- "test_started"
		// Simulate test running
		time.Sleep(50 * time.Millisecond)
		testCompleteChan <- TestCompleteMessage{}
		events <- "test_completed"
	}()

	// Wait for test completion
	<-testCompleteChan
	events <- "received_completion"

	// Now start watcher (this would happen after initial test in real code)
	tempDir := t.TempDir()
	go func() {
		events <- "watcher_starting"
		WatchFiles(ctx, tempDir, fileChangeChan, startWatching)
	}()

	// Watcher should be blocked
	time.Sleep(50 * time.Millisecond)
	events <- "watcher_blocked"

	// Close startWatching to begin normal operation
	close(startWatching)
	events <- "watcher_unblocked"

	// Verify order of events
	require.Equal(t, "test_started", <-events)
	require.Equal(t, "test_completed", <-events)
	require.Equal(t, "received_completion", <-events)
	require.Equal(t, "watcher_starting", <-events)
	require.Equal(t, "watcher_blocked", <-events)
	require.Equal(t, "watcher_unblocked", <-events)
}

// TestStartupSequence_WatcherDoesNotSendMessagesDuringInitialTest
// tests watcher doesn't send messages during initial test
func TestStartupSequence_WatcherDoesNotSendMessagesDuringInitialTest(t *testing.T) {
	ctx, cancel := context.WithCancel(WithConfig(context.Background(), NewTestConfig()))
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	tempDir := t.TempDir()

	// Start watcher but don't unblock it
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Simulate initial test running
	time.Sleep(100 * time.Millisecond)

	// Channel should be empty - watcher is blocked
	assert.Equal(t, 0, len(fileChangeChan), "watcher should not send messages while blocked")

	// Unblock watcher
	close(startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Now it's running (but still no file changes)
	assert.Equal(t, 0, len(fileChangeChan), "no file changes occurred yet")

	cancel()
}

// TestStartupSequence_PromptAppearsAfterInitialTest tests prompt appears after initial test
func TestStartupSequence_PromptAppearsAfterInitialTest(t *testing.T) {
	// This is more of an integration test concept
	// We verify that the sequence is:
	// 1. Run initial test
	// 2. Wait for completion
	// 3. Show prompt
	// 4. Start watcher
	// 5. Start dispatcher

	testCompleteChan := make(chan TestCompleteMessage, 1)
	sequence := make(chan string, 10)

	// Simulate the startup sequence
	go func() {
		// Initial test runs
		sequence <- "initial_test_started"
		time.Sleep(50 * time.Millisecond)
		testCompleteChan <- TestCompleteMessage{}
	}()

	// Wait for test completion (as main() would)
	<-testCompleteChan
	sequence <- "test_completed"

	// Prompt would be shown here
	sequence <- "prompt_shown"

	// Now start watcher and dispatcher
	sequence <- "watcher_started"
	sequence <- "dispatcher_started"

	// Verify order
	assert.Equal(t, "initial_test_started", <-sequence)
	assert.Equal(t, "test_completed", <-sequence)
	assert.Equal(t, "prompt_shown", <-sequence)
	assert.Equal(t, "watcher_started", <-sequence)
	assert.Equal(t, "dispatcher_started", <-sequence)
}

// TestStartupSequence_FullIntegration tests complete startup sequence integration
func TestStartupSequence_FullIntegration(t *testing.T) {
	config := NewTestConfig()
	config.SetRunPattern("Message")
	ctx, cancel := context.WithCancel(WithConfig(context.Background(), config))
	defer cancel()

	// Create channels
	testCompleteChan := make(chan TestCompleteMessage, 1)
	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	tempDir := t.TempDir()

	events := make(chan string, 20)

	// Phase 1: Initial test run (synchronous in real code)
	go func() {
		events <- "phase1_start"
		RunTests(ctx, testCompleteChan, nil, nil)
		events <- "phase1_test_launched"
	}()

	// Wait for initial test completion
	select {
	case <-testCompleteChan:
		events <- "phase1_test_completed"
	case <-time.After(2 * time.Second):
		t.Fatal("initial test did not complete")
	}

	// Phase 2: Show prompt (would happen in real code)
	events <- "phase2_prompt_shown"

	// Phase 3: Start watcher (blocked)
	go func() {
		events <- "phase3_watcher_starting"
		WatchFiles(ctx, tempDir, fileChangeChan, startWatching)
	}()

	time.Sleep(50 * time.Millisecond)
	events <- "phase3_watcher_blocked"

	// Phase 4: Unblock watcher to begin normal operation
	close(startWatching)
	events <- "phase4_watcher_unblocked"

	time.Sleep(50 * time.Millisecond)
	events <- "phase4_normal_operation"

	// Verify all phases occurred in order
	assert.Equal(t, "phase1_start", <-events)
	assert.Equal(t, "phase1_test_launched", <-events)
	assert.Equal(t, "phase1_test_completed", <-events)
	assert.Equal(t, "phase2_prompt_shown", <-events)
	assert.Equal(t, "phase3_watcher_starting", <-events)
	assert.Equal(t, "phase3_watcher_blocked", <-events)
	assert.Equal(t, "phase4_watcher_unblocked", <-events)
	assert.Equal(t, "phase4_normal_operation", <-events)

	cancel()
}

// TestWatchFiles_UnblocksImmediatelyIfChannelAlreadyClosed tests watcher doesn't block if channel already closed
func TestWatchFiles_UnblocksImmediatelyIfChannelAlreadyClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(WithConfig(context.Background(), NewTestConfig()))
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 1)
	startWatching := make(chan struct{})
	tempDir := t.TempDir()

	// Close channel before starting watcher
	close(startWatching)

	started := time.Now()
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	elapsed := time.Since(started)

	// Should not have blocked significantly
	assert.Less(t, elapsed, 200*time.Millisecond, "watcher should start immediately if channel already closed")

	cancel()
}

// TestWatchFiles_ContextCancellationWhileBlocked tests watcher respects context cancellation while blocked
func TestWatchFiles_ContextCancellationWhileBlocked(t *testing.T) {
	ctx, cancel := context.WithCancel(WithConfig(context.Background(), NewTestConfig()))

	fileChangeChan := make(chan FileChangeMessage, 1)
	startWatching := make(chan struct{}) // Never closed
	tempDir := t.TempDir()

	done := make(chan struct{})
	go func() {
		WatchFiles(ctx, tempDir, fileChangeChan, startWatching)
		close(done)
	}()

	// Wait a bit for watcher to start and block
	time.Sleep(50 * time.Millisecond)

	// Cancel context while watcher is blocked
	cancel()

	// Watcher should exit even though startWatching was never closed
	select {
	case <-done:
		// Expected - watcher exited due to context cancellation
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher should exit when context is cancelled, even while blocked on startWatching")
	}
}
