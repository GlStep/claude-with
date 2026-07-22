package wrapper

import (
	"os"
	"os/exec"
)

func Run(claudeBin string, args []string, env []string) (int, error) {
	cmd := exec.Command(claudeBin, args...)
	cmd.Env = env
	
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return -1, err
	}
	
	return 0, nil
}
