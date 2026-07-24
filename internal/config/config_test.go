package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	cfg := &Config{
		DefaultProfile: "local",
		Profiles: map[string]Profile{
			"local":  {BaseURL: "http://localhost:11434/v1", Model: "llama3.1"},
			"remote": {BaseURL: "https://example.com/v1", Model: "big-model"},
		},
	}

	t.Run("named profile", func(t *testing.T) {
		name, p, err := cfg.Resolve("remote")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "remote" || p.Model != "big-model" {
			t.Errorf("got name=%q model=%q, want remote/big-model", name, p.Model)
		}
	})

	t.Run("falls back to default", func(t *testing.T) {
		name, p, err := cfg.Resolve("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "local" || p.Model != "llama3.1" {
			t.Errorf("got name=%q model=%q, want local/llama3.1", name, p.Model)
		}
	})

	t.Run("unknown profile", func(t *testing.T) {
		if _, _, err := cfg.Resolve("nope"); err == nil {
			t.Error("expected error for unknown profile, got nil")
		}
	})

	t.Run("no name and no default", func(t *testing.T) {
		empty := &Config{Profiles: cfg.Profiles}
		if _, _, err := empty.Resolve(""); err == nil {
			t.Error("expected error when no profile given and no default set, got nil")
		}
	})
}

func TestResolveAPIKey(t *testing.T) {
	t.Run("literal key", func(t *testing.T) {
		p := Profile{APIKey: "sk-literal"}
		if got := p.ResolveAPIKey(); got != "sk-literal" {
			t.Errorf("got %q, want sk-literal", got)
		}
	})

	t.Run("env var wins over literal", func(t *testing.T) {
		t.Setenv("CCW_TEST_API_KEY", "sk-from-env")
		p := Profile{APIKey: "sk-literal", APIKeyEnv: "CCW_TEST_API_KEY"}
		if got := p.ResolveAPIKey(); got != "sk-from-env" {
			t.Errorf("got %q, want sk-from-env", got)
		}
	})

	t.Run("env var set but empty", func(t *testing.T) {
		p := Profile{APIKey: "sk-literal", APIKeyEnv: "CCW_TEST_UNSET_KEY"}
		if got := p.ResolveAPIKey(); got != "" {
			t.Errorf("got %q, want empty string when APIKeyEnv is unset", got)
		}
	})
}

func TestPath(t *testing.T) {
	t.Run("explicit override", func(t *testing.T) {
		t.Setenv("CLAUDE_WITH_CONFIG", "/some/where/config.toml")
		got, err := Path()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/some/where/config.toml" {
			t.Errorf("got %q, want /some/where/config.toml", got)
		}
	})

	t.Run("XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("CLAUDE_WITH_CONFIG", "")
		t.Setenv("XDG_CONFIG_HOME", "/xdg")
		got, err := Path()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join("/xdg", "claude-with", "config.toml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("home fallback", func(t *testing.T) {
		t.Setenv("CLAUDE_WITH_CONFIG", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		got, err := Path()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "claude-with", "config.toml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		content := `default_profile = "local"

[profiles.local]
base_url = "http://localhost:11434/v1"
model = "llama3.1"

[profiles.local.env]
FOO = "bar"
`
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProfile != "local" {
			t.Errorf("DefaultProfile = %q, want local", cfg.DefaultProfile)
		}
		p, ok := cfg.Profiles["local"]
		if !ok {
			t.Fatal("profile local missing")
		}
		if p.BaseURL != "http://localhost:11434/v1" || p.Model != "llama3.1" {
			t.Errorf("unexpected profile: %+v", p)
		}
		if p.Env["FOO"] != "bar" {
			t.Errorf("Env[FOO] = %q, want bar", p.Env["FOO"])
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := Load(filepath.Join(t.TempDir(), "nope.toml")); err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})

	t.Run("malformed file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := os.WriteFile(path, []byte("not [valid toml"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Error("expected error for malformed file, got nil")
		}
	})
}
