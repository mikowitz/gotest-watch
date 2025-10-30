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

- [ ] Define CommandHandler function type: `func(*TestConfig, []string) error`
- [ ] Create global commandRegistry variable: `map[string]CommandHandler`
- [ ] Implement initRegistry() function
  - [ ] Initialize empty commandRegistry map
  - [ ] Call from init() or main()
- [ ] Implement handleCommand() function
  - [ ] Accept command name, config pointer, args
  - [ ] Look up handler in registry
  - [ ] Return error if command not found
  - [ ] Call handler with config and args
  - [ ] Return handler result
- [ ] Write tests:
  - [ ] Test looking up existing command (use mock handler)
  - [ ] Test looking up non-existent command (verify error)
  - [ ] Test handler is called with correct arguments
  - [ ] Test handler errors are propagated
  - [ ] Create test mock handlers for verification
- [ ] Run tests: `go test -v`

## Phase 2: Core Components

### Step 4: Simple Command Handlers

- [ ] Implement handleVerbose() CommandHandler
  - [ ] Toggle config.Verbose
  - [ ] Print "Verbose: enabled" or "Verbose: disabled"
  - [ ] Return nil
- [ ] Implement handleClear() CommandHandler
  - [ ] Reset TestPath to "./..."
  - [ ] Reset Verbose to false
  - [ ] Reset RunPattern to ""
  - [ ] Print "All parameters cleared"
  - [ ] Return nil
- [ ] Implement handleHelp() CommandHandler
  - [ ] Print help text with all commands
  - [ ] Format as per spec (detailed format)
  - [ ] Return nil
- [ ] Update initRegistry() to register handlers:
  - [ ] Register "v" -> handleVerbose
  - [ ] Register "clear" -> handleClear
  - [ ] Register "help" -> handleHelp
- [ ] Write tests for handleVerbose:
  - [ ] Test toggles from false to true
  - [ ] Test toggles from true to false
  - [ ] Capture stdout and verify acknowledgment message
- [ ] Write tests for handleClear:
  - [ ] Set non-default values
  - [ ] Call handler
  - [ ] Verify all fields reset to defaults
  - [ ] Capture stdout and verify acknowledgment
- [ ] Write tests for handleHelp:
  - [ ] Capture stdout
  - [ ] Verify help text contains all commands
  - [ ] Verify formatting is correct
- [ ] Update main() to demonstrate handlers
  - [ ] Create TestConfig
  - [ ] Call a few handlers manually
  - [ ] Verify output
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`

### Step 5: Parameter Command Handlers

- [ ] Implement handleRunPattern() CommandHandler
  - [ ] If no args: clear RunPattern, print "Run pattern: cleared"
  - [ ] If args provided: set RunPattern to args[0]
  - [ ] Print "Run pattern: <pattern>"
  - [ ] Ignore extra arguments
  - [ ] Return nil
- [ ] Implement handleTestPath() CommandHandler
  - [ ] Require exactly 1 argument
  - [ ] Return error if no argument provided
  - [ ] Validate path exists with os.Stat()
  - [ ] Validate path is directory with FileInfo.IsDir()
  - [ ] If valid: set TestPath, print "Test path: <path>"
  - [ ] If invalid: return descriptive error
- [ ] Implement handleCls() CommandHandler
  - [ ] Print ANSI escape sequence: "\033[H\033[2J"
  - [ ] Return nil
- [ ] Implement handleRun() CommandHandler (stub)
  - [ ] Do nothing for now
  - [ ] Return nil
- [ ] Update initRegistry() to add:
  - [ ] Register "r" -> handleRunPattern
  - [ ] Register "p" -> handleTestPath
  - [ ] Register "cls" -> handleCls
  - [ ] Register "run" -> handleRun
- [ ] Write tests for handleRunPattern:
  - [ ] Test with pattern argument
  - [ ] Test without arguments (clears pattern)
  - [ ] Test with multiple arguments (uses first, ignores rest)
  - [ ] Capture stdout and verify acknowledgment
- [ ] Write tests for handleTestPath:
  - [ ] Create temporary test directory
  - [ ] Test with valid directory path
  - [ ] Test with invalid/non-existent path
  - [ ] Test with file path (not directory)
  - [ ] Test with no arguments (error)
  - [ ] Verify error messages
- [ ] Write tests for handleCls:
  - [ ] Capture stdout
  - [ ] Verify ANSI escape sequence printed
- [ ] Write tests for handleRun:
  - [ ] Verify it returns nil (stub)
- [ ] Update main() to demonstrate new handlers
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`

### Step 6: Stdin Reader

- [ ] Define channel types (at package level):
  - [ ] commandChan: chan CommandMessage
  - [ ] helpChan: chan HelpMessage
  - [ ] readyChan: chan bool (unbuffered)
- [ ] Implement parseCommand() helper function
  - [ ] Accept input string
  - [ ] Use strings.TrimSpace to clean input
  - [ ] Use strings.Fields to split into parts
  - [ ] Return command (first part) and args (rest)
  - [ ] Handle empty input (return empty string, nil slice)
- [ ] Implement readStdin() function
  - [ ] Accept commandChan, helpChan, readyChan as parameters
  - [ ] Create bufio.Scanner from os.Stdin
  - [ ] Loop: read lines with Scanner.Scan()
  - [ ] Use select to check readyChan before processing
  - [ ] If readyChan receives false: block until receives true
  - [ ] Parse each line with parseCommand()
  - [ ] If command == "help": send HelpMessage to helpChan
  - [ ] Otherwise: send CommandMessage to commandChan
  - [ ] Handle empty lines (ignore)
  - [ ] Handle Scanner errors
- [ ] Write tests for parseCommand:
  - [ ] Test command only (no args)
  - [ ] Test command with single arg
  - [ ] Test command with multiple args
  - [ ] Test empty string
  - [ ] Test whitespace only
  - [ ] Test leading/trailing whitespace
- [ ] Write tests for readStdin:
  - [ ] Mock stdin with bytes.Buffer
  - [ ] Test "help" command sends HelpMessage
  - [ ] Test regular command sends CommandMessage
  - [ ] Test command parsing (verify Command and Args fields)
  - [ ] Test ready channel blocking (mock channel, verify behavior)
  - [ ] Test empty lines are ignored
- [ ] Add demo in main():
  - [ ] Create channels
  - [ ] Start readStdin goroutine
  - [ ] Send ready=true
  - [ ] Simulate a few commands with mock stdin
  - [ ] Receive and print messages
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`

### Step 7: File Watcher with Debounce

- [ ] Add fsnotify dependency: `go get github.com/fsnotify/fsnotify`
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
- [ ] Write tests:
  - [ ] Create temporary directory with subdirectories
  - [ ] Create .go files in temp directory
  - [ ] Start watcher in test
  - [ ] Modify a .go file
  - [ ] Verify FileChangeMessage received
  - [ ] Test debouncing: modify multiple files rapidly
  - [ ] Verify only one message received after 200ms
  - [ ] Test hidden directory exclusion
  - [ ] Create .hidden directory with .go files
  - [ ] Verify changes in hidden dir don't trigger messages
  - [ ] Test timer reset: multiple changes extend wait time
  - [ ] Test context cancellation stops watcher
- [ ] Add demo in main():
  - [ ] Create context with 5 second timeout
  - [ ] Create fileChangeChan
  - [ ] Start watchFiles goroutine
  - [ ] Wait for messages or timeout
  - [ ] Print any received messages
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`

### Step 8: Test Runner

- [ ] Define channel type:
  - [ ] testCompleteChan: chan TestCompleteMessage
- [ ] Implement streamOutput() helper
  - [ ] Accept bufio.Scanner and io.Writer
  - [ ] Accept sync.WaitGroup
  - [ ] Defer wg.Done()
  - [ ] Loop: Scanner.Scan()
  - [ ] Write each line to output
  - [ ] Handle Scanner errors
- [ ] Implement runTests() function
  - [ ] Accept ctx, config, testCompleteChan, readyChan
  - [ ] Get command args from config.BuildCommand()
  - [ ] Print full command: "go test <args...>"
  - [ ] Create exec.Command("go", args...)
  - [ ] Get stdout pipe with cmd.StdoutPipe()
  - [ ] Get stderr pipe with cmd.StderrPipe()
  - [ ] Start command with cmd.Start()
  - [ ] Create WaitGroup with count 2
  - [ ] Launch goroutine for stdout streaming
    - [ ] Create bufio.Scanner from stdout pipe
    - [ ] Call streamOutput with scanner, os.Stdout, wg
  - [ ] Launch goroutine for stderr streaming
    - [ ] Create bufio.Scanner from stderr pipe
    - [ ] Call streamOutput with scanner, os.Stderr, wg
  - [ ] Wait for both scanner goroutines (wg.Wait())
  - [ ] Wait for command to finish (cmd.Wait())
  - [ ] Send TestCompleteMessage to testCompleteChan
- [ ] Write tests:
  - [ ] Create test script that prints to both stdout and stderr
  - [ ] Mock TestConfig with known values
  - [ ] Capture stdout/stderr during test
  - [ ] Call runTests with test command
  - [ ] Verify both stdout and stderr are captured
  - [ ] Verify TestCompleteMessage is sent
  - [ ] Verify command is built correctly from TestConfig
  - [ ] Test with command that exits with error (non-zero)
  - [ ] Verify test completes and sends message anyway
  - [ ] Test WaitGroup waits for both scanners
- [ ] Add demo in main():
  - [ ] Create TestConfig
  - [ ] Create channels
  - [ ] Run "go version" as simple test command
  - [ ] Wait for TestCompleteMessage
  - [ ] Print completion
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`

## Phase 3: Integration

### Step 9: Dispatcher & Integration

- [ ] Implement dispatcher() function
  - [ ] Accept ctx, config, all channel types
  - [ ] Initialize testRunning bool to false
  - [ ] Create infinite loop with select:
    - [ ] Case fileChangeChan receive:
      - [ ] If !testRunning:
        - [ ] Spawn runTests goroutine
        - [ ] Set testRunning = true
        - [ ] Send false to readyChan
    - [ ] Case commandChan receive:
      - [ ] Call handleCommand with msg.Command, config, msg.Args
      - [ ] If error: print to stderr
      - [ ] If !testRunning:
        - [ ] Spawn runTests goroutine
        - [ ] Set testRunning = true
        - [ ] Send false to readyChan
    - [ ] Case helpChan receive:
      - [ ] Call handleCommand("help", config, nil)
      - [ ] Do NOT spawn test runner
      - [ ] Do NOT change testRunning
    - [ ] Case testCompleteChan receive:
      - [ ] Set testRunning = false
      - [ ] Send true to readyChan
      - [ ] Print blank line
      - [ ] Print prompt "> "
    - [ ] Case ctx.Done():
      - [ ] If testRunning: wait for TestCompleteMessage
      - [ ] Print shutdown message
      - [ ] Return
- [ ] Update main() to wire everything:
  - [ ] Create context with signal.NotifyContext (SIGINT, SIGTERM)
  - [ ] Defer cancel()
  - [ ] Initialize TestConfig with defaults
  - [ ] Create all channels:
    - [ ] fileChangeChan
    - [ ] commandChan
    - [ ] helpChan
    - [ ] testCompleteChan
    - [ ] readyChan
  - [ ] Start watchFiles goroutine
  - [ ] Start readStdin goroutine
  - [ ] Call dispatcher (blocks in main goroutine)
- [ ] Write tests:
  - [ ] Mock all channels
  - [ ] Test FileChangeMessage spawns test runner
  - [ ] Test FileChangeMessage ignored when testRunning=true
  - [ ] Test CommandMessage calls handler
  - [ ] Test CommandMessage spawns test runner
  - [ ] Test CommandMessage ignored when testRunning=true
  - [ ] Test HelpMessage doesn't spawn test runner
  - [ ] Test TestCompleteMessage updates state and re-enables stdin
  - [ ] Test ctx.Done() waits for tests to finish
  - [ ] Test ready channel receives correct values
- [ ] Run tests: `go test -v`
- [ ] Build and run: `go build && ./gotest-watch`
- [ ] Manual test: start tool, type commands, save files, verify behavior

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
