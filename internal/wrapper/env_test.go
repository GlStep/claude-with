package wrapper

import (
	"reflect"
	"testing"

	"github.com/glstep/claude-with/internal/config"
)

func TestBuildEnv(t *testing.T) {
	t.Run("sets profile vars on top of base", func(t *testing.T) {
		base := []string{"PATH=/usr/bin", "HOME=/home/u"}
		p := config.Profile{
			BaseURL: "http://localhost:11434/v1",
			Model:   "llama3.1",
			APIKey:  "sk-test",
		}

		got := BuildEnv(base, p)
		want := []string{
			"PATH=/usr/bin",
			"HOME=/home/u",
			"ANTHROPIC_BASE_URL=http://localhost:11434/v1",
			"ANTHROPIC_MODEL=llama3.1",
			"ANTHROPIC_API_KEY=sk-test",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("skips empty values", func(t *testing.T) {
		got := BuildEnv(nil, config.Profile{Model: "llama3.1"})
		want := []string{"ANTHROPIC_MODEL=llama3.1"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("extra env vars are sorted", func(t *testing.T) {
		p := config.Profile{
			Env: map[string]string{"ZED": "3", "ALPHA": "1", "MID": "2"},
		}
		got := BuildEnv(nil, p)
		want := []string{"ALPHA=1", "MID=2", "ZED=3"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("does not mutate base", func(t *testing.T) {
		base := []string{"PATH=/usr/bin"}
		BuildEnv(base, config.Profile{Model: "m"})
		if !reflect.DeepEqual(base, []string{"PATH=/usr/bin"}) {
			t.Errorf("base was mutated: %v", base)
		}
	})
}
