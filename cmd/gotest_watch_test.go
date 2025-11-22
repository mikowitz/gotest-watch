package cmd

import (
	"testing"

	"github.com/mikowitz/gotest-watch/internal"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// createTestCommand creates a fresh command with all flags for isolated testing
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "gotest-watch",
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose test output")
	cmd.Flags().StringVarP(&runPattern, "run", "r", "", "run tests that match this pattern")
	cmd.Flags().StringVarP(&skipPattern, "skip", "s", "", "skip tests that match this pattern")
	cmd.Flags().IntVarP(&count, "count", "n", 0, "number of times to run each test")
	cmd.Flags().BoolVarP(&clearScreen, "cls", "l", false, "clear the screen before each test run")
	cmd.Flags().BoolVarP(&color, "color", "c", false, "ANSI color output")
	cmd.Flags().StringVarP(&commandBase, "cmd", "m", "go test", "base command to run")
	cmd.Flags().StringVarP(&testPath, "path", "p", "./...", "directory to run tests in")
	return cmd
}

func TestOverrideConfig(t *testing.T) {
	t.Run("unset flags preserve config values", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)
		config.SetRunPattern("TestFoo")
		config.SetSkipPattern("TestBar")
		config.SetCount(5)
		config.ClearScreen = true
		config.Color = true
		config.SetCommandBase([]string{"richgo", "test"})
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{}) // No flags set

		overrideConfig(config, cmd)

		// All config values should be preserved
		assert.True(t, config.GetVerbose())
		assert.Equal(t, "TestFoo", config.GetRunPattern())
		assert.Equal(t, "TestBar", config.GetSkipPattern())
		assert.Equal(t, 5, config.GetCount())
		assert.True(t, config.GetClearScreen())
		assert.True(t, config.GetColor())
		assert.Equal(t, []string{"richgo", "test"}, config.GetCommandBase())
		assert.Equal(t, "./pkg/...", config.GetTestPath())
	})

	t.Run("only explicitly set flags override config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)
		config.SetRunPattern("TestFoo")
		config.SetSkipPattern("TestBar")
		config.SetCount(5)
		config.ClearScreen = true
		config.Color = true
		config.SetCommandBase([]string{"richgo", "test"})
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--run=TestNew", "--count=10"}) // Only set run and count

		overrideConfig(config, cmd)

		// Explicitly set flags should override
		assert.Equal(t, "TestNew", config.GetRunPattern())
		assert.Equal(t, 10, config.GetCount())

		// All other config values should be preserved
		assert.True(t, config.GetVerbose())
		assert.Equal(t, "TestBar", config.GetSkipPattern())
		assert.True(t, config.GetClearScreen())
		assert.True(t, config.GetColor())
		assert.Equal(t, []string{"richgo", "test"}, config.GetCommandBase())
		assert.Equal(t, "./pkg/...", config.GetTestPath())
	})

	t.Run("can set boolean flags to false explicitly", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)
		config.ClearScreen = true
		config.Color = true

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--verbose=false"}) // Explicitly set to false

		overrideConfig(config, cmd)

		// Explicitly set to false should override
		assert.False(t, config.GetVerbose())
		// Non-set booleans should preserve config values
		assert.True(t, config.GetClearScreen())
		assert.True(t, config.GetColor())
	})

	t.Run("empty string flag clears config value when explicitly set", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetRunPattern("TestFoo")
		config.SetSkipPattern("TestBar")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--run="}) // Explicitly set to empty

		overrideConfig(config, cmd)

		// Run pattern should be cleared
		assert.Equal(t, "", config.GetRunPattern())
		// Skip pattern should be preserved
		assert.Equal(t, "TestBar", config.GetSkipPattern())
	})

	t.Run("all flags override when all are set", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)
		config.SetRunPattern("TestFoo")
		config.SetSkipPattern("TestBar")
		config.SetCount(5)
		config.ClearScreen = true
		config.Color = true
		config.SetCommandBase([]string{"richgo", "test"})
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{
			"--verbose=false",
			"--run=TestCLI",
			"--skip=TestSkipCLI",
			"--count=1",
			"--cls=false",
			"--color=false",
			"--cmd=go test -tags integration",
			"--path=./cli/...",
		})

		overrideConfig(config, cmd)

		// All values should be overridden
		assert.False(t, config.GetVerbose())
		assert.Equal(t, "TestCLI", config.GetRunPattern())
		assert.Equal(t, "TestSkipCLI", config.GetSkipPattern())
		assert.Equal(t, 1, config.GetCount())
		assert.False(t, config.GetClearScreen())
		assert.False(t, config.GetColor())
		assert.Equal(t, []string{"go", "test", "-tags", "integration"}, config.GetCommandBase())
		assert.Equal(t, "./cli/...", config.GetTestPath())
	})
}

func TestVerboseFlag(t *testing.T) {
	t.Run("no flag with default config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(false)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.False(t, config.GetVerbose())
	})

	t.Run("no flag preserves true config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.True(t, config.GetVerbose())
	})

	t.Run("short flag overrides false config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(false)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"-v"})

		overrideConfig(config, cmd)

		assert.True(t, config.GetVerbose())
	})

	t.Run("long flag overrides false config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(false)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--verbose"})

		overrideConfig(config, cmd)

		assert.True(t, config.GetVerbose())
	})

	t.Run("explicit false flag overrides true config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetVerbose(true)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--verbose=false"})

		overrideConfig(config, cmd)

		assert.False(t, config.GetVerbose())
	})
}

func TestRunPatternFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetRunPattern("TestFoo")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.Equal(t, "TestFoo", config.GetRunPattern())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetRunPattern("TestFoo")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--run=TestBar"})

		overrideConfig(config, cmd)

		assert.Equal(t, "TestBar", config.GetRunPattern())
	})

	t.Run("short flag works", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetRunPattern("TestFoo")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"-r", "TestBar"})

		overrideConfig(config, cmd)

		assert.Equal(t, "TestBar", config.GetRunPattern())
	})
}

func TestSkipPatternFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetSkipPattern("TestFoo")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.Equal(t, "TestFoo", config.GetSkipPattern())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetSkipPattern("TestFoo")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--skip=TestBar"})

		overrideConfig(config, cmd)

		assert.Equal(t, "TestBar", config.GetSkipPattern())
	})
}

func TestCountFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetCount(5)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.Equal(t, 5, config.GetCount())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetCount(5)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--count=10"})

		overrideConfig(config, cmd)

		assert.Equal(t, 10, config.GetCount())
	})

	t.Run("short flag works", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetCount(5)

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"-n", "10"})

		overrideConfig(config, cmd)

		assert.Equal(t, 10, config.GetCount())
	})
}

func TestClearScreenFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.ClearScreen = true

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.True(t, config.GetClearScreen())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.ClearScreen = false

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--cls"})

		overrideConfig(config, cmd)

		assert.True(t, config.GetClearScreen())
	})

	t.Run("explicit false overrides true config", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.ClearScreen = true

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--cls=false"})

		overrideConfig(config, cmd)

		assert.False(t, config.GetClearScreen())
	})
}

func TestColorFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.Color = true

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.True(t, config.GetColor())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.Color = false

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--color"})

		overrideConfig(config, cmd)

		assert.True(t, config.GetColor())
	})
}

func TestCommandBaseFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetCommandBase([]string{"richgo", "test"})

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.Equal(t, []string{"richgo", "test"}, config.GetCommandBase())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetCommandBase([]string{"richgo", "test"})

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--cmd=go test -tags integration"})

		overrideConfig(config, cmd)

		assert.Equal(t, []string{"go", "test", "-tags", "integration"}, config.GetCommandBase())
	})
}

func TestTestPathFlag(t *testing.T) {
	t.Run("no flag preserves config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{})

		overrideConfig(config, cmd)

		assert.Equal(t, "./pkg/...", config.GetTestPath())
	})

	t.Run("flag overrides config value", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"--path=./cli/..."})

		overrideConfig(config, cmd)

		assert.Equal(t, "./cli/...", config.GetTestPath())
	})

	t.Run("short flag works", func(t *testing.T) {
		config := internal.NewTestConfig()
		config.SetTestPath("./pkg/...")

		cmd := createTestCommand()
		cmd.ParseFlags([]string{"-p", "./cli/..."})

		overrideConfig(config, cmd)

		assert.Equal(t, "./cli/...", config.GetTestPath())
	})
}
