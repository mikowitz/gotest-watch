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
		expectedCmd string
	}{
		{"default configuration", "./...", false, "", "go test ./..."},
		{"verbose enabled", "./...", true, "", "go test ./... -v"},
		{"run pattern set", "./...", false, "MyTest", "go test ./... -run=MyTest"},
		{"specific test path", "./testing", false, "", "go test ./testing"},
		{"multiple test paths", "./testing ./integration", false, "", "go test ./testing ./integration"},
		{"everything configured", "./mytests", true, "MyTest", "go test ./mytests -v -run=MyTest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := TestConfig{
				TestPath:   tc.testPath,
				Verbose:    tc.verbose,
				RunPattern: tc.runPattern,
			}

			cmd := config.BuildCommand()

			assert.Equal(t, tc.expectedCmd, cmd, "expected command string to match for "+tc.name)
		})
	}
}
