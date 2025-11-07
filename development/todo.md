# gotest-watch Implementation Checklist

## Phase 1: Foundation

### Step 1: Project Setup & TestConfig

- [x] Initialize Go module with `go mod init github.com/mikowitz/gotest-watch`
- [x] Set minimum Go version to 1.23 in go.mod
- [x] Create main.go with basic main function
- [x] Add "Test watcher started" print statement
- [x] Define TestConfig struct with fields:
  - [x] TestPath string
  - [x] Verbose bool
  - [x] RunPattern string
- [x] Implement BuildCommand() method on TestConfig
  - [x] Return string with concatenated command
  - [x] Order: "test", TestPath, boolean flags, flags with values
  - [x] Add -v if Verbose is true
  - [x] Add -run=<pattern> if RunPattern is not empty
- [x] Write tests for BuildCommand():
  - [x] Test default configuration (no flags)
  - [x] Test with Verbose enabled
  - [x] Test with RunPattern set
  - [x] Test with both Verbose and RunPattern
  - [x] Test with different test paths
  - [x] Verify proper flag ordering
- [x] Run tests: `go test -v`
- [x] Build and verify: `go build`

### Step 2: Message Types

- [x] Define MessageType as string type
- [x] Create MessageType constants:
  - [x] FileChangeMsg
  - [x] CommandMsg
  - [x] HelpMsg
  - [x] TestCompleteMsg
- [x] Define Message interface with Type() method
- [x] Implement FileChangeMessage struct
  - [x] Add Type() method returning FileChangeMsg
- [x] Implement CommandMessage struct
  - [x] Add Command string field
  - [x] Add Args []string field
  - [x] Add Type() method returning CommandMsg
- [x] Implement HelpMessage struct
  - [x] Add Type() method returning HelpMsg
- [x] Implement TestCompleteMessage struct
  - [x] Add Type() method returning TestCompleteMsg
- [x] Write tests:
  - [x] Verify FileChangeMessage.Type() returns correct type
  - [x] Verify CommandMessage.Type() returns correct type
  - [x] Verify HelpMessage.Type() returns correct type
  - [x] Verify TestCompleteMessage.Type() returns correct type
  - [x] Test CommandMessage stores command and args correctly
  - [x] Test all types satisfy Message interface (type assertions)
- [x] Run tests: `go test -v`

### Step 3: Command Registry Foundation

- [x] Define CommandHandler function type: `func(*TestConfig, []string) error`
- [x] Create global commandRegistry variable: `map[string]CommandHandler`
- [x] Implement initRegistry() function
  - [x] Initialize empty commandRegistry map
  - [x] Call from init() or main()
- [x] Implement handleCommand() function
  - [x] Accept command name, config pointer, args
  - [x] Look up handler in registry
  - [x] Return error if command not found
  - [x] Call handler with config and args
  - [x] Return handler result
- [x] Write tests:
  - [x] Test looking up existing command (use mock handler)
  - [x] Test looking up non-existent command (verify error)
  - [x] Test handler is called with correct arguments
  - [x] Test handler errors are propagated
  - [x] Create test mock handlers for verification
- [x] Run tests: `go test -v`

## Phase 2: Core Components

### Step 4: Simple Command Handlers

- [x] Implement handleVerbose() CommandHandler
  - [x] Toggle config.Verbose
  - [x] Print "Verbose: enabled" or "Verbose: disabled"
  - [x] Return nil
- [x] Implement handleClear() CommandHandler
  - [x] Reset TestPath to "./..."
  - [x] Reset Verbose to false
  - [x] Reset RunPattern to ""
  - [x] Print "All parameters cleared"
  - [x] Return nil
- [x] Implement handleHelp() CommandHandler
  - [x] Print help text with all commands
  - [x] Format as per spec (detailed format)
  - [x] Return nil
- [x] Update initRegistry() to register handlers:
  - [x] Register "v" -> handleVerbose
  - [x] Register "clear" -> handleClear
  - [x] Register "h" -> handleHelp
- [x] Write tests for handleVerbose:
  - [x] Test toggles from false to true
  - [x] Test toggles from true to false
  - [x] Capture stdout and verify acknowledgment message
- [x] Write tests for handleClear:
  - [x] Set non-default values
  - [x] Call handler
  - [x] Verify all fields reset to defaults
  - [x] Capture stdout and verify acknowledgment
- [x] Write tests for handleHelp:
  - [x] Capture stdout
  - [x] Verify help text contains all commands
  - [x] Verify formatting is correct
- [x] Update main() to demonstrate handlers
  - [x] Create TestConfig
  - [x] Call a few handlers manually
  - [x] Verify output
- [x] Run tests: `go test -v`
- [x] Build and run: `go build && ./gotest-watch`

### Step 5: Parameter Command Handlers

- [x] Implement handleRunPattern() CommandHandler
  - [x] If no args: clear RunPattern, print "Run pattern: cleared"
  - [x] If args provided: set RunPattern to args[0]
  - [x] Print "Run pattern: <pattern>"
  - [x] Ignore extra arguments
  - [x] Return nil
- [x] Implement handleTestPath() CommandHandler
  - [x] Require exactly 1 argument
  - [x] Return error if no argument provided
  - [x] Validate path exists with os.Stat()
  - [x] Validate path is directory with FileInfo.IsDir()
  - [x] If valid: set TestPath, print "Test path: <path>"
  - [x] If invalid: return descriptive error
- [x] Implement handleCls() CommandHandler
  - [x] Print ANSI escape sequence: "\033[H\033[2J"
  - [x] Return nil
- [x] Implement handleRun() CommandHandler (stub)
  - [x] Do nothing for now
  - [x] Return nil
- [x] Update initRegistry() to add:
  - [x] Register "r" -> handleRunPattern
  - [x] Register "p" -> handleTestPath
  - [x] Register "cls" -> handleCls
  - [x] Register "f" -> handleRun
- [x] Write tests for handleRunPattern:
  - [x] Test with pattern argument
  - [x] Test without arguments (clears pattern)
  - [x] Test with multiple arguments (uses first, ignores rest)
  - [x] Capture stdout and verify acknowledgment
- [x] Write tests for handleTestPath:
  - [x] Create temporary test directory
  - [x] Test with valid directory path
  - [x] Test with invalid/non-existent path
  - [x] Test with file path (not directory)
  - [x] Test with no arguments (error)
  - [x] Verify error messages
- [x] Write tests for handleCls:
  - [x] Capture stdout
  - [x] Verify ANSI escape sequence printed
- [x] Write tests for handleRun:
  - [x] Verify it returns nil (stub)
- [x] Update main() to demonstrate new handlers
- [x] Run tests: `go test -v`
- [x] Build and run: `go build && ./gotest-watch`

### Step 6: Stdin Reader

- [x] Define channel types (at package level):
  - [x] commandChan: chan CommandMessage
  - [x] helpChan: chan HelpMessage
  - [x] readyChan: chan bool (buffered, capacity 1)
- [x] Implement parseCommand() helper function
  - [x] Accept input string
  - [x] Use strings.TrimSpace to clean input
  - [x] Use strings.Fields to split into parts
  - [x] Return command (first part) and args (rest)
  - [x] Handle empty input (return empty string, nil slice)
- [x] Implement readStdin() function
  - [x] Accept commandChan, helpChan, readyChan as parameters
  - [x] Create bufio.Scanner from io.Reader (generic, not os.Stdin specific)
  - [x] Loop: read lines with Scanner.Scan()
  - [x] Use select to check readyChan before processing
  - [x] If readyChan receives false: block until receives true
  - [x] Parse each line with parseCommand()
  - [x] If command == "help": send HelpMessage to helpChan
  - [x] Otherwise: send CommandMessage to commandChan
  - [x] Handle empty lines (ignore)
  - [x] Handle Scanner errors
  - [x] Support context cancellation throughout
- [x] Write tests for parseCommand:
  - [x] Test command only (no args)
  - [x] Test command with single arg
  - [x] Test command with multiple args
  - [x] Test empty string
  - [x] Test whitespace only
  - [x] Test leading/trailing whitespace
  - [x] Additional edge cases (newlines, carriage returns, mixed whitespace)
  - [x] Real-world command examples
- [x] Write tests for readStdin:
  - [x] Mock stdin with strings.NewReader and io.Pipe
  - [x] Test "help" command sends HelpMessage
  - [x] Test regular command sends CommandMessage
  - [x] Test command parsing (verify Command and Args fields)
  - [x] Test ready channel blocking (with proper buffered channel)
  - [x] Test empty lines are ignored
  - [x] Test multiple commands
  - [x] Test context cancellation
- [x] Add demo in main():
  - [x] Create channels (buffered to prevent blocking)
  - [x] Start readStdin goroutine
  - [x] Send ready=true
  - [x] Create TestConfig for command handlers
  - [x] Receive and handle messages in infinite loop
  - [x] Execute commands via handleCommand registry
  - [x] Handle help messages via handleHelp
- [x] Run tests: `go test -v` (61 tests passing)
- [x] Build and run: `go build && ./gotest-watch` (interactive CLI working)

### Step 7: File Watcher with Debounce

- [x] Add fsnotify dependency: `go get github.com/fsnotify/fsnotify`
- [ ] Define channel type:
  - [ ] fileChangeChan: chan FileChangeMessage
- [ ] Implement isGoFile() helper
  - [ ] Accept file path string
  - [ ] Return true if filepath.Ext(path) == ".go"
- [ ] Implement addWatchRecursive() helper
  - [ ] Accept watcher and root path
  - [ ] Use filepath.WalkDir to walk directory tree
  - [ ] Skip hidden files/directories (check strings.HasPrefix(name, "."))
  - [ ] Add each non-hidden directory to watcher
  - [ ] Return error if walking fails
- [ ] Implement watchFiles() function
  - [ ] Accept ctx context.Context and fileChangeChan
  - [ ] Create fsnotify.Watcher
  - [ ] Defer watcher.Close()
  - [ ] Call addWatchRecursive() for current directory
  - [ ] Initialize debounce timer as nil
  - [ ] Loop with select:
    - [ ] Handle watcher.Events:
      - [ ] Check if event is for .go file (isGoFile)
      - [ ] On Create, Write, Remove, Rename:
        - [ ] If timer is nil: create 200ms timer
        - [ ] If timer exists: stop and reset to 200ms
    - [ ] Handle timer.C (when timer fires):
      - [ ] Send FileChangeMessage to fileChangeChan
      - [ ] Set timer back to nil
    - [ ] Handle watcher.Errors:
      - [ ] Log error but continue (graceful handling)
    - [ ] Handle ctx.Done():
      - [ ] Stop timer if exists
      - [ ] Return
- [x] Write tests (18 test functions in file_watcher_test.go):
  - [x] Test isGoFile with .go extensions
  - [x] Test isGoFile with non-.go extensions
  - [x] Test isGoFile edge cases (hidden files, case sensitivity, etc.)
  - [x] Test addWatchRecursive with simple directory
  - [x] Test addWatchRecursive with nested directories
  - [x] Test addWatchRecursive excludes hidden directories
  - [x] Test addWatchRecursive error handling
  - [x] Test watchFiles detects .go file creation
  - [x] Test watchFiles detects .go file modification
  - [x] Test watchFiles ignores non-.go files
  - [x] Test watchFiles debouncing (multiple rapid changes)
  - [x] Test watchFiles timer reset on subsequent changes
  - [x] Test watchFiles handles nested directories
  - [x] Test watchFiles ignores hidden directories
  - [x] Test watchFiles context cancellation
  - [x] Test watchFiles with mixed file types
  - [x] Test watchFiles file removal detection
- [x] Add integration in main():
  - [x] Create fileChangeChan (buffered, capacity 10)
  - [x] Start watchFiles goroutine watching current directory
  - [x] Add case in select loop to handle FileChangeMessage
  - [x] Print message when file changes detected
- [x] Run tests: `go test -v` (78 tests passing, including 18 file watcher tests)
- [x] Build and run: `go build && ./gotest-watch` (builds successfully)

### Step 8: Test Runner

- [x] Define channel type:
  - [x] testCompleteChan: chan TestCompleteMessage
- [x] Implement streamOutput() helper
  - [x] Accept bufio.Scanner and io.Writer
  - [x] Accept sync.WaitGroup
  - [x] Defer wg.Done()
  - [x] Loop: Scanner.Scan()
  - [x] Write each line to output
  - [x] Handle Scanner errors
- [x] Implement runTests() function
  - [x] Accept ctx, config, testCompleteChan, readyChan
  - [x] Get command args from config.BuildCommand()
  - [x] Print full command: "go test <args...>"
  - [x] Create exec.Command("go", args...)
  - [x] Get stdout pipe with cmd.StdoutPipe()
  - [x] Get stderr pipe with cmd.StderrPipe()
  - [x] Start command with cmd.Start()
  - [x] Create WaitGroup with count 2
  - [x] Launch goroutine for stdout streaming
    - [x] Create bufio.Scanner from stdout pipe
    - [x] Call streamOutput with scanner, os.Stdout, wg
  - [x] Launch goroutine for stderr streaming
    - [x] Create bufio.Scanner from stderr pipe
    - [x] Call streamOutput with scanner, os.Stderr, wg
  - [x] Wait for both scanner goroutines (wg.Wait())
  - [x] Wait for command to finish (cmd.Wait())
  - [x] Send TestCompleteMessage to testCompleteChan
- [x] Write tests (20 test functions in test_runner_test.go):
  - [x] Test streamOutput reads all lines
  - [x] Test streamOutput calls wg.Done()
  - [x] Test streamOutput handles empty input
  - [x] Test streamOutput preserves line content
  - [x] Test streamOutput handles scanner errors
  - [x] Test streamOutput writes line by line
  - [x] Test streamOutput concurrent safety
  - [x] Test runTests sends TestCompleteMessage
  - [x] Test runTests builds correct command from config
  - [x] Test runTests streams stdout and stderr
  - [x] Test runTests handles command failure (non-zero exit)
  - [x] Test runTests waits for both streamers (WaitGroup)
  - [x] Test runTests displays command before running
  - [x] Test runTests context cancellation
  - [x] Test runTests uses correct go command
  - [x] Test runTests creates stdout pipe
  - [x] Test runTests creates stderr pipe
  - [x] Test runTests integration with various TestConfig combinations
- [x] Add demo in main():
  - [x] Create TestConfig
  - [x] Create channels
  - [x] Run "go version" as simple test command
  - [x] Wait for TestCompleteMessage
  - [x] Print completion
- [x] Run tests: `go test -v`
- [x] Build and run: `go build && ./gotest-watch`

## Phase 3: Integration

### Step 9: Dispatcher & Integration

- [x] Implement dispatcher() function
  - [x] Accept ctx, config, all channel types
  - [x] Initialize testRunning bool to false
  - [x] Create infinite loop with select:
    - [x] Case fileChangeChan receive:
      - [x] If !testRunning:
        - [x] Spawn runTests goroutine
        - [x] Set testRunning = true
        - [x] Send false to readyChan
    - [x] Case commandChan receive:
      - [x] Call handleCommand with msg.Command, config, msg.Args
      - [x] If error: print to stderr
      - [x] If !testRunning:
        - [x] Spawn runTests goroutine
        - [x] Set testRunning = true
        - [x] Send false to readyChan
    - [x] Case helpChan receive:
      - [x] Call handleCommand("help", config, nil)
      - [x] Do NOT spawn test runner
      - [x] Do NOT change testRunning
    - [x] Case testCompleteChan receive:
      - [x] Set testRunning = false
      - [x] Send true to readyChan
      - [x] Print blank line
      - [x] Print prompt "> "
    - [x] Case ctx.Done():
      - [x] If testRunning: wait for TestCompleteMessage
      - [x] Print shutdown message
      - [x] Return
- [x] Update main() to wire everything:
  - [x] Create context with signal.NotifyContext (SIGINT, SIGTERM)
  - [x] Defer cancel()
  - [x] Initialize TestConfig with defaults
  - [x] Create all channels:
    - [x] fileChangeChan
    - [x] commandChan
    - [x] helpChan
    - [x] testCompleteChan
    - [x] readyChan
  - [x] Start watchFiles goroutine
  - [x] Start readStdin goroutine
  - [x] Call dispatcher (blocks in main goroutine)
- [x] Write tests:
  - [x] Mock all channels
  - [x] Test FileChangeMessage spawns test runner
  - [x] Test FileChangeMessage ignored when testRunning=true
  - [x] Test CommandMessage calls handler
  - [x] Test CommandMessage spawns test runner
  - [x] Test CommandMessage ignored when testRunning=true
  - [x] Test HelpMessage doesn't spawn test runner
  - [x] Test TestCompleteMessage updates state and re-enables stdin
  - [x] Test ctx.Done() waits for tests to finish
  - [x] Test ready channel receives correct values
- [x] Run tests: `go test -v`
- [x] Build and run: `go build && ./gotest-watch`
- [x] Manual test: start tool, type commands, save files, verify behavior

### Step 10: Context & Lifecycle

- [ ] Define context key type: `type configKey struct{}`
- [ ] Implement withConfig() function
  - [ ] Accept ctx and config pointer
  - [ ] Return context.WithValue(ctx, configKey{}, config)
- [ ] Implement getConfig() function
  - [ ] Accept ctx
  - [ ] Get value with ctx.Value(configKey{})
  - [ ] Type assert to *TestConfig
  - [ ] Return config or nil if not found
- [ ] Implement setupSignalHandler() function
  - [ ] Create context.WithCancel(context.Background())
  - [ ] Create signal channel with signal.Notify
  - [ ] Listen for SIGINT and SIGTERM
  - [ ] Start goroutine:
    - [ ] Wait for signal
    - [ ] Print "Shutting down..."
    - [ ] Call cancel()
  - [ ] Return context and cancel function
- [ ] Update watchFiles() signature
  - [ ] Accept context as first parameter
  - [ ] Remove any redundant config parameter
  - [ ] Get config from context if needed
- [ ] Update readStdin() signature
  - [ ] Accept context as first parameter
  - [ ] Get config from context if needed
- [ ] Update runTests() signature
  - [ ] Accept context as first parameter
  - [ ] Get config from context at start of function
  - [ ] Use this config throughout
- [ ] Update dispatcher() to handle graceful shutdown
  - [ ] In ctx.Done() case:
    - [ ] If testRunning: enter wait loop
    - [ ] Create timeout (5 seconds) with time.After
    - [ ] Select between testCompleteChan and timeout
    - [ ] If timeout: force exit with os.Exit(1)
    - [ ] If TestCompleteMessage: clean exit
  - [ ] Print shutdown message
  - [ ] Return normally
- [ ] Update main() to use new context system
  - [ ] Call setupSignalHandler() for context
  - [ ] Store config in context with withConfig()
  - [ ] Defer cancel()
  - [ ] Pass context to all component functions
  - [ ] Remove redundant config parameters from calls
- [ ] Write tests:
  - [ ] Test withConfig stores config in context
  - [ ] Test getConfig retrieves config from context
  - [ ] Test getConfig returns nil if not in context
  - [ ] Mock signal and test setupSignalHandler
  - [ ] Test dispatcher waits for test completion on shutdown
  - [ ] Test graceful shutdown timeout (force exit)
  - [ ] Integration test: start, signal, verify clean shutdown
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`
- [ ] Test signal handling: start tool, press Ctrl+C

### Step 11: Startup Behavior

- [ ] Update watchFiles() signature
  - [ ] Add startWatching <-chan struct{} parameter
  - [ ] At start of function: block on <-startWatching
  - [ ] Then proceed with normal watching logic
- [ ] Update main() for startup sequence
  - [ ] Print "Test watcher started" with log/slog
  - [ ] Create TestConfig with defaults
  - [ ] Create context with setupSignalHandler()
  - [ ] Store config in context with withConfig()
  - [ ] Create startWatching channel (unbuffered)
  - [ ] Create testCompleteChan for initial run
  - [ ] Create readyChan (but don't use for startup)
  - [ ] Call runTests() synchronously (not as goroutine)
  - [ ] Wait for TestCompleteMessage
  - [ ] Print blank line
  - [ ] Print prompt "> "
  - [ ] Create all channels for main loop
  - [ ] Start watchFiles goroutine with startWatching param
  - [ ] Start readStdin goroutine
  - [ ] Close startWatching channel (unblocks watcher)
  - [ ] Call dispatcher
- [ ] Write tests:
  - [ ] Test watcher blocks until startWatching closes
  - [ ] Test initial test runs before watcher starts
  - [ ] Test watcher doesn't send messages during initial test
  - [ ] Test prompt appears after initial test
  - [ ] Integration test of full startup sequence
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`
- [ ] Manual test: verify initial test runs, then prompt appears

## Phase 4: Polish

### Step 12: Output Formatting & Final Polish

- [ ] Implement displayPrompt() function
  - [ ] Print "\n> " (blank line then prompt)
  - [ ] Flush output with os.Stdout.Sync() if needed
- [ ] Implement displayCommand() function
  - [ ] Accept []string of command parts
  - [ ] Join with spaces: strings.Join(parts, " ")
  - [ ] Print "go <joined parts>"
- [ ] Update runTests() to use displayCommand
  - [ ] Build command args with config.BuildCommand()
  - [ ] Call displayCommand(args) before executing
  - [ ] Remove inline command printing
- [ ] Update dispatcher() to use displayPrompt
  - [ ] Call displayPrompt() in TestCompleteMessage case
  - [ ] Remove inline prompt printing
- [ ] Update main() to use displayPrompt
  - [ ] Call displayPrompt() after initial test
  - [ ] Remove inline prompt printing
- [ ] Configure slog for startup
  - [ ] Create text handler without timestamps
  - [ ] Set as default logger
  - [ ] Use for startup message only
- [ ] Add usage/help flag handling
  - [ ] Check os.Args for "-h" or "--help"
  - [ ] If found: print usage message and exit
  - [ ] Usage: "gotest-watch [options]"
  - [ ] Note: currently no options supported
- [ ] Final cleanup pass:
  - [ ] Review all print statements
  - [ ] Ensure acknowledgments go to stdout
  - [ ] Ensure errors go to stderr
  - [ ] Remove any debug prints
  - [ ] Verify spacing is consistent
  - [ ] Check all error messages are clear
- [ ] Write tests:
  - [ ] Test displayPrompt output format
  - [ ] Test displayCommand output format
  - [ ] Capture output and verify exact format
  - [ ] Integration test of complete output flow
- [ ] Run tests: `go test -v`
- [ ] Build: `go build`
- [ ] Manual testing checklist:
  - [ ] Start tool, verify startup message
  - [ ] Verify initial test runs
  - [ ] Verify prompt appears after initial test
  - [ ] Type "v", verify toggle and test run
  - [ ] Type "r TestFoo", verify pattern set and test run
  - [ ] Type "r", verify pattern cleared and test run
  - [ ] Type "p .", verify path set and test run
  - [ ] Type "p /invalid", verify error message
  - [ ] Type "clear", verify reset and test run
  - [ ] Type "cls", verify screen clears
  - [ ] Type "help", verify help displays (no test run)
  - [ ] Type "run", verify tests run
  - [ ] Type "invalid", verify error message
  - [ ] Save a .go file, verify file change triggers test
  - [ ] Verify input blocked while tests run
  - [ ] Press Ctrl+C during test, verify waits then exits
  - [ ] Press Ctrl+C while idle, verify immediate exit
  - [ ] Verify all spacing is correct
  - [ ] Verify prompt appears at right times

## Final Steps

### Code Quality

- [ ] Run `go fmt ./...`
- [ ] Run `go vet ./...`
- [ ] Run `golangci-lint run` (if available)
- [ ] Review all error handling
- [ ] Review all resource cleanup (defer statements)
- [ ] Check for potential race conditions
- [ ] Verify all goroutines properly terminated on shutdown

### Documentation

- [ ] Add package documentation comment in main.go
- [ ] Add function documentation comments
- [ ] Create README.md with:
  - [ ] Project description
  - [ ] Installation instructions
  - [ ] Usage examples
  - [ ] Available commands
  - [ ] Build instructions
- [ ] Add LICENSE file if needed
- [ ] Add .gitignore file

### Testing

- [ ] Achieve >80% test coverage
- [ ] Run `go test -race ./...` to check for race conditions
- [ ] Run `go test -cover ./...` to check coverage
- [ ] Test on different platforms if possible (Linux, macOS, Windows)

### Release Preparation

- [ ] Tag first version: `git tag v0.1.0`
- [ ] Build for multiple platforms if needed
- [ ] Test binary distribution
- [ ] Consider adding install script
- [ ] Update documentation with version info

## Future Enhancements (Not in Initial Version)

- [ ] Add command aliases (f for run, h for help)
- [ ] Add output colorization
- [ ] Add -race flag support
- [ ] Add -cover flag support
- [ ] Add configuration file support (.gotest-watch.yaml)
- [ ] Add file logging option
- [ ] Add verbose logging mode
- [ ] Add custom test command support
- [ ] Add timestamps on messages
- [ ] Add test result summaries
- [ ] Add watch pattern customization
- [ ] Add exclude pattern support
- [ ] Add notification support (desktop notifications)
- [ ] Add web UI dashboard
