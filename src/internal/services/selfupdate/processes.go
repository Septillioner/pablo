package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	processExitPollAttempts = 10
	processExitPollDelay    = 200 * time.Millisecond
)

type processInfo struct {
	PID  int
	Name string
}

func normalizeExecutablePath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}

	abs, err := filepath.Abs(resolved)
	if err != nil {
		return "", err
	}

	return filepath.Clean(abs), nil
}

func executablePathsMatch(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if left == right {
		return true
	}

	return strings.EqualFold(left, right)
}

func formatProcessList(processes []processInfo) string {
	if len(processes) == 0 {
		return ""
	}

	lines := make([]string, 0, len(processes))
	for _, process := range processes {
		lines = append(lines, fmt.Sprintf("  %d %s", process.PID, process.Name))
	}
	return strings.Join(lines, "\n")
}

func formatProcessesInUseMessage(processes []processInfo) string {
	return fmt.Sprintf("Pablo is in use by other processes:\n%s", formatProcessList(processes))
}

func currentProcessID() int {
	return os.Getpid()
}

func sleep(duration time.Duration) {
	time.Sleep(duration)
}
