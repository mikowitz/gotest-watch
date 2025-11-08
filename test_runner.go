package main

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

func streamOutput(r *bufio.Scanner, w io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	for r.Scan() {
		err := r.Err()
		if err != nil {
			log.Println(err)
			return
		}
		_, err = w.Write([]byte(r.Text()))
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
func runTests(
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
	testCommand := config.BuildCommand()
	fields := strings.Fields(testCommand)

	displayCommand(fields[1:])

	// Use CommandContext to support cancellation via context
	//nolint:gosec // TODO: sanitize input
	cmd := exec.CommandContext(ctx, "go", fields[1:]...)

	// Set working directory if specified
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

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
		streamOutput(r, stdoutWriter, &wg)
	}()

	go func() {
		r := bufio.NewScanner(stderr)
		streamOutput(r, stderrWriter, &wg)
	}()

	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}

	completeChan <- TestCompleteMessage{}
}
