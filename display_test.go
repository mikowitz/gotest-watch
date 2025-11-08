package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDisplayPrompt_OutputFormat tests that displayPrompt prints the correct format
func TestDisplayPrompt_OutputFormat(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Call the function
	displayPrompt()

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	// Verify exact format: "> "
	assert.Equal(t, "> ", buf.String())
}

// TestDisplayPrompt_DoesNotPanic tests that displayPrompt doesn't panic
func TestDisplayPrompt_DoesNotPanic(t *testing.T) {
	// Capture stdout to suppress output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Should not panic
	assert.NotPanics(t, func() {
		displayPrompt()
	})

	// Cleanup
	_ = w.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(io.Discard, r)
}

// TestDisplayCommand_OutputFormat tests that displayCommand prints correct format
func TestDisplayCommand_OutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple command",
			args:     []string{"test", "./..."},
			expected: "go test ./...\n",
		},
		{
			name:     "command with single flag",
			args:     []string{"test", "-v", "./..."},
			expected: "go test -v ./...\n",
		},
		{
			name:     "command with multiple flags",
			args:     []string{"test", "-v", "-race", "-run=TestFoo", "./..."},
			expected: "go test -v -race -run=TestFoo ./...\n",
		},
		{
			name:     "command with only subcommand",
			args:     []string{"test"},
			expected: "go test\n",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "go \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w

			// Call the function
			displayCommand(tt.args)

			// Restore stdout
			_ = w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			require.NoError(t, err)

			// Verify exact format
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

// TestDisplayCommand_DoesNotPanic tests that displayCommand doesn't panic with various inputs
func TestDisplayCommand_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "normal args",
			args: []string{"test", "./..."},
		},
		{
			name: "empty slice",
			args: []string{},
		},
		{
			name: "nil slice",
			args: nil,
		},
		{
			name: "args with empty strings",
			args: []string{"test", "", "./..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to suppress output
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w

			// Should not panic
			assert.NotPanics(t, func() {
				displayCommand(tt.args)
			})

			// Cleanup
			_ = w.Close()
			os.Stdout = oldStdout
			_, _ = io.Copy(io.Discard, r)
		})
	}
}

// TestDisplayCommand_JoinsWithSpaces tests that displayCommand joins args with spaces
func TestDisplayCommand_JoinsWithSpaces(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Call with multiple args
	displayCommand([]string{"test", "-v", "-race", "./..."})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	// Verify spaces between each part
	assert.Contains(t, buf.String(), "go test -v -race ./...")
}

// TestDisplayCommand_PrintsToStdout tests that displayCommand writes to stdout, not stderr
func TestDisplayCommand_PrintsToStdout(t *testing.T) {
	// Capture both stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = wOut
	os.Stderr = wErr

	// Call the function
	displayCommand([]string{"test", "./..."})

	// Restore stdout/stderr
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	// Should write to stdout, not stderr
	assert.NotEmpty(t, bufOut.String(), "should write to stdout")
	assert.Empty(t, bufErr.String(), "should not write to stderr")
}

// TestDisplayPrompt_PrintsToStdout tests that displayPrompt writes to stdout, not stderr
func TestDisplayPrompt_PrintsToStdout(t *testing.T) {
	// Capture both stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = wOut
	os.Stderr = wErr

	// Call the function
	displayPrompt()

	// Restore stdout/stderr
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	// Should write to stdout, not stderr
	assert.NotEmpty(t, bufOut.String(), "should write to stdout")
	assert.Empty(t, bufErr.String(), "should not write to stderr")
}

// TestDisplayCommand_WithRealCommandFormat tests displayCommand with realistic command formats
func TestDisplayCommand_WithRealCommandFormat(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "test with verbose",
			args:     []string{"test", "-v", "./..."},
			contains: "go test -v ./...",
		},
		{
			name:     "test with pattern",
			args:     []string{"test", "-run=TestFoo", "./..."},
			contains: "go test -run=TestFoo ./...",
		},
		{
			name:     "test with verbose and pattern",
			args:     []string{"test", "-v", "-run=TestBar", "./..."},
			contains: "go test -v -run=TestBar ./...",
		},
		{
			name:     "test with specific path",
			args:     []string{"test", "./pkg/..."},
			contains: "go test ./pkg/...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w

			// Call the function
			displayCommand(tt.args)

			// Restore stdout
			_ = w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			require.NoError(t, err)

			// Verify output contains expected command
			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}
