package internal

import (
	"io"
	"os"
	"sync"
	"testing"
)

var stdoutMu sync.Mutex

// captureStdout captures stdout during test execution
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	old := os.Stdout
	defer func() { os.Stdout = old }()

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = w

	f()

	os.Stdout = old
	_ = w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(out)
}
