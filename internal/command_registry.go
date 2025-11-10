package internal

import "fmt"

type CommandHandler func(*TestConfig, []string) error

var commandRegistry map[Command]CommandHandler

func InitRegistry() {
	initRegistry()
}

func initRegistry() {
	commandRegistry = make(map[Command]CommandHandler)
	commandRegistry[VerboseCmd] = handleVerbose
	commandRegistry[HelpCmd] = handleHelp
	commandRegistry[ClearCmd] = handleClear
	commandRegistry[SetPatternCmd] = handleRunPattern
	commandRegistry[SetSkipCmd] = handleSkipPattern
	commandRegistry[SetPathCmd] = handleTestPath
	commandRegistry[ClearScreenCmd] = handleCls
	commandRegistry[ForceRunCmd] = handleForceRun
	commandRegistry[SetCommandBaseCmd] = handleCommandBase
	commandRegistry[RaceCmd] = handleRace
	commandRegistry[FailFastCmd] = handleFailFast
}

func handleCommand(command Command, config *TestConfig, args []string) error {
	handler, ok := commandRegistry[command]

	if !ok {
		return fmt.Errorf("unknown command: %q", command)
	}
	return handler(config, args)
}
