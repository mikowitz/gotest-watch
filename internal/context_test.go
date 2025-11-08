package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithConfig_StoresConfigInContext tests that WithConfig stores config in context
func TestWithConfig_StoresConfigInContext(t *testing.T) {
	config := &TestConfig{
		TestPath:    "./test",
		Verbose:     true,
		RunPattern:  "TestFoo",
		SkipPattern: "TestBar",
	}

	ctx := context.Background()
	ctxWithConfig := WithConfig(ctx, config)

	// Verify context is not nil
	require.NotNil(t, ctxWithConfig, "context should not be nil")

	// Verify we can retrieve the config
	retrievedConfig := getConfig(ctxWithConfig)
	require.NotNil(t, retrievedConfig, "retrieved config should not be nil")

	// Verify it's the same config
	assert.Equal(t, config, retrievedConfig, "retrieved config should be the same as stored config")
	assert.Equal(t, "./test", retrievedConfig.GetTestPath(), "test path should match")
	assert.True(t, retrievedConfig.GetVerbose(), "verbose should match")
	assert.Equal(t, "TestFoo", retrievedConfig.GetRunPattern(), "run pattern should match")
	assert.Equal(t, "TestBar", retrievedConfig.GetSkipPattern(), "skip pattern should match")
}

// TestGetConfig_RetrievesConfigFromContext tests that getConfig retrieves config from context
func TestGetConfig_RetrievesConfigFromContext(t *testing.T) {
	config := NewTestConfig()
	config.SetVerbose(true)
	config.SetRunPattern("TestExample")

	ctx := WithConfig(context.Background(), config)
	retrieved := getConfig(ctx)

	require.NotNil(t, retrieved, "retrieved config should not be nil")
	assert.Equal(t, config, retrieved, "should retrieve the same config instance")
	assert.True(t, retrieved.GetVerbose(), "verbose should be true")
	assert.Equal(t, "TestExample", retrieved.GetRunPattern(), "run pattern should match")
}

// TestGetConfig_ReturnsNilIfNotInContext tests that getConfig returns nil if config not in context
func TestGetConfig_ReturnsNilIfNotInContext(t *testing.T) {
	ctx := context.Background()
	retrieved := getConfig(ctx)

	assert.Nil(t, retrieved, "should return nil when config not in context")
}

// TestGetConfig_ReturnsNilForWrongValueType tests that getConfig returns nil for wrong type
func TestGetConfig_ReturnsNilForWrongValueType(t *testing.T) {
	// Store a string instead of *TestConfig
	ctx := context.WithValue(context.Background(), configKey{}, "not a config")
	retrieved := getConfig(ctx)

	assert.Nil(t, retrieved, "should return nil when value is wrong type")
}

// TestWithConfig_PreservesParentContext tests that WithConfig preserves parent context
func TestWithConfig_PreservesParentContext(t *testing.T) {
	type testKey struct{}
	parentCtx := context.WithValue(context.Background(), testKey{}, "parent value")

	config := NewTestConfig()
	ctxWithConfig := WithConfig(parentCtx, config)

	// Should still be able to get parent value
	parentValue := ctxWithConfig.Value(testKey{})
	assert.Equal(t, "parent value", parentValue, "should preserve parent context values")

	// Should also be able to get config
	retrievedConfig := getConfig(ctxWithConfig)
	assert.Equal(t, config, retrievedConfig, "should have config value")
}

// TestWithConfig_CanBeChained tests that multiple WithConfig calls can be chained
func TestWithConfig_CanBeChained(t *testing.T) {
	config1 := NewTestConfig()
	config1.SetVerbose(true)

	config2 := NewTestConfig()
	config2.SetVerbose(false)
	config2.SetRunPattern("TestNew")

	ctx1 := WithConfig(context.Background(), config1)
	ctx2 := WithConfig(ctx1, config2)

	// The second config should override the first
	retrieved := getConfig(ctx2)
	assert.Equal(t, config2, retrieved, "should get the most recent config")
	assert.False(t, retrieved.GetVerbose(), "should have config2's verbose setting")
	assert.Equal(t, "TestNew", retrieved.GetRunPattern(), "should have config2's run pattern")
}
