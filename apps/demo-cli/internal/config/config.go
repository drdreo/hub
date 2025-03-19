package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Step represents a single demo step
type Step struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`    // "generate", "modify", "execute"
	Command  string `yaml:"command"` // For execute type
	Template string `yaml:"template"`
	Target   string `yaml:"target"`
	Match    string `yaml:"match"`   // For modify type
	Replace  string `yaml:"replace"` // For modify type
}

// Config represents the demo configuration
type Config struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}

// LoadConfig loads and parses a demo config file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
