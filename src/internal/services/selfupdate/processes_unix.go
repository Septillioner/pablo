//go:build !windows

package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func findProcessesUsingExecutable(exePath string) ([]processInfo, error) {
	targetPath, err := normalizeExecutablePath(exePath)
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "linux":
		return findProcessesLinux(targetPath)
	case "darwin":
		return findProcessesDarwin(targetPath)
	default:
		return nil, nil
	}
}

func findProcessesLinux(targetPath string) ([]processInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	var processes []processInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid == currentProcessID() {
			continue
		}

		exeLink, err := os.Readlink(filepath.Join("/proc", entry.Name(), "exe"))
		if err != nil {
			continue
		}

		exeLink = strings.TrimSuffix(exeLink, " (deleted)")
		if !executablePathsMatch(exeLink, targetPath) {
			continue
		}

		name, err := processNameUnix(pid)
		if err != nil || name == "" {
			name = filepath.Base(targetPath)
		}

		processes = append(processes, processInfo{PID: pid, Name: name})
	}

	return processes, nil
}

func findProcessesDarwin(targetPath string) ([]processInfo, error) {
	cmd := exec.Command("lsof", "-t", "--", targetPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("list processes with lsof: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var processes []processInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pid, err := strconv.Atoi(line)
		if err != nil || pid == currentProcessID() {
			continue
		}

		name, err := processNameUnix(pid)
		if err != nil || name == "" {
			name = filepath.Base(targetPath)
		}

		processes = append(processes, processInfo{PID: pid, Name: name})
	}

	return processes, nil
}

func processNameUnix(pid int) (string, error) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func terminateProcesses(processes []processInfo) error {
	for _, process := range processes {
		proc, err := os.FindProcess(process.PID)
		if err != nil {
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}
	}

	sleep(processExitPollDelay)

	for _, process := range processes {
		if !processAlive(process.PID) {
			continue
		}

		proc, err := os.FindProcess(process.PID)
		if err != nil {
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}
	}

	return waitForProcessesExit(processes)
}

func waitForProcessesExit(processes []processInfo) error {
	for range processExitPollAttempts {
		remaining := make([]processInfo, 0, len(processes))
		for _, process := range processes {
			if processAlive(process.PID) {
				remaining = append(remaining, process)
			}
		}
		if len(remaining) == 0 {
			return nil
		}
		processes = remaining
		sleep(processExitPollDelay)
	}

	return fmt.Errorf("processes still running: %s", formatProcessList(processes))
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
