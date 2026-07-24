package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/glstep/claude-with/internal/config"
)

const configTemplate = `default_profile = "local"

[profiles.local]
base_url = "http://localhost:11434/v1"
model = "llama3.1"
# api_key_env = "OLLAMA_API_KEY"  # reads the key from this env var at runtime
# api_key = "sk-..."              # avoid this if you plan to commit this file
`

func runInit(args []string, dryRun bool) int {
	force := false
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printInitHelp()
			return 0
		case "--force", "-f":
			force = true
		default:
			errMsg("init: unknown argument %q (see `ccw init --help`)", arg)
			return 1
		}
	}

	path, err := config.Path()
	if err != nil {
		errMsg("getting config path: %v", err)
		return 1
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			errMsg("config already exists at %s (use --force to overwrite)", path)
			return 1
		}
	}

	if dryRun {
		fmt.Printf("Dry run mode. Would create config at %s with contents:\n\n%s", path, configTemplate)
		return 0
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		errMsg("creating config directory: %v", err)
		return 1
	}

	if err := os.WriteFile(path, []byte(configTemplate), 0o644); err != nil {
		errMsg("writing config: %v", err)
		return 1
	}

	fmt.Printf("Created config at %s\n", path)
	return 0
}

func printInitHelp() {
	fmt.Println("Usage: ccw init [options]")
	fmt.Println("\nOptions:")

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  -h, --help\tShow this help message")
	fmt.Fprintln(w, "  -f, --force\tForce initialization even if config already exists")
	fmt.Fprintln(w, "  --dry-run\tShow what would be written without creating the file")
	w.Flush()
}
