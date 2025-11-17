package internal

import (
	"strconv"
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
	FailFast    bool
	Count       int
	ClearScreen bool
	Cover       bool
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
	if tc.FailFast {
		b.WriteString(" -failfast")
	}
	if tc.Cover {
		b.WriteString(" -cover")
	}
	if tc.Count > 0 {
		b.WriteString(" -count=")
		b.WriteString(strconv.Itoa(tc.Count))
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

func (tc *TestConfig) GetClearScreen() bool {
	tc.RLock()
	defer tc.RUnlock()
	return tc.ClearScreen
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

func (tc *TestConfig) GetFailFast() bool {
	tc.RLock()
	defer tc.RUnlock()
	return tc.FailFast
}

func (tc *TestConfig) GetCount() int {
	tc.RLock()
	defer tc.RUnlock()
	return tc.Count
}

func (tc *TestConfig) GetCover() bool {
	tc.RLock()
	defer tc.RUnlock()
	return tc.Cover
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

func (tc *TestConfig) SetCount(count int) {
	tc.Lock()
	defer tc.Unlock()
	tc.Count = count
}

func (tc *TestConfig) ToggleVerbose() {
	tc.Lock()
	defer tc.Unlock()
	tc.Verbose = !tc.Verbose
}

func (tc *TestConfig) ToggleClearScreen() {
	tc.Lock()
	defer tc.Unlock()
	tc.ClearScreen = !tc.ClearScreen
}

func (tc *TestConfig) ToggleRace() {
	tc.Lock()
	defer tc.Unlock()
	tc.Race = !tc.Race
}

func (tc *TestConfig) ToggleFailFast() {
	tc.Lock()
	defer tc.Unlock()
	tc.FailFast = !tc.FailFast
}

func (tc *TestConfig) ToggleCover() {
	tc.Lock()
	defer tc.Unlock()
	tc.Cover = !tc.Cover
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
	tc.FailFast = false
	tc.Count = 0
	tc.Cover = false
}
