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

func runTests(ctx context.Context, config *TestConfig, completeChan chan TestCompleteMessage, readyChan chan bool) {
	testCommand := config.BuildCommand()
	fields := strings.Fields(testCommand)

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
		streamOutput(r, os.Stdout, &wg)
	}()

	go func() {
		r := bufio.NewScanner(stderr)
		streamOutput(r, os.Stderr, &wg)
	}()

	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}

	completeChan <- TestCompleteMessage{}
	readyChan <- true
}
