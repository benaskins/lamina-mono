package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LLM configures the optional LLM integration for doc review.
type LLM struct {
	Provider  string `yaml:"provider"`    // "anthropic", "ollama", "openai"
	Model     string `yaml:"model"`       // model name
	APIKeyEnv string `yaml:"api_key_env"` // env var holding the API key
}

// APIKey reads the API key from the configured environment variable.
func (l *LLM) APIKey() (string, error) {
	key := os.Getenv(l.APIKeyEnv)
	if key == "" {
		return "", fmt.Errorf("environment variable %s is not set", l.APIKeyEnv)
	}
	return key, nil
}

// Config holds lamina workspace configuration.
type Config struct {
	LLM *LLM `yaml:"llm,omitempty"`
}

// LLMConfigured returns true if all required LLM fields are set.
func (c *Config) LLMConfigured() bool {
	return c.LLM != nil && c.LLM.Provider != "" && c.LLM.Model != "" && c.LLM.APIKeyEnv != ""
}

// DefaultPath returns the default config file path: ~/.config/lamina/config.yaml.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lamina", "config.yaml")
}

// Load reads a YAML config file. Returns an empty Config if the file doesn't exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
