package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromYAML(t *testing.T) {
	t.Run("loads all fields correctly from valid YAML", func(t *testing.T) {
		yamlContent := `---
commandBase: [go, test]
testPath: ./pkg/...
verbose: true
runPattern: TestFoo
skipPattern: TestBar
race: true
cover: true
failfast: true
clearScreen: true
color: true
workingDir: /tmp/test
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, []string{"go", "test"}, config.CommandBase)
		assert.Equal(t, "./pkg/...", config.TestPath)
		assert.True(t, config.Verbose)
		assert.Equal(t, "TestFoo", config.RunPattern)
		assert.Equal(t, "TestBar", config.SkipPattern)
		assert.True(t, config.Race)
		assert.True(t, config.Cover)
		assert.True(t, config.FailFast)
		assert.True(t, config.ClearScreen)
		assert.True(t, config.Color)
		assert.Equal(t, "/tmp/test", config.WorkingDir)
	})

	t.Run("handles empty strings correctly", func(t *testing.T) {
		yamlContent := `---
commandBase:
- go
- test
testPath: ./...
verbose: false
runPattern: ""
skipPattern: ""
race: false
cover: false
failfast: false
clearScreen: false
color: false
workingDir: ""
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, []string{"go", "test"}, config.CommandBase)
		assert.Equal(t, "./...", config.TestPath)
		assert.False(t, config.Verbose)
		assert.Equal(t, "", config.RunPattern)
		assert.Equal(t, "", config.SkipPattern)
		assert.False(t, config.Race)
		assert.False(t, config.Cover)
		assert.False(t, config.FailFast)
		assert.False(t, config.ClearScreen)
		assert.False(t, config.Color)
		assert.Equal(t, "", config.WorkingDir)
	})

	t.Run("handles commandBase with custom command", func(t *testing.T) {
		yamlContent := `---
commandBase: [richgo, test]
testPath: ./...
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, []string{"richgo", "test"}, config.CommandBase)
	})

	t.Run("handles commandBase with additional flags", func(t *testing.T) {
		yamlContent := `---
commandBase:
- go
- test
- -tags
- integration
testPath: ./...
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, []string{"go", "test", "-tags", "integration"}, config.CommandBase)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadConfigFromYAML("/path/that/does/not/exist.yml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		yamlContent := `---
this is not: valid: yaml: structure
	bad indentation
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		_, err := LoadConfigFromYAML(tmpFile)
		assert.Error(t, err)
	})

	t.Run("merges with defaults for missing fields", func(t *testing.T) {
		yamlContent := `---
testPath: ./custom/...
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		// Explicitly set field should be overridden
		assert.Equal(t, "./custom/...", config.TestPath)
		// Missing fields should use defaults from NewTestConfig()
		assert.Equal(t, []string{"go", "test"}, config.CommandBase)
		assert.False(t, config.Verbose)
		assert.False(t, config.Race)
		assert.False(t, config.Cover)
	})

	t.Run("handles partial configuration", func(t *testing.T) {
		yamlContent := `---
commandBase: ["go", "test"]
verbose: true
cover: true
`
		tmpFile := createTempYAMLFile(t, yamlContent)
		defer os.Remove(tmpFile)

		config, err := LoadConfigFromYAML(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, []string{"go", "test"}, config.CommandBase)
		assert.True(t, config.Verbose)
		assert.True(t, config.Cover)
		assert.False(t, config.Race)
		// TestPath should use default since not specified
		assert.Equal(t, "./...", config.TestPath)
	})
}

func TestFindConfigFile(t *testing.T) {
	t.Run("finds .gotest-watch.yml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		err := os.WriteFile(configPath, []byte("test: true"), 0o600)
		require.NoError(t, err)

		found, err := FindConfigFile(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, configPath, found)
	})

	t.Run("finds .gotest-watch.yaml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gotest-watch.yaml")
		err := os.WriteFile(configPath, []byte("test: true"), 0o600)
		require.NoError(t, err)

		found, err := FindConfigFile(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, configPath, found)
	})

	t.Run("prefers .yml over .yaml when both exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		ymlPath := filepath.Join(tmpDir, ".gotest-watch.yml")
		yamlPath := filepath.Join(tmpDir, ".gotest-watch.yaml")

		err := os.WriteFile(ymlPath, []byte("yml: true"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(yamlPath, []byte("yaml: true"), 0o600)
		require.NoError(t, err)

		found, err := FindConfigFile(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, ymlPath, found)
	})

	t.Run("returns error when no config file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := FindConfigFile(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("doesn't search subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir", "nested")
		err := os.MkdirAll(subDir, 0o700)
		require.NoError(t, err)

		configPath := filepath.Join(subDir, ".gotest-watch.yml")
		err = os.WriteFile(configPath, []byte("test: true"), 0o600)
		require.NoError(t, err)

		found, err := FindConfigFile(tmpDir)
		require.Error(t, err)
		assert.Empty(t, found)
	})
}

// createTempYAMLFile creates a temporary YAML file with the given content
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}
