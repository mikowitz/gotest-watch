package internal

import "log"

func LoadOrDefaultConfig(dirpath string) *TestConfig {
	filepath, err := FindConfigFile(dirpath)
	if err != nil {
		return NewTestConfig()
	}

	config, err := LoadConfigFromYAML(filepath)
	if err != nil {
		log.Printf("Warning: failed to parse config file %s: %v", filepath, err)
		return NewTestConfig()
	}

	return config
}
