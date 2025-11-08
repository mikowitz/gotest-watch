package internal

type MessageType string

const (
	MessageTypeFileChange   MessageType = "FileChange"
	MessageTypeCommand      MessageType = "Command"
	MessageTypeHelp         MessageType = "Help"
	MessageTypeTestComplete MessageType = "TestComplete"
)

type Command string

const (
	VerboseCmd     Command = "v"
	SetPathCmd     Command = "p"
	SetPatternCmd  Command = "r"
	SetSkipCmd     Command = "s"
	HelpCmd        Command = "h"
	ClearCmd       Command = "clear"
	ClearScreenCmd Command = "cls"
	ForceRunCmd    Command = "f"
)

type Message interface {
	Type() MessageType
}

type (
	FileChangeMessage struct{}
	CommandMessage    struct {
		Command Command
		Args    []string
	}
	HelpMessage         struct{}
	TestCompleteMessage struct{}
)

func (m *FileChangeMessage) Type() MessageType {
	return MessageTypeFileChange
}

func NewCommandMessage(cmd Command, args []string) *CommandMessage {
	return &CommandMessage{
		Command: cmd,
		Args:    args,
	}
}

func (m *CommandMessage) Type() MessageType {
	return MessageTypeCommand
}

func (m *HelpMessage) Type() MessageType {
	return MessageTypeHelp
}

func (m *TestCompleteMessage) Type() MessageType {
	return MessageTypeTestComplete
}
