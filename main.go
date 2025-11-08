package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mikowitz/gotest-watch/internal"
)

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

func main() {
	internal.InitRegistry()

	// Create a cancellable context for graceful shutdown
	ctx, _ := internal.SetupSignalHandler()

	// Create test config for command handlers
	config := &internal.TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}

	// Store config in context
	ctx = internal.WithConfig(ctx, config)

	logger := slog.New(slog.NewTextHandler(getLoggerDest(), nil))
	logger.Log(ctx, slog.LevelInfo, "gotest-watch starting...")

	cmdChan := make(chan internal.CommandMessage, 10)
	helpChan := make(chan internal.HelpMessage, 10)
	fileChangeChan := make(chan internal.FileChangeMessage, 10)
	testCompleteChan := make(chan internal.TestCompleteMessage, 10)

	// Start file watcher in background
	root, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

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
