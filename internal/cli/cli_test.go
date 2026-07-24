package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExtractFlag(t *testing.T) {
	t.Run("flag at the front", func(t *testing.T) {
		found, rest := extractFlag([]string{"--dry-run", "local", "hi"}, "--dry-run")
		if !found {
			t.Error("expected flag to be found")
		}
		if want := []string{"local", "hi"}; !reflect.DeepEqual(rest, want) {
			t.Errorf("rest = %v, want %v", rest, want)
		}
	})

	t.Run("flag in the middle", func(t *testing.T) {
		found, rest := extractFlag([]string{"local", "--dry-run", "hi"}, "--dry-run")
		if !found {
			t.Error("expected flag to be found")
		}
		if want := []string{"local", "hi"}; !reflect.DeepEqual(rest, want) {
			t.Errorf("rest = %v, want %v", rest, want)
		}
	})

	t.Run("flag absent", func(t *testing.T) {
		args := []string{"local", "hi"}
		found, rest := extractFlag(args, "--dry-run")
		if found {
			t.Error("expected flag not to be found")
		}
		if !reflect.DeepEqual(rest, args) {
			t.Errorf("rest = %v, want %v", rest, args)
		}
	})

	t.Run("does not mutate input", func(t *testing.T) {
		args := []string{"local", "--dry-run", "hi"}
		extractFlag(args, "--dry-run")
		if want := []string{"local", "--dry-run", "hi"}; !reflect.DeepEqual(args, want) {
			t.Errorf("input was mutated: %v, want %v", args, want)
		}
	})
}

func TestRedactEnv(t *testing.T) {
	tests := []struct {
		entry string
		want  string
	}{
		{"ANTHROPIC_API_KEY=sk-secret", "ANTHROPIC_API_KEY=<REDACTED>"},
		{"OPENAI_API_KEY=sk-other", "OPENAI_API_KEY=<REDACTED>"},
		{"HF_TOKEN=hf_abc", "HF_TOKEN=<REDACTED>"},
		{"my_secret=hunter2", "my_secret=<REDACTED>"},
		{"DB_PASSWORD=hunter2", "DB_PASSWORD=<REDACTED>"},
		{"ANTHROPIC_BASE_URL=http://localhost:11434/v1", "ANTHROPIC_BASE_URL=http://localhost:11434/v1"},
		{"ANTHROPIC_MODEL=llama3.1", "ANTHROPIC_MODEL=llama3.1"},
		{"PATH=/usr/bin", "PATH=/usr/bin"},
		// value containing a sensitive word is fine; only the name matters
		{"GREETING=my token of appreciation", "GREETING=my token of appreciation"},
		{"malformed-no-equals", "malformed-no-equals"},
	}
	for _, tt := range tests {
		if got := redactEnv(tt.entry); got != tt.want {
			t.Errorf("redactEnv(%q) = %q, want %q", tt.entry, got, tt.want)
		}
	}
}

func TestRunHelpAndVersionWithoutConfig(t *testing.T) {
	// --help and --version must work even when no config file exists.
	t.Setenv("CLAUDE_WITH_CONFIG", "/nonexistent/ccw-test/config.toml")

	for _, args := range [][]string{{"--help"}, {"-h"}, {"--version"}, {"-v"}} {
		if code := Run(args); code != 0 {
			t.Errorf("Run(%v) = %d, want 0", args, code)
		}
	}
}

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{nil, "claude"},
		{[]string{"hi"}, "claude hi"},
		{[]string{"-p", "hi there"}, `claude -p "hi there"`},
		{[]string{""}, `claude ""`},
	}
	for _, tt := range tests {
		if got := formatCommand("claude", tt.args); got != tt.want {
			t.Errorf("formatCommand(claude, %v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

func TestRunInit(t *testing.T) {
	// Points the config path into a fresh temp dir and returns it.
	setup := func(t *testing.T) string {
		path := filepath.Join(t.TempDir(), "config.toml")
		t.Setenv("CLAUDE_WITH_CONFIG", path)
		return path
	}

	t.Run("creates config", func(t *testing.T) {
		path := setup(t)
		if code := runInit(nil, false); code != 0 {
			t.Fatalf("runInit = %d, want 0", code)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("config not written: %v", err)
		}
		if !strings.Contains(string(data), "default_profile") {
			t.Error("written config does not look like the template")
		}
	})

	t.Run("refuses to overwrite without force", func(t *testing.T) {
		path := setup(t)
		if err := os.WriteFile(path, []byte("mine"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code := runInit(nil, false); code != 1 {
			t.Errorf("runInit = %d, want 1", code)
		}
		data, _ := os.ReadFile(path)
		if string(data) != "mine" {
			t.Error("existing config was overwritten")
		}
	})

	t.Run("force overwrites", func(t *testing.T) {
		path := setup(t)
		if err := os.WriteFile(path, []byte("mine"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code := runInit([]string{"--force"}, false); code != 0 {
			t.Errorf("runInit = %d, want 0", code)
		}
		data, _ := os.ReadFile(path)
		if string(data) != configTemplate {
			t.Error("config was not overwritten with the template")
		}
	})

	t.Run("dry run writes nothing", func(t *testing.T) {
		path := setup(t)
		if code := runInit(nil, true); code != 0 {
			t.Errorf("runInit = %d, want 0", code)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("dry run created the config file")
		}
	})

	t.Run("unknown argument errors and writes nothing", func(t *testing.T) {
		path := setup(t)
		if code := runInit([]string{"--forc"}, false); code != 1 {
			t.Errorf("runInit = %d, want 1", code)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("config was created despite the unknown argument")
		}
	})
}
