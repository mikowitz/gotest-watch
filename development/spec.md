# gotest-watch Specification

## Project Overview

A Golang CLI tool for interactive test running that watches Go files and provides an interactive command interface for controlling test execution.

## Project Details

- **Binary Name**: `gotest-watch`
- **Module Path**: `github.com/mikowitz/gotest-watch`
- **Minimum Go Version**: 1.23
- **Project Structure**: Single `main.go` file with all logic (to be refactored into internal packages as needed)

## Core Behavior

### File Watching
- Watch all `*.go` files recursively in the current directory
- Exclude hidden files and directories (those starting with `.`)
- Ignore symbolic links
- Use `fsnotify/fsnotify` library for file system events
- File additions, modifications, AND deletions trigger test runs
- Gracefully handle file/directory removals or permission changes (no errors)
- 200ms debounce window: reset timer on each file change to batch rapid successive changes
- Start watching after initial test run completes

### Test Execution
- Default command: `go test <path> <flags>`
- Default test path: `./...`
- Stream test output to stdout/stderr in real-time using line-by-line reading
- Read stdout and stderr in separate goroutines
- Wait for both scanner goroutines AND `cmd.Wait()` before considering test complete
- Display full test command before each run
- Automatically run tests once on startup with default settings
- Block all stdin input while tests are running
- Ignore file changes while tests are running (no queuing)
- One blank line between test output and next prompt

### Test Configuration
- Parameters accumulate until explicitly cleared
- Build commands in order: `go test <path> <boolean-flags> <flags-with-values>`
- Use `=` syntax for flags with values (e.g., `-run=TestFoo`)
- Configuration stored in TestConfig struct:
  ```go
  type TestConfig struct {
      TestPath   string  // default: "./..."
      Verbose    bool    // default: false
      RunPattern string  // default: ""
  }
  ```
- TestConfig has method `BuildCommand() []string` to generate command arguments

## Architecture

### Concurrency Model
- Use goroutines with channels for communication
- Context-based lifecycle management (context.Context passed throughout)
- TestConfig pointer stored in context
- Graceful shutdown on SIGINT/SIGTERM:
  - Wait for current test to complete
  - Clean up resources
  - Print shutdown message
  - Exit

### Message Types
All message types implement a common interface:

```go
type Message interface {
    Type() MessageType
}
```

Concrete message types (minimal data, no timestamps initially):
- `FileChangeMessage` - sent by watcher after debounce expires
- `CommandMessage` - sent by stdin reader with parsed command
  - Fields: `Command string`, `Args []string`
  - Args split by simple whitespace (strings.Fields)
- `HelpMessage` - sent by stdin reader for help command
- `TestCompleteMessage` - sent when test execution finishes (no data)

### Channel Architecture
- Separate channel for each message type
- Dispatcher uses `select` statement to handle all channels
- Additional channel for stdin ready state (unbuffered, receive-only, carries bool)
  - Stdin reader uses `select` to monitor both stdin and ready channel
  - True = stdin enabled, false = stdin blocked

### Components

#### File Watcher
- Runs in dedicated goroutine
- Manages 200ms debounce timer internally
- Resets timer on each file change event
- Sends `FileChangeMessage` only after debounce expires
- Handles fsnotify errors gracefully

#### Stdin Reader
- Runs in dedicated goroutine
- Reads from stdin line by line
- Parses commands: splits on whitespace into command name and args
- Case-sensitive command names
- Sends `CommandMessage` or `HelpMessage` to dispatcher
- Blocks/unblocks based on ready channel state

#### Dispatcher
- Central event loop using `select` on multiple channels
- Receives all message types
- Maintains state: is test currently running?
- Handles message-specific logic:
  - `FileChangeMessage`: spawn test runner if not already running
  - `CommandMessage`: execute command handler, spawn test runner
  - `HelpMessage`: display help, do NOT spawn test runner
  - `TestCompleteMessage`: update running state, re-enable stdin

#### Test Runner
- Spawned as new goroutine for each test run
- Reads TestConfig pointer from context when starting
- Builds command using `TestConfig.BuildCommand()`
- Uses `exec.Command` with separate goroutines for stdout/stderr streaming
- Uses `bufio.Scanner` for line-by-line reading
- Writes output directly to os.Stdout/os.Stderr (allows future colorization)
- Sends `TestCompleteMessage` when complete

#### Command Registry
- Map of command names to handler functions: `map[string]func(*TestConfig, []string) error`
- Canonical command names only (no aliases initially)
- Handlers print acknowledgment messages and return errors
- Errors written to stderr

## Command Interface

### User Input
- Prompt: `>` (inline where user types)
- Commands are case-sensitive
- Input blocked during test execution (ready channel controls this)

### Commands

#### `v` - Toggle Verbose
- Toggles `-v` flag on/off
- Acknowledgment: "Verbose: enabled" or "Verbose: disabled"
- Triggers test run

#### `r <pattern>` - Set Run Pattern
- Sets `-run=<pattern>` parameter
- Acknowledgment: "Run pattern: <pattern>"
- Triggers test run

#### `r` - Clear Run Pattern
- Clears `-run` parameter
- Acknowledgment: "Run pattern: cleared"
- Triggers test run

#### `p <path>` - Set Test Path
- Sets test path (replaces `./...`)
- Validates path exists in current project root
- Validates it's a directory/package
- Let `go test` handle validation of whether it contains tests
- Acknowledgment: "Test path: <path>"
- Triggers test run
- If validation fails: write error to stderr, continue running

#### `clear` - Clear All Parameters
- Resets all parameters to defaults
- Acknowledgment: "All parameters cleared"
- Triggers test run

#### `cls` - Clear Screen
- Clears the terminal screen
- No test run triggered

#### `run` - Force Test Run
- Immediately runs tests with current parameters
- No acknowledgment message
- Note: input is blocked during test runs, so this only works when idle

#### `help` - Show Help
- Displays help text with available commands
- Format:
  ```
  Available commands:
    v              Toggle verbose mode (-v flag)
    r <pattern>    Set test run pattern (-run=<pattern>)
    r              Clear run pattern
    p <path>       Set test path (default: ./...)
    clear          Clear all parameters
    cls            Clear screen
    run            Force test run
    help           Show this help
  ```
- No test run triggered

### Command Execution Flow
1. User types command
2. Stdin reader parses and sends appropriate message
3. Dispatcher receives message
4. Command handler executes:
   - Validates input (if needed)
   - Updates TestConfig
   - Prints acknowledgment
   - Returns error or nil
5. If error: write to stderr and continue
6. If CommandMessage (not HelpMessage): dispatcher spawns test runner
7. Test runner disables stdin, runs tests, re-enables stdin when complete

## Output Format

### Startup
```
Test watcher started
go test ./...
[test output]

> 
```

### After Command
```
> v
Verbose: enabled
go test ./... -v
[test output]

> 
```

### Error Handling
- Startup errors (watcher init, etc.): print error and exit
- Test command failures (non-zero exit): display output, continue
- Invalid commands: write error to stderr, continue
- Path validation failures: write error to stderr, continue

## Logging

- Use `log/slog` for structured logging
- Initial version: minimal logging to stdout
- Log startup message: "Test watcher started"
- Log parameter changes via command acknowledgments
- Future: configurable file logging and verbose output

## Implementation Notes

### Dependencies
- `github.com/fsnotify/fsnotify` - file system watching
- Standard library for everything else

### Error Handling
- Fatal startup errors: log and exit immediately
- Runtime errors: log to stderr and continue
- Graceful handling of file system changes (deletions, permission changes)

### Testing Strategy
- Unit tests for command parsing
- Unit tests for TestConfig.BuildCommand()
- Integration tests for file watching and test execution flow
- Manual testing for terminal interaction and output formatting

### Future Enhancements (Not in Initial Version)
- Command aliases (`f` for `run`, `h` for `help`)
- Output colorization
- Additional test flags (e.g., `-race`, `-cover`)
- Configuration file support
- File/verbose logging options
- Custom test commands (replace `go test` entirely)
- Timestamps on messages
- Test result summaries
- Watch specific file patterns beyond `*.go`

## Development Checklist

1. Initialize Go module
2. Implement TestConfig struct and BuildCommand method
3. Implement message types and interfaces
4. Implement command registry and handlers
5. Implement file watcher with debounce
6. Implement stdin reader with blocking control
7. Implement test runner with streaming output
8. Implement central dispatcher
9. Wire up context, channels, and graceful shutdown
10. Add startup behavior and initial test run
11. Test and refine
12. Refactor into internal packages as code grows
