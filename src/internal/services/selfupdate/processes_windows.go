//go:build windows

package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func findProcessesUsingExecutable(exePath string) ([]processInfo, error) {
	targetPath, err := normalizeExecutablePath(exePath)
	if err != nil {
		return nil, err
	}

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("enumerate processes: %w", err)
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	var processes []processInfo

	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		pid := int(entry.ProcessID)
		if pid != currentProcessID() {
			name := strings.TrimSpace(windows.UTF16ToString(entry.ExeFile[:]))
			imagePath, imageErr := processImagePath(uint32(pid))
			if imageErr == nil && executablePathsMatch(imagePath, targetPath) {
				if name == "" {
					name = filepath.Base(imagePath)
				}
				processes = append(processes, processInfo{PID: pid, Name: name})
			}
		}

		err = windows.Process32Next(snapshot, &entry)
	}
	if err != nil && err != syscall.ERROR_NO_MORE_FILES {
		return nil, fmt.Errorf("enumerate processes: %w", err)
	}

	return processes, nil
}

func processImagePath(pid uint32) (string, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)

	var buffer [windows.MAX_PATH]uint16
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(handle, 0, &buffer[0], &size); err != nil {
		return "", err
	}

	return filepath.Clean(windows.UTF16ToString(buffer[:size])), nil
}

func terminateProcesses(processes []processInfo) error {
	for _, process := range processes {
		handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(process.PID))
		if err != nil {
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}

		if err := windows.TerminateProcess(handle, 1); err != nil {
			windows.CloseHandle(handle)
			return fmt.Errorf("terminate process %d (%s): %w", process.PID, process.Name, err)
		}
		windows.CloseHandle(handle)
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
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	windows.CloseHandle(handle)

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
