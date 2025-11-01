package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout captures stdout during test execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// TestHandleVerbose_TogglesFromFalseToTrue tests verbose toggle from false to true
func TestHandleVerbose_TogglesFromFalseToTrue(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleVerbose(config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.Verbose, "Verbose should be toggled to true")
	assert.Equal(t, "Verbose: enabled\n", output, "Should print enabled message")
}

// TestHandleVerbose_TogglesFromTrueToFalse tests verbose toggle from true to false
func TestHandleVerbose_TogglesFromTrueToFalse(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    true,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleVerbose(config, []string{})
		require.NoError(t, err)
	})

	assert.False(t, config.Verbose, "Verbose should be toggled to false")
	assert.Equal(t, "Verbose: disabled\n", output, "Should print disabled message")
}

// TestHandleVerbose_TogglesMultipleTimes tests verbose toggle multiple times
func TestHandleVerbose_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Toggle on
	err := handleVerbose(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.Verbose)

	// Toggle off
	err = handleVerbose(config, []string{})
	require.NoError(t, err)
	assert.False(t, config.Verbose)

	// Toggle on again
	err = handleVerbose(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.Verbose)
}

// TestHandleVerbose_IgnoresArguments tests that handleVerbose ignores any arguments
func TestHandleVerbose_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleVerbose(config, []string{"arg1", "arg2"})
	require.NoError(t, err)
	assert.True(t, config.Verbose, "Should toggle regardless of arguments")
}

// TestHandleClear_ResetsAllFields tests that handleClear resets all config fields
func TestHandleClear_ResetsAllFields(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./custom/path",
		Verbose:    true,
		RunPattern: "TestFoo",
	}

	output := captureStdout(func() {
		err := handleClear(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.TestPath, "TestPath should be reset to default")
	assert.False(t, config.Verbose, "Verbose should be reset to false")
	assert.Equal(t, "", config.RunPattern, "RunPattern should be reset to empty")
	assert.Equal(t, "All parameters cleared\n", output, "Should print cleared message")
}

// TestHandleClear_WorksWithDefaultValues tests clear when already at defaults
func TestHandleClear_WorksWithDefaultValues(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleClear(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.TestPath)
	assert.False(t, config.Verbose)
	assert.Equal(t, "", config.RunPattern)
	assert.Equal(t, "All parameters cleared\n", output)
}

// TestHandleClear_IgnoresArguments tests that handleClear ignores any arguments
func TestHandleClear_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestBar",
	}

	err := handleClear(config, []string{"arg1", "arg2"})
	require.NoError(t, err)
	assert.Equal(t, "./...", config.TestPath, "Should reset regardless of arguments")
}

// TestHandleHelp_DisplaysAllCommands tests that help displays all available commands
func TestHandleHelp_DisplaysAllCommands(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleHelp(config, []string{})
		require.NoError(t, err)
	})

	// Verify help header
	assert.Contains(t, output, "Available commands:", "Should have header")

	// Verify all commands are listed
	assert.Contains(t, output, "v", "Should list v command")
	assert.Contains(t, output, "Toggle verbose mode", "Should describe v command")
	assert.Contains(t, output, "-v flag", "Should mention -v flag")

	assert.Contains(t, output, "r <pattern>", "Should list r command with pattern")
	assert.Contains(t, output, "Set test run pattern", "Should describe r command")
	assert.Contains(t, output, "-run=<pattern>", "Should mention -run flag")

	assert.Contains(t, output, "r  ", "Should list r command without args")
	assert.Contains(t, output, "Clear run pattern", "Should describe r clear")

	assert.Contains(t, output, "p <path>", "Should list p command")
	assert.Contains(t, output, "Set test path", "Should describe p command")
	assert.Contains(t, output, "default: ./...", "Should mention default path")

	assert.Contains(t, output, "clear", "Should list clear command")
	assert.Contains(t, output, "Clear all parameters", "Should describe clear command")

	assert.Contains(t, output, "cls", "Should list cls command")
	assert.Contains(t, output, "Clear screen", "Should describe cls command")

	assert.Contains(t, output, "f", "Should list force run command")
	assert.Contains(t, output, "Force test run", "Should describe run command")

	assert.Contains(t, output, "h", "Should list help command")
	assert.Contains(t, output, "Show this help", "Should describe help command")
}

// TestHandleHelp_FormattingIsCorrect tests the exact formatting of help output
func TestHandleHelp_FormattingIsCorrect(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleHelp(config, []string{})
		require.NoError(t, err)
	})

	// Should start with header
	assert.True(t, strings.HasPrefix(output, "Available commands:\n"), "Should start with header")

	// Each command should be on its own line with proper indentation
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Greater(t, len(lines), 1, "Should have multiple lines")

	// First line should be the header
	assert.Equal(t, "Available commands:", lines[0])

	// Remaining lines should have proper indentation (2 spaces)
	for i := 1; i < len(lines); i++ {
		if lines[i] != "" {
			assert.True(t, strings.HasPrefix(lines[i], "  "),
				"Command line %d should be indented: %q", i, lines[i])
		}
	}
}

// TestHandleHelp_DoesNotModifyConfig tests that help doesn't change config
func TestHandleHelp_DoesNotModifyConfig(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestFoo",
	}

	originalPath := config.TestPath
	originalVerbose := config.Verbose
	originalPattern := config.RunPattern

	err := handleHelp(config, []string{})
	require.NoError(t, err)

	assert.Equal(t, originalPath, config.TestPath, "TestPath should not change")
	assert.Equal(t, originalVerbose, config.Verbose, "Verbose should not change")
	assert.Equal(t, originalPattern, config.RunPattern, "RunPattern should not change")
}

// TestHandleHelp_IgnoresArguments tests that help ignores any arguments
func TestHandleHelp_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleHelp(config, []string{"arg1", "arg2"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Available commands:", "Should display help regardless of arguments")
}

// TestInitRegistry_RegistersSimpleHandlers tests that initRegistry registers the simple handlers
func TestInitRegistry_RegistersSimpleHandlers(t *testing.T) {
	initRegistry()

	// Verify handlers are registered
	_, hasVerbose := commandRegistry[Command("v")]
	assert.True(t, hasVerbose, "Should register 'v' command")

	_, hasClear := commandRegistry[Command("clear")]
	assert.True(t, hasClear, "Should register 'clear' command")

	_, hasHelp := commandRegistry[Command("h")]
	assert.True(t, hasHelp, "Should register 'help' command")
}

// TestHandleVerbose_WorksViaRegistry tests verbose command through the registry
func TestHandleVerbose_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("v"), config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.Verbose, "Should toggle verbose via registry")
	assert.Equal(t, "Verbose: enabled\n", output)
}

// TestHandleClear_WorksViaRegistry tests clear command through the registry
func TestHandleClear_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestFoo",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("clear"), config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.TestPath)
	assert.False(t, config.Verbose)
	assert.Equal(t, "", config.RunPattern)
	assert.Equal(t, "All parameters cleared\n", output)
}

// TestHandleHelp_WorksViaRegistry tests help command through the registry
func TestHandleHelp_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("h"), config, []string{})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Available commands:", "Should display help via registry")
}

// ============================================================================
// Step 5: Parameter Command Handlers Tests
// ============================================================================

// TestHandleRunPattern_WithPattern tests setting a run pattern
func TestHandleRunPattern_WithPattern(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleRunPattern(config, []string{"TestFoo"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestFoo", config.RunPattern, "Should set run pattern")
	assert.Equal(t, "Run pattern: TestFoo\n", output, "Should print pattern message")
}

// TestHandleRunPattern_WithoutArgs tests clearing the run pattern
func TestHandleRunPattern_WithoutArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "TestBar",
	}

	output := captureStdout(func() {
		err := handleRunPattern(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.RunPattern, "Should clear run pattern")
	assert.Equal(t, "Run pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleRunPattern_WithNilArgs tests clearing with nil args
func TestHandleRunPattern_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "TestBaz",
	}

	output := captureStdout(func() {
		err := handleRunPattern(config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.RunPattern, "Should clear run pattern with nil args")
	assert.Equal(t, "Run pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleRunPattern_WithMultipleArgs tests that only first arg is used
func TestHandleRunPattern_WithMultipleArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleRunPattern(config, []string{"TestFirst", "TestSecond", "TestThird"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestFirst", config.RunPattern, "Should use only first argument")
	assert.Equal(t, "Run pattern: TestFirst\n", output, "Should print first argument")
}

// TestHandleRunPattern_TogglesMultipleTimes tests setting and clearing multiple times
func TestHandleRunPattern_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Set pattern
	err := handleRunPattern(config, []string{"TestOne"})
	require.NoError(t, err)
	assert.Equal(t, "TestOne", config.RunPattern)

	// Clear pattern
	err = handleRunPattern(config, []string{})
	require.NoError(t, err)
	assert.Equal(t, "", config.RunPattern)

	// Set different pattern
	err = handleRunPattern(config, []string{"TestTwo"})
	require.NoError(t, err)
	assert.Equal(t, "TestTwo", config.RunPattern)
}

// TestHandleTestPath_WithValidDirectory tests setting a valid test path
func TestHandleTestPath_WithValidDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleTestPath(config, []string{tempDir})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.TestPath, "Should set test path")
	assert.Equal(t, "Test path: "+tempDir+"\n", output, "Should print path message")
}

// TestHandleTestPath_WithCurrentDirectory tests setting path to current directory
func TestHandleTestPath_WithCurrentDirectory(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleTestPath(config, []string{"."})
		require.NoError(t, err)
	})

	assert.Equal(t, ".", config.TestPath, "Should set path to current directory")
	assert.Equal(t, "Test path: .\n", output, "Should print path message")
}

// TestHandleTestPath_WithInvalidPath tests error handling for non-existent path
func TestHandleTestPath_WithInvalidPath(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleTestPath(config, []string{"/nonexistent/path/that/does/not/exist"})

	require.Error(t, err, "Should return error for invalid path")
	assert.Contains(t, err.Error(), "path does not exist", "Error should mention path doesn't exist")
	assert.Equal(t, "./...", config.TestPath, "TestPath should not change on error")
}

// TestHandleTestPath_WithFile tests error handling for file path (not directory)
func TestHandleTestPath_WithFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0o644)
	require.NoError(t, err)

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err = handleTestPath(config, []string{tempFile})

	require.Error(t, err, "Should return error for file path")
	assert.Contains(t, err.Error(), "not a directory", "Error should mention it's not a directory")
	assert.Equal(t, "./...", config.TestPath, "TestPath should not change on error")
}

// TestHandleTestPath_WithNoArgs tests error handling for missing argument
func TestHandleTestPath_WithNoArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleTestPath(config, []string{})

	require.Error(t, err, "Should return error when no args provided")
	assert.Contains(t, err.Error(), "path argument required", "Error should mention required argument")
	assert.Equal(t, "./...", config.TestPath, "TestPath should not change on error")
}

// TestHandleTestPath_WithNilArgs tests error handling for nil arguments
func TestHandleTestPath_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleTestPath(config, nil)

	require.Error(t, err, "Should return error when nil args provided")
	assert.Contains(t, err.Error(), "path argument required", "Error should mention required argument")
	assert.Equal(t, "./...", config.TestPath, "TestPath should not change on error")
}

// TestHandleTestPath_IgnoresExtraArgs tests that only first arg is used
func TestHandleTestPath_IgnoresExtraArgs(t *testing.T) {
	tempDir := t.TempDir()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleTestPath(config, []string{tempDir, "extra", "args"})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.TestPath, "Should use only first argument")
	assert.Equal(t, "Test path: "+tempDir+"\n", output, "Should print first argument")
}

// TestHandleCls_PrintsAnsiEscapeSequence tests cls prints correct escape sequence
func TestHandleCls_PrintsAnsiEscapeSequence(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCls(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "\033[H\033[2J", output, "Should print ANSI escape sequence")
}

// TestHandleCls_DoesNotModifyConfig tests that cls doesn't change config
func TestHandleCls_DoesNotModifyConfig(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestFoo",
	}

	originalPath := config.TestPath
	originalVerbose := config.Verbose
	originalPattern := config.RunPattern

	err := handleCls(config, []string{})
	require.NoError(t, err)

	assert.Equal(t, originalPath, config.TestPath, "TestPath should not change")
	assert.Equal(t, originalVerbose, config.Verbose, "Verbose should not change")
	assert.Equal(t, originalPattern, config.RunPattern, "RunPattern should not change")
}

// TestHandleCls_IgnoresArguments tests that cls ignores any arguments
func TestHandleCls_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCls(config, []string{"arg1", "arg2"})
		require.NoError(t, err)
	})

	assert.Equal(t, "\033[H\033[2J", output, "Should print escape sequence regardless of args")
}

// TestHandleRun_ReturnsNil tests that run handler is a stub returning nil
func TestHandleRun_ReturnsNil(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleRun(config, []string{})
	require.NoError(t, err, "Run handler should return nil (stub)")
}

// TestHandleRun_DoesNotModifyConfig tests that run doesn't change config
func TestHandleRun_DoesNotModifyConfig(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestFoo",
	}

	originalPath := config.TestPath
	originalVerbose := config.Verbose
	originalPattern := config.RunPattern

	err := handleRun(config, []string{})
	require.NoError(t, err)

	assert.Equal(t, originalPath, config.TestPath, "TestPath should not change")
	assert.Equal(t, originalVerbose, config.Verbose, "Verbose should not change")
	assert.Equal(t, originalPattern, config.RunPattern, "RunPattern should not change")
}

// TestHandleRun_IgnoresArguments tests that run ignores any arguments
func TestHandleRun_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleRun(config, []string{"arg1", "arg2"})
	require.NoError(t, err, "Should succeed regardless of arguments")
}

// TestInitRegistry_RegistersParameterHandlers tests that initRegistry registers parameter handlers
func TestInitRegistry_RegistersParameterHandlers(t *testing.T) {
	initRegistry()

	// Verify handlers are registered
	_, hasRunPattern := commandRegistry[Command("r")]
	assert.True(t, hasRunPattern, "Should register 'r' command")

	_, hasTestPath := commandRegistry[Command("p")]
	assert.True(t, hasTestPath, "Should register 'p' command")

	_, hasCls := commandRegistry[Command("cls")]
	assert.True(t, hasCls, "Should register 'cls' command")

	_, hasRun := commandRegistry[Command("f")]
	assert.True(t, hasRun, "Should register 'f' command")
}

// TestHandleRunPattern_WorksViaRegistry tests run pattern through the registry
func TestHandleRunPattern_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("r"), config, []string{"TestViaRegistry"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestViaRegistry", config.RunPattern)
	assert.Equal(t, "Run pattern: TestViaRegistry\n", output)
}

// TestHandleTestPath_WorksViaRegistry tests test path through the registry
func TestHandleTestPath_WorksViaRegistry(t *testing.T) {
	initRegistry()
	tempDir := t.TempDir()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("p"), config, []string{tempDir})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.TestPath)
	assert.Equal(t, "Test path: "+tempDir+"\n", output)
}

// TestHandleCls_WorksViaRegistry tests cls through the registry
func TestHandleCls_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(func() {
		err := handleCommand(Command("cls"), config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "\033[H\033[2J", output)
}

// TestHandleRun_WorksViaRegistry tests run through the registry
func TestHandleRun_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleCommand(Command("f"), config, []string{})
	require.NoError(t, err)
}
