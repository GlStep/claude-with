package cli

import (
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/glstep/claude-with/internal/config"
	"github.com/glstep/claude-with/internal/wrapper"
)

// version is overridden at release time via
// -ldflags "-X github.com/glstep/claude-with/internal/cli.version=v1.2.3".
var version = "dev"

func Run(args []string) int {
	var dryRun bool
	dryRun, args = extractFlag(args, "--dry-run")

	// Handle everything that doesn't need a config file before loading it,
	// so --help and --version work on a fresh install.
	if len(args) > 0 {
		switch args[0] {
		case "init":
			return runInit(args[1:], dryRun)
		case "--help", "-h":
			printHelp()
			return 0
		case "--version", "-v":
			printVersion()
			return 0
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		errMsg("loading config: %v", err)
		return 1
	}

	if len(args) > 0 && (args[0] == "--list" || args[0] == "-l") {
		printProfiles(cfg)
		return 0
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
		fmt.Println("Command: " + formatCommand(claudeBin, args))
		fmt.Println("Environment variables:")
		for _, env := range partEnv {
			fmt.Println("  " + redactEnv(env))
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
			rest := make([]string, 0, len(args)-1)
			rest = append(rest, args[:i]...)
			rest = append(rest, args[i+1:]...)
			return true, rest
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
	fmt.Fprintln(w, "  --dry-run\tPrint what would happen without doing it (works with init too).")
	fmt.Fprintln(w, "  init\tCreate a starter config file.")
	w.Flush()

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ccw local \"hi\"")
	fmt.Println("  ccw local -p \"hi\"")
	fmt.Println("  ccw")
}

func printVersion() {
	v := version
	if v == "dev" {
		// Not a release build; fall back to the module version stamped by
		// `go install github.com/glstep/claude-with/cmd/ccw@vX.Y.Z`.
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}
	fmt.Printf("ccw version %s\n", v)
}

// formatCommand renders a command line for display, quoting only the
// arguments that need it.
func formatCommand(bin string, args []string) string {
	parts := []string{bin}
	for _, a := range args {
		if a == "" || strings.ContainsAny(a, " \t\"'") {
			a = strconv.Quote(a)
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}

// sensitiveEnvFragments marks env-var names whose values must not appear in
// dry-run output. Matched case-insensitively against the variable name, so it
// also covers user-defined vars from the [profiles.NAME.env] table.
var sensitiveEnvFragments = []string{"KEY", "TOKEN", "SECRET", "PASSWORD"}

func redactEnv(entry string) string {
	name, _, ok := strings.Cut(entry, "=")
	if !ok {
		return entry
	}
	upper := strings.ToUpper(name)
	for _, frag := range sensitiveEnvFragments {
		if strings.Contains(upper, frag) {
			return name + "=<REDACTED>"
		}
	}
	return entry
}

func errMsg(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
