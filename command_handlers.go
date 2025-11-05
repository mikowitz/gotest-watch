package main

import (
	"fmt"
	"os"
)

func handleVerbose(config *TestConfig, _ []string) error {
	config.Verbose = !config.Verbose
	if config.Verbose {
		fmt.Println("Verbose: enabled")
	} else {
		fmt.Println("Verbose: disabled")
	}
	return nil
}

func handleClear(config *TestConfig, _ []string) error {
	config.TestPath = "./..."
	config.Verbose = false
	config.RunPattern = ""
	fmt.Println("All parameters cleared")
	return nil
}

func handleRunPattern(config *TestConfig, args []string) error {
	if len(args) == 0 {
		config.RunPattern = ""
		fmt.Println("Run pattern: cleared")
		return nil
	}
	pattern := args[0]
	config.RunPattern = pattern
	fmt.Printf("Run pattern: %s\n", pattern)
	return nil
}

func handleTestPath(config *TestConfig, args []string) error {
	var path string
	if len(args) == 0 {
		path = "./..."
	} else {
		path = args[0]
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("path does not exist: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path %q is not a directory", path)
		}
	}
	config.TestPath = path
	fmt.Println("Test path:", path)
	return nil
}

func handleCls(_ *TestConfig, _ []string) error {
	fmt.Print("\x1b[H\x1b[2J")
	return nil
}

func handleRun(_ *TestConfig, _ []string) error {
	// This handler is no longer used - the force run command is handled
	// directly in main.go because it needs access to channels
	// Keeping this for backwards compatibility with tests
	return nil
}

func handleHelp(_ *TestConfig, _ []string) error {
	fmt.Println("Available commands:")
	fmt.Println("  v            Toggle verbose mode (-v flag)")
	fmt.Println("  r <pattern>  Set test run pattern (-run=<pattern>)")
	fmt.Println("  r            Clear run pattern")
	fmt.Println("  p <path>     Set test path (default: ./...")
	fmt.Println("  p            Set test path to default (./...)")
	fmt.Println("  clear        Clear all parameters")
	fmt.Println("  cls          Clear screen")
	fmt.Println("  f            Force test run")
	fmt.Println("  h            Show this help")
	return nil
}
