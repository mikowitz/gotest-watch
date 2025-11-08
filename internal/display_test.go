package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDisplayPrompt_OutputFormat tests that displayPrompt prints the correct format
func TestDisplayPrompt_OutputFormat(t *testing.T) {
	// Call the function
	actual := captureStdout(t, func() {
		displayPrompt()
	})

	// Verify exact format: "> "
	assert.Equal(t, "> ", actual)
}

// TestDisplayPrompt_DoesNotPanic tests that displayPrompt doesn't panic
func TestDisplayPrompt_DoesNotPanic(t *testing.T) {
	// Should not panic
	assert.NotPanics(t, func() {
		captureStdout(t, func() {
			displayPrompt()
		})
	})
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
			actual := captureStdout(t, func() {
				displayCommand(tt.args)
			})
			assert.Equal(t, tt.expected, actual)
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
			assert.NotPanics(t, func() {
				captureStdout(t, func() {
					displayCommand(tt.args)
				})
			})
		})
	}
}

// TestDisplayCommand_JoinsWithSpaces tests that displayCommand joins args with spaces
func TestDisplayCommand_JoinsWithSpaces(t *testing.T) {
	actual := captureStdout(t, func() {
		displayCommand([]string{"test", "-v", "-race", "./..."})
	})

	// Verify spaces between each part
	assert.Contains(t, actual, "go test -v -race ./...")
}

// TestDisplayCommand_PrintsToStdout tests that displayCommand writes to stdout, not stderr
func TestDisplayCommand_PrintsToStdout(t *testing.T) {
	actual := captureStdout(t, func() {
		displayCommand([]string{"test", "./..."})
	})

	// Should write to stdout, not stderr
	assert.NotEmpty(t, actual, "should write to stdout")
}

// TestDisplayPrompt_PrintsToStdout tests that displayPrompt writes to stdout, not stderr
func TestDisplayPrompt_PrintsToStdout(t *testing.T) {
	actual := captureStdout(t, func() {
		displayPrompt()
	})

	// Should write to stdout, not stderr
	assert.NotEmpty(t, actual, "should write to stdout")
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
			actual := captureStdout(t, func() {
				displayCommand(tt.args)
			})

			// Verify output contains expected command
			assert.Contains(t, actual, tt.contains)
		})
	}
}
