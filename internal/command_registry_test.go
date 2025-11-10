package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that initRegistry creates a valid registry map
func TestInitRegistry(t *testing.T) {
	// Clear any existing registry
	commandRegistry = nil

	initRegistry()

	require.NotNil(t, commandRegistry, "initRegistry() did not initialize commandRegistry")
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

	_, hasSkipPattern := commandRegistry[Command("s")]
	assert.True(t, hasSkipPattern, "Should register 's' command")

	_, hasCommandBase := commandRegistry[Command("cmd")]
	assert.True(t, hasCommandBase, "Should register 'cmd' command")
}

// Test that handleCommand returns an error for unknown commands
func TestHandleCommand_UnknownCommand(t *testing.T) {
	initRegistry()

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleCommand(Command("nonexistent"), config, []string{})

	require.Error(t, err, "expected error for unknown command")
	assert.EqualError(t, err, "unknown command: \"nonexistent\"")
}

// Test that handleCommand executes a registered handler
func TestHandleCommand_ExecutesHandler(t *testing.T) {
	initRegistry()

	// Track if handler was called
	handlerCalled := false

	// Create a mock handler
	mockHandler := func(cfg *TestConfig, args []string) error {
		handlerCalled = true
		return nil
	}

	// Register the mock handler
	commandRegistry[Command("test")] = mockHandler

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleCommand(Command("test"), config, []string{})

	require.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
}

// Test that handleCommand passes correct arguments to handler
func TestHandleCommand_PassesCorrectArguments(t *testing.T) {
	initRegistry()

	var receivedConfig *TestConfig
	var receivedArgs []string

	// Create a mock handler that captures arguments
	mockHandler := func(cfg *TestConfig, args []string) error {
		receivedConfig = cfg
		receivedArgs = args
		return nil
	}

	// Register the mock handler
	commandRegistry[Command("test")] = mockHandler

	config := &TestConfig{
		TestPath:   "./custom",
		Verbose:    true,
		RunPattern: "TestFoo",
	}
	args := []string{"arg1", "arg2"}

	err := handleCommand(Command("test"), config, args)

	require.NoError(t, err)
	assert.Same(t, config, receivedConfig, "handler did not receive correct config pointer")
	assert.Equal(t, args, receivedArgs, "handler did not receive correct arguments")
}

// Test that handleCommand propagates handler errors
func TestHandleCommand_PropagatesErrors(t *testing.T) {
	initRegistry()

	expectedError := errors.New("handler error")

	// Create a mock handler that returns an error
	mockHandler := func(cfg *TestConfig, args []string) error {
		return expectedError
	}

	// Register the mock handler
	commandRegistry[Command("test")] = mockHandler

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleCommand(Command("test"), config, []string{})

	require.Error(t, err, "expected error to be propagated")
	assert.Equal(t, expectedError, err, "expected exact error to be propagated")
}

// Test that handleCommand works with nil args
func TestHandleCommand_WithNilArgs(t *testing.T) {
	initRegistry()

	var receivedArgs []string

	// Create a mock handler that captures arguments
	mockHandler := func(cfg *TestConfig, args []string) error {
		receivedArgs = args
		return nil
	}

	// Register the mock handler
	commandRegistry[Command("test")] = mockHandler

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	err := handleCommand(Command("test"), config, nil)

	require.NoError(t, err)
	assert.Nil(t, receivedArgs, "expected nil args to be passed to handler")
}

// Test that handleCommand works with empty args
func TestHandleCommand_WithEmptyArgs(t *testing.T) {
	initRegistry()

	var receivedArgs []string

	// Create a mock handler that captures arguments
	mockHandler := func(cfg *TestConfig, args []string) error {
		receivedArgs = args
		return nil
	}

	// Register the mock handler
	commandRegistry[Command("test")] = mockHandler

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	emptyArgs := []string{}
	err := handleCommand(Command("test"), config, emptyArgs)

	require.NoError(t, err)
	assert.Empty(t, receivedArgs, "expected empty args to be passed to handler")
}

// Test that multiple handlers can be registered
func TestHandleCommand_MultipleHandlers(t *testing.T) {
	initRegistry()

	handler1Called := false
	handler2Called := false

	// Create mock handlers
	mockHandler1 := func(cfg *TestConfig, args []string) error {
		handler1Called = true
		return nil
	}

	mockHandler2 := func(cfg *TestConfig, args []string) error {
		handler2Called = true
		return nil
	}

	// Register both handlers
	commandRegistry[Command("cmd1")] = mockHandler1
	commandRegistry[Command("cmd2")] = mockHandler2

	config := &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Call first handler
	err := handleCommand(Command("cmd1"), config, []string{})
	require.NoError(t, err)
	assert.True(t, handler1Called, "handler1 was not called")
	assert.False(t, handler2Called, "handler2 should not have been called")

	// Reset and call second handler
	handler1Called = false
	err = handleCommand(Command("cmd2"), config, []string{})
	require.NoError(t, err)
	assert.False(t, handler1Called, "handler1 should not have been called")
	assert.True(t, handler2Called, "handler2 was not called")
}
