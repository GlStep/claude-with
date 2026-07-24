package wrapper

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run executes claudeBin with the given args and environment, wired to this
// process's stdio, and returns the child's exit code.
func Run(claudeBin string, args []string, env []string) (int, error) {
	cmd := exec.Command(claudeBin, args...)
	cmd.Env = env

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ctrl+C sends SIGINT to the whole foreground process group — ccw and
	// claude alike. claude traps it (to interrupt the current generation and
	// keep running), so ccw must not die and orphan it; ignore the signal and
	// wait for claude to exit on its own terms.
	signal.Ignore(os.Interrupt, syscall.SIGQUIT)
	defer signal.Reset(os.Interrupt, syscall.SIGQUIT)

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				// Shell convention for a signal-terminated child.
				return 128 + int(status.Signal()), nil
			}
			return exitError.ExitCode(), nil
		}
		return -1, err
	}

	return 0, nil
}
