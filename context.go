package main

import (
	"context"
)

type configKey struct{}

func withConfig(ctx context.Context, config *TestConfig) context.Context {
	return context.WithValue(ctx, configKey{}, config)
}

func getConfig(ctx context.Context) *TestConfig {
	if config := ctx.Value(configKey{}); config != nil {
		switch cfg := config.(type) {
		case *TestConfig:
			return cfg
		default:
			return nil
		}
	}
	return nil
}
