package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func streamOutput(r *bufio.Scanner, w io.Writer, wg *sync.WaitGroup, colorize bool) {
	defer wg.Done()

	for r.Scan() {
		err := r.Err()
		if err != nil {
			log.Println(err)
			return
		}

		output := r.Text()
		if colorize {
			var colorizer string
			reset := "\033[0m"
			if strings.HasPrefix(output, "?") || strings.Contains(output, "SKIP") || strings.HasPrefix(output, "=== RUN") {
				colorizer = "\033[33;1m"
			}
			if strings.HasPrefix(output, "ok") || strings.Contains(output, "PASS") {
				colorizer = "\033[32;1m"
			}
			if strings.HasPrefix(output, "FAIL") {
				colorizer = "\033[31;1m"
			}
			output = fmt.Sprintf("%s%s%s", colorizer, output, reset)
		}
		_, err = w.Write([]byte(output))
		if err != nil {
			log.Println(err)
		}
		_, err = w.Write([]byte("\n"))
		if err != nil {
			log.Println(err)
		}
	}
}

//nolint:funlen
func RunTests(
	ctx context.Context,
	completeChan chan TestCompleteMessage,
	stdoutWriter io.Writer,
	stderrWriter io.Writer,
) {
	// Default to os.Stdout/Stderr if nil
	if stdoutWriter == nil {
		stdoutWriter = os.Stdout
	}
	if stderrWriter == nil {
		stderrWriter = os.Stderr
	}

	config := getConfig(ctx)
	if config == nil {
		fmt.Fprintln(os.Stderr, "Error: config not found in context")
		return
	}

	if config.GetClearScreen() {
		fmt.Print("\x1b[H\x1b[2J")
	}
	testCommand := config.BuildCommand()
	fields := strings.Fields(testCommand)

	displayCommand(fields)

	// Use CommandContext to support cancellation via context
	//nolint:gosec // TODO: sanitize input
	cmd := exec.CommandContext(ctx, "go", fields[1:]...)

	// Set working directory if specified
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	colorize := config.GetColor()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		r := bufio.NewScanner(stdout)
		streamOutput(r, stdoutWriter, &wg, colorize)
	}()

	go func() {
		r := bufio.NewScanner(stderr)
		streamOutput(r, stderrWriter, &wg, colorize)
	}()

	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}

	completeChan <- TestCompleteMessage{}
}
