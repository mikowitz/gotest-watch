package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name        string
		testPath    string
		verbose     bool
		runPattern  string
		commandBase []string
		expectedCmd string
	}{
		{"default configuration", "./...", false, "", []string{"go", "test"}, "go test ./..."},
		{"verbose enabled", "./...", true, "", []string{"go", "test"}, "go test ./... -v"},
		{"run pattern set", "./...", false, "MyTest", []string{"go", "test"}, "go test ./... -run=MyTest"},
		{"specific test path", "./testing", false, "", []string{"go", "test"}, "go test ./testing"},
		{
			"multiple test paths",
			"./testing ./integration",
			false, "",
			[]string{"go", "test"},
			"go test ./testing ./integration",
		},
		{"everything configured", "./mytests", true, "MyTest", []string{"go", "test"}, "go test ./mytests -v -run=MyTest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := TestConfig{
				TestPath:    tc.testPath,
				Verbose:     tc.verbose,
				RunPattern:  tc.runPattern,
				CommandBase: tc.commandBase,
			}

			cmd := config.BuildCommand()

			assert.Equal(t, tc.expectedCmd, cmd, "expected command string to match for "+tc.name)
		})
	}
}

func TestBuildCommand_WithCover(t *testing.T) {
	tests := []struct {
		name        string
		cover       bool
		expectedCmd string
	}{
		{"cover disabled", false, "go test ./..."},
		{"cover enabled", true, "go test ./... -cover"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := TestConfig{
				TestPath:    "./...",
				CommandBase: []string{"go", "test"},
				Cover:       tc.cover,
			}

			cmd := config.BuildCommand()

			assert.Equal(t, tc.expectedCmd, cmd)
		})
	}
}

func TestBuildCommand_CoverWithOtherFlags(t *testing.T) {
	config := TestConfig{
		TestPath:    "./...",
		CommandBase: []string{"go", "test"},
		Verbose:     true,
		Cover:       true,
		Race:        true,
	}

	cmd := config.BuildCommand()

	assert.Equal(t, "go test ./... -v -race -cover", cmd)
}

func TestGetCover(t *testing.T) {
	config := &TestConfig{
		Cover: true,
	}

	assert.True(t, config.GetCover())

	config.Cover = false
	assert.False(t, config.GetCover())
}

func TestToggleCover(t *testing.T) {
	config := &TestConfig{
		Cover: false,
	}

	config.ToggleCover()
	assert.True(t, config.GetCover(), "Cover should toggle from false to true")

	config.ToggleCover()
	assert.False(t, config.GetCover(), "Cover should toggle from true to false")
}
