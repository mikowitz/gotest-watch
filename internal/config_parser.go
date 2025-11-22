package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadConfigFromYAML(file string) (*TestConfig, error) {
	file = filepath.Clean(file)
	config, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	tc := NewTestConfig()
	err = yaml.Unmarshal(config, tc)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

func FindConfigFile(dirpath string) (string, error) {
	ymlPath := filepath.Join(dirpath, ".gotest-watch.yml")
	if _, err := os.Stat(ymlPath); err == nil {
		return ymlPath, nil
	}
	yamlPath := filepath.Join(dirpath, ".gotest-watch.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return yamlPath, nil
	}
	return "", fmt.Errorf("gotest-watch config file not found")
}
