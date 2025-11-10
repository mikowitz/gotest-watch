package internal

import (
	"strings"
	"sync"
)

type TestConfig struct {
	sync.RWMutex
	TestPath    string
	Verbose     bool
	RunPattern  string
	SkipPattern string
	CommandBase []string
	Race        bool
	WorkingDir  string // Optional: if set, tests will run in this directory
}

func NewTestConfig() *TestConfig {
	return &TestConfig{
		TestPath:    "./...",
		CommandBase: []string{"go", "test"},
	}
}

func (tc *TestConfig) BuildCommand() string {
	tc.RLock()
	defer tc.RUnlock()

	var b strings.Builder
	b.WriteString(strings.Join(tc.CommandBase, " "))
	b.WriteString(" ")
	b.WriteString(tc.TestPath)
	if tc.Verbose {
		b.WriteString(" -v")
	}
	if tc.Race {
		b.WriteString(" -race")
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

func (tc *TestConfig) GetVerbose() bool {
	tc.RLock()
	defer tc.RUnlock()
	return tc.Verbose
}

func (tc *TestConfig) GetTestPath() string {
	tc.RLock()
	defer tc.RUnlock()
	return tc.TestPath
}

func (tc *TestConfig) GetRunPattern() string {
	tc.RLock()
	defer tc.RUnlock()
	return tc.RunPattern
}

func (tc *TestConfig) GetSkipPattern() string {
	tc.RLock()
	defer tc.RUnlock()
	return tc.SkipPattern
}

func (tc *TestConfig) GetCommandBase() []string {
	tc.RLock()
	defer tc.RUnlock()
	return tc.CommandBase
}

func (tc *TestConfig) GetRace() bool {
	tc.RLock()
	defer tc.RUnlock()
	return tc.Race
}

// Safe setters
func (tc *TestConfig) SetVerbose(v bool) {
	tc.Lock()
	defer tc.Unlock()
	tc.Verbose = v
}

func (tc *TestConfig) SetTestPath(path string) {
	tc.Lock()
	defer tc.Unlock()
	tc.TestPath = path
}

func (tc *TestConfig) SetRunPattern(pattern string) {
	tc.Lock()
	defer tc.Unlock()
	tc.RunPattern = pattern
}

func (tc *TestConfig) SetSkipPattern(pattern string) {
	tc.Lock()
	defer tc.Unlock()
	tc.SkipPattern = pattern
}

func (tc *TestConfig) SetCommandBase(commandBase []string) {
	tc.Lock()
	defer tc.Unlock()
	tc.CommandBase = commandBase
}

func (tc *TestConfig) ToggleVerbose() {
	tc.Lock()
	defer tc.Unlock()
	tc.Verbose = !tc.Verbose
}

func (tc *TestConfig) ToggleRace() {
	tc.Lock()
	defer tc.Unlock()
	tc.Race = !tc.Race
}

func (tc *TestConfig) Clear() {
	tc.Lock()
	defer tc.Unlock()
	tc.TestPath = "./..."
	tc.Verbose = false
	tc.RunPattern = ""
	tc.SkipPattern = ""
	tc.CommandBase = []string{"go", "test"}
	tc.Race = false
}
