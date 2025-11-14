package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleVerbose_TogglesFromFalseToTrue tests verbose toggle from false to true
func TestHandleVerbose_TogglesFromFalseToTrue(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleVerbose(config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.GetVerbose(), "Verbose should be toggled to true")
	assert.Equal(t, "Verbose: enabled\n", output, "Should print enabled message")
}

// TestHandleVerbose_TogglesFromTrueToFalse tests verbose toggle from true to false
func TestHandleVerbose_TogglesFromTrueToFalse(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    true,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleVerbose(config, []string{})
		require.NoError(t, err)
	})

	assert.False(t, config.GetVerbose(), "Verbose should be toggled to false")
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
	assert.True(t, config.GetVerbose())

	// Toggle off
	err = handleVerbose(config, []string{})
	require.NoError(t, err)
	assert.False(t, config.GetVerbose())

	// Toggle on again
	err = handleVerbose(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.GetVerbose())
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
	assert.True(t, config.GetVerbose(), "Should toggle regardless of arguments")
}

// TestHandleClear_ResetsAllFields tests that handleClear resets all config fields
func TestHandleClear_ResetsAllFields(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./custom/path",
		Verbose:     true,
		RunPattern:  "TestFoo",
		SkipPattern: "FooBar",
	}

	output := captureStdout(t, func() {
		err := handleClear(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.GetTestPath(), "TestPath should be reset to default")
	assert.False(t, config.GetVerbose(), "Verbose should be reset to false")
	assert.Equal(t, "", config.GetRunPattern(), "RunPattern should be reset to empty")
	assert.Equal(t, "", config.GetSkipPattern(), "SkipPattern should be reset to empty")
	assert.Equal(t, "All parameters cleared\n", output, "Should print cleared message")
}

// TestHandleClear_WorksWithDefaultValues tests clear when already at defaults
func TestHandleClear_WorksWithDefaultValues(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleClear(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.GetTestPath())
	assert.False(t, config.GetVerbose())
	assert.Equal(t, "", config.GetRunPattern())
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
	assert.Equal(t, "./...", config.GetTestPath(), "Should reset regardless of arguments")
}

// TestHandleHelp_DisplaysAllCommands tests that help displays all available commands
func TestHandleHelp_DisplaysAllCommands(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
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

	assert.Contains(t, output, "p    ", "Should list p command without args")
	assert.Contains(t, output, "Set test path to default", "Should describe p command without args")
	assert.Contains(t, output, "(./...)", "Should mention default path")

	assert.Contains(t, output, "clear", "Should list clear command")
	assert.Contains(t, output, "Clear all parameters", "Should describe clear command")

	assert.Contains(t, output, "cls", "Should list cls command")
	assert.Contains(t, output, "Clear screen", "Should describe cls command")

	assert.Contains(t, output, "f", "Should list force run command")
	assert.Contains(t, output, "Force test run", "Should describe run command")

	assert.Contains(t, output, "h ", "Should list help command")
	assert.Contains(t, output, "Show this help", "Should describe help command")
}

// TestHandleHelp_FormattingIsCorrect tests the exact formatting of help output
func TestHandleHelp_FormattingIsCorrect(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
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

	originalPath := config.GetTestPath()
	originalVerbose := config.GetVerbose()
	originalPattern := config.GetRunPattern()

	err := handleHelp(config, []string{})
	require.NoError(t, err)

	assert.Equal(t, originalPath, config.GetTestPath(), "TestPath should not change")
	assert.Equal(t, originalVerbose, config.GetVerbose(), "Verbose should not change")
	assert.Equal(t, originalPattern, config.GetRunPattern(), "RunPattern should not change")
}

// TestHandleHelp_IgnoresArguments tests that help ignores any arguments
func TestHandleHelp_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleHelp(config, []string{"arg1", "arg2"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Available commands:", "Should display help regardless of arguments")
}

// TestHandleVerbose_WorksViaRegistry tests verbose command through the registry
func TestHandleVerbose_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleCommand(Command("v"), config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.GetVerbose(), "Should toggle verbose via registry")
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

	output := captureStdout(t, func() {
		err := handleCommand(Command("clear"), config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.GetTestPath())
	assert.False(t, config.GetVerbose())
	assert.Equal(t, "", config.GetRunPattern())
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

	output := captureStdout(t, func() {
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

	output := captureStdout(t, func() {
		err := handleRunPattern(config, []string{"TestFoo"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestFoo", config.GetRunPattern(), "Should set run pattern")
	assert.Equal(t, "Run pattern: TestFoo\n", output, "Should print pattern message")
}

// TestHandleRunPattern_WithoutArgs tests clearing the run pattern
func TestHandleRunPattern_WithoutArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "TestBar",
	}

	output := captureStdout(t, func() {
		err := handleRunPattern(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.GetRunPattern(), "Should clear run pattern")
	assert.Equal(t, "Run pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleRunPattern_WithNilArgs tests clearing with nil args
func TestHandleRunPattern_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "TestBaz",
	}

	output := captureStdout(t, func() {
		err := handleRunPattern(config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.GetRunPattern(), "Should clear run pattern with nil args")
	assert.Equal(t, "Run pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleRunPattern_WithMultipleArgs tests that only first arg is used
func TestHandleRunPattern_WithMultipleArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleRunPattern(config, []string{"TestFirst", "TestSecond", "TestThird"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestFirst", config.GetRunPattern(), "Should use only first argument")
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
	assert.Equal(t, "TestOne", config.GetRunPattern())

	// Clear pattern
	err = handleRunPattern(config, []string{})
	require.NoError(t, err)
	assert.Equal(t, "", config.GetRunPattern())

	// Set different pattern
	err = handleRunPattern(config, []string{"TestTwo"})
	require.NoError(t, err)
	assert.Equal(t, "TestTwo", config.GetRunPattern())
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

	output := captureStdout(t, func() {
		err := handleTestPath(config, []string{tempDir})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.GetTestPath(), "Should set test path")
	assert.Equal(t, "Test path: "+tempDir+"\n", output, "Should print path message")
}

// TestHandleTestPath_WithCurrentDirectory tests setting path to current directory
func TestHandleTestPath_WithCurrentDirectory(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleTestPath(config, []string{"."})
		require.NoError(t, err)
	})

	assert.Equal(t, ".", config.GetTestPath(), "Should set path to current directory")
	assert.Equal(t, "Test path: .\n", output, "Should print path message")
}

// TestHandleTestPath_WithNoArgs that that handling 0 arguments resets TestPath to ./...
func TestHandleTestPath_WithNoArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./foo",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleTestPath(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.GetTestPath(), "TestPath should reset on blank input")
	assert.Equal(t, "Test path: ./...\n", output, "Should print path message")
}

// TestHandleTestPath_WithNilArgs tests that a nil input resets TestPath
func TestHandleTestPath_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./foo",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleTestPath(config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "./...", config.GetTestPath(), "TestPath should reset on nil input")
	assert.Equal(t, "Test path: ./...\n", output, "Should print path message")
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
	assert.Equal(t, "./...", config.GetTestPath(), "TestPath should not change on error")
}

// TestHandleTestPath_WithFile tests error handling for file path (not directory)
func TestHandleTestPath_WithFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0o600)
	require.NoError(t, err)

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err = handleTestPath(config, []string{tempFile})

	require.Error(t, err, "Should return error for file path")
	assert.Contains(t, err.Error(), "not a directory", "Error should mention it's not a directory")
	assert.Equal(t, "./...", config.GetTestPath(), "TestPath should not change on error")
}

// TestHandleTestPath_IgnoresExtraArgs tests that only first arg is used
func TestHandleTestPath_IgnoresExtraArgs(t *testing.T) {
	tempDir := t.TempDir()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleTestPath(config, []string{tempDir, "extra", "args"})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.GetTestPath(), "Should use only first argument")
	assert.Equal(t, "Test path: "+tempDir+"\n", output, "Should print first argument")
}

func TestHandleCls_UpdatesConfig(t *testing.T) {
	config := NewTestConfig()

	clearA := config.GetClearScreen()

	err := handleCls(config, []string{})
	require.NoError(t, err)

	clearB := config.GetClearScreen()

	err = handleCls(config, []string{})
	require.NoError(t, err)

	clearC := config.GetClearScreen()

	assert.False(t, clearA, "initial config should not clear screen before test runs")
	assert.True(t, clearB, "handling the command should toggle clearing the screen")
	assert.False(t, clearC, "handling the command should toggle clearing the screen")
}

// TestHandleRunPattern_WorksViaRegistry tests run pattern through the registry
func TestHandleRunPattern_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleCommand(Command("r"), config, []string{"TestViaRegistry"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestViaRegistry", config.GetRunPattern())
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

	output := captureStdout(t, func() {
		err := handleCommand(Command("p"), config, []string{tempDir})
		require.NoError(t, err)
	})

	assert.Equal(t, tempDir, config.GetTestPath())
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

	output := captureStdout(t, func() {
		err := handleCommand(Command("cls"), config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "Clear screen before each run: enabled\n", output)
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

// ============================================================================
// Step 6: Skip Pattern Command Handler Tests
// ============================================================================

// TestHandleSkipPattern_WithPattern tests setting a skip pattern
func TestHandleSkipPattern_WithPattern(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleSkipPattern(config, []string{"TestSkip"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestSkip", config.GetSkipPattern(), "Should set skip pattern")
	assert.Equal(t, "Skip pattern: TestSkip\n", output, "Should print skip pattern message")
}

// TestHandleSkipPattern_WithoutArgs tests clearing the skip pattern
func TestHandleSkipPattern_WithoutArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "TestOld",
	}

	output := captureStdout(t, func() {
		err := handleSkipPattern(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.GetSkipPattern(), "Should clear skip pattern")
	assert.Equal(t, "Skip pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleSkipPattern_WithNilArgs tests clearing with nil args
func TestHandleSkipPattern_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "TestSomething",
	}

	output := captureStdout(t, func() {
		err := handleSkipPattern(config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "", config.GetSkipPattern(), "Should clear skip pattern with nil args")
	assert.Equal(t, "Skip pattern: cleared\n", output, "Should print cleared message")
}

// TestHandleSkipPattern_WithMultipleArgs tests that only first arg is used
func TestHandleSkipPattern_WithMultipleArgs(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleSkipPattern(config, []string{"TestFirst", "TestSecond", "TestThird"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestFirst", config.GetSkipPattern(), "Should use only first argument")
	assert.Equal(t, "Skip pattern: TestFirst\n", output, "Should print first argument")
}

// TestHandleSkipPattern_TogglesMultipleTimes tests setting and clearing multiple times
func TestHandleSkipPattern_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "",
	}

	// Set pattern
	err := handleSkipPattern(config, []string{"TestOne"})
	require.NoError(t, err)
	assert.Equal(t, "TestOne", config.GetSkipPattern())

	// Clear pattern
	err = handleSkipPattern(config, []string{})
	require.NoError(t, err)
	assert.Equal(t, "", config.GetSkipPattern())

	// Set different pattern
	err = handleSkipPattern(config, []string{"TestTwo"})
	require.NoError(t, err)
	assert.Equal(t, "TestTwo", config.GetSkipPattern())
}

// TestHandleSkipPattern_WorksViaRegistry tests skip pattern through the registry
func TestHandleSkipPattern_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleCommand(Command("s"), config, []string{"TestViaRegistry"})
		require.NoError(t, err)
	})

	assert.Equal(t, "TestViaRegistry", config.GetSkipPattern())
	assert.Equal(t, "Skip pattern: TestViaRegistry\n", output)
}

func TestHandleCommandBase_WithCommand(t *testing.T) {
	initRegistry()

	config := NewTestConfig()

	output := captureStdout(t, func() {
		err := handleCommand(Command("cmd"), config, []string{"grc", "go", "test"})
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"grc", "go", "test"}, config.GetCommandBase())
	assert.Equal(t, "Test command: grc go test\n", output)
}

func TestHandleCommandBase_WithEmptyArgs(t *testing.T) {
	initRegistry()

	config := NewTestConfig()

	output := captureStdout(t, func() {
		err := handleCommand(Command("cmd"), config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"go", "test"}, config.GetCommandBase(), "command base resets on blank input")
	assert.Equal(t, "Test command: go test\n", output)
}

func TestHandleCommandBase_WithNilArgs(t *testing.T) {
	initRegistry()

	config := NewTestConfig()

	output := captureStdout(t, func() {
		err := handleCommand(Command("cmd"), config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"go", "test"}, config.GetCommandBase(), "command base resets on blank input")
	assert.Equal(t, "Test command: go test\n", output)
}

func TestHandleRace_TogglesFromFalseToTrue(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Race:       false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleRace(config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.GetRace(), "Race should be toggled to true")
	assert.Equal(t, "Race: enabled\n", output, "Should print enabled message")
}

// TestHandleRace_TogglesFromTrueToFalse tests verbose toggle from true to false
func TestHandleRace_TogglesFromTrueToFalse(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Race:       true,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleRace(config, []string{})
		require.NoError(t, err)
	})

	assert.False(t, config.GetRace(), "Race should be toggled to false")
	assert.Equal(t, "Race: disabled\n", output, "Should print disabled message")
}

// TestHandleRace_TogglesMultipleTimes tests verbose toggle multiple times
func TestHandleRace_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Race:       false,
		RunPattern: "",
	}

	// Toggle on
	err := handleRace(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.GetRace())

	// Toggle off
	err = handleRace(config, []string{})
	require.NoError(t, err)
	assert.False(t, config.GetRace())

	// Toggle on again
	err = handleRace(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.GetRace())
}

// TestHandleRace_IgnoresArguments tests that handleRace ignores any arguments
func TestHandleRace_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		Race:       false,
		RunPattern: "",
	}

	err := handleRace(config, []string{"arg1", "arg2"})
	require.NoError(t, err)
	assert.True(t, config.GetRace(), "Should toggle regardless of arguments")
}

func TestHandleFailFast_TogglesFromFalseToTrue(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		FailFast:   false,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleFailFast(config, []string{})
		require.NoError(t, err)
	})

	assert.True(t, config.GetFailFast(), "FailFast should be toggled to true")
	assert.Equal(t, "FailFast: enabled\n", output, "Should print enabled message")
}

func TestHandleFailFast_TogglesFromTrueToFalse(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		FailFast:   true,
		RunPattern: "",
	}

	output := captureStdout(t, func() {
		err := handleFailFast(config, []string{})
		require.NoError(t, err)
	})

	assert.False(t, config.GetFailFast(), "FailFast should be toggled to false")
	assert.Equal(t, "FailFast: disabled\n", output, "Should print disabled message")
}

func TestHandleFailFast_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		FailFast:   false,
		RunPattern: "",
	}

	// Toggle on
	err := handleFailFast(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.GetFailFast())

	// Toggle off
	err = handleFailFast(config, []string{})
	require.NoError(t, err)
	assert.False(t, config.GetFailFast())

	// Toggle on again
	err = handleFailFast(config, []string{})
	require.NoError(t, err)
	assert.True(t, config.GetFailFast())
}

func TestHandleFailFast_IgnoresArguments(t *testing.T) {
	config := &TestConfig{
		TestPath:   "./...",
		FailFast:   false,
		RunPattern: "",
	}

	err := handleFailFast(config, []string{"arg1", "arg2"})
	require.NoError(t, err)
	assert.True(t, config.GetFailFast(), "Should toggle regardless of arguments")
}

func TestHandleCount_WithValidPositiveNumber(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    0,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"5"})
		require.NoError(t, err)
	})

	assert.Equal(t, 5, config.GetCount(), "Should set count to 5")
	assert.Equal(t, "Count: 5\n", output, "Should print count message")
}

func TestHandleCount_WithZero(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    10,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"0"})
		require.NoError(t, err)
	})

	assert.Equal(t, 0, config.GetCount(), "Should set count to 0")
	assert.Equal(t, "Count: cleared\n", output, "Should print cleared message")
}

func TestHandleCount_WithoutArgs(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    10,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{})
		require.NoError(t, err)
	})

	assert.Equal(t, 0, config.GetCount(), "Should clear count")
	assert.Equal(t, "Count: cleared\n", output, "Should print cleared message")
}

func TestHandleCount_WithNilArgs(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    10,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, 0, config.GetCount(), "Should clear count with nil args")
	assert.Equal(t, "Count: cleared\n", output, "Should print cleared message")
}

func TestHandleCount_WithNegativeNumber(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    5,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"-5"})
		require.NoError(t, err)
	})

	assert.Equal(t, 5, config.GetCount(), "Count should remain unchanged")
	assert.Contains(t, output, "Error: count value must be non-negative (got -5)", "Should print error message")
}

func TestHandleCount_WithInvalidString(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    5,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"abc"})
		require.NoError(t, err)
	})

	assert.Equal(t, 5, config.GetCount(), "Count should remain unchanged")
	assert.Contains(t, output, "Error: invalid count value", "Should print error message")
	assert.Contains(t, output, "must be a non-negative integer", "Should explain requirement")
}

func TestHandleCount_WithFloat(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    5,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"3.14"})
		require.NoError(t, err)
	})

	assert.Equal(t, 5, config.GetCount(), "Count should remain unchanged")
	assert.Contains(t, output, "Error: invalid count value", "Should print error message")
}

func TestHandleCount_WithEmptyString(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    5,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{""})
		require.NoError(t, err)
	})

	assert.Equal(t, 5, config.GetCount(), "Count should remain unchanged")
	assert.Contains(t, output, "Error: invalid count value", "Should print error message")
}

func TestHandleCount_WithMultipleArgs(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    0,
	}

	output := captureStdout(t, func() {
		err := handleCount(config, []string{"10", "20", "30"})
		require.NoError(t, err)
	})

	assert.Equal(t, 10, config.GetCount(), "Should use only first argument")
	assert.Equal(t, "Count: 10\n", output, "Should print first argument")
}

func TestHandleCount_TogglesMultipleTimes(t *testing.T) {
	config := &TestConfig{
		TestPath: "./...",
		Count:    0,
	}

	// Set to 5
	err := handleCount(config, []string{"5"})
	require.NoError(t, err)
	assert.Equal(t, 5, config.GetCount())

	// Change to 10
	err = handleCount(config, []string{"10"})
	require.NoError(t, err)
	assert.Equal(t, 10, config.GetCount())

	// Clear
	err = handleCount(config, []string{})
	require.NoError(t, err)
	assert.Equal(t, 0, config.GetCount())

	// Set to 3
	err = handleCount(config, []string{"3"})
	require.NoError(t, err)
	assert.Equal(t, 3, config.GetCount())
}

func TestHandleCount_WorksViaRegistry(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath: "./...",
		Count:    0,
	}

	output := captureStdout(t, func() {
		err := handleCommand(Command("count"), config, []string{"7"})
		require.NoError(t, err)
	})

	assert.Equal(t, 7, config.GetCount())
	assert.Equal(t, "Count: 7\n", output)
}
