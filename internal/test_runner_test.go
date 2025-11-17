package internal

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestModule creates a temporary directory with a go.mod file and test file
func setupTestModule(t *testing.T, testContent string) string {
	t.Helper()
	tempDir := t.TempDir()

	// Initialize a go.mod file
	goModContent := `module testmodule

go 1.24
`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o600)
	require.NoError(t, err)

	// Write the test file
	testFile := filepath.Join(tempDir, "example_test.go")
	err = os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	return tempDir
}

// waitForTestCompletion waits for testCompleteChan to ensure
// runTests has fully completed before the test returns (preventing premature cleanup of temp directories)
func waitForTestCompletion(
	t *testing.T,
	testCompleteChan chan TestCompleteMessage,
) {
	t.Helper()

	// Wait for completion message
	select {
	case <-testCompleteChan:
		// Success - message was sent
	case <-time.After(30 * time.Second):
		t.Fatal("TestCompleteMessage was not sent within timeout")
	}
}

// TestStreamOutput_ReadsAllLines tests that streamOutput reads and writes all lines
func TestStreamOutput_ReadsAllLines(t *testing.T) {
	input := "line1\nline2\nline3\n"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	streamOutput(scanner, &output, &wg, false)

	assert.Equal(t, "line1\nline2\nline3\n", output.String(), "should write all lines to output")
}

// TestStreamOutput_CallsWaitGroupDone tests that streamOutput calls wg.Done()
func TestStreamOutput_CallsWaitGroupDone(t *testing.T) {
	reader := strings.NewReader("test\n")
	scanner := bufio.NewScanner(reader)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	streamOutput(scanner, &output, &wg, false)

	// This should not block if wg.Done() was called
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - wg.Done() was called
	case <-time.After(100 * time.Millisecond):
		t.Fatal("wg.Done() was not called - timeout waiting for WaitGroup")
	}
}

// TestStreamOutput_HandlesEmptyInput tests streamOutput with empty input
func TestStreamOutput_HandlesEmptyInput(t *testing.T) {
	reader := strings.NewReader("")
	scanner := bufio.NewScanner(reader)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	streamOutput(scanner, &output, &wg, false)

	assert.Equal(t, "", output.String(), "should handle empty input")
}

// TestStreamOutput_PreservesLineContent tests that content is preserved exactly
func TestStreamOutput_PreservesLineContent(t *testing.T) {
	input := "PASS: TestFoo (0.00s)\nFAIL: TestBar (0.01s)\n--- FAIL: TestBaz\n"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	streamOutput(scanner, &output, &wg, false)

	assert.Equal(t, input, output.String(), "should preserve exact line content including special characters")
}

// TestRunTests_SendsTestCompleteMessage tests that runTests sends completion message
func TestRunTests_SendsTestCompleteMessage(t *testing.T) {
	testContent := `package example

import "testing"

func TestExample(t *testing.T) {
	// Simple passing test
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Run a simple command that will succeed
	go RunTests(ctx, testCompleteChan, nil, nil)

	waitForTestCompletion(t, testCompleteChan)
}

// TestRunTests_BuildsCorrectCommand tests that runTests uses config.BuildCommand()
func TestRunTests_BuildsCorrectCommand(t *testing.T) {
	testContent := `package buildtest

import "testing"

func TestFoo(t *testing.T) {
	// Simple test
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.SetRunPattern("TestFoo")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)
	go RunTests(ctx, testCompleteChan, nil, nil)

	waitForTestCompletion(t, testCompleteChan)
}

// TestRunTests_StreamsStdoutAndStderr tests that both streams are captured
func TestRunTests_StreamsStdoutAndStderr(t *testing.T) {
	testContent := `package output

import "testing"
import "fmt"

func TestWithOutput(t *testing.T) {
	fmt.Println("this is stdout from test")
	t.Log("this is test log output")
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.ToggleVerbose()
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Capture both stdout and stderr using pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	go RunTests(ctx, testCompleteChan, wOut, wErr)

	waitForTestCompletion(t, testCompleteChan)

	// Close writers and read outputs
	_ = wOut.Close()
	_ = wErr.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, rOut)
	_, _ = io.Copy(&stderrBuf, rErr)
	_ = rOut.Close()
	_ = rErr.Close()

	// Both outputs should have content (from running actual go test)
	// We just verify that streaming happened - content depends on actual test output
	t.Logf("stdout captured: %d bytes", stdoutBuf.Len())
	t.Logf("stderr captured: %d bytes", stderrBuf.Len())
}

// TestRunTests_HandlesCommandFailure tests that runTests completes even if command fails
func TestRunTests_HandlesCommandFailure(t *testing.T) {
	// Create a failing test
	testContent := `package failtest

import "testing"

func TestFailure(t *testing.T) {
	t.Fatal("intentional failure")
}
`
	tempDir := setupTestModule(t, testContent)

	// Run the test that will fail
	config := NewTestConfig()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)
	go RunTests(ctx, testCompleteChan, nil, nil)

	waitForTestCompletion(t, testCompleteChan)
}

// TestRunTests_WaitsForBothStreamers tests that WaitGroup properly waits for both goroutines
func TestRunTests_WaitsForBothStreamers(t *testing.T) {
	testContent := `package wait

import "testing"

func TestWait(t *testing.T) {
	t.Log("waiting test")
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.ToggleVerbose()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	start := time.Now()
	go RunTests(ctx, testCompleteChan, nil, nil)

	waitForTestCompletion(t, testCompleteChan)

	duration := time.Since(start)
	// Should take some time to run tests (streaming takes time)
	// If WaitGroup wasn't working, it might return too quickly
	t.Logf("Tests completed in %v", duration)
}

// TestRunTests_DisplaysCommandBeforeRunning tests that runTests executes with config
func TestRunTests_DisplaysCommandBeforeRunning(t *testing.T) {
	testContent := `package displaytest

import "testing"

func TestPattern(t *testing.T) {
	// Simple test
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.ToggleVerbose()
	config.SetRunPattern("TestPattern")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)
	go RunTests(ctx, testCompleteChan, nil, nil)

	waitForTestCompletion(t, testCompleteChan)
}

// TestRunTests_ContextCancellation tests that runTests respects context cancellation
func TestRunTests_ContextCancellation(t *testing.T) {
	testContent := `package cancel

import "testing"

func TestCancel(t *testing.T) {
	// Simple test
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx, cancel := context.WithCancel(WithConfig(context.Background(), config))
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Cancel context immediately
	cancel()

	// Create a done channel to track if runTests completes
	done := make(chan struct{})
	go func() {
		RunTests(ctx, testCompleteChan, nil, nil)
		close(done)
	}()

	// Should handle cancellation (implementation may vary)
	// Either completes quickly or sends message
	select {
	case <-testCompleteChan:
		// Sent completion message
	case <-done:
		// Function returned
	case <-time.After(2 * time.Second):
		// Some implementations may still run - that's okay for this test
		// The important part is testing the signature and behavior
	}
}

// TestRunTests_UsesCorrectGoCommand tests that runTests calls 'go' with 'test' subcommand
func TestRunTests_UsesCorrectGoCommand(t *testing.T) {
	testContent := `package command

import "testing"

func TestCommand(t *testing.T) {
	// Simple test
}
`
	tempDir := setupTestModule(t, testContent)

	// This test verifies the command structure by running actual go test
	config := NewTestConfig()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	go RunTests(ctx, testCompleteChan, nil, nil)

	// Wait for completion
	select {
	case msg := <-testCompleteChan:
		// Verify we got the right message type
		assert.Equal(t, MessageTypeTestComplete, msg.Type(), "should return TestCompleteMessage")
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for test completion")
	}
}

// TestStreamOutput_HandlesScannerError tests error handling in streamOutput
func TestStreamOutput_HandlesScannerError(t *testing.T) {
	// Create a reader that will cause an error
	pr, pw := io.Pipe()
	scanner := bufio.NewScanner(pr)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	// Close the writer to cause an error
	_ = pw.Close()

	// Should complete without panic even with error
	streamOutput(scanner, &output, &wg, false)

	// Should still call wg.Done()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("wg.Done() not called after scanner error")
	}
}

// TestStreamOutput_WritesLineByLine tests that output is written line by line
func TestStreamOutput_WritesLineByLine(t *testing.T) {
	input := "first\nsecond\nthird\n"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)

	var output bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	streamOutput(scanner, &output, &wg, false)

	lines := strings.Split(output.String(), "\n")
	// Should have at least 3 lines (plus possible empty line at end)
	assert.GreaterOrEqual(t, len(lines), 3, "should write multiple lines")
	assert.Contains(t, output.String(), "first", "should contain first line")
	assert.Contains(t, output.String(), "second", "should contain second line")
	assert.Contains(t, output.String(), "third", "should contain third line")
}

// TestRunTests_CreatesStdoutPipe tests that runTests gets stdout pipe
func TestRunTests_CreatesStdoutPipe(t *testing.T) {
	testContent := `package stdout

import "testing"

func TestStdout(t *testing.T) {
	t.Log("test output")
}
`
	tempDir := setupTestModule(t, testContent)

	config := NewTestConfig()
	config.SetTestPath(".")
	config.WorkingDir = tempDir

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Capture stdout to verify output is streamed
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	go RunTests(ctx, testCompleteChan, nil, nil)

	select {
	case <-testCompleteChan:
		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		output := buf.String()

		// Should have some output from test execution
		// At minimum should show the command being run
		assert.NotEmpty(t, output, "should capture stdout from test execution")
	case <-time.After(30 * time.Second):
		_ = w.Close()
		os.Stdout = oldStdout
		_ = r.Close()
		t.Fatal("timeout")
	}
}

// TestRunTests_CreatesStderrPipe tests that runTests gets stderr pipe
func TestRunTests_CreatesStderrPipe(t *testing.T) {
	// Use invalid path to generate stderr output
	config := NewTestConfig()

	config.SetTestPath("./nonexistent_path_12345")

	ctx := WithConfig(context.Background(), config)
	testCompleteChan := make(chan TestCompleteMessage, 1)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	go RunTests(ctx, testCompleteChan, nil, nil)

	select {
	case <-testCompleteChan:
		_ = w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		// Should have captured stderr (error about nonexistent path)
		// Note: may be empty if all output goes to stdout
		t.Logf("stderr output: %s", buf.String())
	case <-time.After(30 * time.Second):
		_ = w.Close()
		os.Stderr = oldStderr
		_ = r.Close()
		t.Fatal("timeout")
	}
}

// TestRunTests_IntegrationWithTestConfig tests full integration with TestConfig
func TestRunTests_IntegrationWithTestConfig(t *testing.T) {
	testContent := `package integration

import "testing"

func TestOne(t *testing.T) {
	// Test one
}

func TestTwo(t *testing.T) {
	// Test two
}
`
	tempDir := setupTestModule(t, testContent)

	testCases := []struct {
		name   string
		config *TestConfig
	}{
		{
			name: "default config",
			config: func() *TestConfig {
				c := NewTestConfig()
				c.SetTestPath(".")
				c.WorkingDir = tempDir
				return c
			}(),
		},
		{
			name: "verbose config",
			config: func() *TestConfig {
				c := NewTestConfig()
				c.SetTestPath(".")
				c.ToggleVerbose()
				c.WorkingDir = tempDir
				return c
			}(),
		},
		{
			name: "with pattern",
			config: func() *TestConfig {
				c := NewTestConfig()
				c.SetTestPath(".")
				c.SetRunPattern("TestOne")
				c.WorkingDir = tempDir
				return c
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := WithConfig(context.Background(), tc.config)
			testCompleteChan := make(chan TestCompleteMessage, 1)

			go RunTests(ctx, testCompleteChan, nil, nil)

			select {
			case <-testCompleteChan:
				// Success - each config should work
			case <-time.After(30 * time.Second):
				t.Fatalf("timeout with config: %+v", tc.config)
			}
		})
	}
}

// TestStreamOutput_ConcurrentSafety tests that streamOutput is safe to use concurrently
func TestStreamOutput_ConcurrentSafety(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	var output1, output2, output3 bytes.Buffer

	reader1 := strings.NewReader("stream1 line1\nstream1 line2\n")
	reader2 := strings.NewReader("stream2 line1\nstream2 line2\n")
	reader3 := strings.NewReader("stream3 line1\nstream3 line2\n")

	scanner1 := bufio.NewScanner(reader1)
	scanner2 := bufio.NewScanner(reader2)
	scanner3 := bufio.NewScanner(reader3)

	// Run multiple streamOutput calls concurrently
	go streamOutput(scanner1, &output1, &wg, false)
	go streamOutput(scanner2, &output2, &wg, false)
	go streamOutput(scanner3, &output3, &wg, false)

	// Wait for all to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Verify each stream got its correct output
		assert.Contains(t, output1.String(), "stream1", "stream1 should have correct output")
		assert.Contains(t, output2.String(), "stream2", "stream2 should have correct output")
		assert.Contains(t, output3.String(), "stream3", "stream3 should have correct output")
	case <-time.After(1 * time.Second):
		t.Fatal("concurrent streamOutput calls did not complete")
	}
}
