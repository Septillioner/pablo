// Package pathutil provides path helpers that always produce POSIX-style
// paths for remote (SSH) targets, independent of the local operating system.
//
// filepath.Join uses the local OS separator, so a Windows host deploying to a
// Linux target would emit backslashes. Use these helpers for any path that is
// evaluated on the remote host.
package pathutil

import "strings"

const remoteSeparator = "/"

// JoinRemote joins path elements with forward slashes, suitable for a POSIX
// remote host regardless of the caller's operating system. Backslashes in the
// input are normalized and duplicate slashes are collapsed while a leading
// slash (absolute path) is preserved.
func JoinRemote(elem ...string) string {
	var parts []string
	for _, e := range elem {
		if e == "" {
			continue
		}
		parts = append(parts, strings.ReplaceAll(e, "\\", remoteSeparator))
	}
	if len(parts) == 0 {
		return ""
	}

	joined := strings.Join(parts, remoteSeparator)
	leading := strings.HasPrefix(joined, remoteSeparator)

	var clean []string
	for _, segment := range strings.Split(joined, remoteSeparator) {
		if segment != "" {
			clean = append(clean, segment)
		}
	}

	result := strings.Join(clean, remoteSeparator)
	if leading {
		result = remoteSeparator + result
	}
	return result
}

// DirRemote returns the parent directory of a POSIX remote path, mirroring
// filepath.Dir but always using forward slashes.
func DirRemote(path string) string {
	path = strings.ReplaceAll(path, "\\", remoteSeparator)
	idx := strings.LastIndex(path, remoteSeparator)
	switch {
	case idx < 0:
		return "."
	case idx == 0:
		return remoteSeparator
	default:
		return path[:idx]
	}
}
