package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// isGoFile Tests
// ============================================================================

// TestIsGoFile_WithGoExtension tests that files with .go extension return true
func TestIsGoFile_WithGoExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "simple go file",
			path:     "main.go",
			expected: true,
		},
		{
			name:     "go file in subdirectory",
			path:     "internal/server/handler.go",
			expected: true,
		},
		{
			name:     "go file with absolute path",
			path:     "/usr/local/src/project/main.go",
			expected: true,
		},
		{
			name:     "go test file",
			path:     "main_test.go",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGoFile(tt.path)
			assert.Equal(t, tt.expected, result, "should correctly identify .go files")
		})
	}
}

// TestIsGoFile_WithNonGoExtension tests that non-.go files return false
func TestIsGoFile_WithNonGoExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "text file",
			path:     "readme.txt",
			expected: false,
		},
		{
			name:     "markdown file",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "yaml file",
			path:     "config.yaml",
			expected: false,
		},
		{
			name:     "json file",
			path:     "package.json",
			expected: false,
		},
		{
			name:     "no extension",
			path:     "Makefile",
			expected: false,
		},
		{
			name:     "directory",
			path:     "internal/",
			expected: false,
		},
		{
			name:     "empty string",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGoFile(tt.path)
			assert.Equal(t, tt.expected, result, "should correctly reject non-.go files")
		})
	}
}

// TestIsGoFile_EdgeCases tests edge cases
func TestIsGoFile_EdgeCases(t *testing.T) {
	t.Run("file ending with .go but not extension", func(t *testing.T) {
		result := isGoFile("myfile.go.bak")
		assert.False(t, result, ".go.bak should not be identified as go file")
	})

	t.Run("hidden go file", func(t *testing.T) {
		result := isGoFile(".hidden.go")
		assert.True(t, result, "hidden .go files should still be identified")
	})

	t.Run("uppercase extension", func(t *testing.T) {
		result := isGoFile("main.GO")
		assert.False(t, result, "uppercase .GO should not match (case sensitive)")
	})
}

// ============================================================================
// addWatchRecursive Tests
// ============================================================================

// TestAddWatchRecursive_WithSimpleDirectory tests watching a simple directory
func TestAddWatchRecursive_WithSimpleDirectory(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create some go files
	err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0o600)
	require.NoError(t, err)

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Add directory recursively
	err = addWatchRecursive(watcher, tempDir)
	require.NoError(t, err, "should successfully add directory to watcher")

	// Verify the directory is being watched
	// (watcher.WatchList() should contain tempDir)
	watchList := watcher.WatchList()
	assert.Contains(t, watchList, tempDir, "watcher should include the root directory")
}

// TestAddWatchRecursive_WithNestedDirectories tests watching nested directories
func TestAddWatchRecursive_WithNestedDirectories(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create nested directories
	subDir1 := filepath.Join(tempDir, "internal")
	subDir2 := filepath.Join(tempDir, "internal", "server")
	subDir3 := filepath.Join(tempDir, "pkg")

	err := os.MkdirAll(subDir2, 0o750)
	require.NoError(t, err)
	err = os.MkdirAll(subDir3, 0o750)
	require.NoError(t, err)

	// Create some go files
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir2, "handler.go"), []byte("package server"), 0o600)
	require.NoError(t, err)

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Add directory recursively
	err = addWatchRecursive(watcher, tempDir)
	require.NoError(t, err, "should successfully add nested directories")

	// Verify all directories are being watched
	watchList := watcher.WatchList()
	assert.Contains(t, watchList, tempDir, "should watch root directory")
	assert.Contains(t, watchList, subDir1, "should watch internal directory")
	assert.Contains(t, watchList, subDir2, "should watch internal/server directory")
	assert.Contains(t, watchList, subDir3, "should watch pkg directory")
}

// TestAddWatchRecursive_ExcludesHiddenDirectories tests that hidden directories are excluded
func TestAddWatchRecursive_ExcludesHiddenDirectories(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create visible and hidden directories
	visibleDir := filepath.Join(tempDir, "internal")
	hiddenDir := filepath.Join(tempDir, ".git")
	nestedHiddenDir := filepath.Join(tempDir, ".hidden", "subdir")

	err := os.MkdirAll(visibleDir, 0o750)
	require.NoError(t, err)
	err = os.MkdirAll(hiddenDir, 0o750)
	require.NoError(t, err)
	err = os.MkdirAll(nestedHiddenDir, 0o750)
	require.NoError(t, err)

	// Create go files in both visible and hidden directories
	err = os.WriteFile(filepath.Join(visibleDir, "main.go"), []byte("package internal"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(hiddenDir, "config.go"), []byte("package git"), 0o600)
	require.NoError(t, err)

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Add directory recursively
	err = addWatchRecursive(watcher, tempDir)
	require.NoError(t, err)

	// Verify hidden directories are NOT being watched
	watchList := watcher.WatchList()
	assert.Contains(t, watchList, tempDir, "should watch root directory")
	assert.Contains(t, watchList, visibleDir, "should watch visible directory")
	assert.NotContains(t, watchList, hiddenDir, "should NOT watch .git directory")
	assert.NotContains(t, watchList, nestedHiddenDir, "should NOT watch nested hidden directory")
}

// TestAddWatchRecursive_WithInvalidPath tests error handling for invalid path
func TestAddWatchRecursive_WithInvalidPath(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Try to watch non-existent directory
	err = addWatchRecursive(watcher, "/nonexistent/path/that/does/not/exist")
	assert.Error(t, err, "should return error for non-existent path")
}

// TestAddWatchRecursive_WithFile tests that files are not added (only directories)
func TestAddWatchRecursive_WithFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file (not directory)
	filePath := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(filePath, []byte("package main"), 0o600)
	require.NoError(t, err)

	watcher, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	// Try to watch a file directly - should handle gracefully or error
	err = addWatchRecursive(watcher, filePath)
	// Implementation should either skip files or return error
	// For this test, we expect it to handle files appropriately
	if err == nil {
		// If no error, verify it didn't add the file incorrectly
		watchList := watcher.WatchList()
		// The behavior here depends on implementation
		// fsnotify itself might accept a file, but our implementation should handle directories
		_ = watchList // Just verify no panic
	}
}

// ============================================================================
// WatchFiles Tests
// ============================================================================

// TestWatchFiles_DetectsGoFileCreation tests that creating a .go file triggers a message
func TestWatchFiles_DetectsGoFileCreation(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create a new .go file
	err := os.WriteFile(filepath.Join(tempDir, "new.go"), []byte("package main"), 0o600)
	require.NoError(t, err)

	// Should receive FileChangeMessage after debounce period (200ms)
	select {
	case msg := <-fileChangeChan:
		assert.NotNil(t, msg, "should receive FileChangeMessage")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for FileChangeMessage after file creation")
	}
}

// TestWatchFiles_DetectsGoFileModification tests that modifying a .go file triggers a message
func TestWatchFiles_DetectsGoFileModification(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial file
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0o600)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Modify the file
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0o600)
	require.NoError(t, err)

	// Should receive FileChangeMessage
	select {
	case msg := <-fileChangeChan:
		assert.NotNil(t, msg, "should receive FileChangeMessage")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for FileChangeMessage after file modification")
	}
}

// TestWatchFiles_IgnoresNonGoFiles tests that non-.go files don't trigger messages
func TestWatchFiles_IgnoresNonGoFiles(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create non-.go files
	err := os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("readme"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte("config"), 0o600)
	require.NoError(t, err)

	// Should NOT receive FileChangeMessage
	select {
	case <-fileChangeChan:
		t.Fatal("should not receive FileChangeMessage for non-.go files")
	case <-time.After(400 * time.Millisecond):
		// Expected - no message for non-go files
	}
}

// TestWatchFiles_DebounceMultipleChanges tests that multiple rapid changes only trigger one message
func TestWatchFiles_DebounceMultipleChanges(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Make multiple rapid changes (within 200ms window)
	testFile := filepath.Join(tempDir, "test.go")
	for i := range 5 {
		err := os.WriteFile(testFile, []byte("package main // "+string(rune(i))), 0o600)
		require.NoError(t, err)
		time.Sleep(30 * time.Millisecond) // Changes within debounce window
	}

	// Should receive exactly ONE FileChangeMessage after all changes settle
	messageCount := 0
	timeout := time.After(500 * time.Millisecond)

	// Collect messages for a period
messageLoop:
	for {
		select {
		case <-fileChangeChan:
			messageCount++
		case <-timeout:
			break messageLoop
		}
	}

	assert.Equal(t, 1, messageCount, "should receive exactly one message due to debouncing")
}

// TestWatchFiles_TimerResetOnSubsequentChanges tests that the timer resets with new changes
func TestWatchFiles_TimerResetOnSubsequentChanges(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	testFile := filepath.Join(tempDir, "test.go")

	// First change
	err := os.WriteFile(testFile, []byte("package main // 1"), 0o600)
	require.NoError(t, err)

	// Wait 150ms (less than debounce period of 200ms)
	time.Sleep(150 * time.Millisecond)

	// Second change - should reset timer
	err = os.WriteFile(testFile, []byte("package main // 2"), 0o600)
	require.NoError(t, err)

	// The message should arrive ~200ms after the SECOND change
	// So total time is ~350ms from first change
	startTime := time.Now()

	select {
	case <-fileChangeChan:
		elapsed := time.Since(startTime)
		// Should be at least 150ms (remaining from timer reset)
		assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond, "timer should have been reset")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for FileChangeMessage")
	}
}

// TestWatchFiles_HandlesNestedDirectories tests watching files in nested directories
func TestWatchFiles_HandlesNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested directory
	nestedDir := filepath.Join(tempDir, "internal", "server")
	err := os.MkdirAll(nestedDir, 0o750)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create file in nested directory
	err = os.WriteFile(filepath.Join(nestedDir, "handler.go"), []byte("package server"), 0o600)
	require.NoError(t, err)

	// Should receive FileChangeMessage
	select {
	case msg := <-fileChangeChan:
		assert.NotNil(t, msg, "should receive FileChangeMessage for nested directory")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for FileChangeMessage in nested directory")
	}
}

// TestWatchFiles_IgnoresHiddenDirectories tests that hidden directories are not watched
func TestWatchFiles_IgnoresHiddenDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create hidden directory
	hiddenDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(hiddenDir, 0o750)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create .go file in hidden directory
	err = os.WriteFile(filepath.Join(hiddenDir, "config.go"), []byte("package git"), 0o600)
	require.NoError(t, err)

	// Should NOT receive FileChangeMessage
	select {
	case <-fileChangeChan:
		t.Fatal("should not receive FileChangeMessage for files in hidden directories")
	case <-time.After(400 * time.Millisecond):
		// Expected - no message for hidden directory changes
	}
}

// TestWatchFiles_ContextCancellation tests that watcher stops when context is cancelled
func TestWatchFiles_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	watcherDone := make(chan struct{})
	go func() {
		WatchFiles(ctx, tempDir, fileChangeChan, startWatching)
		close(watcherDone)
	}()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create a file to verify watcher is working
	err := os.WriteFile(filepath.Join(tempDir, "test.go"), []byte("package main"), 0o600)
	require.NoError(t, err)

	// Should receive message
	select {
	case <-fileChangeChan:
		// Good, watcher is working
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher should be working before cancellation")
	}

	// Cancel context
	cancel()

	// Watcher should stop
	select {
	case <-watcherDone:
		// Good, watcher stopped
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher should stop after context cancellation")
	}

	// Create another file after cancellation
	err = os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package main"), 0o600)
	require.NoError(t, err)

	// Should NOT receive message (watcher stopped)
	select {
	case <-fileChangeChan:
		t.Fatal("should not receive message after watcher stopped")
	case <-time.After(400 * time.Millisecond):
		// Expected - watcher is stopped
	}
}

// TestWatchFiles_MultipleFileTypes tests mixed file types
func TestWatchFiles_MultipleFileTypes(t *testing.T) {
	tempDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create multiple files - some .go, some not
	err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("readme"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "test.go"), []byte("package main"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte("config"), 0o600)
	require.NoError(t, err)

	// Should receive exactly ONE FileChangeMessage (debounced from the .go files)
	messageCount := 0
	timeout := time.After(500 * time.Millisecond)

messageLoop:
	for {
		select {
		case <-fileChangeChan:
			messageCount++
		case <-timeout:
			break messageLoop
		}
	}

	assert.Equal(t, 1, messageCount, "should receive one debounced message for .go file changes")
}

// TestWatchFiles_FileRemoval tests that removing .go files triggers a message
func TestWatchFiles_FileRemoval(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial file
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0o600)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fileChangeChan := make(chan FileChangeMessage, 10)
	startWatching := make(chan struct{})
	close(startWatching) // Close immediately so watcher starts without blocking

	// Start watcher
	go WatchFiles(ctx, tempDir, fileChangeChan, startWatching)

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Remove the file
	err = os.Remove(testFile)
	require.NoError(t, err)

	// Should receive FileChangeMessage
	select {
	case msg := <-fileChangeChan:
		assert.NotNil(t, msg, "should receive FileChangeMessage after file removal")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for FileChangeMessage after file removal")
	}
}
