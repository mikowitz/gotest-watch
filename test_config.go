package main

import "strings"

type TestConfig struct {
	TestPath   string
	Verbose    bool
	RunPattern string
}

func NewTestConfig() *TestConfig {
	return &TestConfig{
		TestPath:   "./...",
		Verbose:    false,
		RunPattern: "",
	}
}

func (tc *TestConfig) BuildCommand() string {
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
	return b.String()
}
