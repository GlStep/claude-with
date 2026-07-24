package wrapper

import (
	"runtime"
	"testing"
)

func TestRunForwardsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}

	t.Run("success", func(t *testing.T) {
		code, err := Run("sh", []string{"-c", "exit 0"}, nil)
		if err != nil || code != 0 {
			t.Errorf("got code=%d err=%v, want 0/nil", code, err)
		}
	})

	t.Run("nonzero exit", func(t *testing.T) {
		code, err := Run("sh", []string{"-c", "exit 3"}, nil)
		if err != nil || code != 3 {
			t.Errorf("got code=%d err=%v, want 3/nil", code, err)
		}
	})

	t.Run("signal-terminated child", func(t *testing.T) {
		// The child kills itself with SIGKILL (9); shell convention is 128+9.
		code, err := Run("sh", []string{"-c", "kill -9 $$"}, nil)
		if err != nil || code != 137 {
			t.Errorf("got code=%d err=%v, want 137/nil", code, err)
		}
	})
}

func TestRunCommandNotFound(t *testing.T) {
	if _, err := Run("ccw-test-no-such-binary", nil, nil); err == nil {
		t.Error("expected error for missing binary, got nil")
	}
}
