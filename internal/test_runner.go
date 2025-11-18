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

const (
	Red     = "31;1"
	Green   = "32;1"
	Yellow  = "33;1"
	Magenta = "35;1"
	White   = "37;1"
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
			output = colorizeOutput(output)
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

func selectColorizer(line string) string {
	if strings.HasPrefix(line, "?") || strings.Contains(line, "SKIP") { // || strings.HasPrefix(line, "=== RUN") {
		return Yellow
	}
	if strings.HasPrefix(line, "ok") || strings.Contains(line, "PASS") {
		return Green
	}
	if strings.HasPrefix(line, "FAIL") {
		return Red
	}
	if strings.Contains(line, ".go:") {
		return Magenta
	}
	return White
}

func colorizeOutput(output string) string {
	reset := "\033[0m"
	colorizer := selectColorizer(output)
	return fmt.Sprintf("\033[%sm%s%s", colorizer, output, reset)
}
