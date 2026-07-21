// Package config loads claude-with's TOML config, which defines named
// profiles (base URL, model, API key) for local model backends.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Profile describes one local-model backend that claude-with can point
// the claude CLI at.
type Profile struct {
	BaseURL   string            `toml:"base_url"`
	Model     string            `toml:"model"`
	APIKey    string            `toml:"api_key"`
	APIKeyEnv string            `toml:"api_key_env"`
	Env       map[string]string `toml:"env"`
}

// Config is the top-level shape of config.toml.
type Config struct {
	DefaultProfile string             `toml:"default_profile"`
	Profiles       map[string]Profile `toml:"profiles"`
}

// ResolveAPIKey returns the profile's API key, preferring a value read
// from APIKeyEnv (so secrets don't have to live in the config file) over
// the literal APIKey field.
func (p Profile) ResolveAPIKey() string {
	if p.APIKeyEnv != "" {
		return os.Getenv(p.APIKeyEnv)
	}
	return p.APIKey
}

// Path returns the config file location: $CLAUDE_WITH_CONFIG if set,
// otherwise $XDG_CONFIG_HOME/claude-with/config.toml, otherwise
// ~/.config/claude-with/config.toml.
func Path() (string, error) {
	if p := os.Getenv("CLAUDE_WITH_CONFIG"); p != "" {
		return p, nil
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "claude-with", "config.toml"), nil
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	return &cfg, nil
}

// Resolve picks the named profile, falling back to DefaultProfile when
// name is empty. It errors if no profile can be determined.
func (c *Config) Resolve(name string) (string, Profile, error) {
	if name == "" {
		name = c.DefaultProfile
	}
	if name == "" {
		return "", Profile{}, fmt.Errorf("no profile given and no default_profile set in config")
	}
	p, ok := c.Profiles[name]
	if !ok {
		return "", Profile{}, fmt.Errorf("unknown profile %q", name)
	}
	return name, p, nil
}
