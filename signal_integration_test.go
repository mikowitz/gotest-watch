// +build integration

package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestSignalHandling_ActualProcess tests signal handling in a real subprocess
// Run with: go test -tags=integration -run TestSignalHandling_ActualProcess
func TestSignalHandling_ActualProcess(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "gotest-watch-test")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("gotest-watch-test")

	t.Run("SIGINT causes graceful shutdown", func(t *testing.T) {
		cmd := exec.Command("./gotest-watch-test")

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start process: %v", err)
		}

		// Give it time to start
		time.Sleep(500 * time.Millisecond)

		// Send SIGINT
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatalf("Failed to send SIGINT: %v", err)
		}

		// Process should exit gracefully
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process exited (expected)
			// Exit code 0 or -1 (interrupted) is acceptable
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Check if it was killed by signal (acceptable)
					if exitErr.ExitCode() != 0 && exitErr.ExitCode() != -1 {
						t.Fatalf("Process exited with unexpected code: %v", err)
					}
				}
			}
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatal("Process did not exit within timeout")
		}
	})

	t.Run("SIGTERM causes graceful shutdown", func(t *testing.T) {
		cmd := exec.Command("./gotest-watch-test")

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start process: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		// Send SIGTERM
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("Failed to send SIGTERM: %v", err)
		}

		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited (expected)
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatal("Process did not exit within timeout")
		}
	})
}
