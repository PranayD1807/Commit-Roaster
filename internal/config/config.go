package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the persistent user preferences for commit-roaster.
type Config struct {
	AIEnabled bool   `json:"ai_enabled"`
	Provider  string `json:"provider"`  // "gemini", "claude", or "chatgpt"
	Model     string `json:"model"`     // e.g. "gemini-2.5-flash", "claude-sonnet-4-20250514", "gpt-4.1"
	APIKey    string `json:"api_key"`
}

// ConfigDir returns the directory for commit-roaster configuration.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "commit-roaster"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config from disk. Returns a zero-value Config if the file
// does not exist. Returns an error only for unexpected I/O or parse failures.
func Load() (Config, error) {
	var cfg Config

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config yet — all defaults
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err // corrupt file — return zero config + error
	}

	return cfg, nil
}

// Save writes the config to disk with restrictive permissions.
// The config directory is created if it does not exist.
func Save(cfg Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	// 0700: owner-only read/write/execute on directory
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")

	// 0600: owner-only read/write — protects the API key
	return os.WriteFile(path, data, 0600)
}

// MaskKey returns a masked version of the API key for display.
// Shows only the first 4 and last 4 characters.
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// ProviderModels returns the available models for each provider.
func ProviderModels(provider string) []string {
	switch provider {
	case "gemini":
		return []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-1.5-flash", "gemini-1.5-pro"}
	case "claude":
		return []string{"claude-3-5-sonnet-latest", "claude-3-5-haiku-latest", "claude-sonnet-4-20250514", "claude-haiku-4-20250414"}
	case "chatgpt":
		return []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4.1"}
	default:
		return nil
	}
}

// ProviderDisplayName returns a human-friendly name for a provider.
func ProviderDisplayName(provider string) string {
	switch provider {
	case "gemini":
		return "Google Gemini"
	case "claude":
		return "Claude"
	case "chatgpt":
		return "ChatGPT"
	default:
		return provider
	}
}
