# gotest-watch Implementation Blueprint

## Overview
This blueprint breaks down the implementation into small, testable steps that build incrementally. Each step adds functionality while maintaining a working program that can be tested.

---

## Phase 1: Foundation (Steps 1-3)
**Goal**: Establish project structure, basic types, and test configuration

## Phase 2: Core Components (Steps 4-7)
**Goal**: Build individual components in isolation with tests

## Phase 3: Integration (Steps 8-10)
**Goal**: Wire components together into a working system

## Phase 4: Polish (Steps 11-12)
**Goal**: Add remaining commands and refinements

---

# Detailed Step Breakdown

## Step 1: Project Initialization & TestConfig
- Initialize Go module
- Create TestConfig struct with defaults
- Implement BuildCommand() method
- Write comprehensive tests for command building

## Step 2: Message Types & Interfaces
- Define Message interface and MessageType enum
- Implement all message type structs
- Write tests for message type identification

## Step 3: Command Registry Foundation
- Create command registry map structure
- Implement registry lookup function
- Add basic error handling for unknown commands
- Test registry with mock handlers

## Step 4: Command Handlers - Part 1 (Simple Commands)
- Implement `v` (toggle verbose) handler
- Implement `clear` (clear all params) handler
- Implement `help` handler
- Test each handler in isolation
- Wire handlers into registry

## Step 5: Command Handlers - Part 2 (Parameter Commands)
- Implement `r <pattern>` handler with argument parsing
- Implement `p <path>` handler with validation
- Implement `cls` handler
- Implement `run` handler (stub for now)
- Test handlers with various inputs
- Wire into registry

## Step 6: Stdin Reader
- Implement stdin reader goroutine
- Add command parsing (split into command + args)
- Implement ready channel blocking logic
- Send CommandMessage or HelpMessage based on input
- Test with mock channels

## Step 7: File Watcher with Debounce
- Implement fsnotify-based watcher
- Add recursive directory walking (excluding hidden)
- Implement 200ms debounce timer with reset logic
- Send FileChangeMessage after debounce
- Test with temporary file system

## Step 8: Test Runner
- Implement test runner goroutine
- Build command from TestConfig
- Execute with exec.Command
- Stream stdout/stderr with separate scanner goroutines
- Send TestCompleteMessage when done
- Test with mock commands

## Step 9: Dispatcher & Channel Wiring
- Implement central dispatcher with select loop
- Create all channels
- Wire stdin reader to dispatcher
- Wire file watcher to dispatcher
- Wire test runner to dispatcher
- Add "tests running" state tracking
- Test message routing

## Step 10: Context & Lifecycle Management
- Add context with TestConfig
- Implement signal handling (SIGINT/SIGTERM)
- Add graceful shutdown logic
- Wire context through all components
- Test shutdown behavior

## Step 11: Startup & Initial Test Run
- Implement startup sequence
- Run initial test with defaults
- Start file watcher after initial run
- Display startup message
- Test complete startup flow

## Step 12: Output Formatting & Polish
- Add prompt display logic
- Implement proper spacing (blank line before prompt)
- Add command output display (full command before tests)
- Final integration testing
- Manual testing of full workflow

---

# Prompt Sequence for LLM Code Generation

## Prompt 1: Project Setup & TestConfig

```
Create a new Go project for gotest-watch with the following requirements:

1. Initialize a Go module at github.com/mikowitz/gotest-watch with minimum Go version 1.23
2. Create main.go with a basic main function that prints "Test watcher started"
3. Implement a TestConfig struct with these fields:
   - TestPath string (default "./...")
   - Verbose bool (default false)
   - RunPattern string (default "")
4. Add a BuildCommand() method on TestConfig that returns []string with the command arguments in this order:
   - First element: "test"
   - Second element: TestPath
   - Boolean flags (-v) if Verbose is true
   - Flags with values (-run=<pattern>) if RunPattern is not empty

Write comprehensive tests for BuildCommand() covering:
- Default configuration
- Verbose enabled
- RunPattern set
- Both verbose and run pattern
- Different test paths

Ensure the test output shows the command is built correctly with proper ordering.
```

## Prompt 2: Message Types

```
Building on the previous implementation, add message types to main.go:

1. Define a MessageType as a string type with constants:
   - FileChangeMsg
   - CommandMsg
   - HelpMsg
   - TestCompleteMsg

2. Define a Message interface with one method:
   - Type() MessageType

3. Implement these concrete message types that satisfy the Message interface:
   - FileChangeMessage (no fields)
   - CommandMessage (fields: Command string, Args []string)
   - HelpMessage (no fields)
   - TestCompleteMessage (no fields)

Each type should have a Type() method that returns the appropriate MessageType constant.

Write tests that:
- Verify each message type returns the correct MessageType
- Verify CommandMessage properly stores command and args
- Verify all types satisfy the Message interface using type assertions
```

## Prompt 3: Command Registry Foundation

```
Building on the existing code, implement the command registry system:

1. Define a CommandHandler type:
   - Type: func(*TestConfig, []string) error

2. Create a global commandRegistry variable:
   - Type: map[string]CommandHandler

3. Implement an initRegistry() function that:
   - Initializes the commandRegistry map
   - Should be called from init() or early in main()

4. Implement a handleCommand() function that:
   - Takes: command name (string), config (*TestConfig), args ([]string)
   - Looks up the handler in the registry
   - Returns error if command not found
   - Calls the handler and returns its result

Write tests for:
- Looking up a command that exists (use a mock handler)
- Looking up a command that doesn't exist
- Handler being called with correct arguments
- Handler errors being propagated

Don't implement actual command handlers yet - use test mocks.
```

## Prompt 4: Simple Command Handlers

```
Now implement the actual command handlers and wire them into the registry:

1. Implement handleVerbose() CommandHandler:
   - Toggles config.Verbose
   - Prints "Verbose: enabled" or "Verbose: disabled" to stdout
   - Returns nil

2. Implement handleClear() CommandHandler:
   - Resets config to defaults (TestPath: "./...", Verbose: false, RunPattern: "")
   - Prints "All parameters cleared" to stdout
   - Returns nil

3. Implement handleHelp() CommandHandler:
   - Prints help text showing all commands (format as per spec)
   - Returns nil

4. Update initRegistry() to register these handlers:
   - "v" -> handleVerbose
   - "clear" -> handleClear
   - "help" -> handleHelp

Write tests for each handler:
- Test initial state and state after handler execution
- Capture stdout to verify acknowledgment messages
- Test handleVerbose toggles correctly (on->off->on)
- Test handleClear resets all fields

Update main() to demonstrate handlers working (manually call a few).
```

## Prompt 5: Parameter Command Handlers

```
Implement the remaining command handlers with argument handling:

1. Implement handleRunPattern() CommandHandler:
   - If len(args) == 0: clear RunPattern, print "Run pattern: cleared"
   - If len(args) > 0: set RunPattern to args[0], print "Run pattern: <pattern>"
   - Ignore extra arguments beyond first
   - Returns nil

2. Implement handleTestPath() CommandHandler:
   - Requires exactly 1 argument, return error if not provided
   - Validate path exists using os.Stat()
   - Validate path is a directory using FileInfo.IsDir()
   - If valid: set TestPath, print "Test path: <path>"
   - If invalid: return error describing the issue
   - Returns nil on success, error on validation failure

3. Implement handleCls() CommandHandler:
   - Print ANSI escape sequence to clear screen: "\033[H\033[2J"
   - Returns nil

4. Implement handleRun() CommandHandler:
   - Does nothing yet (stub)
   - Returns nil

5. Update initRegistry() to add:
   - "r" -> handleRunPattern
   - "p" -> handleTestPath
   - "cls" -> handleCls
   - "run" -> handleRun

Write tests:
- handleRunPattern with pattern, without pattern, with multiple args
- handleTestPath with valid directory, invalid path, non-directory
- handleCls clears screen
- Test path validation against real temporary directories

Update main() to demonstrate all handlers.
```

## Prompt 6: Stdin Reader

```
Implement the stdin reader component:

1. Create channel types:
   - commandChan: chan CommandMessage
   - helpChan: chan HelpMessage
   - readyChan: chan bool (unbuffered)

2. Implement readStdin() function:
   - Parameters: commandChan, helpChan, readyChan
   - Runs in a goroutine
   - Uses bufio.Scanner to read from os.Stdin line by line
   - Uses select to check readyChan before processing each line
   - If readyChan receives false: block until receives true
   - Parse line using strings.Fields into command and args
   - If command == "help": send HelpMessage to helpChan
   - Otherwise: send CommandMessage{Command: cmd, Args: args} to commandChan
   - Handle empty lines gracefully (ignore)

3. Implement parseCommand() helper function:
   - Takes input string
   - Returns command (string) and args ([]string)
   - Use strings.TrimSpace and strings.Fields

Write tests:
- parseCommand with various inputs (command only, command + args, empty, whitespace)
- Mock stdin reader using bytes.Buffer and test message sending
- Test ready channel blocking behavior with mock channels
- Test help command sends HelpMessage
- Test regular commands send CommandMessage with correct parsing

Add a demo in main() that:
- Creates channels
- Starts readStdin goroutine
- Simulates ready channel states
- Reads a few mock commands
```

## Prompt 7: File Watcher with Debounce

```
Implement the file watcher with debouncing:

1. Add dependency: go get github.com/fsnotify/fsnotify

2. Create channel type:
   - fileChangeChan: chan FileChangeMessage

3. Implement watchFiles() function:
   - Parameters: ctx context.Context, fileChangeChan
   - Creates fsnotify.Watcher
   - Walks current directory recursively using filepath.WalkDir
   - Skips hidden files/directories (strings.HasPrefix(name, "."))
   - Adds directories (not individual files) to watcher
   - Maintains a *time.Timer for debouncing (nil initially)
   - On file events (Create, Write, Remove, Rename):
     - If event is for *.go file:
       - If timer == nil: create new timer (200ms)
       - If timer exists: reset timer to 200ms
   - When timer fires:
     - Send FileChangeMessage to fileChangeChan
     - Set timer back to nil
   - Handle context cancellation and cleanup watcher

4. Implement isGoFile() helper:
   - Returns true if filepath.Ext(path) == ".go"

5. Implement addWatchRecursive() helper:
   - Walks directory tree
   - Adds non-hidden directories to watcher
   - Returns error if walking fails

Write tests:
- Create temporary directory structure with .go files
- Start watcher and verify it sends FileChangeMessage after changes
- Test debouncing: multiple quick changes only trigger one message
- Test hidden directory exclusion
- Test timer reset behavior with rapid file changes
- Test graceful shutdown via context cancellation

Add demo in main() that watches current directory for 5 seconds.
```

## Prompt 8: Test Runner

```
Implement the test runner component:

1. Create channel type:
   - testCompleteChan: chan TestCompleteMessage

2. Implement runTests() function:
   - Parameters: ctx context.Context, config *TestConfig, testCompleteChan, readyChan
   - Gets command args from config.BuildCommand()
   - Prints full command to stdout: "go test <args...>"
   - Creates exec.Command("go", args...)
   - Sets up stdout/stderr pipes
   - Starts command
   - Launches two goroutines:
     - One for streaming stdout using bufio.Scanner
     - One for streaming stderr using bufio.Scanner
   - Uses sync.WaitGroup to wait for both scanner goroutines
   - Calls cmd.Wait() to wait for process completion
   - Sends TestCompleteMessage when everything is done

3. Implement streamOutput() helper:
   - Parameters: scanner *bufio.Scanner, output io.Writer
   - Reads lines from scanner
   - Writes each line to output (os.Stdout or os.Stderr)
   - Signals WaitGroup when done

Write tests:
- Mock exec.Command using a test helper that runs a simple script
- Verify stdout/stderr are properly streamed
- Verify TestCompleteMessage is sent after completion
- Verify command built correctly from TestConfig
- Test with command that has both stdout and stderr output
- Test WaitGroup properly waits for both scanners

Add demo in main() that:
- Creates test config
- Runs tests with "go version" as a simple test
- Waits for completion message
```

## Prompt 9: Dispatcher & Integration

```
Wire all components together with the central dispatcher:

1. Implement dispatcher() function:
   - Parameters: ctx, config *TestConfig, all channel types
   - Maintains testRunning bool flag
   - Uses select statement to receive from:
     - fileChangeChan
     - commandChan
     - helpChan
     - testCompleteChan
     - ctx.Done()
   - On FileChangeMessage:
     - If !testRunning: spawn runTests goroutine, set testRunning=true, send false to readyChan
   - On CommandMessage:
     - Call handleCommand(msg.Command, config, msg.Args)
     - If error: print to stderr
     - If !testRunning: spawn runTests goroutine, set testRunning=true, send false to readyChan
   - On HelpMessage:
     - Call handleCommand("help", config, nil)
     - Do NOT spawn test runner
   - On TestCompleteMessage:
     - Set testRunning=false
     - Send true to readyChan
     - Print blank line then prompt "> "
   - On ctx.Done():
     - Wait for testRunning to become false
     - Print shutdown message
     - Return

2. Update main() to:
   - Create context with signal handling (SIGINT, SIGTERM)
   - Initialize TestConfig with defaults
   - Create all channels
   - Start watchFiles goroutine
   - Start readStdin goroutine
   - Start dispatcher (runs in main goroutine)
   - Block until context cancelled

Write tests:
- Test dispatcher handles FileChangeMessage correctly
- Test dispatcher spawns test runner on command
- Test dispatcher doesn't spawn test runner on help
- Test testRunning flag prevents concurrent test runs
- Test ready channel gets correct values
- Mock all goroutines and verify message flow

This should create a working end-to-end system.
```

## Prompt 10: Context & Lifecycle

```
Enhance the context management and graceful shutdown:

1. Create context key type and value for TestConfig:
   - type configKey struct{}
   - Store config pointer in context
   - Add helper functions:
     - getConfig(ctx context.Context) *TestConfig
     - withConfig(ctx context.Context, config *TestConfig) context.Context

2. Update all component functions to:
   - Accept context as first parameter
   - Retrieve TestConfig from context where needed
   - Remove redundant config parameters

3. Implement setupSignalHandler() function:
   - Creates context with cancel
   - Listens for SIGINT and SIGTERM using signal.Notify
   - Returns context and cancel function
   - On signal: calls cancel() and prints "Shutting down..."

4. Update dispatcher() to:
   - Check testRunning in ctx.Done() case
   - If testRunning: wait for TestCompleteMessage before returning
   - Add timeout (5 seconds) for graceful shutdown
   - If timeout: force exit

5. Update main() to:
   - Use setupSignalHandler() for context creation
   - Store config in context using withConfig()
   - Defer cancel() call
   - Pass context to all components

Write tests:
- Test config storage/retrieval in context
- Test signal handling triggers context cancellation
- Test dispatcher waits for test completion on shutdown
- Test graceful shutdown timeout
- Integration test: start system, send signal, verify clean shutdown

Update main() to use proper lifecycle management.
```

## Prompt 11: Startup Behavior

```
Implement the startup sequence and initial test run:

1. Update main() to implement startup flow:
   - Print "Test watcher started" using log/slog
   - Create TestConfig with defaults
   - Store config in context
   - Run initial test synchronously before starting goroutines:
     - Create testCompleteChan (just for this initial run)
     - Create readyChan (but don't use it for startup)
     - Call runTests() directly (not as goroutine)
     - Wait for TestCompleteMessage
   - Print blank line then prompt "> "
   - NOW start all goroutines (watcher, stdin, dispatcher)
   - Wait for context cancellation

2. Update watchFiles() to:
   - Accept a startWatching <-chan struct{} parameter
   - Block on receiving from this channel before beginning to watch
   - This allows us to delay watching until after initial test

3. Update main() to:
   - Create startWatching channel
   - Pass to watchFiles goroutine
   - After initial test completes, close startWatching channel

Write tests:
- Test initial test runs before watcher starts
- Test watcher doesn't send messages until startWatching received
- Test prompt appears after initial test
- Integration test of full startup sequence

Verify the complete startup experience works correctly.
```

## Prompt 12: Output Formatting & Final Polish

```
Add final output formatting and polish:

1. Create displayPrompt() function:
   - Prints "\n> " (blank line then prompt)
   - Flushes output

2. Create displayCommand() function:
   - Takes []string of command parts
   - Prints "go <parts joined by space>"

3. Update runTests() to:
   - Call displayCommand() before executing
   - Don't print command inline anymore

4. Update dispatcher() to:
   - Call displayPrompt() after TestCompleteMessage
   - Call displayPrompt() after initial startup

5. Add slog configuration:
   - Configure slog to print simple text without timestamps
   - Use for startup message only

6. Final cleanup:
   - Ensure all acknowledgment messages go to stdout
   - Ensure all errors go to stderr  
   - Ensure proper spacing around all output
   - Remove any debug prints

7. Add usage instructions:
   - If run with -h or --help flag: print usage and exit
   - Usage: gotest-watch [options]
   - Currently no options, just shows help

Write tests:
- Test displayPrompt output format
- Test displayCommand output format
- Integration test of complete output flow
- Manual testing of full user experience

Perform thorough manual testing:
- Start tool, see initial test run
- Type "v", see verbose toggle and test run
- Type "r TestFoo", see run pattern and test run
- Type "clear", see reset and test run
- Type "help", see help (no test run)
- Save a .go file, see file change trigger test
- Press Ctrl+C, see graceful shutdown
- Verify input blocked during tests
- Verify prompt appears at right times

This completes the implementation!
```

---

# Testing Strategy Notes

Each prompt should result in:
1. **Compilable code** that runs
2. **Passing tests** for the new functionality
3. **Integration** with previous steps
4. **No orphaned code** - everything is wired up

## Key Testing Principles

- **Unit tests first**: Test each component in isolation
- **Mock dependencies**: Use channels, interfaces, temporary files
- **Integration tests**: Test component interactions
- **Manual testing**: Final verification of user experience

## Running Tests Between Steps

After each step, run:
```bash
go test -v ./...
go build
./gotest-watch  # Manual testing
```

This ensures we catch issues early and maintain a working system throughout development.
