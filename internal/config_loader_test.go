package internal

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrDefaultConfig(t *testing.T) {
	t.Run("returns default config when no config file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		config := LoadOrDefaultConfig(tmpDir)

		expected := NewTestConfig()
		assert.Equal(t, expected.TestPath, config.TestPath)
		assert.Equal(t, expected.CommandBase, config.CommandBase)
		assert.Equal(t, expected.Verbose, config.Verbose)
		assert.Equal(t, expected.Race, config.Race)
		assert.Equal(t, expected.Cover, config.Cover)
		assert.Equal(t, expected.FailFast, config.FailFast)
		assert.Equal(t, expected.ClearScreen, config.ClearScreen)
		assert.Equal(t, expected.Color, config.Color)
	})

	t.Run("loads config from .gotest-watch.yml when it exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		yamlContent := `---
commandBase:
- go
- test
testPath: ./pkg/...
verbose: true
race: true
cover: true
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		config := LoadOrDefaultConfig(tmpDir)

		assert.Equal(t, "./pkg/...", config.TestPath)
		assert.Equal(t, []string{"go", "test"}, config.CommandBase)
		assert.True(t, config.Verbose)
		assert.True(t, config.Race)
		assert.True(t, config.Cover)
	})

	t.Run("loads config from .gotest-watch.yaml when it exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yaml")
		yamlContent := `---
testPath: ./internal/...
verbose: true
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		config := LoadOrDefaultConfig(tmpDir)

		assert.Equal(t, "./internal/...", config.TestPath)
		assert.True(t, config.Verbose)
	})

	t.Run("returns default config when config file has invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		invalidYAML := `---
this is: invalid: yaml: content
	bad indentation
`
		err := os.WriteFile(configPath, []byte(invalidYAML), 0o600)
		require.NoError(t, err)

		config := LoadOrDefaultConfig(tmpDir)

		// Should fall back to defaults
		expected := NewTestConfig()
		assert.Equal(t, expected.TestPath, config.TestPath)
		assert.Equal(t, expected.CommandBase, config.CommandBase)
	})

	t.Run("prefers .yml over .yaml when both exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		ymlPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		yamlPath := filepath.Join(tmpDir, ".gotest-watch.yaml")

		ymlContent := `---
testPath: ./from-yml/...
`
		yamlContent := `---
testPath: ./from-yaml/...
`
		err := os.WriteFile(ymlPath, []byte(ymlContent), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		config := LoadOrDefaultConfig(tmpDir)

		assert.Equal(t, "./from-yml/...", config.TestPath)
	})

	t.Run("loads all config fields correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		yamlContent := `---
commandBase:
- richgo
- test
- -tags
- integration
testPath: ./custom/...
verbose: true
runPattern: TestFoo
skipPattern: TestBar
race: true
cover: true
failfast: true
count: 5
clearScreen: true
color: true
workingDir: /tmp/work
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		config := LoadOrDefaultConfig(tmpDir)

		assert.Equal(t, []string{"richgo", "test", "-tags", "integration"}, config.CommandBase)
		assert.Equal(t, "./custom/...", config.TestPath)
		assert.True(t, config.Verbose)
		assert.Equal(t, "TestFoo", config.RunPattern)
		assert.Equal(t, "TestBar", config.SkipPattern)
		assert.True(t, config.Race)
		assert.True(t, config.Cover)
		assert.True(t, config.FailFast)
		assert.Equal(t, 5, config.Count)
		assert.True(t, config.ClearScreen)
		assert.True(t, config.Color)
		assert.Equal(t, "/tmp/work", config.WorkingDir)
	})

	t.Run("handles empty directory path", func(t *testing.T) {
		config := LoadOrDefaultConfig("")

		expected := NewTestConfig()
		assert.Equal(t, expected.TestPath, config.TestPath)
		assert.Equal(t, expected.CommandBase, config.CommandBase)
	})

	t.Run("logs warning when config file has invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		invalidYAML := `---
this is: invalid: yaml: content
	bad indentation
`
		err := os.WriteFile(configPath, []byte(invalidYAML), 0o600)
		require.NoError(t, err)

		// Capture log output
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		config := LoadOrDefaultConfig(tmpDir)

		// Should still return defaults
		expected := NewTestConfig()
		assert.Equal(t, expected.TestPath, config.TestPath)

		// Should log a warning about the invalid config
		logOutput := buf.String()
		assert.True(t, strings.Contains(logOutput, "Warning"), "Expected log to contain 'Warning', got: %s", logOutput)
		assert.True(t, strings.Contains(logOutput, configPath), "Expected log to contain config path, got: %s", logOutput)
	})

	t.Run("does not log when config file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Capture log output
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		config := LoadOrDefaultConfig(tmpDir)

		// Should return defaults
		expected := NewTestConfig()
		assert.Equal(t, expected.TestPath, config.TestPath)

		// Should NOT log anything when file simply doesn't exist
		logOutput := buf.String()
		assert.Empty(t, logOutput, "Expected no log output when config file doesn't exist, got: %s", logOutput)
	})
}
