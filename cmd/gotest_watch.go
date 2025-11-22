package cmd

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mikowitz/gotest-watch/internal"
	"github.com/spf13/cobra"
)

var (
	commandBase string
	testPath    string
	verbose     bool
	runPattern  string
	skipPattern string
	count       int
	clearScreen bool
	color       bool
)

var gotestWatchCmd = &cobra.Command{
	Use:   "gotest-watch",
	Short: "An interactive command line tool for running `go test`",
	Long:  "An interactive command line tool for running `go test`. It watches *.go files in your project for changes, and can be customized between runs to specify many of the flags that can be set for `go test`.",
	Args:  cobra.NoArgs,
	Run:   gotestWatch,
}

func gotestWatch(cmd *cobra.Command, args []string) {
	internal.InitRegistry()

	fmt.Println("commandBase", commandBase, cmd.Flags().Lookup("cmd").Changed)
	fmt.Println("testPath", testPath, cmd.Flags().Lookup("verbose").Changed)
	fmt.Println("verbose", verbose)
	fmt.Println("runPattern", runPattern)
	fmt.Println("skipPattern", skipPattern)
	fmt.Println("count", count)
	fmt.Println("clearScreen", clearScreen)
	fmt.Println("color", color)

	// Create a cancellable context for graceful shutdown
	ctx, _ := internal.SetupSignalHandler()

	// Get working directory for config lookup
	root, err := os.Getwd()
	if err != nil {
		log.Println(err)
		root = "."
	}

	// Create test config from file or defaults
	config := internal.LoadOrDefaultConfig(root)
	overrideConfig(config, cmd)

	// Store config in context
	ctx = internal.WithConfig(ctx, config)

	logger := slog.New(slog.NewTextHandler(getLoggerDest(), nil))
	logger.Log(ctx, slog.LevelInfo, "gotest-watch starting...")

	cmdChan := make(chan internal.CommandMessage, 10)
	helpChan := make(chan internal.HelpMessage, 10)
	fileChangeChan := make(chan internal.FileChangeMessage, 10)
	testCompleteChan := make(chan internal.TestCompleteMessage, 10)

	// Start file watcher in background
	startWatching := make(chan struct{})

	go internal.WatchFiles(ctx, root, fileChangeChan, startWatching)

	// Start stdin reader in background
	go internal.ReadStdin(ctx, os.Stdin, cmdChan, helpChan)

	fmt.Println("Running tests...")
	internal.RunTests(ctx, testCompleteChan, nil, nil)

	select {
	case <-testCompleteChan:
		close(startWatching)
	case <-ctx.Done():
		return
	}

	// Start dispatcher (blocks until context is cancelled)
	internal.Dispatcher(ctx, fileChangeChan, cmdChan, helpChan, testCompleteChan)
}

func getLoggerDest() io.Writer {
	usr, _ := user.Current()
	logDir := filepath.Join(usr.HomeDir, ".local/state/gotest-watch")
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		fmt.Printf("Could not find directory")
		return io.Discard
	}
	if f, err := os.OpenFile(
		filepath.Join(filepath.Clean(logDir), "gotest-watch.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o600,
	); err != nil {
		return io.Discard
	} else {
		return f
	}
}

func Execute() {
	if err := gotestWatchCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run() {
	Execute()
}

func init() {
	gotestWatchCmd.Flags().StringVarP(&commandBase, "cmd", "m", "go test", "base command to run (e.g. `go test`)")
	gotestWatchCmd.Flags().StringVarP(&testPath, "path", "p", "./...", "directory to run tests in")
	gotestWatchCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose test output")
	gotestWatchCmd.Flags().StringVarP(&runPattern, "run", "r", "", "run tests that match this pattern")
	gotestWatchCmd.Flags().StringVarP(&skipPattern, "skip", "s", "", "skip tests that match this pattern")
	gotestWatchCmd.Flags().IntVarP(&count, "count", "n", 0, "number of times to run each test")
	gotestWatchCmd.Flags().BoolVarP(&clearScreen, "cls", "l", false, "clear the screen before each test run")
	gotestWatchCmd.Flags().BoolVarP(&color, "color", "c", false, "ANSI color output")
}

func overrideConfig(config *internal.TestConfig, cmd *cobra.Command) {
	if cmd.Flags().Lookup("cmd").Changed {
		config.SetCommandBase(strings.Fields(commandBase))
	}
	if cmd.Flags().Lookup("path").Changed {
		config.SetTestPath(testPath)
	}
	if cmd.Flags().Lookup("verbose").Changed {
		config.SetVerbose(verbose)
	}
	if cmd.Flags().Lookup("run").Changed {
		config.SetRunPattern(runPattern)
	}
	if cmd.Flags().Lookup("skip").Changed {
		config.SetSkipPattern(skipPattern)
	}
	if cmd.Flags().Lookup("count").Changed {
		config.SetCount(count)
	}
	if cmd.Flags().Lookup("cls").Changed {
		config.SetClearScreen(clearScreen)
	}
	if cmd.Flags().Lookup("color").Changed {
		config.SetColor(color)
	}
}
