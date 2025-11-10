package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func handleVerbose(config *TestConfig, _ []string) error {
	config.ToggleVerbose()
	if config.GetVerbose() {
		fmt.Println("Verbose: enabled")
	} else {
		fmt.Println("Verbose: disabled")
	}
	return nil
}

func handleRace(config *TestConfig, _ []string) error {
	config.ToggleRace()
	if config.GetRace() {
		fmt.Println("Race: enabled")
	} else {
		fmt.Println("Race: disabled")
	}
	return nil
}

func handleFailFast(config *TestConfig, _ []string) error {
	config.ToggleFailFast()
	if config.GetFailFast() {
		fmt.Println("FailFast: enabled")
	} else {
		fmt.Println("FailFast: disabled")
	}
	return nil
}

func handleCount(config *TestConfig, args []string) error {
	if len(args) == 0 {
		config.SetCount(0)
		fmt.Println("Count: cleared")
		return nil
	}

	countStr := args[0]
	count, err := strconv.Atoi(countStr)
	if err != nil {
		fmt.Printf("Error: invalid count value %q (must be a non-negative integer)\n", countStr)
		return nil // Don't return error to avoid breaking the flow
	}

	if count < 0 {
		fmt.Printf("Error: count value must be non-negative (got %d)\n", count)
		return nil
	}

	config.SetCount(count)
	if count == 0 {
		fmt.Println("Count: cleared")
	} else {
		fmt.Printf("Count: %d\n", count)
	}
	return nil
}

func handleClear(config *TestConfig, _ []string) error {
	config.Clear()
	fmt.Println("All parameters cleared")
	return nil
}

func handleRunPattern(config *TestConfig, args []string) error {
	if len(args) == 0 {
		config.SetRunPattern("")
		fmt.Println("Run pattern: cleared")
		return nil
	}
	pattern := args[0]
	config.SetRunPattern(pattern)
	fmt.Println("Run pattern:", pattern)
	return nil
}

func handleSkipPattern(config *TestConfig, args []string) error {
	if len(args) == 0 {
		config.SetSkipPattern("")
		fmt.Println("Skip pattern: cleared")
		return nil
	}
	pattern := args[0]
	config.SetSkipPattern(pattern)
	fmt.Println("Skip pattern:", pattern)
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
	config.SetTestPath(path)
	fmt.Println("Test path:", path)
	return nil
}

func handleCls(_ *TestConfig, _ []string) error {
	fmt.Print("\x1b[H\x1b[2J")
	return nil
}

func handleForceRun(_ *TestConfig, _ []string) error {
	return nil
}

func handleCommandBase(config *TestConfig, args []string) error {
	var cmdBase []string
	if len(args) == 0 {
		cmdBase = []string{"go", "test"}
	} else {
		cmdBase = args
	}
	config.SetCommandBase(cmdBase)
	fmt.Println("Test command:", strings.Join(cmdBase, " "))
	return nil
}

func handleHelp(_ *TestConfig, _ []string) error {
	fmt.Println("Available commands:")
	fmt.Println("  v            Toggle verbose mode (-v flag)")
	fmt.Println("  race         Toggle race mode (-race flag)")
	fmt.Println("  ff           Toggle failfast mode (-failfast flag)")
	fmt.Println("  count <n>    Set test count (-count=<n>, n > 0)")
	fmt.Println("  count        Clear count")
	fmt.Println("  r <pattern>  Set test run pattern (-run=<pattern>)")
	fmt.Println("  r            Clear run pattern")
	fmt.Println("  s <pattern>  Set test skip pattern (-skip=<pattern>)")
	fmt.Println("  s            Clear skip pattern")
	fmt.Println("  p <path>     Set test path (default: ./...")
	fmt.Println("  p            Set test path to default (./...)")
	fmt.Println("  cmd          Set the base command to run (default: go test)")
	fmt.Println("  clear        Clear all parameters")
	fmt.Println("  cls          Clear screen")
	fmt.Println("  f            Force test run")
	fmt.Println("  h            Show this help")
	return nil
}
