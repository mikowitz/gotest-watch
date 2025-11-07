package main

import (
	"strings"
	"sync"
)

type TestConfig struct {
	mu          sync.RWMutex
	TestPath    string
	Verbose     bool
	RunPattern  string
	SkipPattern string
	WorkingDir  string // Optional: if set, tests will run in this directory
}

func NewTestConfig() *TestConfig {
	return &TestConfig{
		TestPath:    "./...",
		Verbose:     false,
		RunPattern:  "",
		SkipPattern: "",
	}
}

func (tc *TestConfig) BuildCommand() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var b strings.Builder
	b.WriteString("go test ")
	b.WriteString(tc.TestPath)
	if tc.Verbose {
		b.WriteString(" -v")
	}
	if tc.RunPattern != "" {
		b.WriteString(" -run=")
		b.WriteString(tc.RunPattern)
	}
	if tc.SkipPattern != "" {
		b.WriteString(" -skip=")
		b.WriteString(tc.SkipPattern)
	}
	return b.String()
}

// Safe getters
func (tc *TestConfig) GetVerbose() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.Verbose
}

func (tc *TestConfig) GetTestPath() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.TestPath
}

func (tc *TestConfig) GetRunPattern() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.RunPattern
}

func (tc *TestConfig) GetSkipPattern() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.SkipPattern
}

// Safe setters
func (tc *TestConfig) SetVerbose(v bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Verbose = v
}

func (tc *TestConfig) SetTestPath(path string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.TestPath = path
}

func (tc *TestConfig) SetRunPattern(pattern string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.RunPattern = pattern
}

func (tc *TestConfig) SetSkipPattern(pattern string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.SkipPattern = pattern
}

func (tc *TestConfig) ToggleVerbose() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Verbose = !tc.Verbose
}

func (tc *TestConfig) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.TestPath = "./..."
	tc.Verbose = false
	tc.RunPattern = ""
	tc.SkipPattern = ""
}
