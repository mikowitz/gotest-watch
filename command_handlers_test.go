package main

import (
	"bytes"
	"io"
	"os"
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

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
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
