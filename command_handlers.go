package main

import "fmt"

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

func handleHelp(_ *TestConfig, _ []string) error {
	fmt.Println("Available commands:")
	fmt.Println("  v            Toggle verbose mode (-v flag)")
	fmt.Println("  r <pattern>  Set test run pattern (-run=<pattern>)")
	fmt.Println("  r            Clear run pattern")
	fmt.Println("  p <path>     Set test path (default: ./...")
	fmt.Println("  clear        Clear all parameters")
	fmt.Println("  cls          Clear screen")
	fmt.Println("  f            Force test run")
	fmt.Println("  help         Show this help")
	return nil
}
