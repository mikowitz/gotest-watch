package main

import "fmt"

type CommandHandler func(*TestConfig, []string) error

var commandRegistry map[Command]CommandHandler

func initRegistry() {
	commandRegistry = make(map[Command]CommandHandler)
	commandRegistry[VerboseCmd] = handleVerbose
	commandRegistry[HelpCmd] = handleHelp
	commandRegistry[ClearCmd] = handleClear
	commandRegistry[SetPatternCmd] = handleRunPattern
	commandRegistry[SetPathCmd] = handleTestPath
	commandRegistry[ClearScreenCmd] = handleCls
	commandRegistry[ForceRunCmd] = handleRun
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
