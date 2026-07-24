package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/glstep/claude-with/internal/config"
	"github.com/glstep/claude-with/internal/wrapper"
)

var version = "dev"

func Run(args []string) int {
	var dryRun bool
	dryRun, args = extractFlag(args, "--dry-run")

	if len(args) > 0 && args[0] == "init" {
		return runInit(args[1:])
	}

	cfg, err := loadConfig()
	if err != nil {
		errMsg("loading config: %v", err)
		return 1
	}

	if len(args) > 0 {
		if args[0] == "--help" || args[0] == "-h" {
			printHelp()
			return 0
		}

		if args[0] == "--list" || args[0] == "-l" {
			printProfiles(cfg)
			return 0
		}

		if args[0] == "--version" || args[0] == "-v" {
			printVersion()
			return 0
		}
	}

	profileName := ""
	if len(args) > 0 {
		profileName = args[0]
		args = args[1:]
	}

	_, profile, err := cfg.Resolve(profileName)
	if err != nil {
		errMsg("resolving profile: %v", err)
		return 1
	}

	claudeBin := "claude"

	if dryRun {
		emptyEnv := []string(nil)
		partEnv := wrapper.BuildEnv(emptyEnv, profile)

		fmt.Println("Dry run mode. The following command would be executed:")
		fmt.Printf("Command: %s %v\n", claudeBin, args)
		fmt.Println("Environment variables:")
		for _, env := range partEnv {
			if strings.HasPrefix(env, "ANTHROPIC_API_KEY=") {
				fmt.Println("  " + "ANTHROPIC_API_KEY=<REDACTED>")
			} else {
				fmt.Println("  " + env)
			}
		}
		return 0
	}

	baseEnv := os.Environ()
	fullEnv := wrapper.BuildEnv(baseEnv, profile)

	exitCode, err := wrapper.Run(claudeBin, args, fullEnv)
	if err != nil {
		errMsg("running claude: %v", err)
		return 1
	}

	return exitCode
}

func loadConfig() (*config.Config, error) {
	path, err := config.Path()
	if err != nil {
		return nil, err
	}
	return config.Load(path)
}

func extractFlag(args []string, flag string) (bool, []string) {
	for i, arg := range args {
		if arg == flag {
			return true, append(args[:i], args[i+1:]...)
		}
	}
	return false, args
}

func printProfiles(cfg *config.Config) {
	profiles := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}
	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
		return
	}
	sort.Strings(profiles)

	fmt.Println("Available profiles:")
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tBASE URL\tMODEL")
	for _, name := range profiles {
		p := cfg.Profiles[name]
		if name == cfg.DefaultProfile {
			name += " (default)"
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\n", name, p.BaseURL, p.Model)
	}
	w.Flush()
}

func printHelp() {
	fmt.Println("Usage: ccw [profile_name] [args...]")
	fmt.Println()
	fmt.Println("Commands:")

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  [profile_name]\tThe profile to use. Defaults to default_profile in config if omitted.")
	fmt.Fprintln(w, "  [args...]\tThe arguments to pass to the claude binary.")
	fmt.Fprintln(w, "  --help, -h\tShow this help message.")
	fmt.Fprintln(w, "  --version, -v\tShow the version of ccw.")
	fmt.Fprintln(w, "  --list, -l\tList all available profiles.")
	fmt.Fprintln(w, "  --dry-run\tPrint the resolved command and env vars without running claude.")
	fmt.Fprintln(w, "  init\tCreate a starter config file.")
	w.Flush()

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ccw local --chat \"hi\"")
	fmt.Println("  ccw")
}

func printVersion() {
	fmt.Printf("ccw version %s\n", version)
}

func errMsg(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
