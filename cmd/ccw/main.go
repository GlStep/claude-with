package main

import (
	"fmt"
	"os"

	"github.com/glstep/claude-with/internal/config"
	"github.com/glstep/claude-with/internal/wrapper"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "No argument detected, please provide at least one argument.")
		os.Exit(1)
		return
	}
	
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println("Usage: ccw <profile_name> [args...]")
		fmt.Println("Example: ccw my_profile --option value")
		os.Exit(0)
		return
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Println("ccw version 1.0.0")
		os.Exit(0)
		return
	}
	
	profileName := os.Args[1]
	args := os.Args[2:]

	path, err := config.Path()
	if err != nil {
		errMsg("determining config path: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		errMsg("loading config: %v", err)
	}

	_, profile, err := cfg.Resolve(profileName)
	if err != nil {
		errMsg("resolving profile: %v", err)
	}

	claudeBin := "claude"

	baseEnv := os.Environ()

	fullEnv := wrapper.BuildEnv(baseEnv, profile)

	exitCode, err := wrapper.Run(claudeBin, args, fullEnv)
	if err != nil {
		errMsg("running claude: %v", err)
	}
	os.Exit(exitCode)
}

// errMsg prints an error to stderr and exits the process with status 1.
func errMsg(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
