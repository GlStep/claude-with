package wrapper

import (
	"sort"

	"github.com/glstep/claude-with/internal/config"
)

func BuildEnv(base []string, p config.Profile) []string {
	env := append([]string(nil), base...)

	set := func(key, value string) {
		if value != "" {
			env = append(env, key+"="+value)
		}
	}

	set("ANTHROPIC_BASE_URL", p.BaseURL)
	set("ANTHROPIC_MODEL", p.Model)
	set("ANTHROPIC_API_KEY", p.ResolveAPIKey())

	keys := make([]string, 0, len(p.Env))
	for k := range p.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		set(k, p.Env[k])
	}

	return env
}
