package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AddPath adds the specified folder to the environment PATH.
// scope can be "user" or "system".
func AddPath(newPath string, scope string, projectName string) error {
	newPath = strings.TrimSuffix(newPath, "\\")
	newPath = strings.TrimSuffix(newPath, "/")

	switch runtime.GOOS {
	case "windows":
		return addPathWindows(newPath, scope)
	case "darwin", "linux":
		return addPathUnix(newPath, scope, projectName)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func addPathWindows(newPath string, scope string) error {
	// Map internal scope names to PowerShell [EnvironmentVariableTarget] values
	targetScope := "User"
	if strings.ToLower(scope) == "system" || strings.ToLower(scope) == "machine" {
		targetScope = "Machine"
	}

	psScript := fmt.Sprintf(`
		$target = "%s"
		$scope = "%s"
		$path = [Environment]::GetEnvironmentVariable("Path", $scope)
		$paths = $path -split ";" | Where-Object { $_ -ne "" }
		if ($paths -notcontains $target) {
			$newPath = ($paths + $target) -join ";"
			[Environment]::SetEnvironmentVariable("Path", $newPath, $scope)
			Write-Output "ADDED"
		} else {
			Write-Output "EXISTS"
		}
	`, newPath, targetScope)

	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell execution failed: %w (output: %s)", err, string(output))
	}

	return nil
}

func addPathUnix(newPath string, scope string, projectName string) error {
	if strings.ToLower(scope) == "system" {
		// System scope behavior (Linux/macOS differences could be handled here)
		if runtime.GOOS == "darwin" {
			return addPathMacOSSystem(newPath)
		}
		return addPathLinuxSystem(newPath)
	}

	// User scope: Update shell configs
	return addPathUnixUser(newPath, projectName)
}

// addPathLinuxSystem registers a system-wide PATH entry via /etc/profile.d,
// the standard drop-in location sourced by login shells on Linux.
func addPathLinuxSystem(newPath string) error {
	dir := "/etc/profile.d"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("path registration failed: %s does not exist", dir)
	}

	filePath := systemDropInPath()

	content, err := os.ReadFile(filePath)
	if err == nil && strings.Contains(string(content), newPath) {
		return nil
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("permission denied: failed to write to %s (try running with sudo or use 'user' scope)", filePath)
	}
	defer f.Close()

	exportLine := fmt.Sprintf("export PATH=\"$PATH:%s\"\n", newPath)
	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filePath, err)
	}
	return nil
}

func addPathMacOSSystem(newPath string) error {
	// For system scope on macOS, we use /etc/paths.d/
	dir := "/etc/paths.d"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("path registration failed: %s does not exist", dir)
	}

	// Sanitize filename
	name := "pablo"
	filePath := filepath.Join(dir, name)

	// Check if already exists
	content, err := os.ReadFile(filePath)
	if err == nil && strings.Contains(string(content), newPath) {
		return nil
	}

	// Needs sudo to write here usually, but we try anyway
	err = os.WriteFile(filePath, []byte(newPath+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("permission denied: failed to write to %s (try running with sudo or use 'user' scope)", filePath)
	}
	return nil
}

func addPathUnixUser(newPath string, projectName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine user home directory: %w", err)
	}

	// Common shell config files
	configFiles := []string{".zshrc", ".bashrc", ".bash_profile", ".profile"}

	commentTag := fmt.Sprintf("# Added by pablo for %s", projectName)
	exportLine := fmt.Sprintf("\n%s\nexport PATH=\"$PATH:%s\"\n", commentTag, newPath)
	updated := false

	for _, cfg := range configFiles {
		path := filepath.Join(homeDir, cfg)

		// Skip if file doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable files
		}

		// Check idempotency using the unique project tag
		if strings.Contains(string(content), commentTag) {
			// Already present for this project
			updated = true
			fmt.Printf("Path already registered for project '%s' in %s\n", projectName, cfg)
			continue
		}

		// Append
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		if _, err := f.WriteString(exportLine); err != nil {
			f.Close()
			continue
		}
		f.Close()
		updated = true
		fmt.Printf("Added path to %s\n", cfg)
	}

	if !updated {
		return fmt.Errorf("no suitable shell config found (checked %v)", configFiles)
	}

	return nil
}

// RemovePath removes PATH entries added by pablo. On Unix user scope, entries
// are located by the project comment tag; Windows and Unix system scope require
// the concrete targetPath (and scope) that was registered, since those stores
// hold raw paths without a project marker.
func RemovePath(projectName, targetPath, scope string) error {
	switch runtime.GOOS {
	case "windows":
		return removePathWindows(targetPath, scope)
	case "darwin", "linux":
		if strings.ToLower(scope) == "system" {
			return removePathUnixSystem(targetPath)
		}
		return removePathUnixUser(projectName)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// removePathWindows removes the given path from the Windows PATH environment
// variable in the matching scope, mirroring addPathWindows.
func removePathWindows(targetPath, scope string) error {
	targetPath = strings.TrimSuffix(targetPath, "\\")
	targetPath = strings.TrimSuffix(targetPath, "/")

	targetScope := "User"
	if strings.ToLower(scope) == "system" || strings.ToLower(scope) == "machine" {
		targetScope = "Machine"
	}

	psScript := fmt.Sprintf(`
		$target = "%s"
		$scope = "%s"
		$path = [Environment]::GetEnvironmentVariable("Path", $scope)
		if ($null -eq $path) { Write-Output "EMPTY"; return }
		$paths = $path -split ";" | Where-Object { $_ -ne "" -and $_ -ne $target }
		[Environment]::SetEnvironmentVariable("Path", ($paths -join ";"), $scope)
		Write-Output "REMOVED"
	`, targetPath, targetScope)

	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell execution failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// removePathUnixSystem removes the pablo entry matching targetPath from the
// system-wide drop-in file (/etc/paths.d on macOS, /etc/profile.d on Linux).
func removePathUnixSystem(targetPath string) error {
	targetPath = strings.TrimSuffix(targetPath, "/")
	targetPath = strings.TrimSuffix(targetPath, "\\")

	filePath := systemDropInPath()

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var kept []string
	for _, line := range strings.Split(string(content), "\n") {
		if targetPath != "" && strings.Contains(line, targetPath) {
			continue
		}
		kept = append(kept, line)
	}

	remaining := strings.TrimSpace(strings.Join(kept, "\n"))
	if remaining == "" {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("permission denied: failed to remove %s (try running with sudo)", filePath)
		}
		fmt.Printf("Removed system PATH file %s\n", filePath)
		return nil
	}

	if err := os.WriteFile(filePath, []byte(remaining+"\n"), 0644); err != nil {
		return fmt.Errorf("permission denied: failed to update %s (try running with sudo)", filePath)
	}
	fmt.Printf("Removed system PATH entry from %s\n", filePath)
	return nil
}

// systemDropInPathOverride redirects system-scope drop-in file paths in tests.
var systemDropInPathOverride string

func systemDropInPath() string {
	if systemDropInPathOverride != "" {
		return systemDropInPathOverride
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join("/etc/paths.d", "pablo")
	}
	return "/etc/profile.d/pablo.sh"
}

func removePathUnixUser(projectName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine user home directory: %w", err)
	}

	configFiles := []string{".zshrc", ".bashrc", ".bash_profile", ".profile"}
	commentTag := fmt.Sprintf("# Added by pablo for %s", projectName)
	removed := false

	for _, cfg := range configFiles {
		path := filepath.Join(homeDir, cfg)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		var newLines []string
		skipNext := false

		for _, line := range lines {
			if skipNext {
				skipNext = false
				continue
			}
			if strings.Contains(line, commentTag) {
				skipNext = true
				removed = true
				fmt.Printf("Removed PATH entry from %s\n", cfg)
				continue
			}
			newLines = append(newLines, line)
		}

		if removed {
			newContent := strings.Join(newLines, "\n")
			if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to update %s: %w", cfg, err)
			}
		}
	}

	if !removed {
		fmt.Printf("No PATH entries found for project '%s'\n", projectName)
	}

	return nil
}
