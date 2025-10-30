package main

import "fmt"

type CommandHandler func(*TestConfig, []string) error

var commandRegistry map[Command]CommandHandler

func initRegistry() {
	commandRegistry = make(map[Command]CommandHandler)
}

func handleCommand(command Command, config *TestConfig, args []string) error {
	handler, ok := commandRegistry[command]

	if !ok {
		return fmt.Errorf("unknown command: %q", command)
	}
	return handler(config, args)
}

func init() {
	initRegistry()
}
