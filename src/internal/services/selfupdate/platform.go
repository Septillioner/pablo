package selfupdate

import (
	"fmt"
	"runtime"
)

func platformAssetName() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "windows":
		switch arch {
		case "amd64", "arm64":
			return fmt.Sprintf("pablo-windows-%s.exe", arch), nil
		}
	case "darwin":
		switch arch {
		case "amd64", "arm64":
			return fmt.Sprintf("pablo-darwin-%s", arch), nil
		}
	case "linux":
		if arch == "amd64" {
			return "pablo-linux-amd64", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", osName, arch)
}

func normalizeVersion(version string) string {
	if len(version) > 0 && (version[0] == 'v' || version[0] == 'V') {
		return version[1:]
	}
	return version
}

func releaseTag(version string) string {
	if version == "" {
		return ""
	}
	if version[0] == 'v' || version[0] == 'V' {
		return version
	}
	return "v" + version
}
