package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageInterfaceSatisfaction(t *testing.T) {
	var _ Message = (*FileChangeMessage)(nil)
	var _ Message = (*CommandMessage)(nil)
	var _ Message = (*HelpMessage)(nil)
	var _ Message = (*TestCompleteMessage)(nil)
}

// var _ Surface = (*BaseSurface)(nil)

func TestType(t *testing.T) {
	tests := []struct {
		name         string
		message      Message
		expectedType MessageType
	}{
		{"FileChangeMessage", &FileChangeMessage{}, MessageTypeFileChange},
		{"CommandMessage", &CommandMessage{}, MessageTypeCommand},
		{"HelpMessage", &HelpMessage{}, MessageTypeHelp},
		{"TestCompleteMessage", &TestCompleteMessage{}, MessageTypeTestComplete},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualType := tc.message.Type()

			assert.Equal(
				t, actualType, tc.expectedType,
				fmt.Sprintf("Expected %s to have message type %s, got %s", tc.name, tc.expectedType, actualType),
			)
		})
	}
}

func TestNewCommandMessage(t *testing.T) {
	t.Run("verbose", func(t *testing.T) {
		m := NewCommandMessage(VerboseCmd, []string{})

		assert.Equal(t, VerboseCmd, m.Command, "expected verbose command to have command type VerboseCmd")
		assert.NotEmpty(t, m.Args, "expected verbose command to have no args")
	})

	t.Run("set path with single path", func(t *testing.T) {
		m := NewCommandMessage(SetPathCmd, []string{"./cmd"})

		assert.Equal(t, SetPathCmd, m.Command, "expected set path command to have command type SetPathCmd")
		assert.Equal(t, m.Args, []string{"./cmd"})
	})

	t.Run("set path with multiple paths", func(t *testing.T) {
		m := NewCommandMessage(SetPathCmd, []string{"./cmd", "./integration"})

		assert.Equal(t, SetPathCmd, m.Command, "expected set path command to have command type SetPathCmd")
		assert.Equal(t, m.Args, []string{"./cmd", "./integration"})
	})

	t.Run("set pattern", func(t *testing.T) {
		m := NewCommandMessage(SetPatternCmd, []string{"MyTest"})

		assert.Equal(t, SetPatternCmd, m.Command, "expected set pattern command to have command type SetPatternCmd")
		assert.Equal(t, m.Args, []string{"MyTest"})
	})
}
